package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	assuranceCommandUsage  = "runectx assurance <enable|backfill|capture>"
	assuranceEnableUsage   = "runectx assurance enable verified [--path PATH] [path]"
	assuranceBackfillUsage = "runectx assurance backfill [--json] [--non-interactive] [--dry-run] [--explain] [--path PATH] [path]"
	assuranceCaptureUsage  = "runectx assurance capture context-pack [--json] [--non-interactive] [--dry-run] [--explain] [--path PATH] <bundle-id>..."
)

type assuranceEnableRequest struct {
	root         string
	explicitRoot bool
}

func runAssurance(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance", assuranceCommandUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if len(remaining) > 0 && isHelpToken(remaining[0]) {
		if len(remaining) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance", assuranceCommandUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return exitUsage
		}
		emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "assurance"}, {"usage", assuranceCommandUsage}}, machine), exitOK, failureClassNone)
		return exitOK
	}
	if len(remaining) == 0 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance", assuranceCommandUsage, fmt.Errorf("missing subcommand")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	switch remaining[0] {
	case "enable":
		return runAssuranceEnableWithMachine(remaining[1:], stdout, stderr, machine)
	case "backfill":
		return runAssuranceBackfillWithMachine(remaining[1:], stdout, stderr, machine)
	case "capture":
		return runAssuranceCaptureWithMachine(remaining[1:], stdout, stderr, machine)
	default:
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance", assuranceCommandUsage, fmt.Errorf("unknown subcommand %q", remaining[0])), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
}

func runAssuranceEnable(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance enable", assuranceEnableUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	return runAssuranceEnableWithMachine(remaining, stdout, stderr, machine)
}

func runAssuranceEnableWithMachine(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	request, err := parseAssuranceEnableArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance enable", assuranceEnableUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "assurance enable", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	resolvedRoot := projectRootForAssurance(project)
	plans := []string{
		fmt.Sprintf("update %s", filepath.Join(resolvedRoot, "runecontext.yaml")),
		fmt.Sprintf("write %s", filepath.Join(resolvedRoot, "assurance", "baseline.yaml")),
	}
	if machine.dryRun {
		return emitAssuranceEnableDryRun(stdout, stderr, machine, resolvedRoot, plans)
	}
	return executeAssuranceEnable(stdout, stderr, machine, resolvedRoot, project.loaded)
}

func projectRootForAssurance(project *cliProject) string {
	if project != nil && project.loaded != nil && project.loaded.Resolution != nil && project.loaded.Resolution.ProjectRoot != "" {
		return project.loaded.Resolution.ProjectRoot
	}
	if project != nil {
		return project.absRoot
	}
	return "."
}

func emitAssuranceEnableDryRun(stdout, stderr io.Writer, machine machineOptions, root string, plans []string) int {
	output := []line{{"result", "ok"}, {"command", "assurance enable"}, {"root", root}, {"mode", "verified"}}
	output = appendStringItems(output, "plan_action", plans)
	if machine.explain {
		output = appendAssuranceEnableExplainLines(output, root, plans)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintln(stderr, "Dry run: would enable Verified mode and write baseline")
		for _, action := range plans {
			fmt.Fprintf(stderr, "  - %s\n", action)
		}
	}
	return exitOK
}

func executeAssuranceEnable(stdout, stderr io.Writer, machine machineOptions, root string, loaded *contracts.LoadedProject) int {
	ctx, err := newAssuranceEnableContext(root, loaded)
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
	if machine.explain {
		plans := []string{
			fmt.Sprintf("update %s", filepath.Join(root, "runecontext.yaml")),
			fmt.Sprintf("write %s", filepath.Join(root, "assurance", "baseline.yaml")),
		}
		output = appendAssuranceEnableExplainLines(output, root, plans)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Enabled Verified mode and wrote baseline at %s\n", result.baselinePath)
	}
	return exitOK
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
