package contracts

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func snapshotLocalTree(sourceRoot string) (*LocalSourceTree, error) {
	realRoot, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		return nil, err
	}
	tempRoot, err := os.MkdirTemp("", "runectx-local-source-")
	if err != nil {
		return nil, err
	}
	snapshotRoot := filepath.Join(tempRoot, "snapshot")
	if err := copyResolvedTree(realRoot, snapshotRoot, realRoot, map[string]struct{}{}, localSnapshotLimits, &snapshotState{}, 0); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, err
	}
	return &LocalSourceTree{Root: snapshotRoot, SnapshotKind: "snapshot_copy", cleanupRoot: tempRoot}, nil
}

func copyResolvedTree(sourcePath, destPath, root string, active map[string]struct{}, limits snapshotLimits, state *snapshotState, depth int) error {
	resolved, err := validateResolvedSnapshotPath(sourcePath, root, active, depth, limits.MaxDepth)
	if err != nil {
		return err
	}
	defer delete(active, resolved)
	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyResolvedDirectory(resolved, destPath, root, active, limits, state, depth)
	}
	return copyResolvedFile(resolved, destPath, root, limits, state, info)
}

func validateResolvedSnapshotPath(sourcePath, root string, active map[string]struct{}, depth, maxDepth int) (string, error) {
	if depth > maxDepth {
		return "", fmt.Errorf("local source tree exceeds maximum depth of %d", maxDepth)
	}
	resolved, err := filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return "", err
	}
	if !isWithinRoot(root, resolved) {
		return "", fmt.Errorf("resolved path %q escapes declared local source tree", resolved)
	}
	if _, ok := active[resolved]; ok {
		return "", fmt.Errorf("symlink cycle detected at %q", resolved)
	}
	active[resolved] = struct{}{}
	return resolved, nil
}

func copyResolvedDirectory(resolved, destPath, root string, active map[string]struct{}, limits snapshotLimits, state *snapshotState, depth int) error {
	if depth > 0 {
		if _, excluded := limits.Excludes[filepath.Base(resolved)]; excluded {
			return nil
		}
	}
	if err := os.MkdirAll(destPath, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := copyResolvedTree(filepath.Join(resolved, entry.Name()), filepath.Join(destPath, entry.Name()), root, active, limits, state, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func copyResolvedFile(resolved, destPath, root string, limits snapshotLimits, state *snapshotState, info os.FileInfo) error {
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported non-regular file %q in local source tree", resolved)
	}
	if err := validateOpenPathWithinRoot(resolved, root); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := updateSnapshotState(state, limits, info.Size()); err != nil {
		return err
	}
	return copyResolvedFileContents(resolved, destPath, info.Mode().Perm())
}

func updateSnapshotState(state *snapshotState, limits snapshotLimits, size int64) error {
	state.files++
	if state.files > limits.MaxFiles {
		return fmt.Errorf("local source tree exceeds maximum file count of %d", limits.MaxFiles)
	}
	state.bytes += size
	if state.bytes > limits.MaxBytes {
		return fmt.Errorf("local source tree exceeds maximum snapshot size of %d bytes", limits.MaxBytes)
	}
	return nil
}

func copyResolvedFileContents(resolved, destPath string, perm fs.FileMode) error {
	src, err := os.Open(resolved)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	closeErr := dst.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func validateOpenPathWithinRoot(resolvedPath, root string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if !isWithinRoot(resolvedRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes declared local source tree", resolvedPath)
	}
	return nil
}

func walkRuneContextFiles(root string, visit func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return visit(path, d)
	})
}
