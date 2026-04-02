package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func copyPassthroughTree(sourceRoot, targetRoot string, excludes map[string]struct{}) error {
	return filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, skip, err := passthroughRelativePath(sourceRoot, path, excludes)
		if err != nil || skip {
			return err
		}
		targetPath := filepath.Join(targetRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("passthrough source entry %q must not be a symlink", rel)
		}
		return copyFileWithMode(path, targetPath)
	})
}

func passthroughRelativePath(sourceRoot, path string, excludes map[string]struct{}) (string, bool, error) {
	rel, err := filepath.Rel(sourceRoot, path)
	if err != nil {
		return "", false, err
	}
	if rel == "." {
		return "", true, nil
	}
	rel = filepath.ToSlash(rel)
	_, excluded := excludes[rel]
	return rel, excluded, nil
}

func copyFileWithMode(source, target string) error {
	lstatInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if lstatInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("passthrough source file %q must not be a symlink", source)
	}
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()
	fstatInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(lstatInfo, fstatInfo) {
		return fmt.Errorf("passthrough source file %q changed during copy", source)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, fstatInfo.Mode().Perm())
}
