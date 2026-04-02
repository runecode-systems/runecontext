package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func runStandardCreate(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	request, err := parseStandardCreateArgs(args)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines(standardCreateCommand, standardCreateUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, standardCreateCommand, machine)
	if code != exitOK {
		return code
	}
	defer project.close()

	result, err := runStandardCreateMutation(project, machine, request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines(standardCreateCommand, project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := buildStandardMutationOutput(project.absRoot, project.loaded, standardCreateCommand, result)
	if machine.explain {
		output = appendStandardMutationExplainLines(output, standardCreateCommand, result)
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runStandardCreateMutation(project *cliProject, machine machineOptions, request standardCreateRequest) (*contracts.StandardMutationResult, error) {
	return runChangeOperation(project, machine, func(v *contracts.Validator, loaded *contracts.LoadedProject) (*contracts.StandardMutationResult, error) {
		return contracts.CreateStandard(v, loaded, contracts.StandardCreateOptions{
			Path:                    request.path,
			ID:                      request.id,
			Title:                   request.title,
			Status:                  contracts.StandardStatus(strings.TrimSpace(request.status)),
			ReplacedBy:              request.replacedBy,
			Aliases:                 request.aliases,
			SuggestedContextBundles: request.suggestedContextBundles,
			Body:                    request.body,
		})
	})
}

func parseStandardCreateArgs(args []string) (standardCreateRequest, error) {
	request := standardCreateRequest{root: "."}
	positionals := make([]string, 0, len(args))
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return parseStandardCreateFlag(args, flag, &request)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return standardCreateRequest{}, err
	}
	if len(positionals) > 0 {
		return standardCreateRequest{}, fmt.Errorf("standard create does not accept positional arguments")
	}
	if err := validateStandardCreateRequest(request); err != nil {
		return standardCreateRequest{}, err
	}
	return request, nil
}

func parseStandardCreateFlag(args []string, flag parsedFlag, request *standardCreateRequest) (int, error) {
	switch flag.name {
	case "--path":
		return assignStringFlag(args, flag, &request.path)
	case "--id":
		return assignStringFlag(args, flag, &request.id)
	case "--title":
		return assignStringFlag(args, flag, &request.title)
	case "--status":
		return assignStringFlag(args, flag, &request.status)
	case "--replaced-by":
		return assignStringFlag(args, flag, &request.replacedBy)
	case "--alias":
		return appendStringFlag(args, flag, &request.aliases)
	case "--suggested-context-bundle":
		return appendStringFlag(args, flag, &request.suggestedContextBundles)
	case "--body":
		return assignStringFlag(args, flag, &request.body)
	case "--project-path":
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	default:
		return flag.next, fmt.Errorf("unknown standard create flag %q", flag.raw)
	}
}

func validateStandardCreateRequest(request standardCreateRequest) error {
	if strings.TrimSpace(request.path) == "" {
		return fmt.Errorf("standard create requires --path")
	}
	if strings.TrimSpace(request.title) == "" {
		return fmt.Errorf("standard create requires --title")
	}
	return nil
}
