package cli

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type staleManagedFile struct {
	absPath string
	relPath string
}

func plannedManagedDeletes(absRoot, managedRoot string, sourceFiles []string) ([]contracts.FileMutation, error) {
	stale, err := collectStaleManagedFiles(managedRoot, sourceFiles)
	if err != nil {
		return nil, err
	}
	deletes := make([]contracts.FileMutation, 0, len(stale))
	for _, file := range stale {
		out, relErr := filepath.Rel(absRoot, file.absPath)
		if relErr != nil {
			return nil, relErr
		}
		deletes = append(deletes, contracts.FileMutation{Path: filepath.ToSlash(out), Action: "deleted"})
	}
	return deletes, nil
}

func collectStaleManagedFiles(managedRoot string, sourceFiles []string) ([]staleManagedFile, error) {
	if !isDirectory(managedRoot) {
		return nil, nil
	}
	allowed := make(map[string]struct{}, len(sourceFiles))
	for _, rel := range sourceFiles {
		allowed[filepath.ToSlash(rel)] = struct{}{}
	}
	stale := make([]staleManagedFile, 0)
	err := filepath.WalkDir(managedRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		rel, err := filepath.Rel(managedRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if _, ok := allowed[rel]; !ok {
			stale = append(stale, staleManagedFile{absPath: path, relPath: rel})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(stale, func(i, j int) bool { return stale[i].relPath < stale[j].relPath })
	return stale, nil
}
