package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type importedGitHistoryRecord struct {
	SchemaVersion  int                        `json:"schema_version"`
	Kind           string                     `json:"kind"`
	Provenance     string                     `json:"provenance"`
	GeneratedAt    int64                      `json:"generated_at"`
	AdoptionCommit string                     `json:"adoption_commit"`
	Commits        []importedGitHistoryCommit `json:"commits"`
}

type importedGitHistoryCommit struct {
	Commit      string `json:"commit"`
	CommittedAt int64  `json:"committed_at"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Subject     string `json:"subject"`
}

func buildImportedGitHistory(root string, adoptionCommit string) ([]importedGitHistoryCommit, error) {
	allCommits, err := gitLogForBackfill(root)
	if err != nil {
		return nil, err
	}
	bounded, found := trimHistoryAtAdoptionCommit(allCommits, adoptionCommit)
	if !found {
		return nil, fmt.Errorf("adoption commit %q was not found in git history", adoptionCommit)
	}
	return bounded, nil
}

func gitLogForBackfill(root string) ([]importedGitHistoryCommit, error) {
	insideRepo, err := isGitWorkTree(root)
	if err != nil {
		return nil, fmt.Errorf("check git repository state: %w", err)
	}
	if !insideRepo {
		return nil, fmt.Errorf("assurance backfill requires a git repository at %q", root)
	}
	const format = "%H%x1f%ct%x1f%an%x1f%ae%x1f%s"
	command := exec.Command("git", "-C", root, "log", "--reverse", "--format="+format, "HEAD")
	// Use CombinedOutput so stderr is captured and propagated in errors.
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, formatGitLogTraversalError(output, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]importedGitHistoryCommit, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		if len(parts) < 5 {
			return nil, fmt.Errorf("unexpected git log output shape")
		}
		committedAt, parseErr := parseUnixTimestamp(parts[1])
		if parseErr != nil {
			return nil, parseErr
		}
		commits = append(commits, importedGitHistoryCommit{
			Commit:      parts[0],
			CommittedAt: committedAt,
			AuthorName:  parts[2],
			AuthorEmail: parts[3],
			Subject:     parts[4],
		})
	}
	return commits, nil
}

func isGitWorkTree(root string) (bool, error) {
	output, err := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

func formatGitLogTraversalError(output []byte, err error) error {
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return fmt.Errorf("git history traversal failed: %w", err)
	}
	return fmt.Errorf("git history traversal failed: %s: %w", trimmed, err)
}

func trimHistoryAtAdoptionCommit(commits []importedGitHistoryCommit, adoptionCommit string) ([]importedGitHistoryCommit, bool) {
	for i, commit := range commits {
		if commit.Commit == adoptionCommit {
			return append([]importedGitHistoryCommit(nil), commits[:i+1]...), true
		}
	}
	return nil, false
}

func parseUnixTimestamp(raw string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse git commit timestamp %q: %w", raw, err)
	}
	return value, nil
}

func writeImportedGitHistory(root, adoptionCommit string, commits []importedGitHistoryCommit) (string, error) {
	// Use the full canonical adoption commit SHA in filenames to avoid
	// collisions that can occur when truncating to a short prefix.
	path := filepath.Join(root, "assurance", "backfill", fmt.Sprintf("imported-git-history-%s.json", adoptionCommit))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create backfill directory: %w", err)
	}
	record := importedGitHistoryRecord{
		SchemaVersion:  1,
		Kind:           "history",
		Provenance:     "imported_git_history",
		GeneratedAt:    time.Now().Unix(),
		AdoptionCommit: adoptionCommit,
		Commits:        commits,
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal imported history: %w", err)
	}
	data = append(data, '\n')
	if err := writeAtomicFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write imported history: %w", err)
	}
	return path, nil
}

func shortenCommit(commit string) string {
	if len(commit) <= 12 {
		return commit
	}
	return commit[:12]
}
