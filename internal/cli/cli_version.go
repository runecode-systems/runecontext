package cli

import (
	"fmt"
	"io"
	"strings"
)

func runVersion(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("version", versionUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) > 0 && isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("version", versionUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "version"}, {"usage", versionUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	if len(remaining) > 0 {
		err := fmt.Errorf("version does not accept positional arguments")
		if strings.HasPrefix(remaining[0], "-") {
			err = fmt.Errorf("unknown version flag %q", remaining[0])
		}
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("version", versionUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	version := normalizedRunecontextVersion()
	emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "version"}, {"version", version}, {"runecontext_version", version}}, machine), exitOK, failureClassNone)
	return exitOK
}
