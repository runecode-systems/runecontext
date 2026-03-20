package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runChange(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeCommandUsageError(stderr, "change", changeUsage, fmt.Errorf("change subcommand is required"))
		return exitUsage
	}

	switch args[0] {
	case "new":
		return runChangeNew(args[1:], stdout, stderr)
	case "shape":
		return runChangeShape(args[1:], stdout, stderr)
	case "close":
		return runChangeClose(args[1:], stdout, stderr)
	case "reallocate":
		return runChangeReallocate(args[1:], stdout, stderr)
	default:
		writeCommandUsageError(stderr, "change", changeUsage, fmt.Errorf("unknown change subcommand %q", args[0]))
		return exitUsage
	}
}

func runChangeNew(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeNewArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "change_new", changeNewUsage, err)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_new")
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := contracts.CreateChange(project.validator, project.loaded, contracts.ChangeCreateOptions{
		Title:          request.title,
		Type:           request.changeType,
		Size:           request.size,
		Description:    request.description,
		ContextBundles: request.contextBundles,
		RequestedMode:  contracts.ChangeMode(request.mode),
	})
	if err != nil {
		writeCommandInvalid(stderr, "change_new", project.absRoot, err)
		return exitInvalid
	}
	writeLines(stdout, buildChangeNewOutput(project.absRoot, project.loaded, result)...)
	return exitOK
}

func runChangeShape(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeShapeArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "change_shape", changeShapeUsage, err)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_shape")
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := contracts.ShapeChange(project.validator, project.loaded, request.changeID, contracts.ChangeShapeOptions{
		Design:       request.design,
		Verification: request.verification,
		Tasks:        request.tasks,
		References:   request.references,
	})
	if err != nil {
		writeCommandInvalid(stderr, "change_shape", project.absRoot, err)
		return exitInvalid
	}
	writeLines(stdout, buildChangeShapeOutput(project.absRoot, project.loaded, result)...)
	return exitOK
}

func runChangeClose(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeCloseArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "change_close", changeCloseUsage, err)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_close")
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := contracts.CloseChange(project.validator, project.loaded, request.changeID, contracts.ChangeCloseOptions{
		VerificationStatus: request.verificationStatus,
		ClosedAt:           request.closedAt,
		SupersededBy:       request.supersededBy,
	})
	if err != nil {
		writeCommandInvalid(stderr, "change_close", project.absRoot, err)
		return exitInvalid
	}
	writeLines(stdout, buildChangeCloseOutput(project.absRoot, project.loaded, result)...)
	return exitOK
}

func runChangeReallocate(args []string, stdout, stderr io.Writer) int {
	request, err := parseChangeReallocateArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "change_reallocate", changeReallocateUsage, err)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_reallocate")
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := contracts.ReallocateChange(project.validator, project.loaded, request.changeID, contracts.ChangeReallocateOptions{})
	if err != nil {
		writeCommandInvalid(stderr, "change_reallocate", project.absRoot, err)
		return exitInvalid
	}
	writeLines(stdout, buildChangeReallocateOutput(project.absRoot, project.loaded, result)...)
	return exitOK
}

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

func selectedConfigPath(loaded *contracts.LoadedProject) string {
	if loaded == nil || loaded.Resolution == nil {
		return ""
	}
	return loaded.Resolution.SelectedConfigPath
}
