package cli

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func plannedHostNativeWrites(absRoot string, artifacts []hostNativeArtifact) ([]contracts.FileMutation, error) {
	plan := make([]contracts.FileMutation, 0, len(artifacts))
	for _, artifact := range artifacts {
		action, err := plannedHostNativeFileAction(absRoot, artifact)
		if err != nil {
			return nil, err
		}
		if action == "" {
			continue
		}
		plan = append(plan, contracts.FileMutation{Path: artifact.relPath, Action: action})
	}
	return plan, nil
}

func plannedHostNativeDeletes(absRoot, tool string, previousPaths []string, artifacts []hostNativeArtifact) ([]contracts.FileMutation, error) {
	if len(previousPaths) == 0 {
		return nil, nil
	}
	desired := desiredHostNativePaths(artifacts)
	plan := make([]contracts.FileMutation, 0)
	for _, rel := range previousPaths {
		rel = filepath.ToSlash(strings.TrimSpace(rel))
		if rel == "" {
			continue
		}
		if _, ok := desired[rel]; ok {
			continue
		}
		deleteMutation, err := plannedHostNativeDelete(absRoot, rel, tool)
		if err != nil {
			return nil, err
		}
		if deleteMutation.Path == "" {
			continue
		}
		plan = append(plan, deleteMutation)
	}
	sort.Slice(plan, func(i, j int) bool { return plan[i].Path < plan[j].Path })
	return plan, nil
}

func desiredHostNativePaths(artifacts []hostNativeArtifact) map[string]struct{} {
	desired := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		desired[artifact.relPath] = struct{}{}
	}
	return desired
}

func plannedHostNativeDelete(absRoot, rel, tool string) (contracts.FileMutation, error) {
	absPath := filepath.Join(absRoot, filepath.FromSlash(rel))
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return contracts.FileMutation{}, nil
		}
		return contracts.FileMutation{}, err
	}
	if err := validateHostNativeOwnershipForDelete(data, rel, tool); err != nil {
		return contracts.FileMutation{}, err
	}
	return contracts.FileMutation{Path: rel, Action: "deleted"}, nil
}

func plannedHostNativeFileAction(absRoot string, artifact hostNativeArtifact) (string, error) {
	path := filepath.Join(absRoot, filepath.FromSlash(artifact.relPath))
	current, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "created", nil
		}
		return "", err
	}
	if string(current) == string(artifact.content) {
		return "", nil
	}
	rel := filepath.ToSlash(artifact.relPath)
	if err := validateHostNativeOwnershipForWrite(current, rel, artifact); err != nil {
		return "", err
	}
	return "updated", nil
}
