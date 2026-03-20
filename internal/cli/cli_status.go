package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStatus(args []string, stdout, stderr io.Writer) int {
	request, err := parseStatusArgs(args)
	if err != nil {
		writeCommandUsageError(stderr, "status", statusUsage, err)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "status")
	if code != exitOK {
		return code
	}
	defer project.close()
	summary, err := contracts.BuildProjectStatusSummary(project.validator, project.loaded)
	if err != nil {
		writeCommandInvalid(stderr, "status", project.absRoot, err)
		return exitInvalid
	}
	writeLines(stdout, buildStatusOutput(project.absRoot, summary)...)
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
