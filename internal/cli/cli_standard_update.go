package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStandardUpdate(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseStandardUpdateArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(standardUpdateCommand, standardUpdateUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, standardUpdateCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := runStandardUpdateMutation(project, machine, request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(standardUpdateCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildStandardMutationOutput(project.absRoot, project.loaded, standardUpdateCommand, result)
	if machine.explain {
		output = appendStandardMutationExplainLines(output, standardUpdateCommand, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runStandardUpdateMutation(project *cliProject, machine machineOptions, request standardUpdateRequest) (*contracts.StandardMutationResult, error) {
	return runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.StandardMutationResult, error) {
		return contracts.UpdateStandard(v, loaded, contracts.StandardUpdateOptions{
			Path:                           request.path,
			Title:                          request.title,
			Status:                         request.status,
			ReplacedBy:                     request.replacedBy,
			ClearReplacedBy:                request.clearReplacedBy,
			Aliases:                        request.aliases,
			ReplaceAliases:                 request.replaceAliases,
			SuggestedContextBundles:        request.suggestedContextBundles,
			ReplaceSuggestedContextBundles: request.replaceSuggestedContextBundles,
		})
	})
}

func parseStandardUpdateArgs(args []string) (standardUpdateRequest, error) {
	request := standardUpdateRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return parseStandardUpdateFlag(args, flag, &request)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return standardUpdateRequest{}, err
	}
	if len(positionals) > 0 {
		return standardUpdateRequest{}, fmt.Errorf("standard update does not accept positional arguments")
	}
	if err := validateStandardUpdateRequest(request); err != nil {
		return standardUpdateRequest{}, err
	}
	return request, nil
}

func parseStandardUpdateFlag(args []string, flag parsedFlag, request *standardUpdateRequest) (int, error) {
	switch flag.name {
	case "--path":
		return assignStringFlag(args, flag, &request.path)
	case "--title":
		return assignStringFlag(args, flag, &request.title)
	case "--status":
		return assignStringFlag(args, flag, &request.status)
	case "--replaced-by":
		return assignStringFlag(args, flag, &request.replacedBy)
	case "--clear-replaced-by":
		return assignNoValueBoolFlag(flag, &request.clearReplacedBy)
	case "--replace-aliases":
		return assignNoValueBoolFlag(flag, &request.replaceAliases)
	case "--alias":
		return appendStringFlag(args, flag, &request.aliases)
	case "--replace-suggested-context-bundles":
		return assignNoValueBoolFlag(flag, &request.replaceSuggestedContextBundles)
	case "--suggested-context-bundle":
		return appendStringFlag(args, flag, &request.suggestedContextBundles)
	case "--project-path":
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	default:
		return flag.next, fmt.Errorf("unknown standard update flag %q", flag.raw)
	}
}

func validateStandardUpdateRequest(request standardUpdateRequest) error {
	if strings.TrimSpace(request.path) == "" {
		return fmt.Errorf("standard update requires --path")
	}
	if request.replacedBy != "" && request.clearReplacedBy {
		return fmt.Errorf("--replaced-by and --clear-replaced-by cannot be used together")
	}
	return nil
}
