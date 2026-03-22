package cli

import (
	"fmt"
	"strings"
)

type machineOptions struct {
	jsonOutput     bool
	nonInteractive bool
	dryRun         bool
	explain        bool
}

type machineFlagConfig struct {
	allowDryRun  bool
	allowExplain bool
}

type machineFlagHandler func(hasValue bool, config machineFlagConfig, options *machineOptions) error

func parseMachineFlags(args []string, config machineFlagConfig) (machineOptions, []string, error) {
	options := machineOptions{}
	for _, arg := range args {
		name, _, _ := strings.Cut(arg, "=")
		if name == "--json" {
			options.jsonOutput = true
		}
	}
	remaining := make([]string, 0, len(args))
	for _, arg := range args {
		handled, err := applyMachineFlag(arg, config, &options)
		if err != nil {
			return options, remaining, err
		}
		if handled {
			continue
		}
		remaining = append(remaining, arg)
	}
	return options, remaining, nil
}

func applyMachineFlag(arg string, config machineFlagConfig, options *machineOptions) (bool, error) {
	name, _, hasValue := strings.Cut(arg, "=")
	handler, ok := machineFlagHandlers[name]
	if !ok {
		return false, nil
	}
	if err := handler(hasValue, config, options); err != nil {
		return true, err
	}
	return true, nil
}

var machineFlagHandlers = map[string]machineFlagHandler{
	"--json": func(hasValue bool, _ machineFlagConfig, options *machineOptions) error {
		if err := requireNoValue("--json", hasValue); err != nil {
			return err
		}
		options.jsonOutput = true
		return nil
	},
	"--non-interactive": func(hasValue bool, _ machineFlagConfig, options *machineOptions) error {
		if err := requireNoValue("--non-interactive", hasValue); err != nil {
			return err
		}
		options.nonInteractive = true
		return nil
	},
	"--dry-run": func(hasValue bool, config machineFlagConfig, options *machineOptions) error {
		if err := requireNoValue("--dry-run", hasValue); err != nil {
			return err
		}
		if !config.allowDryRun {
			return fmt.Errorf("--dry-run is only supported for write commands")
		}
		options.dryRun = true
		return nil
	},
	"--explain": func(hasValue bool, config machineFlagConfig, options *machineOptions) error {
		if err := requireNoValue("--explain", hasValue); err != nil {
			return err
		}
		if !config.allowExplain {
			return fmt.Errorf("--explain is not supported for this command")
		}
		options.explain = true
		return nil
	},
}

func requireNoValue(name string, hasValue bool) error {
	if hasValue {
		return fmt.Errorf("%s does not accept a value", name)
	}
	return nil
}
