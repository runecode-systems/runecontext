package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runChange(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change", changeUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change", changeUsage, fmt.Errorf("change subcommand is required")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change", changeUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "change"}, {"usage", changeUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}

	switch remaining[0] {
	case "new":
		return runChangeNew(remaining[1:], machine, stdout, stderr)
	case "shape":
		return runChangeShape(remaining[1:], machine, stdout, stderr)
	case "close":
		return runChangeClose(remaining[1:], machine, stdout, stderr)
	case "reallocate":
		return runChangeReallocate(remaining[1:], machine, stdout, stderr)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change", changeUsage, fmt.Errorf("unknown change subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runChangeNew(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeNewArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_new", changeNewUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if err := enforceNonInteractiveChangeNew(machine, request); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_new", changeNewUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_new", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeOperationResult, error) {
		return contracts.CreateChange(v, loaded, contracts.ChangeCreateOptions{
			Title:          request.title,
			Type:           request.changeType,
			Size:           request.size,
			Description:    request.description,
			ContextBundles: request.contextBundles,
			RequestedMode:  contracts.ChangeMode(request.mode),
		})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_new", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeNewOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeNewExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeShape(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeShapeArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_shape", changeShapeUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_shape", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeOperationResult, error) {
		return contracts.ShapeChange(v, loaded, request.changeID, contracts.ChangeShapeOptions{
			Design:       request.design,
			Verification: request.verification,
			Tasks:        request.tasks,
			References:   request.references,
		})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_shape", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeShapeOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeShapeExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeClose(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeCloseArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_close", changeCloseUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_close", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeOperationResult, error) {
		return contracts.CloseChange(v, loaded, request.changeID, contracts.ChangeCloseOptions{
			VerificationStatus: request.verificationStatus,
			ClosedAt:           request.closedAt,
			SupersededBy:       request.supersededBy,
		})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_close", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeCloseOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeCloseExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeReallocate(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeReallocateArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_reallocate", changeReallocateUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_reallocate", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeReallocationResult, error) {
		return contracts.ReallocateChange(v, loaded, request.changeID, contracts.ChangeReallocateOptions{})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_reallocate", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeReallocateOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeReallocateExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
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
