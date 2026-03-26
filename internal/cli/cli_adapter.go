package cli

import (
	"fmt"
	"io"
)

const adapterSyncCommand = "adapter_sync"

func runAdapter(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("adapter", adapterUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("adapter", adapterUsage, fmt.Errorf("adapter subcommand is required")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("adapter", adapterUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "adapter"}, {"usage", adapterUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}

	switch remaining[0] {
	case "sync":
		return runAdapterSync(remaining[1:], machine, stdout, stderr)
	case "render-host-native":
		return runAdapterRenderHostNative(remaining[1:], machine, stdout, stderr)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("adapter", adapterUsage, fmt.Errorf("unknown adapter subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runAdapterSync(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if len(args) == 1 && isHelpToken(args[0]) {
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", adapterSyncCommand}, {"usage", adapterSyncUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	if len(args) > 1 && isHelpToken(args[0]) {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(adapterSyncCommand, adapterSyncUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}

	request, err := parseAdapterSyncArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(adapterSyncCommand, adapterSyncUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}

	state, err := buildAdapterSyncState(request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(adapterSyncCommand, absRootOrFallback(request.root, ""), err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}

	if !machine.dryRun {
		if err := applyAdapterSync(state); err != nil {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(adapterSyncCommand, state.absRoot, err), machine), exitInvalid, failureClassInvalid)
			return exitInvalid
		}
	}

	output := buildAdapterSyncOutput(state, machine.dryRun)
	if machine.explain {
		output = appendAdapterSyncExplainLines(output, state.tool)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}
