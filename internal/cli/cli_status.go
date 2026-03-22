package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStatus(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("status", statusUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseStatusArgs(remaining)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("status", statusUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "status", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	summary, err := contracts.BuildProjectStatusSummary(project.validator, project.loaded)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("status", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildStatusOutput(project.absRoot, summary)
	if machine.explain {
		output = appendStatusExplainLines(output, project.loaded, summary)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
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
