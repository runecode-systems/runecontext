package cli

import (
	"fmt"
	"io"
)

func emitInitSuccess(stdout, stderr io.Writer, machine machineOptions, state initState) int {
	nextActions := []string{
		fmt.Sprintf("runectx validate --path %s", state.absRoot),
		fmt.Sprintf("runectx adapter sync --path %s opencode", state.absRoot),
		"runectx completion bash|zsh|fish",
	}
	output := initSuccessOutput(state, nextActions)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	emitInitSuccessHuman(stderr, machine, state.absRoot, nextActions)
	return exitOK
}

func initSuccessOutput(state initState, nextActions []string) []line {
	output := []line{
		{"result", "ok"},
		{"command", "init"},
		{"root", state.absRoot},
		{"config_path", state.configPath},
		{"bundles_dir", state.bundlesDir},
		{"changes_dir", state.changesDir},
		{"network_access", "false"},
	}
	if state.effectiveMode != "embedded" || state.modeExplicit {
		output = append(output, line{"mode", state.effectiveMode})
	}
	if state.bundlePath != "" {
		output = append(output, line{"seed_bundle_path", state.bundlePath})
	}
	return appendStringItems(output, "next_action", nextActions)
}

func emitInitSuccessHuman(stderr io.Writer, machine machineOptions, absRoot string, nextActions []string) {
	if machine.jsonOutput {
		return
	}
	fmt.Fprintf(stderr, "Initialized RuneContext at %s\n", absRoot)
	for _, action := range nextActions {
		fmt.Fprintf(stderr, "  - Next: %s\n", action)
	}
}

func emitInitUsage(stderr io.Writer, machine machineOptions, err error) int {
	emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("init", initUsage, err), machine), exitUsage, failureClassUsage)
	return exitUsage
}
