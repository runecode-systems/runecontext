package cli

import (
	"fmt"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func buildChangeNewOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeOperationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_new"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"recommended_mode", string(result.RecommendedMode)},
		{"change_status", result.Status},
		{"context_bundle_count", fmt.Sprintf("%d", len(result.ContextBundles))},
	}
	output = appendStringItems(output, "context_bundle", result.ContextBundles)
	output = append(output, line{"applicable_standard_count", fmt.Sprintf("%d", len(result.ApplicableStandards))})
	output = appendStringItems(output, "applicable_standard", result.ApplicableStandards)
	output = append(output, line{"standards_refresh", result.StandardsRefreshAction}, line{"review_diff_required", fmt.Sprintf("%t", result.ReviewDiffRequired)})
	output = appendReasonsAndAssumptions(output, result.Reasons, result.Assumptions)
	return appendChangedFiles(output, result.ChangedFiles)
}

func buildChangeShapeOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeOperationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_shape"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"change_status", result.Status},
		{"applicable_standard_count", fmt.Sprintf("%d", len(result.ApplicableStandards))},
	}
	output = appendStringItems(output, "applicable_standard", result.ApplicableStandards)
	output = append(output, line{"added_standard_count", fmt.Sprintf("%d", len(result.AddedStandards))})
	output = appendStringItems(output, "added_standard", result.AddedStandards)
	output = append(output, line{"standards_refresh", result.StandardsRefreshAction}, line{"review_diff_required", fmt.Sprintf("%t", result.ReviewDiffRequired)})
	output = appendReasonsAndAssumptions(output, result.Reasons, result.Assumptions)
	return appendChangedFiles(output, result.ChangedFiles)
}

func buildChangeCloseOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeOperationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_close"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"change_status", result.Status},
	}
	if result.ClosedAt != "" {
		output = append(output, line{"closed_at", result.ClosedAt})
	}
	if result.Recursive {
		output = append(output, line{"recursive", "true"}, line{"recursive_target_count", fmt.Sprintf("%d", result.RecursiveTargetCount)})
		output = appendStringItems(output, "recursive_target", result.RecursiveTargetIDs)
	}
	return appendChangedFiles(output, result.ChangedFiles)
}

func buildChangeReallocateOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeReallocationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_reallocate"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"old_change_id", result.OldID},
		{"change_id", result.ID},
		{"old_change_path", result.OldChangePath},
		{"change_path", result.ChangePath},
		{"rewritten_reference_count", fmt.Sprintf("%d", result.RewrittenReferenceCount)},
	}
	output = appendWarnings(output, result.Warnings)
	return appendChangedFiles(output, result.ChangedFiles)
}

func buildChangeUpdateOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeOperationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_update"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"change_mode", string(result.Mode)},
		{"change_status", result.Status},
		{"related_change_count", fmt.Sprintf("%d", len(result.RelatedChanges))},
	}
	output = appendStringItems(output, "related_change", result.RelatedChanges)
	if result.Recursive {
		output = append(output, line{"recursive", "true"}, line{"recursive_target_count", fmt.Sprintf("%d", result.RecursiveTargetCount)})
		output = appendStringItems(output, "recursive_target", result.RecursiveTargetIDs)
	}
	return appendChangedFiles(output, result.ChangedFiles)
}

func buildChangeAssessIntakeOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeAssessIntakeResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_assess_intake"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"mutation_performed", "false"},
		{"change_type", result.Type},
		{"change_size", result.Size},
		{"recommended_mode", string(result.RecommendedMode)},
		{"intake_readiness", result.IntakeReadiness},
		{"clarification_needed", boolString(result.ClarificationNeeded)},
		{"decomposition_signal", result.DecompositionSignal},
		{"context_bundle_count", fmt.Sprintf("%d", len(result.ContextBundles))},
	}
	output = appendStringItems(output, "context_bundle", result.ContextBundles)
	output = append(output, line{"applicable_standard_count", fmt.Sprintf("%d", len(result.ApplicableStandards))})
	output = appendStringItems(output, "applicable_standard", result.ApplicableStandards)
	output = append(output, line{"clarification_prompt_count", fmt.Sprintf("%d", len(result.ClarificationPrompts))})
	output = appendStringItems(output, "clarification_prompt", result.ClarificationPrompts)
	return appendReasonsAndAssumptions(output, result.Reasons, result.Assumptions)
}

func buildChangeAssessDecompositionOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeAssessDecompositionResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_assess_decomposition"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"mutation_performed", "false"},
		{"change_id", result.ID},
		{"change_status", result.Status},
		{"change_type", result.Type},
		{"change_size", result.Size},
		{"recommended_mode", string(result.RecommendedMode)},
		{"decomposition_signal", result.DecompositionSignal},
		{"clarification_needed", boolString(result.ClarificationNeeded)},
		{"related_change_count", fmt.Sprintf("%d", len(result.RelatedChanges))},
	}
	output = appendStringItems(output, "related_change", result.RelatedChanges)
	output = append(output, line{"eligible_sub_change_count", fmt.Sprintf("%d", len(result.EligibleSubChangeIDs))})
	output = appendStringItems(output, "eligible_sub_change", result.EligibleSubChangeIDs)
	output = append(output, line{"prerequisite_change_count", fmt.Sprintf("%d", len(result.PrerequisiteChangeIDs))})
	output = appendStringItems(output, "prerequisite_change", result.PrerequisiteChangeIDs)
	output = append(output, line{"clarification_prompt_count", fmt.Sprintf("%d", len(result.ClarificationPrompts))})
	output = appendStringItems(output, "clarification_prompt", result.ClarificationPrompts)
	output = append(output, line{"reason_count", fmt.Sprintf("%d", len(result.Reasons))})
	return appendStringItems(output, "reason", result.Reasons)
}

func buildChangeDecompositionPlanOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeDecompositionPlanResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_decomposition_plan"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"mutation_performed", "false"},
		{"umbrella_change_id", result.UmbrellaID},
		{"graph_node_count", fmt.Sprintf("%d", len(result.NodeIDs))},
	}
	output = appendStringItems(output, "graph_node", result.NodeIDs)
	return appendDecompositionGraphLines(output, result.Graph)
}

func buildChangeDecompositionApplyOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeDecompositionApplyResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "change_decomposition_apply"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"umbrella_change_id", result.UmbrellaID},
		{"graph_node_count", fmt.Sprintf("%d", len(result.NodeIDs))},
	}
	output = appendStringItems(output, "graph_node", result.NodeIDs)
	output = appendDecompositionGraphLines(output, result.Graph)
	return appendChangedFiles(output, result.ChangedFiles)
}

func appendDecompositionGraphLines(output []line, graph map[string]contracts.ChangeGraphLinks) []line {
	for _, nodeID := range contracts.SortedKeys(graph) {
		links := graph[nodeID]
		output = append(output, line{"graph_" + nodeID + "_related_change_count", fmt.Sprintf("%d", len(links.RelatedChanges))})
		output = appendStringItems(output, "graph_"+nodeID+"_related_change", links.RelatedChanges)
		output = append(output, line{"graph_" + nodeID + "_depends_on_count", fmt.Sprintf("%d", len(links.DependsOn))})
		output = appendStringItems(output, "graph_"+nodeID+"_depends_on", links.DependsOn)
	}
	return output
}

func selectedConfigPath(loaded *contracts.LoadedProject) string {
	if loaded == nil || loaded.Resolution == nil {
		return ""
	}
	return loaded.Resolution.SelectedConfigPath
}
