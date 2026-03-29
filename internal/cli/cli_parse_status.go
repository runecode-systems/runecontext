package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type statusRequest struct {
	root            string
	explicitRoot    bool
	historyMode     string
	historyModeSet  bool
	historyLimit    int
	historyLimitSet bool
	verbose         bool
}

func parseStatusArgs(args []string) (statusRequest, error) {
	if len(args) == 1 && isHelpToken(args[0]) {
		return statusRequest{root: args[0], explicitRoot: true}, nil
	}
	if len(args) > 1 && isHelpToken(args[0]) {
		return statusRequest{}, fmt.Errorf("help does not accept additional arguments")
	}
	request := statusRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return handleStatusFlag(args, flag, &request)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return statusRequest{}, err
	}
	if strings.TrimSpace(request.historyMode) == "" {
		request.historyMode = statusHistoryModeRecent
	}
	if request.historyLimit == 0 {
		request.historyLimit = defaultStatusHistoryLimit
	}
	baseRequest, err := finalizeOptionalPath(request.root, request.explicitRoot, positionals)
	if err != nil {
		return statusRequest{}, err
	}
	request.root = baseRequest.root
	request.explicitRoot = baseRequest.explicitRoot
	return request, nil
}

func handleStatusFlag(args []string, flag parsedFlag, request *statusRequest) (int, error) {
	switch flag.name {
	case "--path":
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	case "--history":
		next, err := assignStatusHistoryModeFlag(args, flag, &request.historyMode)
		if err == nil {
			request.historyModeSet = true
		}
		return next, err
	case "--history-limit":
		next, err := assignStatusHistoryLimitFlag(args, flag, &request.historyLimit)
		if err == nil {
			request.historyLimitSet = true
		}
		return next, err
	case "--verbose":
		if err := requireNoValue("--verbose", flag.hasValue); err != nil {
			return flag.next, err
		}
		request.verbose = true
		return flag.next, nil
	default:
		return flag.next, fmt.Errorf("unknown status flag %q", flag.raw)
	}
}

func assignStatusHistoryModeFlag(args []string, flag parsedFlag, target *string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	mode := strings.ToLower(strings.TrimSpace(value))
	switch mode {
	case statusHistoryModeRecent, statusHistoryModeAll, statusHistoryModeNone:
		*target = mode
		return next, nil
	default:
		return flag.next, fmt.Errorf("--history must be one of recent, all, or none")
	}
}

func assignStatusHistoryLimitFlag(args []string, flag parsedFlag, target *int) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return flag.next, fmt.Errorf("--history-limit must be a positive integer")
	}
	*target = parsed
	return next, nil
}
