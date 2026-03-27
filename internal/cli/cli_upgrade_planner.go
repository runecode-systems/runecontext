package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

var collectUpgradeAdapterPlansFn = collectUpgradeAdapterPlans

func buildUpgradePlan(project *cliProject, requestedTarget string) (upgradePlan, error) {
	plan := basePlanFromProject(project, requestedTarget)
	if done, err := classifyUpgradePlanCommon(&plan, "choose a supported --target-version for this runectx release", "resolve conflicts, then rerun `runectx upgrade`"); done || err != nil {
		return plan, err
	}
	finalizeUpgradeVersionState(&plan, "run `runectx upgrade apply --target-version current` after reviewing stale-file plan")
	return plan, nil
}

func buildUpgradeReadinessFromIndex(absRoot string, index *contracts.ProjectIndex) (upgradePlan, error) {
	plan := basePlanFromIndex(absRoot, index)
	if done, err := classifyUpgradePlanCommon(&plan, "choose a supported upgrade path for this runectx release", "resolve managed ownership conflicts and rerun upgrade"); done || err != nil {
		if err != nil && isOptionalAdapterPackUnavailableError(err) {
			plan.Warnings = append(plan.Warnings, err.Error())
			return plan, nil
		}
		return plan, err
	}
	if plan.TargetVersion == plan.CurrentVersion && hasAdapterMutations(plan.AdapterPlans) {
		plan.State = upgradeStateMixedOrStaleTree
		plan.NextActions = append(plan.NextActions, "rerun `runectx upgrade apply --target-version current` to refresh stale managed artifacts")
		return plan, nil
	}
	if plan.TargetVersion != plan.CurrentVersion {
		plan.State = upgradeStateUpgradeable
		plan.NextActions = append(plan.NextActions, fmt.Sprintf("run `runectx upgrade apply --target-version %s`", plan.TargetVersion))
	}
	return plan, nil
}

func basePlanFromProject(project *cliProject, requestedTarget string) upgradePlan {
	current := strings.TrimSpace(fmt.Sprint(project.loaded.RootConfig["runecontext_version"]))
	target, network := resolveUpgradeTargetVersion(current, requestedTarget)
	plan := upgradePlan{
		State:          upgradeStateCurrent,
		CurrentVersion: current,
		TargetVersion:  target,
		NetworkAccess:  network,
		ConfigPath:     selectedConfigPath(project.loaded),
		ProjectRoot:    project.absRoot,
		AdapterPlans:   map[string]adapterSyncState{},
	}
	if project.loaded != nil && project.loaded.Resolution != nil && strings.TrimSpace(project.loaded.Resolution.ProjectRoot) != "" {
		plan.ProjectRoot = project.loaded.Resolution.ProjectRoot
	}
	if source, ok := project.loaded.RootConfig["source"].(map[string]any); ok {
		plan.SourceType = strings.TrimSpace(fmt.Sprint(source["type"]))
	}
	return plan
}

func basePlanFromIndex(absRoot string, index *contracts.ProjectIndex) upgradePlan {
	current := ""
	if index != nil {
		current = readRunecontextVersionFromConfig(index.RootConfigPath)
	}
	target := normalizedRunecontextVersion()
	if target == "" || target == "0.0.0-dev" {
		target = current
	}
	plan := upgradePlan{
		State:          upgradeStateCurrent,
		CurrentVersion: current,
		TargetVersion:  target,
		NetworkAccess:  false,
		AdapterPlans:   map[string]adapterSyncState{},
		ProjectRoot:    absRoot,
	}
	if index != nil && index.Resolution != nil {
		plan.SourceType = string(index.Resolution.SourceMode)
		plan.ConfigPath = index.RootConfigPath
	}
	return plan
}

func classifyUpgradePlanCommon(plan *upgradePlan, edgeAction, conflictAction string) (bool, error) {
	registry := defaultUpgradePlannerRegistry()
	if classifyUnsupportedVersion(plan, registry, "install a compatible runectx release or manually align runecontext_version before retrying upgrade") {
		return true, nil
	}
	if classifyExternallyManagedPath(plan, readSourcePathFromConfig(plan.ConfigPath), "run upgrade in the external source root that owns this path source") {
		return true, nil
	}
	if classifyMissingUpgradeEdge(plan, registry, edgeAction) {
		return true, nil
	}
	if err := classifyAdapterState(plan, plan.TargetVersion != plan.CurrentVersion, conflictAction); err != nil {
		return false, err
	}
	if plan.State == upgradeStateConflicted {
		return true, nil
	}
	return false, nil
}

func classifyUnsupportedVersion(plan *upgradePlan, registry upgradePlannerRegistry, nextAction string) bool {
	if isSupportedProjectVersion(plan.CurrentVersion, registry) {
		return false
	}
	plan.State = upgradeStateUnsupportedProjectVersion
	plan.PlanActions = append(plan.PlanActions, fmt.Sprintf("project runecontext_version %s is unsupported by this runectx release", plan.CurrentVersion))
	plan.NextActions = append(plan.NextActions, nextAction)
	return true
}

func classifyExternallyManagedPath(plan *upgradePlan, sourcePath, fallbackAction string) bool {
	if plan.SourceType != "path" {
		return false
	}
	plan.State = upgradeStateConflicted
	plan.PlanActions = append(plan.PlanActions, "source.type=path is externally managed and cannot be upgraded in place")
	if sourcePath != "" {
		plan.NextActions = append(plan.NextActions, fmt.Sprintf("navigate to %s and run runectx upgrade there", sourcePath))
	} else {
		plan.NextActions = append(plan.NextActions, fallbackAction)
	}
	return true
}

func classifyMissingUpgradeEdge(plan *upgradePlan, registry upgradePlannerRegistry, nextAction string) bool {
	if plan.TargetVersion == plan.CurrentVersion || registry.hasEdge(plan.CurrentVersion, plan.TargetVersion) {
		return false
	}
	plan.State = upgradeStateUnsupportedProjectVersion
	plan.PlanActions = append(plan.PlanActions, fmt.Sprintf("no registered upgrader edge for runecontext_version transition %s -> %s", plan.CurrentVersion, plan.TargetVersion))
	plan.NextActions = append(plan.NextActions, nextAction)
	return true
}

func classifyAdapterState(plan *upgradePlan, includeCreate bool, nextAction string) error {
	adapterPlans, conflicts, warnings, err := collectUpgradeAdapterPlansFn(plan.ProjectRoot, includeCreate)
	if err != nil {
		if isOptionalAdapterPackUnavailableError(err) {
			plan.Warnings = append(plan.Warnings, err.Error())
			return nil
		}
		return err
	}
	plan.AdapterPlans = adapterPlans
	plan.Conflicts = append(plan.Conflicts, conflicts...)
	plan.Warnings = append(plan.Warnings, warnings...)
	if len(plan.Conflicts) == 0 {
		return nil
	}
	plan.State = upgradeStateConflicted
	plan.PlanActions = append(plan.PlanActions, "review and resolve managed artifact ownership conflicts before apply")
	plan.NextActions = append(plan.NextActions, nextAction)
	return nil
}

func isOptionalAdapterPackUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "could not locate installed adapter packs") || strings.Contains(message, "not found in installed adapter packs")
}

func finalizeUpgradeVersionState(plan *upgradePlan, staleAction string) {
	if plan.TargetVersion == plan.CurrentVersion {
		if hasAdapterMutations(plan.AdapterPlans) {
			plan.State = upgradeStateMixedOrStaleTree
			plan.PlanActions = append(plan.PlanActions, collectAdapterPlanActions(plan.AdapterPlans)...)
			plan.NextActions = append(plan.NextActions, staleAction)
			plan.ApplyMutations = append(plan.ApplyMutations, collectAdapterMutationLines(plan.AdapterPlans)...)
		} else {
			plan.PlanActions = append(plan.PlanActions, "no changes required")
		}
		return
	}
	plan.State = upgradeStateUpgradeable
	plan.PlanActions = append(plan.PlanActions, fmt.Sprintf("set runecontext_version to %s", plan.TargetVersion))
	plan.ApplyMutations = append(plan.ApplyMutations, fmt.Sprintf("updated %s", filepath.ToSlash(filepath.Base(plan.ConfigPath))))
	if hasAdapterMutations(plan.AdapterPlans) {
		plan.PlanActions = append(plan.PlanActions, collectAdapterPlanActions(plan.AdapterPlans)...)
		plan.ApplyMutations = append(plan.ApplyMutations, collectAdapterMutationLines(plan.AdapterPlans)...)
	}
}

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
