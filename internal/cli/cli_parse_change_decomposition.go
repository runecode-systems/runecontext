package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type pendingSubChangeCLI struct {
	id        string
	dependsOn []string
}

func parseChangeDecompositionArgs(args []string, commandLabel string) (changeDecompositionRequest, error) {
	request := changeDecompositionRequest{root: "."}
	positionals := make([]string, 0, 1)
	pending := map[string]*pendingSubChangeCLI{}
	order := make([]string, 0)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		return handleDecompositionFlag(args, flag, commandLabel, &request, pending, &order)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeDecompositionRequest{}, err
	}
	umbrellaID, err := requireExactPositional(positionals, fmt.Sprintf("%s requires exactly one umbrella change ID", commandLabel))
	if err != nil {
		return changeDecompositionRequest{}, err
	}
	request.umbrellaID = umbrellaID
	if len(order) == 0 {
		return changeDecompositionRequest{}, fmt.Errorf("%s requires at least one --sub-change ID", commandLabel)
	}
	request.subChanges = make([]contracts.SplitSubChange, 0, len(order))
	for _, id := range order {
		node := pending[id]
		request.subChanges = append(request.subChanges, contracts.SplitSubChange{ID: node.id, DependsOn: node.dependsOn})
	}
	return request, nil
}

func handleDecompositionFlag(args []string, flag parsedFlag, commandLabel string, request *changeDecompositionRequest, pending map[string]*pendingSubChangeCLI, order *[]string) (int, error) {
	switch flag.name {
	case "--sub-change":
		return handleSubChangeFlag(args, flag, commandLabel, pending, order)
	case "--depends-on":
		return handleDependsOnFlag(args, flag, commandLabel, pending, order)
	case "--path":
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	default:
		return flag.next, fmt.Errorf("unknown %s flag %q", commandLabel, flag.raw)
	}
}

func handleSubChangeFlag(args []string, flag parsedFlag, commandLabel string, pending map[string]*pendingSubChangeCLI, order *[]string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	id := strings.TrimSpace(value)
	if id == "" {
		return flag.next, fmt.Errorf("%s requires non-empty --sub-change value", commandLabel)
	}
	ensurePendingSubChange(id, pending, order)
	return next, nil
}

func handleDependsOnFlag(args []string, flag parsedFlag, commandLabel string, pending map[string]*pendingSubChangeCLI, order *[]string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	subID, depID, err := parseDependsOnEdge(value, commandLabel)
	if err != nil {
		return flag.next, err
	}
	node := ensurePendingSubChange(subID, pending, order)
	node.dependsOn = append(node.dependsOn, depID)
	return next, nil
}

func parseDependsOnEdge(raw, commandLabel string) (string, string, error) {
	rawSubID, rawDepID, ok := strings.Cut(raw, ":")
	if !ok {
		return "", "", fmt.Errorf("%s --depends-on must use SUB_CHANGE_ID:CHANGE_ID", commandLabel)
	}
	subID := strings.TrimSpace(rawSubID)
	depID := strings.TrimSpace(rawDepID)
	if subID == "" || depID == "" {
		return "", "", fmt.Errorf("%s --depends-on must use SUB_CHANGE_ID:CHANGE_ID", commandLabel)
	}
	return subID, depID, nil
}

func ensurePendingSubChange(id string, pending map[string]*pendingSubChangeCLI, order *[]string) *pendingSubChangeCLI {
	node, ok := pending[id]
	if ok {
		return node
	}
	node = &pendingSubChangeCLI{id: id}
	pending[id] = node
	*order = append(*order, id)
	return node
}
