package cli

import (
	"fmt"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func buildAdapterSyncOutput(state adapterSyncState, dryRun bool) []line {
	hostNativeFlowAssetCount, hostNativeShimCount := countHostNativeKinds(state.hostNativeFiles)
	output := []line{
		{"result", "ok"},
		{"command", adapterSyncCommand},
		{"root", state.absRoot},
		{"adapter", state.tool},
		{"host_native_file_count", fmt.Sprintf("%d", len(state.hostNativeFiles))},
		{"host_native_flow_asset_count", fmt.Sprintf("%d", hostNativeFlowAssetCount)},
		{"host_native_discoverability_shim_count", fmt.Sprintf("%d", hostNativeShimCount)},
		{"network_access", "false"},
		{"mutation_performed", boolString(!dryRun)},
	}
	output = append(output, line{"plan_action_count", fmt.Sprintf("%d", len(state.plan))})
	output = appendAdapterPlanActions(output, state.plan)
	return appendChangedFiles(output, state.plan)
}

func countHostNativeKinds(artifacts []hostNativeArtifact) (int, int) {
	flowAssets := 0
	shims := 0
	for _, artifact := range artifacts {
		if artifact.shim {
			shims++
			continue
		}
		flowAssets++
	}
	return flowAssets, shims
}

func appendAdapterPlanActions(lines []line, mutations []contracts.FileMutation) []line {
	for i, mutation := range mutations {
		lines = append(lines, line{fmt.Sprintf("plan_action_%d", i+1), mutation.Action + " " + mutation.Path})
	}
	return lines
}

func appendAdapterSyncExplainLines(lines []line, tool string) []line {
	hostBoundary := "host-native artifacts are tool-specific and additive"
	switch tool {
	case "opencode":
		hostBoundary = "host-native writes are additive under .opencode/skills and .opencode/commands"
	case "claude-code":
		hostBoundary = "host-native writes are additive under .claude/skills with optional .claude/commands shim"
	case "codex":
		hostBoundary = "host-native writes are additive under .agents/skills"
	}
	return append(lines,
		line{"explain_scope", "adapter-sync"},
		line{"explain_local_only", "true"},
		line{"explain_network_access", "adapter sync uses installed release contents and never fetches from network"},
		line{"explain_managed_boundary", "writes are limited to tool-specific repo-local host-native artifact roots"},
		line{"explain_host_native_boundary", hostBoundary},
	)
}
