package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type parsedFlag struct {
	raw      string
	name     string
	value    string
	index    int
	next     int
	hasValue bool
}

func consumeArgs(args []string, onFlag func(parsedFlag) (int, error), onPositional func(string) error) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if err := onPositional(arg); err != nil {
				return err
			}
			continue
		}
		flag, err := parseFlag(args, i)
		if err != nil {
			return err
		}
		next, err := onFlag(flag)
		if err != nil {
			return err
		}
		i = next
	}
	return nil
}

func parseFlag(args []string, index int) (parsedFlag, error) {
	arg := args[index]
	if !strings.HasPrefix(arg, "--") {
		return parsedFlag{}, fmt.Errorf("unknown flag %q", arg)
	}
	if name, value, ok := strings.Cut(arg, "="); ok {
		return parsedFlag{raw: arg, name: name, value: strings.TrimSpace(value), index: index, next: index, hasValue: true}, nil
	}
	return parsedFlag{raw: arg, name: arg, index: index, next: index}, nil
}

func (flag parsedFlag) requireValue(args []string) (string, int, error) {
	if flag.hasValue {
		if flag.value == "" {
			return "", flag.next, fmt.Errorf("%s requires a value", flag.name)
		}
		return flag.value, flag.next, nil
	}
	return requireFlagValue(args, flag.index, flag.name)
}

func assignStringFlag(args []string, flag parsedFlag, target *string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	*target = value
	return next, nil
}

func appendStringFlag(args []string, flag parsedFlag, target *[]string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	*target = append(*target, value)
	return next, nil
}

func assignRootFlag(args []string, flag parsedFlag, root *string, explicitRoot *bool) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	setExplicitRoot(root, explicitRoot, value)
	return next, nil
}

func assignClosedAtFlag(args []string, flag parsedFlag, target *time.Time) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	closedAt, err := parseClosedAt(value)
	if err != nil {
		return flag.next, err
	}
	*target = closedAt
	return next, nil
}

func requireFlagValue(args []string, index int, flag string) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	value := strings.TrimSpace(args[index+1])
	if value == "" {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	if strings.HasPrefix(value, "--") {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	return value, index + 1, nil
}

func finalizeOptionalPath(root string, explicitRoot bool, positionals []string) (statusRequest, error) {
	if len(positionals) > 1 {
		return statusRequest{}, fmt.Errorf("expected at most one path argument")
	}
	request := statusRequest{root: root, explicitRoot: explicitRoot}
	if len(positionals) == 1 {
		request.root = positionals[0]
		request.explicitRoot = true
	}
	return request, nil
}

func requireExactPositional(positionals []string, message string) (string, error) {
	if len(positionals) != 1 {
		return "", errors.New(message)
	}
	return positionals[0], nil
}

func parseClosedAt(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("--closed-at must use YYYY-MM-DD")
	}
	return parsed, nil
}

func setExplicitRoot(root *string, explicitRoot *bool, value string) {
	*root = value
	*explicitRoot = true
}
