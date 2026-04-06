package kb

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func seenPath(kb *KB) string {
	return filepath.Join(kb.IssuesDir(), ".seen")
}

// LoadSeen reads the .seen file and returns a set of filenames the user has
// already seen. If the file does not exist, an empty map is returned.
func LoadSeen(kb *KB) (map[string]struct{}, error) {
	data, err := os.ReadFile(seenPath(kb))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			seen[line] = struct{}{}
		}
	}
	return seen, nil
}

// SaveSeen writes the seen set to the .seen file, pruning entries for issues
// that no longer exist in IssuesDir.
func SaveSeen(kb *KB, seen map[string]struct{}) error {
	entries, err := os.ReadDir(kb.IssuesDir())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existing := map[string]struct{}{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			existing[e.Name()] = struct{}{}
		}
	}

	var lines []string
	for name := range seen {
		if _, ok := existing[name]; ok {
			lines = append(lines, name)
		}
	}

	slices.Sort(lines)
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(seenPath(kb), []byte(content), 0644)
}

// UnseenCount returns the number of issues whose filenames are not in seen.
func UnseenCount(issues []IssueMeta, seen map[string]struct{}) int {
	count := 0
	for _, issue := range issues {
		if _, ok := seen[issue.Filename]; !ok {
			count++
		}
	}
	return count
}
