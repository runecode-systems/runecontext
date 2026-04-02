package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

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

func runChangeDecompositionPlan(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if machine.dryRun {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_decomposition_plan", changeDecompPlanUsage, fmt.Errorf("--dry-run is not supported for advisory change decomposition-plan")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseChangeDecompositionArgs(args, "change decomposition-plan")
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_decomposition_plan", changeDecompPlanUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_decomposition_plan", machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := contracts.PlanChangeDecomposition(project.validator, project.loaded, contracts.ChangeDecompositionPlanOptions{
		UmbrellaID: request.umbrellaID,
		SubChanges: request.subChanges,
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_decomposition_plan", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeDecompositionPlanOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeDecompositionPlanExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runChangeDecompositionApply(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseChangeDecompositionArgs(args, "change decomposition-apply")
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("change_decomposition_apply", changeDecompApplyUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "change_decomposition_apply", machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeDecompositionApplyResult, error) {
		return contracts.ApplyChangeDecomposition(v, loaded, contracts.ChangeDecompositionApplyOptions{
			UmbrellaID: request.umbrellaID,
			SubChanges: request.subChanges,
		})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("change_decomposition_apply", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildChangeDecompositionApplyOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendChangeDecompositionApplyExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}
