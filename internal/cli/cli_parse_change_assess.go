package cli

import (
	"fmt"
	"strings"
)

type changeAssessIntakeRequest struct {
	root           string
	explicitRoot   bool
	title          string
	changeType     string
	size           string
	description    string
	contextBundles []string
}

type changeAssessDecompositionRequest struct {
	root         string
	explicitRoot bool
	changeID     string
}

func parseChangeAssessIntakeArgs(args []string) (changeAssessIntakeRequest, error) {
	request := changeAssessIntakeRequest{root: "."}
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--title":
			return assignStringFlag(args, flag, &request.title)
		case "--type":
			return assignStringFlag(args, flag, &request.changeType)
		case "--size":
			return assignStringFlag(args, flag, &request.size)
		case "--description":
			return assignStringFlag(args, flag, &request.description)
		case "--bundle":
			return appendStringFlag(args, flag, &request.contextBundles)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change assess-intake flag %q", flag.raw)
		}
	}, func(arg string) error {
		return fmt.Errorf("unexpected positional argument %q", arg)
	})
	if err != nil {
		return changeAssessIntakeRequest{}, err
	}
	if strings.TrimSpace(request.title) == "" {
		return changeAssessIntakeRequest{}, fmt.Errorf("--title is required")
	}
	if strings.TrimSpace(request.changeType) == "" {
		return changeAssessIntakeRequest{}, fmt.Errorf("--type is required")
	}
	return request, nil
}

func parseChangeAssessDecompositionArgs(args []string) (changeAssessDecompositionRequest, error) {
	request := changeAssessDecompositionRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown change assess-decomposition flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeAssessDecompositionRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change assess-decomposition requires exactly one change ID")
	if err != nil {
		return changeAssessDecompositionRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}
