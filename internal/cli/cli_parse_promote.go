package cli

import (
	"fmt"
	"strings"
)

type promoteRequest struct {
	root         string
	explicitRoot bool
	changeID     string
	accept       bool
	complete     bool
	targets      []string
}

func parsePromoteArgs(args []string) (promoteRequest, error) {
	request := promoteRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return parsePromoteFlag(args, &request, flag)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return promoteRequest{}, err
	}
	if request.accept && request.complete {
		return promoteRequest{}, fmt.Errorf("--accept and --complete cannot be used together")
	}
	if err := validatePromoteTargetArgs(request.targets); err != nil {
		return promoteRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "promote requires exactly one change ID")
	if err != nil {
		return promoteRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}

func parsePromoteFlag(args []string, request *promoteRequest, flag parsedFlag) (int, error) {
	switch flag.name {
	case "--accept":
		if flag.hasValue {
			return flag.next, fmt.Errorf("--accept does not accept a value")
		}
		request.accept = true
		return flag.next, nil
	case "--complete":
		if flag.hasValue {
			return flag.next, fmt.Errorf("--complete does not accept a value")
		}
		request.complete = true
		return flag.next, nil
	case "--target":
		return appendStringFlag(args, flag, &request.targets)
	case "--path":
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	default:
		return flag.next, fmt.Errorf("unknown promote flag %q", flag.raw)
	}
}

func validatePromoteTargetArgs(targets []string) error {
	for _, target := range targets {
		targetType, targetPath, ok := strings.Cut(strings.TrimSpace(target), ":")
		if !ok || strings.TrimSpace(targetType) == "" || strings.TrimSpace(targetPath) == "" {
			return fmt.Errorf("--target must use TYPE:PATH")
		}
	}
	return nil
}
