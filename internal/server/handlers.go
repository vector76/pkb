package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/vector76/pkb/internal/kb"
)

// ── Wiki ─────────────────────────────────────────────────────────────────────

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.renderError(w, http.StatusNotFound, "Page not found.")
		return
	}
	http.Redirect(w, r, "/wiki/index", http.StatusFound)
}

func (s *Server) handleWikiIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/wiki/index", http.StatusFound)
}

func (s *Server) handleWikiPage(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	path := s.kb.WikiPath(page + ".md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.renderError(w, http.StatusNotFound, fmt.Sprintf("Wiki page %q not found.", page))
			return
		}
		s.renderError(w, http.StatusInternalServerError, "Could not read page.")
		return
	}

	body, err := s.renderer.RenderMarkdown(data)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not render page.")
		return
	}

	title := kb.TitleFromBytes(data, page)
	type wikiData struct {
		baseData
		Body template.HTML
	}
	if err := s.tmpl.wiki.ExecuteTemplate(w, "base", wikiData{
		baseData: s.newBaseData(title, "wiki/"+page+".md"),
		Body:     body,
	}); err != nil {
		log.Printf("handleWikiPage: template: %v", err)
	}
}

// ── Conversations ─────────────────────────────────────────────────────────────

func (s *Server) handleConversationList(w http.ResponseWriter, r *http.Request) {
	convs, err := kb.ListConversations(s.kb.ConversationsDir())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not list conversations.")
		return
	}
	ephs, err := kb.ListConversations(s.kb.EphemeralDir())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not list ephemeral conversations.")
		return
	}

	type listData struct {
		baseData
		Conversations []kb.ConversationMeta
		Ephemeral     []kb.ConversationMeta
	}
	if err := s.tmpl.convList.ExecuteTemplate(w, "base", listData{
		baseData:      s.newBaseData("Conversations", ""),
		Conversations: convs,
		Ephemeral:     ephs,
	}); err != nil {
		log.Printf("handleConversationList: template: %v", err)
	}
}

func (s *Server) handleConversation(w http.ResponseWriter, r *http.Request) {
	id, dir := conversationIDAndDir(r)

	path := s.kb.ConversationPath(dir, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.renderError(w, http.StatusNotFound, fmt.Sprintf("Conversation %q not found.", id))
			return
		}
		s.renderError(w, http.StatusInternalServerError, "Could not read conversation.")
		return
	}

	conv, err := kb.ParseConversation(id, dir, data)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not parse conversation.")
		return
	}

	// Render each turn's markdown content.
	type renderedTurn struct {
		Author    string
		Timestamp time.Time
		Body      template.HTML
	}
	var turns []renderedTurn
	for _, t := range conv.Turns {
		body, err := s.renderer.RenderMarkdown([]byte(t.Content))
		if err != nil {
			log.Printf("render turn (%s/%s): %v", dir, id, err)
		}
		turns = append(turns, renderedTurn{
			Author:    t.Author,
			Timestamp: t.Timestamp,
			Body:      body,
		})
	}

	// Check if a reply is pending (signal file exists = waiting).
	replySignal := filepath.Join(s.kb.QueueDir(), "reply", id)
	_, signalErr := os.Stat(replySignal)
	waitingForAgent := signalErr == nil

	draft := kb.LoadDraft(s.kb, dir, id)

	type convData struct {
		baseData
		Conv            *kb.Conversation
		Turns           []renderedTurn
		WaitingForAgent bool
		Draft           string
	}

	if err := s.tmpl.conv.ExecuteTemplate(w, "base", convData{
		baseData:        s.newBaseData(conv.Title, dir+"/"+id+".md"),
		Conv:            conv,
		Turns:           turns,
		WaitingForAgent: waitingForAgent,
		Draft:           draft,
	}); err != nil {
		log.Printf("handleConversation: template: %v", err)
	}
}

func (s *Server) handleNewConversation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderError(w, http.StatusBadRequest, "Bad form data.")
		return
	}
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = "Untitled"
	}
	slug := slugify(title)
	if slug == "" {
		slug = "conversation"
	}
	id := slug + "-" + fmt.Sprintf("%d", time.Now().UnixMilli())

	if err := kb.NewConversationFile(s.kb.EphemeralDir(), id, title); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not create conversation.")
		return
	}
	http.Redirect(w, r, "/ephemeral/"+id, http.StatusSeeOther)
}

func (s *Server) handleReply(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dir := "conversations"
	if r.URL.Query().Get("dir") == "ephemeral" {
		dir = "ephemeral"
	}

	if err := r.ParseForm(); err != nil {
		s.renderError(w, http.StatusBadRequest, "Bad form data.")
		return
	}
	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		s.renderError(w, http.StatusBadRequest, "Message cannot be empty.")
		return
	}

	path := s.kb.ConversationPath(dir, id)
	if err := kb.AppendHumanTurn(path, text, time.Now(), ""); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not save message.")
		return
	}
	kb.DeleteDraft(s.kb, dir, id)
	if err := kb.CreateReplySignal(s.kb, id); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not signal agent — message was saved but agent will not respond.")
		return
	}

	prefix := "/conversations/"
	if dir == "ephemeral" {
		prefix = "/ephemeral/"
	}
	http.Redirect(w, r, prefix+id, http.StatusSeeOther)
}

func (s *Server) handleSaveDraft(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dir := "conversations"
	if r.URL.Query().Get("dir") == "ephemeral" {
		dir = "ephemeral"
	}

	text := r.FormValue("text")
	if strings.TrimSpace(text) == "" {
		kb.DeleteDraft(s.kb, dir, id)
	} else {
		if err := kb.SaveDraft(s.kb, dir, id, text); err != nil {
			http.Error(w, "could not save draft", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := kb.CreateIngestSignal(s.kb, id); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not queue ingest.")
		return
	}
	http.Redirect(w, r, "/conversations/"+id, http.StatusSeeOther)
}

func (s *Server) handlePromote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	src := s.kb.ConversationPath("ephemeral", id)
	dst := s.kb.ConversationPath("conversations", id)

	if err := os.Rename(src, dst); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not promote conversation.")
		return
	}
	// Move draft sidecar if it exists.
	draftSrc := s.kb.DraftPath("ephemeral", id)
	draftDst := s.kb.DraftPath("conversations", id)
	os.Rename(draftSrc, draftDst) // best-effort; no draft is fine
	if err := kb.CreateIngestSignal(s.kb, id); err != nil {
		// Rename already succeeded; try to undo it so state stays consistent.
		if undoErr := os.Rename(dst, src); undoErr != nil {
			log.Printf("handlePromote: undo rename failed: %v", undoErr)
		}
		os.Rename(draftDst, draftSrc) // best-effort undo of draft move
		s.renderError(w, http.StatusInternalServerError, "Could not queue ingest after promotion.")
		return
	}
	http.Redirect(w, r, "/conversations/"+id, http.StatusSeeOther)
}

// ── Queue triggers ────────────────────────────────────────────────────────────

func (s *Server) handleLint(w http.ResponseWriter, r *http.Request) {
	if err := kb.CreateLintSignal(s.kb); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not trigger lint.")
		return
	}
	http.Redirect(w, r, safeRedirectURL(r, "/"), http.StatusSeeOther)
}

func (s *Server) handleCommit(w http.ResponseWriter, r *http.Request) {
	if err := kb.CreateCommitSignal(s.kb); err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not trigger commit.")
		return
	}
	http.Redirect(w, r, safeRedirectURL(r, "/"), http.StatusSeeOther)
}

// ── Attachments ───────────────────────────────────────────────────────────────

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Use the original filename, sanitized.
	name := filepath.Base(header.Filename)
	name = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	name = strings.Trim(name, ".")
	if name == "" {
		name = "attachment"
	}

	dst := filepath.Join(s.kb.AttachmentsDir(), name)
	// If a file with that name exists, add a timestamp prefix.
	if _, err := os.Stat(dst); err == nil {
		name = fmt.Sprintf("%d-%s", time.Now().UnixMilli(), name)
		dst = filepath.Join(s.kb.AttachmentsDir(), name)
	}

	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "could not save file", http.StatusInternalServerError)
		return
	}
	_, copyErr := io.Copy(out, file)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		os.Remove(dst)
		http.Error(w, "could not save file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"name": name,
		"ref":  "../attachments/" + name,
	}); err != nil {
		log.Printf("handleUpload: encode response: %v", err)
	}
}

func (s *Server) handleAttachment(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	// Prevent path traversal.
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	path := filepath.Join(s.kb.AttachmentsDir(), name)
	http.ServeFile(w, r, path)
}

// ── Search ────────────────────────────────────────────────────────────────────

type searchResult struct {
	Title string
	URL   string
	Lines []string
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	var results []searchResult
	if query != "" {
		results = s.search(query)
	}

	type searchData struct {
		baseData
		Query   string
		Results []searchResult
	}
	if err := s.tmpl.search.ExecuteTemplate(w, "base", searchData{
		baseData: s.newBaseData("Search", ""),
		Query:    query,
		Results:  results,
	}); err != nil {
		log.Printf("handleSearch: template: %v", err)
	}
}

func (s *Server) search(query string) []searchResult {
	lower := strings.ToLower(query)
	var results []searchResult

	searchDir := func(dir, urlPrefix string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasSuffix(e.Name(), ".draft.md") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var matchLines []string
			scanner := bufio.NewScanner(bytes.NewReader(data))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(strings.ToLower(line), lower) {
					matchLines = append(matchLines, line)
					if len(matchLines) >= 3 {
						break
					}
				}
			}
			if len(matchLines) == 0 {
				continue
			}

			id := strings.TrimSuffix(e.Name(), ".md")
			title := kb.TitleFromBytes(data, id)
			results = append(results, searchResult{
				Title: title,
				URL:   urlPrefix + id,
				Lines: matchLines,
			})
		}
	}

	searchDir(s.kb.WikiDir(), "/wiki/")
	searchDir(s.kb.ConversationsDir(), "/conversations/")

	return results
}

// ── Log ───────────────────────────────────────────────────────────────────────

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(s.kb.LogPath())
	if err != nil {
		if os.IsNotExist(err) {
			data = []byte("*No log entries yet.*\n")
		} else {
			s.renderError(w, http.StatusInternalServerError, "Could not read log.")
			return
		}
	}
	body, err := s.renderer.RenderMarkdown(data)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, "Could not render log.")
		return
	}
	type logData struct {
		baseData
		Body template.HTML
	}
	if err := s.tmpl.wiki.ExecuteTemplate(w, "base", logData{
		baseData: s.newBaseData("Log", "log.md"),
		Body:     body,
	}); err != nil {
		log.Printf("handleLog: template: %v", err)
	}
}

// ── SSE ───────────────────────────────────────────────────────────────────────

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsub := s.hub.Subscribe()
	defer unsub()

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			if _, err := w.Write(data); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// safeRedirectURL returns the path portion of the request's Referer, or
// fallback if the Referer is absent or unparseable. Using only the path
// (never the full URL) prevents open redirects regardless of whether the
// Referer is absolute (as browsers always send) or relative.
func safeRedirectURL(r *http.Request, fallback string) string {
	ref := r.Referer()
	if ref == "" {
		return fallback
	}
	u, err := url.Parse(ref)
	if err != nil {
		return fallback
	}
	path := u.RequestURI()
	if path == "" || path[0] != '/' || strings.HasPrefix(path, "//") {
		return fallback
	}
	return path
}

// conversationIDAndDir extracts the conversation ID and directory from a request.
func conversationIDAndDir(r *http.Request) (id, dir string) {
	id = r.PathValue("id")
	if strings.HasPrefix(r.URL.Path, "/ephemeral/") {
		dir = "ephemeral"
	} else {
		dir = "conversations"
	}
	return
}

// slugify converts a title to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

