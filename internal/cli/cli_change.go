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
	case "update":
		return runChangeUpdate(remaining[1:], machine, stdout, stderr)
	case "assess-intake":
		return runChangeAssessIntake(remaining[1:], machine, stdout, stderr)
	case "assess-decomposition":
		return runChangeAssessDecomposition(remaining[1:], machine, stdout, stderr)
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
			Recursive:          request.recursive,
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

func runChangeUpdate(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeUpdateArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_update", changeUpdateUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_update", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeOperationResult, error) {
		return contracts.UpdateChange(v, loaded, request.changeID, contracts.ChangeUpdateOptions{Status: request.status, VerificationStatus: request.verificationStatus, Recursive: request.recursive})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_update", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeUpdateOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeUpdateExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeAssessIntake(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if machine.dryRun {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_assess_intake", changeAssessIntakeUsage, fmt.Errorf("--dry-run is not supported for advisory change assess-intake")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseChangeAssessIntakeArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_assess_intake", changeAssessIntakeUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_assess_intake", machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := contracts.AssessChangeIntake(project.validator, project.loaded, contracts.ChangeAssessIntakeOptions{
		Title:          request.title,
		Type:           request.changeType,
		Size:           request.size,
		Description:    request.description,
		ContextBundles: request.contextBundles,
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_assess_intake", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeAssessIntakeOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeAssessIntakeExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeAssessDecomposition(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if machine.dryRun {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_assess_decomposition", changeAssessDecompUsage, fmt.Errorf("--dry-run is not supported for advisory change assess-decomposition")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseChangeAssessDecompositionArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_assess_decomposition", changeAssessDecompUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_assess_decomposition", machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := contracts.AssessChangeDecomposition(project.validator, project.loaded, request.changeID)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_assess_decomposition", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeAssessDecompositionOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeAssessDecompositionExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}
