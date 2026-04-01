package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const maxManagedHostNativeArtifactScanBytes = 1 << 20

func collectUpgradeAdapterPlans(absRoot string) (map[string]adapterSyncState, []string, []string, error) {
	states := map[string]adapterSyncState{}
	conflicts := make([]string, 0)
	warnings := make([]string, 0)
	for _, tool := range []string{"opencode", "claude-code", "codex"} {
		includeCreate, err := hasManagedHostNativeArtifactsForTool(absRoot, tool)
		if err != nil {
			return nil, nil, nil, err
		}
		nextState, nextConflicts, nextWarnings, skip, err := collectSingleUpgradeAdapterPlan(absRoot, tool, includeCreate, conflicts, warnings)
		if err != nil {
			return nil, nil, nil, err
		}
		conflicts = nextConflicts
		warnings = nextWarnings
		if skip {
			continue
		}
		if len(nextState.plan) > 0 {
			states[tool] = nextState
		}
	}
	return states, conflicts, warnings, nil
}

func hasManagedHostNativeArtifactsForTool(absRoot, tool string) (bool, error) {
	for _, relDir := range hostNativeRootsForTool(tool) {
		managed, err := hasManagedHostNativeArtifactsInDir(filepath.Join(absRoot, filepath.FromSlash(relDir)), tool)
		if err != nil {
			return false, err
		}
		if managed {
			return true, nil
		}
	}
	return false, nil
}

func hasManagedHostNativeArtifactsInDir(root, tool string) (bool, error) {
	exists, err := existingHostNativeDir(root)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	managed, err := scanManagedHostNativeArtifactsInDir(root, tool)
	if err != nil {
		return false, err
	}
	return managed, nil
}

func scanManagedHostNativeArtifactsInDir(root, tool string) (bool, error) {
	var managed bool
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if skip, err := skipManagedHostNativeWalkEntry(root, path, entry, walkErr); skip || err != nil {
			return err
		}
		return markManagedHostNativeFile(path, tool, &managed)
	})
	return managed, err
}

func skipManagedHostNativeWalkEntry(root, path string, entry os.DirEntry, walkErr error) (bool, error) {
	if walkErr != nil {
		return true, walkErr
	}
	if entry.IsDir() {
		return true, nil
	}
	if entry.Type()&os.ModeSymlink != 0 {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return true, err
		}
		return true, fmt.Errorf("managed host-native scan rejects symlinked path %s", filepath.ToSlash(rel))
	}
	return false, nil
}

func markManagedHostNativeFile(path, tool string, managed *bool) error {
	if *managed {
		return nil
	}
	owned, err := isManagedHostNativeFileForTool(path, tool)
	if err != nil {
		return err
	}
	if !owned {
		return nil
	}
	*managed = true
	return nil
}

func existingHostNativeDir(root string) (bool, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func isManagedHostNativeFileForTool(path, tool string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.Size() > maxManagedHostNativeArtifactScanBytes {
		return false, fmt.Errorf("managed host-native scan rejects oversized file %s", filepath.ToSlash(path))
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	header, ok := parseHostNativeOwnershipHeader(content)
	return ok && header.Tool == tool, nil
}

func hostNativeRootsForTool(tool string) []string {
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

func filterAdapterMutations(mutations []contracts.FileMutation, includeCreate bool) []contracts.FileMutation {
	result := make([]contracts.FileMutation, 0, len(mutations))
	for _, mutation := range mutations {
		if mutation.Action == "created" && !includeCreate {
			continue
		}
		result = append(result, mutation)
	}
	return result
}

func hasAdapterMutations(states map[string]adapterSyncState) bool {
	for _, state := range states {
		if len(state.plan) > 0 {
			return true
		}
	}
	return false
}

func collectAdapterPlanActions(states map[string]adapterSyncState) []string {
	actions := make([]string, 0)
	for _, tool := range sortedMapKeys(states) {
		for _, mutation := range states[tool].plan {
			actions = append(actions, fmt.Sprintf("refresh host-native %s artifact: %s %s", tool, mutation.Action, mutation.Path))
		}
	}
	return actions
}

func collectAdapterMutationLines(states map[string]adapterSyncState) []string {
	changes := make([]string, 0)
	for _, tool := range sortedMapKeys(states) {
		for _, mutation := range states[tool].plan {
			changes = append(changes, fmt.Sprintf("%s %s", mutation.Action, mutation.Path))
		}
	}
	return changes
}
