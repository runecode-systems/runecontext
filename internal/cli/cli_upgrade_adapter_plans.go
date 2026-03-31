package cli

import (
	"fmt"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func collectUpgradeAdapterPlans(absRoot string, includeCreate bool) (map[string]adapterSyncState, []string, []string, error) {
	states := map[string]adapterSyncState{}
	conflicts := make([]string, 0)
	warnings := make([]string, 0)
	for _, tool := range []string{"opencode", "claude-code", "codex"} {
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
