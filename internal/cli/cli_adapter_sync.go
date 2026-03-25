package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type adapterSyncState struct {
	absRoot      string
	tool         string
	sourceRoot   string
	managedRoot  string
	manifestPath string
	manifest     []byte
	managedFiles []string
	plan         []contracts.FileMutation
}

func buildAdapterSyncState(request adapterRequest) (adapterSyncState, error) {
	absRoot, err := resolveAbsoluteRoot(request.root)
	if err != nil {
		return adapterSyncState{}, err
	}
	adaptersRoot, err := locateAdaptersRoot()
	if err != nil {
		return adapterSyncState{}, err
	}
	sourceRoot := filepath.Join(adaptersRoot, request.tool)
	if !isDirectory(sourceRoot) {
		return adapterSyncState{}, fmt.Errorf("adapter %q not found in installed adapter packs", request.tool)
	}
	managedRoot := filepath.Join(absRoot, ".runecontext", "adapters", request.tool, "managed")
	manifestPath := filepath.Join(absRoot, ".runecontext", "adapters", request.tool, "sync-manifest.yaml")
	managedFiles, err := listRelativeFiles(sourceRoot)
	if err != nil {
		return adapterSyncState{}, err
	}
	manifest := buildAdapterManifest(request.tool, managedFiles)
	plan, err := buildAdapterSyncPlan(absRoot, sourceRoot, managedRoot, manifestPath, managedFiles, manifest)
	if err != nil {
		return adapterSyncState{}, err
	}
	return adapterSyncState{
		absRoot:      absRoot,
		tool:         request.tool,
		sourceRoot:   sourceRoot,
		managedRoot:  managedRoot,
		manifestPath: manifestPath,
		manifest:     manifest,
		managedFiles: managedFiles,
		plan:         plan,
	}, nil
}

func locateAdaptersRoot() (string, error) {
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(schemaRoot, "adapters"),
		filepath.Join(filepath.Dir(schemaRoot), "adapters"),
	}
	for _, candidate := range candidates {
		if isDirectory(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not locate installed adapter packs")
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func listRelativeFiles(root string) ([]string, error) {
	files := make([]string, 0, 8)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("adapter pack contains unsupported symlink at %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func buildAdapterManifest(tool string, managedFiles []string) []byte {
	lines := []string{
		"schema_version: 1",
		"adapter: " + tool,
		"source: local_release",
		"manifest_kind: convenience_metadata",
		fmt.Sprintf("managed_file_count: %d", len(managedFiles)),
		"managed_files:",
	}
	for _, rel := range managedFiles {
		lines = append(lines, "  - managed/"+rel)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func buildAdapterSyncPlan(absRoot, sourceRoot, managedRoot, manifestPath string, sourceFiles []string, manifest []byte) ([]contracts.FileMutation, error) {
	plan, err := plannedManagedWrites(absRoot, sourceRoot, managedRoot, sourceFiles)
	if err != nil {
		return nil, err
	}
	deletes, err := plannedManagedDeletes(absRoot, managedRoot, sourceFiles)
	if err != nil {
		return nil, err
	}
	plan = append(plan, deletes...)
	manifestAction, err := plannedManifestAction(manifestPath, manifest)
	if err != nil {
		return nil, err
	}
	if manifestAction != "" {
		relOut, err := filepath.Rel(absRoot, manifestPath)
		if err != nil {
			return nil, err
		}
		plan = append(plan, contracts.FileMutation{Path: filepath.ToSlash(relOut), Action: manifestAction})
	}
	sort.Slice(plan, func(i, j int) bool {
		if plan[i].Path == plan[j].Path {
			return plan[i].Action < plan[j].Action
		}
		return plan[i].Path < plan[j].Path
	})
	return plan, nil
}

func plannedManagedWrites(absRoot, sourceRoot, managedRoot string, sourceFiles []string) ([]contracts.FileMutation, error) {
	plan := make([]contracts.FileMutation, 0, len(sourceFiles))
	for _, rel := range sourceFiles {
		action, err := plannedFileAction(filepath.Join(sourceRoot, rel), filepath.Join(managedRoot, rel))
		if err != nil {
			return nil, err
		}
		if action == "" {
			continue
		}
		relOut, err := filepath.Rel(absRoot, filepath.Join(managedRoot, rel))
		if err != nil {
			return nil, err
		}
		plan = append(plan, contracts.FileMutation{Path: filepath.ToSlash(relOut), Action: action})
	}
	return plan, nil
}

func plannedFileAction(srcPath, dstPath string) (string, error) {
	srcData, err := os.ReadFile(srcPath)
	if err != nil {
		return "", err
	}
	dstData, err := os.ReadFile(dstPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "created", nil
		}
		return "", err
	}
	if string(srcData) == string(dstData) {
		return "", nil
	}
	return "updated", nil
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

func plannedManifestAction(path string, manifest []byte) (string, error) {
	current, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "created", nil
		}
		return "", err
	}
	if string(current) == string(manifest) {
		return "", nil
	}
	return "updated", nil
}

type staleManagedFile struct {
	absPath string
	relPath string
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
