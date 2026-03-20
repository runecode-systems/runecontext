package contracts

import (
	"fmt"
	"os"
	"path/filepath"
)

func walkBundleFiles(contentRoot, aspectRoot, logicalPath string, active map[string]struct{}, state *bundleWalkState, visit func(string) error) error {
	state = ensureBundleWalkState(state)
	if err := enterBundleWalk(state); err != nil {
		return err
	}
	defer leaveBundleWalk(state)
	resolvedPath, err := resolveBundleWalkPath(logicalPath)
	if err != nil {
		return err
	}
	if err := validateBundleWalkRoot(contentRoot, aspectRoot, resolvedPath); err != nil {
		return err
	}
	info, err := os.Stat(logicalPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return walkBundleDirectory(contentRoot, aspectRoot, logicalPath, resolvedPath, active, state, visit)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("resolved path %q is not a regular file", resolvedPath)
	}
	if err := countBundleWalkFile(state); err != nil {
		return err
	}
	return visit(logicalPath)
}

func ensureBundleWalkState(state *bundleWalkState) *bundleWalkState {
	if state == nil {
		return &bundleWalkState{}
	}
	return state
}

func enterBundleWalk(state *bundleWalkState) error {
	state.depth++
	if state.depth > bundleTraversalLimits.MaxDepth {
		return fmt.Errorf("bundle traversal exceeds maximum depth of %d", bundleTraversalLimits.MaxDepth)
	}
	return nil
}

func leaveBundleWalk(state *bundleWalkState) { state.depth-- }

func resolveBundleWalkPath(logicalPath string) (string, error) {
	resolvedPath, err := filepath.EvalSymlinks(logicalPath)
	if err != nil {
		return "", err
	}
	return resolvedPath, nil
}

func validateBundleWalkRoot(contentRoot, aspectRoot, resolvedPath string) error {
	if resolvedPath == "" {
		return nil
	}
	if !isWithinRoot(contentRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes the RuneContext root", resolvedPath)
	}
	if !isWithinRoot(aspectRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes the selected aspect root", resolvedPath)
	}
	return nil
}

func walkBundleDirectory(contentRoot, aspectRoot, logicalPath, resolvedPath string, active map[string]struct{}, state *bundleWalkState, visit func(string) error) error {
	if resolvedPath == "" {
		return nil
	}
	resolvedKey := filepath.Clean(resolvedPath)
	if _, ok := active[resolvedKey]; ok {
		return fmt.Errorf("encountered a symlink cycle at %q", logicalPath)
	}
	active[resolvedKey] = struct{}{}
	defer delete(active, resolvedKey)
	entries, err := os.ReadDir(logicalPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := walkBundleFiles(contentRoot, aspectRoot, filepath.Join(logicalPath, entry.Name()), active, state, visit); err != nil {
			return err
		}
	}
	return nil
}

func countBundleWalkFile(state *bundleWalkState) error {
	state.files++
	if state.files > bundleTraversalLimits.MaxFiles {
		return fmt.Errorf("bundle traversal exceeds maximum file count of %d", bundleTraversalLimits.MaxFiles)
	}
	return nil
}

func validateResolvedBundlePath(logicalPath, contentRoot, aspectRoot string) error {
	canonicalAspectRoot, err := canonicalContainedRoot(aspectRoot)
	if err != nil {
		return err
	}
	resolvedPath, err := filepath.EvalSymlinks(logicalPath)
	if err != nil {
		return err
	}
	if !isWithinRoot(contentRoot, resolvedPath) {
		return fmt.Errorf("resolves to %q, which escapes the RuneContext root", resolvedPath)
	}
	if !isWithinRoot(canonicalAspectRoot, resolvedPath) {
		return fmt.Errorf("resolves to %q, which escapes the selected aspect root", resolvedPath)
	}
	return nil
}

func canonicalContainedRoot(root string) (string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Clean(root), nil
		}
		return "", fmt.Errorf("resolve root %q: %w", root, err)
	}
	return filepath.Clean(resolvedRoot), nil
}
