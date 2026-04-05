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
// Raymond touches heartbeat.md at the knowledge base root periodically.
// If the file is missing or its mtime is older than staleThreshold,
// Raymond is considered inactive.
func RaymondActive(kb *KB, staleThreshold time.Duration) bool {
	info, err := os.Stat(kb.HeartbeatPath())
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) <= staleThreshold
}
