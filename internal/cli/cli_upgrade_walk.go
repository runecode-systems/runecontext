package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var upgradeWalkExcludeDirs = map[string]struct{}{
	".git":          {},
	".direnv":       {},
	"node_modules":  {},
	".cache":        {},
	"__pycache__":   {},
	".pytest_cache": {},
	".mypy_cache":   {},
	".tox":          {},
}

type upgradeWalkPolicy struct {
	root           string
	protectedPaths []string
	gitIgnoreCheck gitIgnoreChecker
}

type upgradeWalkDecision int

const (
	upgradeWalkProceed upgradeWalkDecision = iota
	upgradeWalkSkip
	upgradeWalkSkipDir
	upgradeWalkDir
)

type gitIgnoreChecker interface {
	IsIgnored(root string, relPath string) bool
}

type gitCheckIgnoreChecker struct{}

var upgradeGitIgnoreChecker gitIgnoreChecker = gitCheckIgnoreChecker{}

func newUpgradeWalkPolicy(root string, plan upgradePlan) upgradeWalkPolicy {
	protected := collectUpgradeProtectedRelPaths(root, plan)
	return upgradeWalkPolicy{
		root:           root,
		protectedPaths: protected,
		gitIgnoreCheck: upgradeGitIgnoreChecker,
	}
}

func collectUpgradeProtectedRelPaths(root string, plan upgradePlan) []string {
	protected := []string{"runecontext.yaml"}
	if rel, ok := upgradeRelPathWithinRoot(root, plan.ConfigPath); ok {
		protected = append(protected, rel)
	}
	if plan.SourceType == "embedded" {
		if rel, ok := upgradeProtectedSourceRelPath(root, plan.ConfigPath); ok {
			protected = append(protected, rel)
		}
	}
	for _, tool := range []string{"opencode", "claude-code", "codex"} {
		protected = append(protected, hostNativeRootsForTool(tool)...)
	}
	return uniqueSortedUpgradePaths(protected)
}

func upgradeProtectedSourceRelPath(root, configPath string) (string, bool) {
	sourcePath := strings.TrimSpace(readSourcePathFromConfig(configPath))
	if sourcePath == "" {
		return "", false
	}
	if filepath.IsAbs(sourcePath) {
		return upgradeRelPathWithinRoot(root, sourcePath)
	}
	return upgradeRelPathWithinRoot(root, filepath.Join(root, filepath.FromSlash(sourcePath)))
}

func uniqueSortedUpgradePaths(paths []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		cleaned := strings.Trim(filepath.ToSlash(filepath.Clean(strings.TrimSpace(path))), "/")
		if cleaned == "." || cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		result = append(result, cleaned)
	}
	return result
}

func upgradeRelPathWithinRoot(root, path string) (string, bool) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(path) == "" {
		return "", false
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." || rel == "" || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}

func (p upgradeWalkPolicy) shouldSkipDir(rel string) bool {
	parts := strings.Split(strings.Trim(filepath.ToSlash(rel), "/"), "/")
	for _, part := range parts {
		if _, ok := upgradeWalkExcludeDirs[part]; ok {
			return true
		}
	}
	return false
}

func (p upgradeWalkPolicy) shouldRejectSymlink(rel string) bool {
	rel = strings.Trim(filepath.ToSlash(rel), "/")
	if rel == "" {
		return false
	}
	for _, protected := range p.protectedPaths {
		if rel == protected || strings.HasPrefix(rel, protected+"/") || strings.HasPrefix(protected, rel+"/") {
			return true
		}
	}
	if p.gitIgnoreCheck != nil && p.gitIgnoreCheck.IsIgnored(p.root, rel) {
		return false
	}
	return true
}

func classifyUpgradeWalkEntry(root, path string, entry os.DirEntry, walkErr error, policy upgradeWalkPolicy) (string, upgradeWalkDecision, error) {
	if walkErr != nil {
		return "", upgradeWalkProceed, walkErr
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", upgradeWalkProceed, err
	}
	if rel != "." && policy.shouldSkipDir(rel) {
		if entry.IsDir() {
			return rel, upgradeWalkSkipDir, nil
		}
		return rel, upgradeWalkSkip, nil
	}
	if entry.IsDir() {
		return rel, upgradeWalkDir, nil
	}
	if entry.Type()&os.ModeSymlink != 0 {
		if policy.shouldRejectSymlink(rel) {
			return rel, upgradeWalkProceed, fmt.Errorf("upgrade staging rejects symlinked path %s", filepath.ToSlash(rel))
		}
		return rel, upgradeWalkSkip, nil
	}
	return rel, upgradeWalkProceed, nil
}

func (gitCheckIgnoreChecker) IsIgnored(root, relPath string) bool {
	relPath = strings.TrimSpace(filepath.ToSlash(relPath))
	if relPath == "" {
		return false
	}
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	cmd := exec.Command("git", "check-ignore", "--quiet", "--", relPath)
	cmd.Dir = root
	cmd.Env = sanitizedUpgradeGitCheckIgnoreEnv()
	err := cmd.Run()
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	return exitErr.ExitCode() == 0
}

func sanitizedUpgradeGitCheckIgnoreEnv() []string {
	env := []string{
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_DISCOVERY_ACROSS_FILESYSTEM=0",
		"LANG=C",
		"LC_ALL=C",
	}
	for _, key := range []string{"PATH", "HOME", "XDG_CONFIG_HOME", "SYSTEMROOT", "TMPDIR", "TMP", "TEMP"} {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}
