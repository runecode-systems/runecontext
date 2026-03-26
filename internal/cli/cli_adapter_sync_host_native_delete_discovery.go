package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func plannedHostNativeDeletesFromExisting(absRoot, tool string, artifacts []hostNativeArtifact) ([]contracts.FileMutation, error) {
	existing, err := listExistingManagedHostNativePaths(absRoot, tool)
	if err != nil {
		return nil, err
	}
	return plannedHostNativeDeletes(absRoot, tool, existing, artifacts)
}

func listExistingManagedHostNativePaths(absRoot, tool string) ([]string, error) {
	roots := hostNativeToolRoots(tool)
	paths := make([]string, 0)
	for _, root := range roots {
		if err := ensureHostNativeRootSafe(absRoot, root); err != nil {
			return nil, err
		}
		items, err := listExistingManagedHostNativePathsUnderRoot(absRoot, root, tool)
		if err != nil {
			return nil, err
		}
		paths = append(paths, items...)
	}
	sort.Strings(paths)
	return paths, nil
}

func ensureHostNativeRootSafe(absRoot, relRoot string) error {
	absPath := filepath.Join(absRoot, filepath.FromSlash(relRoot))
	return ensurePathsSafe(absRoot, absPath)
}

func hostNativeToolRoots(tool string) []string {
	switch tool {
	case "opencode":
		return []string{".opencode/skills", ".opencode/commands"}
	case "claude-code":
		return []string{".claude/skills", ".claude/commands"}
	case "codex":
		return []string{".agents/skills"}
	default:
		return nil
	}
}

func listExistingManagedHostNativePathsUnderRoot(absRoot, relRoot, tool string) ([]string, error) {
	absPath := filepath.Join(absRoot, filepath.FromSlash(relRoot))
	if !isDirectory(absPath) {
		return nil, nil
	}
	paths := make([]string, 0)
	err := filepath.WalkDir(absPath, func(path string, entry fs.DirEntry, walkErr error) error {
		return collectManagedHostNativePath(absRoot, tool, path, entry, walkErr, &paths)
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func collectManagedHostNativePath(absRoot, tool, path string, entry fs.DirEntry, walkErr error, paths *[]string) error {
	if walkErr != nil || entry.IsDir() {
		return walkErr
	}
	rel, err := filepath.Rel(absRoot, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	candidate, err := hostNativeManagedDeleteCandidate(absRoot, rel, tool)
	if err != nil {
		return err
	}
	if candidate != "" {
		*paths = append(*paths, candidate)
	}
	return nil
}

func hostNativeManagedDeleteCandidate(absRoot, rel, tool string) (string, error) {
	data, err := os.ReadFile(filepath.Join(absRoot, filepath.FromSlash(rel)))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	header, ok := parseHostNativeOwnershipHeader(data)
	if !ok || header.Tool != tool {
		return "", nil
	}
	return rel, nil
}
