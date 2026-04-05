package kb

import (
	"os"
	"path/filepath"
	"time"
)

// CreateReplySignal creates queue/reply/<id> to signal Raymond that a conversation
// turn needs an agent response.
func CreateReplySignal(kb *KB, conversationID string) error {
	path := filepath.Join(kb.QueueDir(), "reply", conversationID)
	return os.WriteFile(path, []byte{}, 0644)
}

// CreateIngestSignal creates queue/ingest/<id> to signal Raymond that a conversation
// should be ingested into the wiki.
func CreateIngestSignal(kb *KB, conversationID string) error {
	path := filepath.Join(kb.QueueDir(), "ingest", conversationID)
	return os.WriteFile(path, []byte{}, 0644)
}

// CreateLintSignal creates queue/lint to signal Raymond to run a lint pass.
func CreateLintSignal(kb *KB) error {
	return os.WriteFile(filepath.Join(kb.QueueDir(), "lint"), []byte{}, 0644)
}

// CreateCommitSignal creates queue/commit to signal Raymond to commit changes.
func CreateCommitSignal(kb *KB) error {
	return os.WriteFile(filepath.Join(kb.QueueDir(), "commit"), []byte{}, 0644)
}

// RaymondActive returns true if Raymond appears to be running.
// The heuristic: if any signal file is present and was created more than
// staleThreshold ago without being consumed, Raymond is probably not running.
// If no signal files are pending, we optimistically assume Raymond is available.
func RaymondActive(kb *KB, staleThreshold time.Duration) bool {
	dirs := []string{
		filepath.Join(kb.QueueDir(), "reply"),
		filepath.Join(kb.QueueDir(), "ingest"),
	}
	singletons := []string{
		filepath.Join(kb.QueueDir(), "lint"),
		filepath.Join(kb.QueueDir(), "commit"),
	}

	paths := append([]string{}, singletons...)
	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			paths = append(paths, filepath.Join(d, e.Name()))
		}
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		// A signal file older than the threshold that hasn't been consumed
		// suggests Raymond is not processing work.
		if time.Since(info.ModTime()) > staleThreshold {
			return false
		}
	}

	return true
}
