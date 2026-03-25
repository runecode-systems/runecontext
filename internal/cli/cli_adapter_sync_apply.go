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
	plannedWrites, writeManifest, err := plannedAdapterWrites(state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(state.managedRoot, 0o755); err != nil {
		return err
	}
	if err := copyManagedFiles(state.sourceRoot, state.managedRoot, state.managedFiles, plannedWrites); err != nil {
		return err
	}
	if err := removeStaleManagedFiles(state.managedRoot, state.managedFiles); err != nil {
		return err
	}
	if writeManifest {
		if err := os.MkdirAll(filepath.Dir(state.manifestPath), 0o755); err != nil {
			return err
		}
		if err := writeAtomicFile(state.manifestPath, state.manifest, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func plannedAdapterWrites(state adapterSyncState) (map[string]struct{}, bool, error) {
	managedRelRoot, err := filepath.Rel(state.absRoot, state.managedRoot)
	if err != nil {
		return nil, false, err
	}
	manifestRelPath, err := filepath.Rel(state.absRoot, state.manifestPath)
	if err != nil {
		return nil, false, err
	}
	managedPrefix := filepath.ToSlash(managedRelRoot) + "/"
	manifestRelPath = filepath.ToSlash(manifestRelPath)
	plannedWrites := make(map[string]struct{})
	writeManifest := false
	for _, mutation := range state.plan {
		if mutation.Action != "created" && mutation.Action != "updated" {
			continue
		}
		if mutation.Path == manifestRelPath {
			writeManifest = true
			continue
		}
		if strings.HasPrefix(mutation.Path, managedPrefix) {
			rel := strings.TrimPrefix(mutation.Path, managedPrefix)
			plannedWrites[filepath.ToSlash(rel)] = struct{}{}
		}
	}
	return plannedWrites, writeManifest, nil
}

func copyManagedFiles(sourceRoot, managedRoot string, sourceFiles []string, plannedWrites map[string]struct{}) error {
	for _, rel := range sourceFiles {
		if _, ok := plannedWrites[filepath.ToSlash(rel)]; !ok {
			continue
		}
		if err := copyManagedFile(sourceRoot, managedRoot, rel); err != nil {
			return err
		}
	}
	return nil
}

func copyManagedFile(sourceRoot, managedRoot, rel string) error {
	srcPath := filepath.Join(sourceRoot, rel)
	dstPath := filepath.Join(managedRoot, rel)
	srcInfo, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 || !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("adapter sync does not support non-regular source files: %s", filepath.ToSlash(srcPath))
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}
	return writeAtomicFile(dstPath, data, srcInfo.Mode().Perm())
}

func ensureAdapterSyncPathsSafe(state adapterSyncState) error {
	return ensurePathsSafe(state.absRoot, state.managedRoot, filepath.Dir(state.manifestPath))
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

func removeStaleManagedFiles(managedRoot string, sourceFiles []string) error {
	stale, err := collectStaleManagedFiles(managedRoot, sourceFiles)
	if err != nil {
		return err
	}
	for _, file := range stale {
		if err := os.Remove(file.absPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return pruneEmptyDirs(managedRoot)
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
