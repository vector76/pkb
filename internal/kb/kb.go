package kb

import (
	"fmt"
	"os"
	"path/filepath"
)

// KB holds the root path of the knowledge base and provides path helpers.
type KB struct {
	Root string
}

// New resolves the absolute path of root and initializes the directory structure.
func New(root string) (*KB, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving kb root: %w", err)
	}
	kb := &KB{Root: abs}
	if err := kb.Init(); err != nil {
		return nil, err
	}
	return kb, nil
}

// Init creates the standard directory skeleton if it does not exist.
// It is idempotent.
func (kb *KB) Init() error {
	dirs := []string{
		kb.WikiDir(),
		kb.ConversationsDir(),
		kb.EphemeralDir(),
		kb.AttachmentsDir(),
		filepath.Join(kb.QueueDir(), "reply"),
		filepath.Join(kb.QueueDir(), "ingest"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", d, err)
		}
	}

	// Bootstrap wiki/index.md if absent.
	index := kb.WikiPath("index.md")
	if _, err := os.Stat(index); os.IsNotExist(err) {
		content := "# Knowledge Base\n\nWelcome to your personal knowledge base.\n"
		if err := os.WriteFile(index, []byte(content), 0644); err != nil {
			return fmt.Errorf("creating index.md: %w", err)
		}
	}

	return nil
}

func (kb *KB) WikiDir() string          { return filepath.Join(kb.Root, "wiki") }
func (kb *KB) ConversationsDir() string { return filepath.Join(kb.Root, "conversations") }
func (kb *KB) EphemeralDir() string     { return filepath.Join(kb.Root, "ephemeral") }
func (kb *KB) AttachmentsDir() string   { return filepath.Join(kb.Root, "attachments") }
func (kb *KB) QueueDir() string         { return filepath.Join(kb.Root, "queue") }

func (kb *KB) LogPath() string          { return filepath.Join(kb.Root, "log.md") }
func (kb *KB) HeartbeatPath() string    { return filepath.Join(kb.Root, "heartbeat.md") }

// WikiPath returns the full path to a file within the wiki directory.
func (kb *KB) WikiPath(name string) string {
	return filepath.Join(kb.WikiDir(), name)
}

// ConversationPath returns the full path to a conversation file.
// dir should be "conversations" or "ephemeral".
func (kb *KB) ConversationPath(dir, id string) string {
	return filepath.Join(kb.Root, dir, id+".md")
}

// DraftPath returns the full path to a conversation's draft sidecar file.
func (kb *KB) DraftPath(dir, id string) string {
	return filepath.Join(kb.Root, dir, id+".draft.md")
}

// FilePath joins the KB root with a relative path.
func (kb *KB) FilePath(rel string) string {
	return filepath.Join(kb.Root, rel)
}
