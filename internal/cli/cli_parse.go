package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type statusRequest struct {
	root         string
	explicitRoot bool
}

type changeNewRequest struct {
	root           string
	explicitRoot   bool
	title          string
	changeType     string
	size           string
	description    string
	mode           string
	contextBundles []string
}

type changeShapeRequest struct {
	root         string
	explicitRoot bool
	changeID     string
	design       string
	verification string
	tasks        []string
	references   []string
}

type changeCloseRequest struct {
	root               string
	explicitRoot       bool
	changeID           string
	verificationStatus string
	closedAt           time.Time
	supersededBy       []string
}

type changeReallocateRequest struct {
	root         string
	explicitRoot bool
	changeID     string
}

func parseStatusArgs(args []string) (statusRequest, error) {
	request := statusRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return flag.next, fmt.Errorf("unknown status flag %q", flag.raw)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return statusRequest{}, err
	}
	return finalizeOptionalPath(request.root, request.explicitRoot, positionals)
}

func parseChangeNewArgs(args []string) (changeNewRequest, error) {
	request := changeNewRequest{root: "."}
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
		case "--shape":
			return assignStringFlag(args, flag, &request.mode)
		case "--bundle":
			return appendStringFlag(args, flag, &request.contextBundles)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change new flag %q", flag.raw)
		}
	}, func(arg string) error {
		return fmt.Errorf("unexpected positional argument %q", arg)
	})
	if err != nil {
		return changeNewRequest{}, err
	}
	return finalizeChangeNewRequest(request)
}

func finalizeChangeNewRequest(request changeNewRequest) (changeNewRequest, error) {
	if strings.TrimSpace(request.title) == "" {
		return changeNewRequest{}, fmt.Errorf("--title is required")
	}
	if strings.TrimSpace(request.changeType) == "" {
		return changeNewRequest{}, fmt.Errorf("--type is required")
	}
	if request.mode != "" && request.mode != string(contracts.ChangeModeMinimum) && request.mode != string(contracts.ChangeModeFull) {
		return changeNewRequest{}, fmt.Errorf("--shape must be %q or %q", contracts.ChangeModeMinimum, contracts.ChangeModeFull)
	}
	return request, nil
}

func parseChangeShapeArgs(args []string) (changeShapeRequest, error) {
	request := changeShapeRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--design":
			return assignStringFlag(args, flag, &request.design)
		case "--verification":
			return assignStringFlag(args, flag, &request.verification)
		case "--task":
			return appendStringFlag(args, flag, &request.tasks)
		case "--reference":
			return appendStringFlag(args, flag, &request.references)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change shape flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeShapeRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change shape requires exactly one change ID")
	if err != nil {
		return changeShapeRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}

func parseChangeCloseArgs(args []string) (changeCloseRequest, error) {
	request := changeCloseRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--verification-status":
			return assignStringFlag(args, flag, &request.verificationStatus)
		case "--superseded-by":
			return appendStringFlag(args, flag, &request.supersededBy)
		case "--closed-at":
			return assignClosedAtFlag(args, flag, &request.closedAt)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change close flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeCloseRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change close requires exactly one change ID")
	if err != nil {
		return changeCloseRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}

func parseChangeReallocateArgs(args []string) (changeReallocateRequest, error) {
	request := changeReallocateRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown change reallocate flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeReallocateRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change reallocate requires exactly one change ID")
	if err != nil {
		return changeReallocateRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}
