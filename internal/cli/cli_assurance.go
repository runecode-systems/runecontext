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
	assuranceBackfillUsage = "runectx assurance backfill [--path PATH] [path]"
	assuranceCaptureUsage  = "runectx assurance capture context-pack [--json] [--non-interactive] [--dry-run] [--explain] [--path PATH] <bundle-id>..."
)

type assuranceEnableRequest struct {
	root         string
	explicitRoot bool
}

type assuranceBackfillRequest struct {
	root         string
	explicitRoot bool
}

func runAssurance(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowDryRun: true, allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance", assuranceCommandUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
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

func runAssuranceBackfill(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	return runAssuranceBackfillWithMachine(remaining, stdout, stderr, machine)
}

func runAssuranceBackfillWithMachine(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	if machine.dryRun {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, fmt.Errorf("--dry-run is not supported for assurance backfill")), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, err := parseAssuranceBackfillArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "assurance backfill", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	resolvedRoot := projectRootForAssurance(project)
	result, err := executeAssuranceBackfill(resolvedRoot)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance backfill", resolvedRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	return emitAssuranceBackfillSuccess(stdout, stderr, machine, resolvedRoot, result)
}

func emitAssuranceBackfillSuccess(stdout, stderr io.Writer, machine machineOptions, root string, result assuranceBackfillResult) int {
	output := []line{
		{"result", "ok"},
		{"command", "assurance backfill"},
		{"root", root},
		{"baseline_path", result.baselinePath},
		{"history_path", result.historyPath},
		{"adoption_commit", result.adoptionCommit},
		{"history_commit_count", fmt.Sprintf("%d", result.commitCount)},
	}
	if result.importedAdded {
		output = append(output, line{"imported_evidence_added", "true"})
	} else {
		output = append(output, line{"imported_evidence_added", "false"})
	}
	if machine.explain {
		output = append(output,
			line{"explain_scope", "assurance-backfill"},
			line{"explain_provenance_class", "imported_git_history"},
			line{"explain_receipts_mutation", "none"},
		)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintf(stderr, "Backfilled imported git history at %s and updated %s\n", result.historyPath, result.baselinePath)
	}
	return exitOK
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
