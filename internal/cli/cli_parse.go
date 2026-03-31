package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type changeNewRequest struct {
	root           string
	explicitRoot   bool
	title          string
	changeType     string
	size           string
	sizeProvided   bool
	description    string
	mode           string
	modeProvided   bool
	contextBundles []string
	bundleProvided bool
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
	recursive          bool
}

type changeReallocateRequest struct {
	root         string
	explicitRoot bool
	changeID     string
}

type changeUpdateRequest struct {
	root               string
	explicitRoot       bool
	changeID           string
	status             string
	verificationStatus string
	recursive          bool
}

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

func parseChangeNewArgs(args []string) (changeNewRequest, error) {
	request := changeNewRequest{root: "."}
	err := consumeArgs(args, changeNewFlagHandler(args, &request), func(arg string) error {
		return fmt.Errorf("unexpected positional argument %q", arg)
	})
	if err != nil {
		return changeNewRequest{}, err
	}
	return finalizeChangeNewRequest(request)
}

func changeNewFlagHandler(args []string, request *changeNewRequest) func(parsedFlag) (int, error) {
	return func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--title":
			return assignStringFlag(args, flag, &request.title)
		case "--type":
			return assignStringFlag(args, flag, &request.changeType)
		case "--size":
			return assignStringFlagWithProvided(args, flag, &request.size, &request.sizeProvided)
		case "--description":
			return assignStringFlag(args, flag, &request.description)
		case "--shape":
			return assignStringFlagWithProvided(args, flag, &request.mode, &request.modeProvided)
		case "--bundle":
			return appendStringFlagWithProvided(args, flag, &request.contextBundles, &request.bundleProvided)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change new flag %q", flag.raw)
		}
	}
}

func assignStringFlagWithProvided(args []string, flag parsedFlag, target *string, provided *bool) (int, error) {
	next, err := assignStringFlag(args, flag, target)
	if err == nil {
		*provided = true
	}
	return next, err
}

func appendStringFlagWithProvided(args []string, flag parsedFlag, target *[]string, provided *bool) (int, error) {
	next, err := appendStringFlag(args, flag, target)
	if err == nil {
		*provided = true
	}
	return next, err
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
		case "--recursive":
			return assignNoValueBoolFlag(flag, &request.recursive)
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

func parseChangeUpdateArgs(args []string) (changeUpdateRequest, error) {
	request := changeUpdateRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--status":
			return assignStringFlag(args, flag, &request.status)
		case "--verification-status":
			return assignStringFlag(args, flag, &request.verificationStatus)
		case "--recursive":
			return assignNoValueBoolFlag(flag, &request.recursive)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change update flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeUpdateRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change update requires exactly one change ID")
	if err != nil {
		return changeUpdateRequest{}, err
	}
	if strings.TrimSpace(request.status) == "" {
		return changeUpdateRequest{}, fmt.Errorf("change update requires --status")
	}
	switch strings.TrimSpace(request.status) {
	case "planned", "implemented", "verified":
		request.status = strings.TrimSpace(request.status)
	default:
		return changeUpdateRequest{}, fmt.Errorf("change update --status must be one of planned, implemented, or verified")
	}
	request.verificationStatus = strings.TrimSpace(request.verificationStatus)
	request.changeID = changeID
	return request, nil
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
