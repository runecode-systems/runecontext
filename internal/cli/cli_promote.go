package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runPromote(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("promote", promoteUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parsePromoteArgs(remaining)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("promote", promoteUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "promote", machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.ChangeOperationResult, error) {
		return contracts.PromoteChange(v, loaded, request.changeID, contracts.PromoteOptions{
			Accept:   request.accept,
			Complete: request.complete,
			Targets:  append([]string(nil), request.targets...),
		})
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("promote", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildPromoteOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendPromoteExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func buildPromoteOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.ChangeOperationResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", "promote"},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"change_id", result.ID},
		{"change_path", result.ChangePath},
		{"promotion_status", result.PromotionAssessmentStatus},
		{"target_count", fmt.Sprintf("%d", len(result.SuggestedPromotionTargets))},
	}
	output = appendTargets(output, result.SuggestedPromotionTargets)
	return appendChangedFiles(output, result.ChangedFiles)
}

func appendTargets(lines []line, targets []string) []line {
	for i, target := range targets {
		lines = append(lines, line{fmt.Sprintf("target_%d", i+1), target})
	}
	return lines
}
