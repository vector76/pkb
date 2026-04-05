(function () {
  'use strict';

  // ── SSE ──────────────────────────────────────────────────────────────────
  //
  // Close the EventSource on beforeunload so the browser frees the HTTP/1.1
  // connection slot immediately.  Without this, rapid link clicks can
  // accumulate lingering SSE connections that exhaust the browser's 6-slot
  // per-origin limit, stalling navigation for up to ~60 s.

  var es = null;

  function connectSSE() {
    if (es) es.close();
    es = new EventSource('/events');

    es.onmessage = function (e) {
      try {
        var ev = JSON.parse(e.data);
        handleEvent(ev);
      } catch (_) {}
    };

    es.onerror = function () {
      es.close();
      es = null;
      setTimeout(connectSSE, 5000);
    };
  }

  window.addEventListener('beforeunload', function () {
    if (es) { es.close(); es = null; }
  });

  function handleEvent(ev) {
    if (ev.type === 'conversation_updated') {
      // If we're viewing the conversation that was updated, reload the turns.
      var conv = document.getElementById('conversation');
      if (!conv) return;

      var id  = conv.dataset.id;
      var dir = conv.dataset.dir;
      var expectedPath = dir + '/' + id + '.md';

      if (ev.path === expectedPath) {
        // Reload just the turns section via a fetch, or fall back to full reload.
        reloadTurns(dir, id);
      }
    } else if (ev.type === 'wiki_updated') {
      // Show a stale banner only if this exact page was updated.
      var pagePath = (window.PKB || {}).pagePath || '';
      if (ev.path !== pagePath) return;

      var article = document.querySelector('.wiki-page');
      var content = document.querySelector('.content');
      if (article && content && !document.getElementById('wiki-stale-banner')) {
        var banner = document.createElement('div');
        banner.id = 'wiki-stale-banner';
        banner.style.cssText = 'background:#fffbe6;padding:0.5rem 1rem;font-size:0.88rem;border-bottom:1px solid #e0e0e0;cursor:pointer;';
        banner.textContent = 'This page was updated. Click to reload.';
        banner.onclick = function () { location.reload(); };
        content.prepend(banner);
      }
    }
  }

  function reloadTurns(dir, id) {
    var url = '/' + dir + '/' + id + '?fragment=turns';
    fetch(url)
      .then(function (r) { return r.ok ? r.text() : null; })
      .then(function (html) {
        if (!html) { location.reload(); return; }
        var parser = new DOMParser();
        var doc = parser.parseFromString(html, 'text/html');
        var newTurns = doc.getElementById('turns');
        var newWaiting = doc.getElementById('agent-waiting');
        var oldTurns = document.getElementById('turns');
        var oldWaiting = document.getElementById('agent-waiting');

        if (newTurns && oldTurns) {
          oldTurns.replaceWith(newTurns);
        }
        // Update or remove the waiting indicator.
        if (oldWaiting) oldWaiting.remove();
        if (newWaiting) {
          var form = document.getElementById('reply-form');
          if (form) form.before(newWaiting);
        }
      })
      .catch(function () { location.reload(); });
  }

  // ── Attachment injection ─────────────────────────────────────────────────
  var attachInput = document.getElementById('attachment-input');
  if (attachInput) {
    attachInput.addEventListener('change', function () {
      var file = attachInput.files[0];
      if (!file) return;

      var form = document.getElementById('reply-form');
      var textarea = form ? form.querySelector('textarea') : null;
      if (!textarea) return;

      var fd = new FormData();
      fd.append('file', file);

      fetch('/attachments/upload', { method: 'POST', body: fd })
        .then(function (r) { return r.ok ? r.json() : Promise.reject(r.statusText); })
        .then(function (data) {
          // Inject the markdown reference into the textarea.
          var ref = '\n[' + data.name + '](' + data.ref + ')';
          textarea.value = textarea.value + ref;
          textarea.focus();
        })
        .catch(function (err) {
          alert('Upload failed: ' + err);
        });

      // Reset so the same file can be re-selected.
      attachInput.value = '';
    });
  }

  // ── Save Draft ───────────────────────────────────────────────────────────
  var draftBtn = document.getElementById('save-draft-btn');
  if (draftBtn) {
    draftBtn.addEventListener('click', function () {
      var conv = document.getElementById('conversation');
      if (!conv) return;
      var id = conv.dataset.id;
      var dir = conv.dataset.dir;
      var form = document.getElementById('reply-form');
      var textarea = form ? form.querySelector('textarea') : null;
      if (!textarea) return;

      var fd = new FormData();
      fd.append('text', textarea.value);

      var url = '/conversations/' + id + '/draft';
      if (dir === 'ephemeral') url += '?dir=ephemeral';

      fetch(url, { method: 'POST', body: fd })
        .then(function (r) {
          if (r.ok) {
            draftBtn.textContent = 'Draft Saved';
            setTimeout(function () { draftBtn.textContent = 'Save Draft'; }, 1500);
          }
        })
        .catch(function () {});
    });
  }

  var THEME_KEY = 'pkb-theme';

  function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
  }

  function updateThemeButton(theme) {
    var btn = document.getElementById('theme-toggle');
    if (btn) {
      btn.textContent = theme === 'light' ? '🌙' : '☀️';
    }
  }

  function toggleTheme() {
    var current = document.documentElement.getAttribute('data-theme');
    var newTheme = current === 'dark' ? 'light' : 'dark';
    localStorage.setItem(THEME_KEY, newTheme);
    applyTheme(newTheme);
    updateThemeButton(newTheme);
  }

  // ── Init ─────────────────────────────────────────────────────────────────
  connectSSE();
  var themeBtn = document.getElementById('theme-toggle');
  if (themeBtn) {
    updateThemeButton(document.documentElement.getAttribute('data-theme') || 'light');
    themeBtn.addEventListener('click', toggleTheme);
  }
}());
