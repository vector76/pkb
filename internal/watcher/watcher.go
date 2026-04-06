package watcher

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/vector76/pkb/internal/kb"
	"github.com/vector76/pkb/internal/server"

	"github.com/fsnotify/fsnotify"
)

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
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watcher: failed to create fsnotify watcher: %v", err)
		return
	}
	defer fw.Close()

	dirs := []string{
		w.kb.WikiDir(),
		w.kb.ConversationsDir(),
		w.kb.EphemeralDir(),
		w.kb.IssuesDir(),
	}
	for _, d := range dirs {
		if err := fw.Add(d); err != nil {
			log.Printf("watcher: watching %s: %v", d, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-fw.Events:
			if !ok {
				return
			}
			if e, ok := w.classifyEvent(event.Name); ok {
				if e.Type == "issues_updated" {
					if seen, err := kb.LoadSeen(w.kb); err == nil {
						if issues, err := kb.ListIssues(w.kb); err == nil {
							e.UnseenCount = kb.UnseenCount(issues, seen)
						}
					}
				}
				w.hub.Publish(e)
			}
		case err, ok := <-fw.Errors:
			if !ok {
				return
			}
			log.Printf("watcher: %v", err)
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
