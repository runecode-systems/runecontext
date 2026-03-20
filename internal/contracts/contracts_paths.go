package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validatePathMatchedID(path, root string, rawID any) error {
	id := fmt.Sprint(rawID)
	artifactRoot, err := findNearestArtifactRoot(path, root)
	if err != nil {
		return &ValidationError{Path: path, Message: fmt.Sprintf("path does not live under %s/", root)}
	}
	rel, err := filepath.Rel(artifactRoot, filepath.Clean(path))
	if err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	rel = strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	if rel != id {
		return &ValidationError{Path: path, Message: fmt.Sprintf("frontmatter id %q must match path-relative stem %q", id, rel)}
	}
	return nil
}

func findNearestArtifactRoot(path, root string) (string, error) {
	current := filepath.Clean(filepath.Dir(path))
	for {
		if filepath.Base(current) == root {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", os.ErrNotExist
		}
		current = next
	}
}

func walkChangeDirectories(root string, visit func(changeDir string) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := visit(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func walkProjectFiles(root string, visit func(path string) error) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return &ValidationError{Path: root, Message: "expected a directory root"}
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return &ValidationError{Path: root, Message: err.Error()}
	}
	return walkContainedFiles(resolvedRoot, root, map[string]struct{}{}, visit)
}

func walkContainedFiles(boundaryResolved, currentPath string, active map[string]struct{}, visit func(path string) error) error {
	resolvedPath, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		return &ValidationError{Path: currentPath, Message: err.Error()}
	}
	if !isWithinRoot(boundaryResolved, resolvedPath) {
		return &ValidationError{Path: currentPath, Message: fmt.Sprintf("resolved path %q escapes the selected project subtree", resolvedPath)}
	}
	info, err := os.Stat(currentPath)
	if err != nil {
		return &ValidationError{Path: currentPath, Message: err.Error()}
	}
	if info.IsDir() {
		return walkContainedDirectory(boundaryResolved, currentPath, resolvedPath, active, visit)
	}
	if !info.Mode().IsRegular() {
		return &ValidationError{Path: currentPath, Message: fmt.Sprintf("resolved path %q is not a regular file", resolvedPath)}
	}
	return visit(currentPath)
}

func walkContainedDirectory(boundaryResolved, currentPath, resolvedPath string, active map[string]struct{}, visit func(path string) error) error {
	resolvedKey := filepath.Clean(resolvedPath)
	if _, ok := active[resolvedKey]; ok {
		return &ValidationError{Path: currentPath, Message: fmt.Sprintf("symlink cycle detected at %q", currentPath)}
	}
	active[resolvedKey] = struct{}{}
	defer delete(active, resolvedKey)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return &ValidationError{Path: currentPath, Message: err.Error()}
	}
	for _, entry := range entries {
		if err := walkContainedFiles(boundaryResolved, filepath.Join(currentPath, entry.Name()), active, visit); err != nil {
			return err
		}
	}
	return nil
}

func readProjectFile(boundaryPath, path string) ([]byte, error) {
	resolvedBoundary, resolvedPath, err := resolveProjectFilePaths(boundaryPath, path)
	if err != nil {
		return nil, err
	}
	if !isWithinRoot(resolvedBoundary, resolvedPath) {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("resolved path %q escapes the selected project subtree", resolvedPath)}
	}
	info, err := statProjectFile(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, &ValidationError{Path: path, Message: "expected a file, found a directory"}
	}
	if !info.Mode().IsRegular() {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("resolved path %q is not a regular file", resolvedPath)}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	return data, nil
}

func resolveProjectFilePaths(boundaryPath, path string) (string, string, error) {
	resolvedBoundary, err := filepath.EvalSymlinks(boundaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", os.ErrNotExist
		}
		return "", "", &ValidationError{Path: boundaryPath, Message: err.Error()}
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", os.ErrNotExist
		}
		return "", "", &ValidationError{Path: path, Message: err.Error()}
	}
	return resolvedBoundary, resolvedPath, nil
}

func statProjectFile(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	return info, nil
}
