package contracts

import (
	"fmt"
	"strings"
)

func PlanChangeDecomposition(v *Validator, loaded *LoadedProject, options ChangeDecompositionPlanOptions) (*ChangeDecompositionPlanResult, error) {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()

	plan, err := normalizeDecompositionPlan(options)
	if err != nil {
		return nil, err
	}
	if err := validateDecompositionPlanAgainstProject(index, plan, false); err != nil {
		return nil, err
	}
	graph, err := BuildSplitChangeGraph(plan)
	if err != nil {
		return nil, err
	}
	return buildDecompositionPlanResult(plan, graph), nil
}

func ApplyChangeDecomposition(v *Validator, loaded *LoadedProject, options ChangeDecompositionApplyOptions) (*ChangeDecompositionApplyResult, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()

	plan, err := normalizeDecompositionPlan(ChangeDecompositionPlanOptions(options))
	if err != nil {
		return nil, err
	}
	if err := validateDecompositionPlanAgainstProject(index, plan, true); err != nil {
		return nil, err
	}
	graph, err := BuildSplitChangeGraph(plan)
	if err != nil {
		return nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	writes, changedFiles, err := buildDecompositionStatusWrites(v, index, writableRoot, graph)
	if err != nil {
		return nil, err
	}
	if err := applyFileRewritesTransaction(writes, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	}); err != nil {
		return nil, err
	}
	return buildDecompositionApplyResult(plan, graph, changedFiles), nil
}

func normalizeDecompositionPlan(options ChangeDecompositionPlanOptions) (SplitChangePlan, error) {
	umbrellaID := strings.TrimSpace(options.UmbrellaID)
	if umbrellaID == "" {
		return SplitChangePlan{}, fmt.Errorf("decomposition plan requires umbrella change ID")
	}
	if len(options.SubChanges) == 0 {
		return SplitChangePlan{}, fmt.Errorf("decomposition plan requires at least one sub-change")
	}
	subChanges := make([]SplitSubChange, 0, len(options.SubChanges))
	for _, sub := range options.SubChanges {
		subID := strings.TrimSpace(sub.ID)
		if subID == "" {
			return SplitChangePlan{}, fmt.Errorf("decomposition sub-change ID must not be blank")
		}
		dependsOn := make([]string, 0, len(sub.DependsOn))
		for _, dep := range sub.DependsOn {
			trimmed := strings.TrimSpace(dep)
			if trimmed == "" {
				return SplitChangePlan{}, fmt.Errorf("decomposition depends_on entries must not be blank")
			}
			dependsOn = append(dependsOn, trimmed)
		}
		subChanges = append(subChanges, SplitSubChange{ID: subID, DependsOn: uniqueSortedStrings(dependsOn)})
	}
	return SplitChangePlan{UmbrellaID: umbrellaID, SubChanges: subChanges}, nil
}

func validateDecompositionPlanAgainstProject(index *ProjectIndex, plan SplitChangePlan, forApply bool) error {
	if err := validateDecompositionUmbrella(index, plan.UmbrellaID, forApply); err != nil {
		return err
	}
	for _, sub := range plan.SubChanges {
		if err := validateDecompositionSubChange(index, sub.ID, forApply); err != nil {
			return err
		}
	}
	return nil
}

func validateDecompositionUmbrella(index *ProjectIndex, umbrellaID string, forApply bool) error {
	umbrella := index.Changes[umbrellaID]
	if umbrella == nil {
		return fmt.Errorf("change %q does not exist", umbrellaID)
	}
	if strings.TrimSpace(umbrella.Type) != "project" {
		return fmt.Errorf("change %q is type %q; decomposition umbrella must be type project", umbrella.ID, umbrella.Type)
	}
	if forApply && isTerminalLifecycleStatus(umbrella.Status) {
		return fmt.Errorf("change %q is already in terminal status %q and cannot accept decomposition apply edits", umbrella.ID, umbrella.Status)
	}
	return nil
}

func validateDecompositionSubChange(index *ProjectIndex, subChangeID string, forApply bool) error {
	record := index.Changes[subChangeID]
	if record == nil {
		return fmt.Errorf("change %q does not exist", subChangeID)
	}
	if strings.TrimSpace(record.Type) != "feature" {
		return fmt.Errorf("change %q is type %q; decomposition sub-changes must be type feature", subChangeID, record.Type)
	}
	if forApply && isTerminalLifecycleStatus(record.Status) {
		return fmt.Errorf("change %q is already in terminal status %q and cannot accept decomposition apply edits", subChangeID, record.Status)
	}
	return nil
}

func buildDecompositionPlanResult(plan SplitChangePlan, graph map[string]ChangeGraphLinks) *ChangeDecompositionPlanResult {
	return &ChangeDecompositionPlanResult{
		UmbrellaID: plan.UmbrellaID,
		NodeIDs:    decompositionNodeIDs(plan.UmbrellaID, plan.SubChanges),
		Graph:      cloneChangeGraphLinksMap(graph),
	}
}

func buildDecompositionApplyResult(plan SplitChangePlan, graph map[string]ChangeGraphLinks, changedFiles []FileMutation) *ChangeDecompositionApplyResult {
	return &ChangeDecompositionApplyResult{
		UmbrellaID:   plan.UmbrellaID,
		NodeIDs:      decompositionNodeIDs(plan.UmbrellaID, plan.SubChanges),
		Graph:        cloneChangeGraphLinksMap(graph),
		ChangedFiles: append([]FileMutation(nil), changedFiles...),
	}
}

func decompositionNodeIDs(umbrellaID string, subChanges []SplitSubChange) []string {
	ids := make([]string, 0, len(subChanges)+1)
	ids = append(ids, umbrellaID)
	for _, sub := range subChanges {
		ids = append(ids, sub.ID)
	}
	return uniqueSortedStrings(ids)
}

func cloneChangeGraphLinksMap(input map[string]ChangeGraphLinks) map[string]ChangeGraphLinks {
	if input == nil {
		return nil
	}
	cloned := make(map[string]ChangeGraphLinks, len(input))
	for id, links := range input {
		cloned[id] = ChangeGraphLinks{
			RelatedChanges: append([]string(nil), links.RelatedChanges...),
			DependsOn:      append([]string(nil), links.DependsOn...),
		}
	}
	return cloned
}

func buildDecompositionStatusWrites(v *Validator, index *ProjectIndex, writableRoot string, graph map[string]ChangeGraphLinks) ([]fileRewrite, []FileMutation, error) {
	writes := make([]fileRewrite, 0, len(graph))
	changedFiles := make([]FileMutation, 0, len(graph))
	for _, changeID := range SortedKeys(graph) {
		record := index.Changes[changeID]
		if record == nil {
			return nil, nil, fmt.Errorf("change %q does not exist", changeID)
		}
		links := graph[changeID]
		status := cloneMap(index.StatusFiles[record.StatusPath].Data)
		status["related_changes"] = stringSliceToAny(uniqueSortedStrings(links.RelatedChanges))
		status["depends_on"] = stringSliceToAny(uniqueSortedStrings(links.DependsOn))
		write, _, err := buildPrimaryCloseStatusWrite(v, writableRoot, record, status)
		if err != nil {
			return nil, nil, err
		}
		writes = append(writes, write)
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, record.StatusPath), Action: "updated"})
	}
	sortFileMutations(changedFiles)
	return writes, changedFiles, nil
}
