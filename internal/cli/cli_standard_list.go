package cli

import (
	"fmt"
	"io"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStandardList(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseStandardListArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(standardListCommand, standardListUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, standardListCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := contracts.ListStandards(project.validator, project.loaded, contracts.StandardListOptions{
		ScopePaths: request.scopePaths,
		Focus:      request.focus,
		Statuses:   parseStandardStatuses(request.statuses),
	})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(standardListCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildStandardListOutput(project.absRoot, project.loaded, result)
	if machine.explain {
		output = appendStandardListExplainLines(output, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func parseStandardListArgs(args []string) (standardListRequest, error) {
	request := standardListRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--scope-path":
			return appendStringFlag(args, flag, &request.scopePaths)
		case "--focus":
			return assignStringFlag(args, flag, &request.focus)
		case "--status":
			return appendStringFlag(args, flag, &request.statuses)
		default:
			return flag.next, fmt.Errorf("unknown standard list flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return standardListRequest{}, err
	}
	if len(positionals) > 0 {
		return standardListRequest{}, fmt.Errorf("standard list does not accept positional arguments")
	}
	return request, nil
}

func buildStandardListOutput(absRoot string, loaded *contracts.LoadedProject, result *contracts.StandardListResult) []line {
	output := []line{
		{"result", "ok"},
		{"command", standardListCommand},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
		{"mutation_performed", "false"},
		{"standard_scope_count", fmt.Sprintf("%d", len(result.ScopePaths))},
		{"standard_focus", result.Focus},
		{"standard_status_filter_count", fmt.Sprintf("%d", len(result.Statuses))},
		{"standard_count", fmt.Sprintf("%d", len(result.Standards))},
	}
	for i, scope := range result.ScopePaths {
		output = append(output, line{fmt.Sprintf("standard_scope_%d", i+1), scope})
	}
	for i, status := range result.Statuses {
		output = append(output, line{fmt.Sprintf("standard_status_filter_%d", i+1), string(status)})
	}
	for i, standard := range result.Standards {
		prefix := fmt.Sprintf("standard_%d", i+1)
		output = append(output,
			line{prefix + "_path", standard.Path},
			line{prefix + "_id", standard.ID},
			line{prefix + "_title", standard.Title},
			line{prefix + "_status", string(standard.Status)},
			line{prefix + "_replaced_by", standard.ReplacedBy},
			line{prefix + "_alias_count", fmt.Sprintf("%d", len(standard.Aliases))},
			line{prefix + "_suggested_context_bundle_count", fmt.Sprintf("%d", len(standard.SuggestedContextBundles))},
		)
		for j, alias := range standard.Aliases {
			output = append(output, line{fmt.Sprintf("%s_alias_%d", prefix, j+1), alias})
		}
		for j, bundle := range standard.SuggestedContextBundles {
			output = append(output, line{fmt.Sprintf("%s_suggested_context_bundle_%d", prefix, j+1), bundle})
		}
	}
	return output
}

func appendStandardListExplainLines(lines []line, result *contracts.StandardListResult) []line {
	if result == nil {
		return lines
	}
	return append(lines,
		line{"explain_scope", "standards-list"},
		line{"explain_advisory_only", "true"},
		line{"explain_standard_count_reason", fmt.Sprintf("%d standards matched scope, focus, and lifecycle filters", len(result.Standards))},
	)
}
