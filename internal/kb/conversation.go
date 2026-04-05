package kb

import (
	"bufio"
	"bytes"
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Turn represents one authored segment of a conversation.
type Turn struct {
	Author    string    // "human" or "agent"
	Timestamp time.Time // zero if not present on the author line
	Content   string    // raw markdown text of this turn (trimmed)
}

// Conversation is a parsed conversation file.
type Conversation struct {
	ID    string // filename without extension
	Dir   string // "conversations" or "ephemeral"
	Title string // first H1 heading, or ID if absent
	Turns []Turn
}

// ParseConversation parses a conversation markdown file.
// id is the filename without extension; dir is "conversations" or "ephemeral".
func ParseConversation(id, dir string, data []byte) (*Conversation, error) {
	c := &Conversation{
		ID:  id,
		Dir: dir,
	}

	// Split into raw segments on lines that are exactly "---".
	segments := splitOnHR(string(data))

	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		// The first segment may start with a title heading before the first turn.
		// Extract title from the first segment if it has an H1.
		if i == 0 {
			if title, rest, ok := extractTitle(seg); ok {
				c.Title = title
				seg = strings.TrimSpace(rest)
				if seg == "" {
					continue
				}
			}
		}

		turn, err := parseTurn(seg)
		if err != nil {
			// Tolerate malformed segments rather than failing the whole parse.
			// Treat them as human content.
			c.Turns = append(c.Turns, Turn{Author: "human", Content: seg})
			continue
		}
		c.Turns = append(c.Turns, turn)
	}

	if c.Title == "" {
		c.Title = id
	}

	return c, nil
}

// splitOnHR splits text on lines that are exactly "---" (with optional trailing whitespace).
func splitOnHR(text string) []string {
	var segments []string
	var buf strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimRight(line, " \t") == "---" {
			segments = append(segments, buf.String())
			buf.Reset()
		} else {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	if buf.Len() > 0 {
		segments = append(segments, buf.String())
	}
	return segments
}

// extractTitle looks for a leading "# Title" line and returns (title, rest, true) if found.
func extractTitle(seg string) (string, string, bool) {
	idx := strings.IndexByte(seg, '\n')
	var first, rest string
	if idx == -1 {
		first = seg
	} else {
		first = seg[:idx]
		rest = seg[idx+1:]
	}
	first = strings.TrimSpace(first)
	if strings.HasPrefix(first, "# ") {
		return strings.TrimPrefix(first, "# "), rest, true
	}
	return "", seg, false
}

// parseTurn parses a segment that begins with "human:" or "agent:" (optionally with a timestamp).
func parseTurn(seg string) (Turn, error) {
	idx := strings.IndexByte(seg, '\n')
	var authorLine, content string
	if idx == -1 {
		authorLine = strings.TrimSpace(seg)
	} else {
		authorLine = strings.TrimSpace(seg[:idx])
		content = strings.TrimSpace(seg[idx+1:])
	}

	var author, prefix string
	switch {
	case strings.HasPrefix(authorLine, "human:"):
		author, prefix = "human", "human:"
	case strings.HasPrefix(authorLine, "agent:"):
		author, prefix = "agent", "agent:"
	default:
		return Turn{}, fmt.Errorf("segment does not begin with human: or agent:")
	}

	var ts time.Time
	if rest := strings.TrimSpace(strings.TrimPrefix(authorLine, prefix)); rest != "" {
		if t, err := time.Parse(time.RFC3339, rest); err == nil {
			ts = t
		}
	}

	return Turn{Author: author, Timestamp: ts, Content: content}, nil
}

// AppendHumanTurn atomically appends a human turn to the conversation file at path.
func AppendHumanTurn(path string, text string, ts time.Time) error {
	var buf bytes.Buffer

	// Read existing content.
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading conversation: %w", err)
	}
	buf.Write(existing)

	// Ensure file ends with a newline before the separator.
	if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}

	// Write separator and turn header.
	buf.WriteString("\n---\n")
	if ts.IsZero() {
		buf.WriteString("human:\n")
	} else {
		fmt.Fprintf(&buf, "human: %s\n", ts.UTC().Format(time.RFC3339))
	}
	buf.WriteString(text)
	if len(text) > 0 && text[len(text)-1] != '\n' {
		buf.WriteByte('\n')
	}

	return atomicWrite(path, buf.Bytes())
}

// NewConversationFile creates a new conversation file with a title and no turns yet.
func NewConversationFile(dir, id, title string) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s\n", title)
	return atomicWrite(filepath.Join(dir, id+".md"), buf.Bytes())
}

// atomicWrite writes data to path via a temp file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pkb-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// ListConversations returns all conversation IDs in a directory, sorted newest first.
func ListConversations(dir string) ([]ConversationMeta, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result []ConversationMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.HasSuffix(e.Name(), ".draft.md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		title := titleFromFile(filepath.Join(dir, e.Name()), id)
		result = append(result, ConversationMeta{
			ID:      id,
			Title:   title,
			ModTime: info.ModTime(),
		})
	}
	// Sort newest first.
	slices.SortFunc(result, func(a, b ConversationMeta) int {
		return cmp.Compare(b.ModTime.UnixNano(), a.ModTime.UnixNano())
	})
	return result, nil
}

// ConversationMeta holds lightweight info for listing conversations.
type ConversationMeta struct {
	ID      string
	Title   string
	ModTime time.Time
}

// TitleFromBytes extracts the first H1 heading from markdown bytes,
// or returns fallback if none is found.
func TitleFromBytes(data []byte, fallback string) string {
	return scanTitle(bufio.NewScanner(bytes.NewReader(data)), fallback)
}

// titleFromFile reads the first H1 heading from a file, or returns fallback.
// It reads only as many bytes as needed to find the heading.
func titleFromFile(path, fallback string) string {
	f, err := os.Open(path)
	if err != nil {
		return fallback
	}
	defer f.Close()
	return scanTitle(bufio.NewScanner(f), fallback)
}

// scanTitle scans lines from scanner looking for the first "# Title" heading.
func scanTitle(scanner *bufio.Scanner, fallback string) string {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		// Stop at the first non-blank, non-heading line.
		if line != "" {
			break
		}
	}
	return fallback
}
