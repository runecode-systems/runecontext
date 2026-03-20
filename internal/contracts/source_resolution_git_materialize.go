package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r gitResolver) materializeCommitToTree(commit, subdir string) (*LocalSourceTree, string, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", err
	}
	if err := fetchCommitIntoRepository(repoRoot, commit); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	return r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
}

func fetchCommitIntoRepository(repoRoot, commit string) error {
	for _, args := range [][]string{{"-C", repoRoot, "fetch", "--quiet", "--no-tags", "origin", "+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"}, {"-C", repoRoot, "cat-file", "-e", commit + "^{commit}"}, {"-C", repoRoot, "checkout", "--quiet", "--detach", commit}} {
		if err := runGit(args...); err != nil {
			if strings.Contains(strings.Join(args, " "), "cat-file") {
				return fmt.Errorf("pinned git commit %q was not found after fetching advertised refs: %v", commit, err)
			}
			return err
		}
	}
	return nil
}

func (r gitResolver) initializeRepository() (string, string, error) {
	tempRoot, err := os.MkdirTemp("", "runectx-git-source-")
	if err != nil {
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	repoRoot := filepath.Join(tempRoot, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		_ = os.RemoveAll(tempRoot)
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	for _, args := range [][]string{{"init", "--quiet", repoRoot}, {"-C", repoRoot, "remote", "add", "origin", r.url}} {
		if err := runGit(args...); err != nil {
			_ = os.RemoveAll(tempRoot)
			return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
		}
	}
	return tempRoot, repoRoot, nil
}

func (r gitResolver) finalizeMaterializedTree(tempRoot, repoRoot, subdir string) (*LocalSourceTree, string, error) {
	commitOutput, err := gitOutput("-C", repoRoot, "rev-parse", "HEAD")
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	commit := strings.TrimSpace(commitOutput)
	materializedRoot := filepath.Clean(filepath.Join(repoRoot, filepath.FromSlash(subdir)))
	if !isWithinRoot(repoRoot, materializedRoot) {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("git source subdir %q escapes the fetched repository root", subdir)}
	}
	info, err := os.Stat(materializedRoot)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git source root %q is not a directory", materializedRoot)}
	}
	return &LocalSourceTree{Root: materializedRoot, SnapshotKind: "git_checkout", cleanupRoot: tempRoot}, commit, nil
}
