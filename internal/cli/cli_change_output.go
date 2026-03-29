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
	}
	return appendChangedFiles(output, result.ChangedFiles)
}

func selectedConfigPath(loaded *contracts.LoadedProject) string {
	if loaded == nil || loaded.Resolution == nil {
		return ""
	}
	return loaded.Resolution.SelectedConfigPath
}
