package kb

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// IssueMeta holds lightweight info for listing issues.
type IssueMeta struct {
	Filename string
	Title    string
	ModTime  time.Time
}

// ListIssues returns all issue files in the issues directory, sorted newest-mtime-first.
func ListIssues(kb *KB) ([]IssueMeta, error) {
	dir := kb.IssuesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result []IssueMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		title := titleFromFile(filepath.Join(dir, e.Name()), e.Name())
		result = append(result, IssueMeta{
			Filename: e.Name(),
			Title:    title,
			ModTime:  info.ModTime(),
		})
	}
	slices.SortFunc(result, func(a, b IssueMeta) int {
		return cmp.Compare(b.ModTime.UnixNano(), a.ModTime.UnixNano())
	})
	return result, nil
}

// IssueContent holds the parsed contents of an issue file.
type IssueContent struct {
	Title    string
	Workflow string
	Time     string
	Related  string
	Body     string
}

// ParseIssue reads and parses an issue file from the issues directory.
func ParseIssue(kb *KB, filename string) (IssueContent, error) {
	path := filepath.Join(kb.IssuesDir(), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return IssueContent{}, err
	}

	lines := strings.Split(string(data), "\n")
	var ic IssueContent
	i := 0

	// Extract title from first # heading.
	for i < len(lines) {
		line := lines[i]
		i++
		if strings.HasPrefix(line, "# ") {
			ic.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Scan for bold-label metadata lines; skip blank lines; stop on unrecognized non-blank line.
	for i < len(lines) {
		line := lines[i]
		if line == "" {
			i++
			continue
		}
		if val, ok := boldLabel(line, "Workflow"); ok {
			ic.Workflow = val
			i++
		} else if val, ok := boldLabel(line, "Time"); ok {
			ic.Time = val
			i++
		} else if val, ok := boldLabel(line, "Related"); ok {
			ic.Related = val
			i++
		} else {
			// Not a recognized metadata line — body starts here.
			break
		}
	}

	// Everything from i onward is the body.
	if i < len(lines) {
		ic.Body = strings.Join(lines[i:], "\n")
	}

	return ic, nil
}

// boldLabel checks if line matches "**Label:** value" and returns the value.
func boldLabel(line, label string) (string, bool) {
	prefix := "**" + label + ":** "
	if strings.HasPrefix(line, prefix) {
		return strings.TrimPrefix(line, prefix), true
	}
	return "", false
}
