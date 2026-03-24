package cli

import (
	"fmt"
	"io"
	"path/filepath"
)

type assuranceBackfillRequest struct {
	root         string
	explicitRoot bool
}

type assuranceBackfillContext struct {
	baselinePath string
}

type assuranceBackfillResult struct {
	baselinePath   string
	historyPath    string
	adoptionCommit string
	commitCount    int
	importedAdded  bool
}

func runAssuranceBackfillWithMachine(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	if handled, code := maybeHandleAssuranceBackfillHelp(args, stdout, stderr, machine); handled {
		return code
	}
	if machine.dryRun {
		return runAssuranceBackfillDryRun(args, stdout, stderr, machine)
	}
	return runAssuranceBackfillExecute(args, stdout, stderr, machine)
}

func maybeHandleAssuranceBackfillHelp(args []string, stdout, stderr io.Writer, machine machineOptions) (bool, int) {
	if len(args) == 0 || !isHelpToken(args[0]) {
		return false, exitOK
	}
	if len(args) != 1 {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
		return true, exitUsage
	}
	emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "assurance backfill"}, {"usage", assuranceBackfillUsage}}, machine), exitOK, failureClassNone)
	return true, exitOK
}

func runAssuranceBackfillDryRun(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	project, resolvedRoot, code := loadAssuranceBackfillProject(args, stderr, machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	plan, err := buildAssuranceBackfillDryRunPlan(resolvedRoot)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance backfill", resolvedRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := []line{{"result", "ok"}, {"command", "assurance backfill"}, {"root", resolvedRoot}, {"mode", "imported-git-history"}}
	output = appendStringItems(output, "plan_action", plan)
	if machine.explain {
		output = append(output,
			line{"explain_scope", "assurance-backfill"},
			line{"explain_provenance_class", "imported_git_history"},
			line{"explain_receipts_mutation", "none"},
		)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	if !machine.jsonOutput {
		fmt.Fprintln(stderr, "Dry run: would run assurance backfill validation and planned updates")
		for _, action := range plan {
			fmt.Fprintf(stderr, "  - %s\n", action)
		}
	}
	return exitOK
}

func runAssuranceBackfillExecute(args []string, stdout, stderr io.Writer, machine machineOptions) int {
	project, resolvedRoot, code := loadAssuranceBackfillProject(args, stderr, machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	result, err := executeAssuranceBackfill(resolvedRoot)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("assurance backfill", resolvedRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	return emitAssuranceBackfillSuccess(stdout, stderr, machine, resolvedRoot, result)
}

func loadAssuranceBackfillProject(args []string, stderr io.Writer, machine machineOptions) (*cliProject, string, int) {
	request, err := parseAssuranceBackfillArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("assurance backfill", assuranceBackfillUsage, err), machine), exitUsage, failureClassUsage)
		return nil, "", exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "assurance backfill", machine)
	if code != exitOK {
		return nil, "", code
	}
	resolvedRoot := projectRootForAssurance(project)
	return project, resolvedRoot, exitOK
}

func buildAssuranceBackfillDryRunPlan(root string) ([]string, error) {
	context, baselineMap, adoptionCommit, err := loadBackfillInputs(root)
	if err != nil {
		return nil, err
	}
	if skip, existing := shouldSkipRebuild(root, baselineMap, adoptionCommit); skip {
		return []string{
			fmt.Sprintf("reuse %s", existing),
			fmt.Sprintf("leave %s unchanged", context.baselinePath),
		}, nil
	}
	if _, err := buildImportedGitHistory(root, adoptionCommit); err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("write %s", filepath.Join(root, "assurance", "backfill", fmt.Sprintf("imported-git-history-%s.json", adoptionCommit))),
		fmt.Sprintf("update %s", context.baselinePath),
	}, nil
}

func executeAssuranceBackfill(root string) (assuranceBackfillResult, error) {
	context, baselineMap, adoptionCommit, err := loadBackfillInputs(root)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	if skip, existing := shouldSkipRebuild(root, baselineMap, adoptionCommit); skip {
		return assuranceBackfillResult{
			baselinePath:   context.baselinePath,
			historyPath:    existing,
			adoptionCommit: adoptionCommit,
			commitCount:    0,
			importedAdded:  false,
		}, nil
	}
	commits, err := buildImportedGitHistory(root, adoptionCommit)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	historyPath, err := writeImportedGitHistory(root, adoptionCommit, commits)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	importedAdded, err := appendImportedEvidence(context.baselinePath, baselineMap, root, historyPath)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	return assuranceBackfillResult{
		baselinePath:   context.baselinePath,
		historyPath:    historyPath,
		adoptionCommit: adoptionCommit,
		commitCount:    len(commits),
		importedAdded:  importedAdded,
	}, nil
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
		if result.importedAdded || result.commitCount > 0 {
			fmt.Fprintf(stderr, "Backfilled imported git history at %s and updated %s\n", result.historyPath, result.baselinePath)
		} else {
			fmt.Fprintf(stderr, "Backfill is up to date at %s; no baseline changes were needed\n", result.baselinePath)
		}
	}
	return exitOK
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
