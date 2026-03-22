package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

var dryRunCloneLimits = snapshotLimits{
	MaxFiles: 10000,
	MaxBytes: 128 << 20,
	MaxDepth: 64,
	Excludes: map[string]struct{}{
		".git":          {},
		"node_modules":  {},
		".cache":        {},
		"__pycache__":   {},
		".pytest_cache": {},
		".mypy_cache":   {},
		".tox":          {},
	},
}

type snapshotLimits struct {
	MaxFiles int
	MaxBytes int64
	MaxDepth int
	Excludes map[string]struct{}
}

type snapshotState struct {
	files int
	bytes int64
}

func runChangeOperation[T any](project *cliProject, machine machineOptions, op func(v *contracts.Validator, loaded *contracts.LoadedProject) (T, error)) (T, error) {
	if !machine.dryRun {
		return op(project.validator, project.loaded)
	}
	return runChangeOperationDryRun(dryRunProjectRoot(project), project.explicitRoot, op)
}

func dryRunProjectRoot(project *cliProject) string {
	if project != nil && project.loaded != nil && project.loaded.Resolution != nil && project.loaded.Resolution.ProjectRoot != "" {
		return project.loaded.Resolution.ProjectRoot
	}
	if project != nil {
		return project.absRoot
	}
	return "."
}

func runChangeOperationDryRun[T any](root string, explicitRoot bool, op func(v *contracts.Validator, loaded *contracts.LoadedProject) (T, error)) (T, error) {
	var zero T
	tempRoot, err := cloneDir(root)
	if err != nil {
		return zero, err
	}
	defer os.RemoveAll(tempRoot)

	_, validator, loaded, err := loadProjectForCLI(tempRoot, explicitRoot)
	if err != nil {
		return zero, err
	}
	defer loaded.Close()

	result, err := op(validator, loaded)
	if err != nil {
		return zero, err
	}
	index, err := validator.ValidateLoadedProject(loaded)
	if err != nil {
		return zero, err
	}
	index.Close()
	return result, nil
}

func cloneDir(srcRoot string) (string, error) {
	root := filepath.Clean(srcRoot)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	targetRoot, err := os.MkdirTemp("", "runectx-dry-run-")
	if err != nil {
		return "", err
	}
	state := &snapshotState{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		return cloneWalkEntry(state, dryRunCloneLimits, root, absRoot, targetRoot, path, d, walkErr)
	})
	if err != nil {
		os.RemoveAll(targetRoot)
		return "", err
	}
	return targetRoot, nil
}

func cloneWalkEntry(state *snapshotState, limits snapshotLimits, root, absRoot, targetRoot, path string, d fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if limits.MaxDepth > 0 && dryRunPathDepth(relPath) > limits.MaxDepth {
		return fmt.Errorf("dry-run clone exceeds maximum depth of %d", limits.MaxDepth)
	}
	targetPath := filepath.Join(targetRoot, relPath)
	if d.IsDir() {
		return cloneDirectoryEntry(relPath, d, targetPath, limits)
	}
	return cloneFileEntry(state, limits, absRoot, path, targetPath)
}

func cloneDirectoryEntry(relPath string, d fs.DirEntry, targetPath string, limits snapshotLimits) error {
	info, err := d.Info()
	if err != nil {
		return err
	}
	if relPath != "." {
		if limits.Excludes != nil {
			if _, ok := limits.Excludes[filepath.Base(relPath)]; ok {
				return filepath.SkipDir
			}
		} else {
			if shouldSkipDirForDryRun(filepath.Base(relPath)) {
				return filepath.SkipDir
			}
		}
	}
	return os.MkdirAll(targetPath, info.Mode().Perm())
}

func cloneFileEntry(state *snapshotState, limits snapshotLimits, absRoot, path, targetPath string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		if err := updateSnapshotState(state, limits, 0); err != nil {
			return err
		}
		return cloneSymlink(absRoot, path, targetPath)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("dry-run clone rejects unsupported file type %q", path)
	}
	if err := updateSnapshotState(state, limits, info.Size()); err != nil {
		return err
	}
	return copyFile(path, targetPath, info.Mode().Perm())
}

func cloneSymlink(absRoot, srcPath, targetPath string) error {
	linkTarget, err := os.Readlink(srcPath)
	if err != nil {
		return err
	}
	if filepath.IsAbs(linkTarget) {
		return fmt.Errorf("dry-run clone rejects absolute symlink %q; convert it to a relative link or run without --dry-run", srcPath)
	}
	resolvedTarget := filepath.Join(filepath.Dir(srcPath), linkTarget)
	absResolved, err := filepath.Abs(resolvedTarget)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absRoot, absResolved)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("dry-run clone rejects relative symlink %q resolving outside project root to %q", srcPath, absResolved)
	}
	return os.Symlink(linkTarget, targetPath)
}

func shouldSkipDirForDryRun(name string) bool {
	_, ok := dryRunCloneLimits.Excludes[name]
	return ok
}

func updateSnapshotState(state *snapshotState, limits snapshotLimits, size int64) error {
	state.files++
	if limits.MaxFiles > 0 && state.files > limits.MaxFiles {
		return fmt.Errorf("dry-run clone exceeds maximum file count of %d", limits.MaxFiles)
	}
	state.bytes += size
	if limits.MaxBytes > 0 && state.bytes > limits.MaxBytes {
		return fmt.Errorf("dry-run clone exceeds maximum size of %d bytes", limits.MaxBytes)
	}
	return nil
}

func dryRunPathDepth(relPath string) int {
	if relPath == "." || relPath == "" {
		return 0
	}
	return strings.Count(filepath.ToSlash(relPath), "/") + 1
}

func copyFile(srcPath, targetPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(target, src); err != nil {
		target.Close()
		return err
	}
	return target.Close()
}
