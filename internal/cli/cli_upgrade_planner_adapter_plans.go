package cli

import (
	"fmt"
	"strings"
)

func collectSingleUpgradeAdapterPlan(absRoot, tool string, includeCreate bool, conflicts, warnings []string) (adapterSyncState, []string, []string, bool, error) {
	state, err := buildAdapterSyncState(adapterRequest{root: absRoot, explicitRoot: true, tool: tool})
	if err != nil {
		handled, nextConflicts, nextWarnings, handleErr := handleUpgradeAdapterPlanError(tool, err, conflicts, warnings)
		if handleErr != nil {
			return adapterSyncState{}, conflicts, warnings, false, handleErr
		}
		if handled {
			return adapterSyncState{}, nextConflicts, nextWarnings, true, nil
		}
		return adapterSyncState{}, nextConflicts, nextWarnings, false, err
	}
	state.plan = filterAdapterMutations(state.plan, includeCreate)
	return state, conflicts, warnings, false, nil
}

func handleUpgradeAdapterPlanError(tool string, err error, conflicts, warnings []string) (bool, []string, []string, error) {
	if strings.Contains(err.Error(), "host-native artifact conflict") {
		return true, append(conflicts, err.Error()), warnings, nil
	}
	if strings.Contains(err.Error(), "could not locate installed adapter packs") || strings.Contains(err.Error(), "not found in installed adapter packs") {
		warning := fmt.Sprintf("optional adapter pack unavailable for %s; skipping host-native readiness checks", tool)
		return true, conflicts, append(warnings, warning), nil
	}
	return false, conflicts, warnings, err
}
