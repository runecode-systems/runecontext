package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStatus(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		return emitStatusUsageError(stderr, machine, err)
	}
	request, err := parseStatusArgs(remaining)
	if err != nil {
		return emitStatusUsageError(stderr, machine, err)
	}
	if request.explicitRoot && isHelpToken(request.root) {
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "status"}, {"usage", statusUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "status", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	summary, err := contracts.BuildProjectStatusSummary(project.validator, project.loaded)
	if err != nil {
		return emitStatusInvalid(stderr, machine, project.absRoot, err)
	}
	if err := validateStatusMachineFlags(machine, request); err != nil {
		return emitStatusUsageError(stderr, machine, err)
	}
	if !machine.jsonOutput {
		rendered := renderHumanStatus(project.absRoot, project.loaded, summary, statusRenderOptionsForMachine(stdout, machine, request))
		_, _ = io.WriteString(stdout, rendered)
		return exitOK
	}
	output := buildStatusOutput(project.absRoot, summary)
	if machine.explain {
		output = appendStatusExplainLines(output, project.loaded, summary)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func emitStatusUsageError(w io.Writer, machine machineOptions, err error) int {
	emitOutput(w, machine, appendMachineOptionLines(buildCommandUsageErrorLines("status", statusUsage, err), machine), exitUsage, failureClassUsage)
	return exitUsage
}

func emitStatusInvalid(w io.Writer, machine machineOptions, root string, err error) int {
	emitOutput(w, machine, appendMachineOptionLines(buildCommandInvalidLines("status", root, err), machine), exitInvalid, failureClassInvalid)
	return exitInvalid
}

func validateStatusMachineFlags(machine machineOptions, request statusRequest) error {
	if !machine.jsonOutput {
		return nil
	}
	if !request.historyModeSet && !request.historyLimitSet && !request.verbose {
		return nil
	}
	return fmt.Errorf("--history, --history-limit, and --verbose are only supported for human status output")
}

func statusRenderOptionsForMachine(stdout io.Writer, machine machineOptions, request statusRequest) statusRenderOptions {
	return statusRenderOptions{
		color:        shouldUseStatusColor(stdout),
		explain:      machine.explain,
		historyMode:  request.historyMode,
		historyLimit: request.historyLimit,
		verbose:      request.verbose,
	}
}

func buildStatusOutput(absRoot string, summary *contracts.ProjectStatusSummary) []line {
	output := []line{
		{"result", "ok"},
		{"command", "status"},
		{"root", absRoot},
		{"selected_config_path", summary.SelectedConfigPath},
		{"runecontext_version", summary.RuneContextVersion},
		{"assurance_tier", summary.AssuranceTier},
		{"active_count", fmt.Sprintf("%d", len(summary.Active))},
	}
	output = appendStatusEntries(output, "active", summary.Active)
	output = append(output, line{"closed_count", fmt.Sprintf("%d", len(summary.Closed))})
	output = appendStatusEntries(output, "closed", summary.Closed)
	output = append(output, line{"superseded_count", fmt.Sprintf("%d", len(summary.Superseded))})
	output = appendStatusEntries(output, "superseded", summary.Superseded)
	output = append(output, line{"bundle_count", fmt.Sprintf("%d", len(summary.BundleIDs))})
	for i, bundleID := range summary.BundleIDs {
		output = append(output, line{fmt.Sprintf("bundle_%d", i+1), bundleID})
	}
	return output
}
