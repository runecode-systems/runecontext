package cli

import (
	"fmt"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func buildAdapterSyncOutput(state adapterSyncState, dryRun bool) []line {
	managedRel, _ := filepath.Rel(state.absRoot, state.managedRoot)
	manifestRel, _ := filepath.Rel(state.absRoot, state.manifestPath)
	output := []line{
		{"result", "ok"},
		{"command", adapterSyncCommand},
		{"root", state.absRoot},
		{"adapter", state.tool},
		{"managed_root", filepath.ToSlash(managedRel)},
		{"manifest_path", filepath.ToSlash(manifestRel)},
		{"managed_file_count", fmt.Sprintf("%d", len(state.managedFiles))},
		{"network_access", "false"},
		{"mutation_performed", boolString(!dryRun)},
	}
	output = append(output, line{"plan_action_count", fmt.Sprintf("%d", len(state.plan))})
	output = appendAdapterPlanActions(output, state.plan)
	return appendChangedFiles(output, state.plan)
}

func appendAdapterPlanActions(lines []line, mutations []contracts.FileMutation) []line {
	for i, mutation := range mutations {
		lines = append(lines, line{fmt.Sprintf("plan_action_%d", i+1), mutation.Action + " " + mutation.Path})
	}
	return lines
}

func appendAdapterSyncExplainLines(lines []line) []line {
	return append(lines,
		line{"explain_scope", "adapter-sync"},
		line{"explain_local_only", "true"},
		line{"explain_network_access", "adapter sync uses installed release contents and never fetches from network"},
		line{"explain_managed_boundary", "writes are limited to .runecontext/adapters/<tool>/managed and sync-manifest.yaml"},
		line{"explain_manifest_role", "sync manifest is convenience metadata and not correctness-critical state"},
	)
}
