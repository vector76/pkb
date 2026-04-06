package watcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vector76/pkb/internal/kb"
	"github.com/vector76/pkb/internal/server"
)

// snapshot maps absolute file path to last-observed mtime.
type snapshot map[string]time.Time

// scanDirs scans dirs and returns a snapshot of all classifiable files.
func scanDirs(dirs []string, classify func(string) (server.Event, bool)) snapshot {
	snap := make(snapshot)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			if _, ok := classify(path); !ok {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			snap[path] = info.ModTime()
		}
	}
	return snap
}

// Watcher watches the knowledge base for changes and publishes SSE events.
type Watcher struct {
	hub *server.Hub
	kb  *kb.KB
}

func New(hub *server.Hub, kb *kb.KB) *Watcher {
	return &Watcher{hub: hub, kb: kb}
}

// Start begins watching the knowledge base directories.
// It runs until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) {
	dirs := []string{
		w.kb.WikiDir(),
		w.kb.ConversationsDir(),
		w.kb.EphemeralDir(),
		w.kb.IssuesDir(),
	}
	snap := scanDirs(dirs, w.classifyEvent)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fresh := scanDirs(dirs, w.classifyEvent)
			for path, mtime := range fresh {
				if snap[path] == mtime {
					continue
				}
				e, ok := w.classifyEvent(path)
				if !ok {
					continue
				}
				if e.Type == "issues_updated" {
					if seen, err := kb.LoadSeen(w.kb); err == nil {
						if issues, err := kb.ListIssues(w.kb); err == nil {
							e.UnseenCount = kb.UnseenCount(issues, seen)
						}
					}
				}
				w.hub.Publish(e)
			}
			snap = fresh
		}
	}
}

// classifyEvent maps a filesystem path to an SSE event.
func (w *Watcher) classifyEvent(path string) (server.Event, bool) {
	base := filepath.Base(path)
	// Ignore temp files created by atomicWrite and draft sidecar files.
	if strings.HasPrefix(base, ".pkb-tmp-") || strings.HasSuffix(base, ".draft.md") {
		return server.Event{}, false
	}

	rel, err := filepath.Rel(w.kb.Root, path)
	if err != nil {
		return server.Event{}, false
	}
	rel = filepath.ToSlash(rel)

	switch {
	case strings.HasPrefix(rel, "conversations/"):
		return server.Event{Type: "conversation_updated", Path: rel}, true
	case strings.HasPrefix(rel, "ephemeral/"):
		return server.Event{Type: "conversation_updated", Path: rel}, true
	case strings.HasPrefix(rel, "wiki/"):
		return server.Event{Type: "wiki_updated", Path: rel}, true
	case strings.HasPrefix(rel, "issues/"):
		if strings.HasSuffix(base, ".md") && base != ".seen" {
			return server.Event{Type: "issues_updated"}, true
		}
		return server.Event{}, false
	default:
		return server.Event{}, false
	}
}
