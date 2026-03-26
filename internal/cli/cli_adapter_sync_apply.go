package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func applyAdapterSync(state adapterSyncState) error {
	if err := ensureAdapterSyncPathsSafe(state); err != nil {
		return err
	}
	plannedWrites, err := plannedAdapterWrites(state)
	if err != nil {
		return err
	}
	return applyAdapterSyncMutations(state, plannedWrites)
}

func applyAdapterSyncMutations(state adapterSyncState, plannedWrites map[string]struct{}) error {
	if err := applyHostNativeArtifactWrites(state.absRoot, state.hostNativeFiles, plannedWrites); err != nil {
		return err
	}
	if err := applyHostNativeArtifactDeletes(state.absRoot, state.tool, state.plan); err != nil {
		return err
	}
	return nil
}

func plannedAdapterWrites(state adapterSyncState) (map[string]struct{}, error) {
	plannedWrites := make(map[string]struct{})
	for _, mutation := range state.plan {
		if mutation.Action != "created" && mutation.Action != "updated" {
			continue
		}
		plannedWrites[mutation.Path] = struct{}{}
	}
	return plannedWrites, nil
}

func ensureAdapterSyncPathsSafe(state adapterSyncState) error {
	paths := make([]string, 0, len(state.hostNativeFiles)+len(state.plan))
	for _, artifact := range state.hostNativeFiles {
		paths = append(paths, filepath.Join(state.absRoot, filepath.FromSlash(artifact.relPath)))
	}
	for _, mutation := range state.plan {
		paths = append(paths, filepath.Join(state.absRoot, filepath.FromSlash(mutation.Path)))
	}
	return ensurePathsSafe(state.absRoot, paths...)
}

func ensurePathsSafe(root string, paths ...string) error {
	root = filepath.Clean(root)
	if err := ensureNotSymlink(root); err != nil {
		return err
	}
	for _, path := range paths {
		if err := ensurePathWithinRoot(root, path); err != nil {
			return err
		}
	}
	return nil
}

func ensurePathWithinRoot(root, path string) error {
	fullPath := filepath.Clean(path)
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return err
	}
	if rel == "" {
		rel = "."
	}
	if rel == "." {
		return nil
	}
	if strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || rel == ".." {
		return fmt.Errorf("adapter sync target %s escapes repository root %s", filepath.ToSlash(fullPath), filepath.ToSlash(root))
	}
	return ensureNoSymlinkSegments(root, rel)
}

func ensureNoSymlinkSegments(root, rel string) error {
	current := root
	for _, segment := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, segment)
		if err := ensureNotSymlink(current); err != nil {
			return err
		}
	}
	return nil
}

func ensureNotSymlink(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("adapter sync rejects symlinked path %s", filepath.ToSlash(path))
	}
	return nil
}

func pruneEmptyDirs(root string) error {
	if !isDirectory(root) {
		return nil
	}
	dirs, err := collectDirectoryTree(root)
	if err != nil {
		return err
	}
	for _, path := range dirs {
		if path == root {
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			_ = os.Remove(path)
		}
	}
	return nil
}

func collectDirectoryTree(root string) ([]string, error) {
	dirs := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	return dirs, nil
}
