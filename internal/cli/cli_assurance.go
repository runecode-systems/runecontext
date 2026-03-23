package cli

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
)

const (
	assuranceCommandUsage  = "runectx assurance <enable|backfill>"
	assuranceEnableUsage   = "runectx assurance enable verified [--path PATH] [path]"
	assuranceBackfillUsage = "runectx assurance backfill [--path PATH] [path]"
)

var assuranceTierRegex = regexp.MustCompile(`(?m)^(\s*assurance_tier:\s*).*$`)

type assuranceEnableRequest struct {
	root         string
	explicitRoot bool
}

type assuranceBackfillRequest struct {
	root         string
	explicitRoot bool
}

func runAssurance(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		writeCommandUsageError(stderr, "assurance", assuranceCommandUsage, fmt.Errorf("missing subcommand"))
		return exitUsage
	}
	switch args[0] {
	case "enable":
		return runAssuranceEnable(args[1:], stdout, stderr)
	case "backfill":
		return runAssuranceBackfill(args[1:], stdout, stderr)
	default:
		writeCommandUsageError(stderr, "assurance", assuranceCommandUsage, fmt.Errorf("unknown subcommand %q", args[0]))
		return exitUsage
	}
}

func runAssuranceEnable(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance enable", assuranceEnableUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseAssuranceEnableArgs(remaining)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance enable", assuranceEnableUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	plans := []string{
		fmt.Sprintf("update %s", filepath.Join(request.root, "runecontext.yaml")),
		fmt.Sprintf("write %s", filepath.Join(request.root, "assurance", "baseline.yaml")),
	}
	if machine.dryRun {
		return emitAssuranceEnableDryRun(stdout, stderr, machine, request.root, plans)
	}
	return executeAssuranceEnable(stdout, stderr, machine, request.root)
}

func emitAssuranceEnableDryRun(stdout, stderr io.Writer, machine machineOptions, root string, plans []string) int {
	output := []line{{"result", "ok"}, {"command", "assurance enable"}, {"root", root}, {"mode", "verified"}}
	output = appendStringItems(output, "plan_action", plans)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintln(stderr, "Dry run: would enable Verified mode and write baseline")
		for _, action := range plans {
			fmt.Fprintf(stderr, "  - %s\n", action)
		}
	}
	return exitOK
}

func executeAssuranceEnable(stdout, stderr io.Writer, machine machineOptions, root string) int {
	ctx, err := newAssuranceEnableContext(root)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance enable", root, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	result, err := finalizeAssuranceEnable(ctx)
	if err != nil {
		emitAssuranceEnableError(stderr, machine, root, err)
		return exitInvalid
	}
	output := []line{{"result", "ok"}, {"command", "assurance enable"}, {"root", root}, {"baseline_path", result.baselinePath}}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Enabled Verified mode and wrote baseline at %s\n", result.baselinePath)
	}
	return exitOK
}

func runAssuranceBackfill(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseAssuranceBackfillArgs(remaining)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	msg := "assurance backfill is not implemented yet"
	lines := []line{
		{"result", "not_implemented"},
		{"command", "assurance backfill"},
		{"root", request.root},
		{"error_message", msg},
	}
	emitOutput(stderr, machine, appendMachineOptionLines(lines, machine), exitUsage, failureClassUsage)
	if !machine.jsonOutput {
		fmt.Fprintln(stderr, msg)
	}
	return exitUsage
}

func parseAssuranceEnableArgs(args []string) (assuranceEnableRequest, error) {
	request := assuranceEnableRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown assurance enable flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return assuranceEnableRequest{}, err
	}
	return applyAssuranceEnablePositionals(request, positionals)
}

func applyAssuranceEnablePositionals(request assuranceEnableRequest, positionals []string) (assuranceEnableRequest, error) {
	if len(positionals) == 0 {
		return assuranceEnableRequest{}, fmt.Errorf("assurance enable requires the positional \"verified\"")
	}
	if positionals[0] != "verified" {
		return assuranceEnableRequest{}, fmt.Errorf("expected \"verified\", got %q", positionals[0])
	}
	if len(positionals) == 1 {
		return request, nil
	}
	if len(positionals) > 2 {
		return assuranceEnableRequest{}, fmt.Errorf("assurance enable accepts at most one positional path")
	}
	if request.explicitRoot {
		return assuranceEnableRequest{}, fmt.Errorf("cannot specify both --path and positional path")
	}
	request.root = positionals[1]
	request.explicitRoot = true
	return request, nil
}

func parseAssuranceBackfillArgs(args []string) (assuranceBackfillRequest, error) {
	request := assuranceBackfillRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown assurance backfill flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return assuranceBackfillRequest{}, err
	}
	if len(positionals) > 1 {
		return assuranceBackfillRequest{}, fmt.Errorf("assurance backfill accepts at most one positional path")
	}
	if len(positionals) == 1 {
		if request.explicitRoot {
			return assuranceBackfillRequest{}, fmt.Errorf("cannot specify both --path and positional path")
		}
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func ensureAssuranceTierConfig(data []byte) ([]byte, bool) {
	replaced := false
	replacer := func(match []byte) []byte {
		replaced = true
		submatches := assuranceTierRegex.FindSubmatch(match)
		if len(submatches) >= 2 {
			return append(submatches[1], []byte("verified")...)
		}
		return []byte("assurance_tier: verified")
	}
	updated := assuranceTierRegex.ReplaceAllFunc(data, replacer)
	if replaced {
		return updated, true
	}
	buffer := bytes.NewBuffer(updated)
	if len(updated) > 0 && updated[len(updated)-1] != '\n' {
		buffer.WriteByte('\n')
	}
	buffer.WriteString("assurance_tier: verified\n")
	return buffer.Bytes(), false
}
