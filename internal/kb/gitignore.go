package kb

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// GitignoreStatus reports whether the KB is a git repository and which
// required paths are not covered by .gitignore rules.
type GitignoreStatus struct {
	IsGitRepo bool     // false → KB has no .git directory
	Uncovered []string // nil/empty → all covered; non-empty → these canonical names are not gitignored
}

// CheckGitignore checks whether all required KB paths are covered by
// the repository's .gitignore rules.
func CheckGitignore(kb *KB) GitignoreStatus {
	if _, err := os.Stat(filepath.Join(kb.Root, ".git")); err != nil {
		return GitignoreStatus{IsGitRepo: false}
	}

	required := []struct {
		arg      string // argument passed to git check-ignore
		canonical string // name recorded in results
	}{
		{"ephemeral/", "ephemeral/"},
		{"queue/", "queue/"},
		{"issues/", "issues/"},
		{"heartbeat.md", "heartbeat.md"},
		{"test.draft.md", "*.draft.md"},
	}

	var uncovered []string
	for _, r := range required {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "git", "check-ignore", "-q", r.arg)
		cmd.Dir = kb.Root
		err := cmd.Run()
		cancel()

		if err == nil {
			// exit code 0 → covered
			continue
		}
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// exit code 1 → not ignored
			uncovered = append(uncovered, r.canonical)
			continue
		}
		// any other error → treat as covered, skip
	}

	return GitignoreStatus{IsGitRepo: true, Uncovered: uncovered}
}
