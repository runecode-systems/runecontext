package cli

import (
	"fmt"
	"strings"
)

func parseChangeUpdateArgs(args []string) (changeUpdateRequest, error) {
	request := changeUpdateRequest{root: "."}
	positionals, err := parseChangeUpdateFlagsAndPositionals(args, &request)
	if err != nil {
		return changeUpdateRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change update requires exactly one change ID")
	if err != nil {
		return changeUpdateRequest{}, err
	}
	request.changeID = changeID
	return finalizeChangeUpdateRequest(request)
}

func parseChangeUpdateFlagsAndPositionals(args []string, request *changeUpdateRequest) ([]string, error) {
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, changeUpdateFlagHandler(args, request), func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return positionals, nil
}

func changeUpdateFlagHandler(args []string, request *changeUpdateRequest) func(parsedFlag) (int, error) {
	return func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--status":
			return assignStringFlag(args, flag, &request.status)
		case "--verification-status":
			return assignStringFlag(args, flag, &request.verificationStatus)
		case "--add-related-change":
			return appendStringFlag(args, flag, &request.addRelatedChanges)
		case "--remove-related-change":
			return appendStringFlag(args, flag, &request.removeRelatedChanges)
		case "--recursive":
			return assignNoValueBoolFlag(flag, &request.recursive)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change update flag %q", flag.raw)
		}
	}
}

func finalizeChangeUpdateRequest(request changeUpdateRequest) (changeUpdateRequest, error) {
	status := strings.TrimSpace(request.status)
	if status == "" {
		return changeUpdateRequest{}, fmt.Errorf("change update requires --status")
	}
	switch status {
	case "planned", "implemented", "verified":
		request.status = status
	default:
		return changeUpdateRequest{}, fmt.Errorf("change update --status must be one of planned, implemented, or verified")
	}
	request.verificationStatus = strings.TrimSpace(request.verificationStatus)
	return request, nil
}
