package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var semverLikePattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type upgradeRequest struct {
	root          string
	explicitRoot  bool
	targetVersion string
	apply         bool
}

func runUpgrade(args []string, stdout, stderr io.Writer) int {
	machine, remaining, err := parseMachineFlags(args, machineFlagConfig{allowExplain: true})
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("upgrade", upgradeUsage, err), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	request, usageErr, usage := parseUpgradeRequest(remaining, stdout, machine)
	if usageErr != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("upgrade", usage, usageErr), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if usage == "" {
		return exitOK
	}
	project, code := loadProjectOrReport(request.root, request.explicitRoot, stderr, "upgrade", machine)
	if code != exitOK {
		return code
	}
	defer project.close()
	if request.apply {
		return runUpgradeApply(project, request, machine, stdout, stderr)
	}
	return runUpgradePreview(project, request, machine, stdout)
}

func parseUpgradeRequest(args []string, stdout io.Writer, machine machineOptions) (upgradeRequest, error, string) {
	if handled, err, usage := maybeHandleUpgradeHelp(args, stdout, machine); handled {
		return upgradeRequest{}, err, usage
	}
	if len(args) > 0 && args[0] == "apply" {
		request, err := parseUpgradeApplyArgs(args[1:])
		return request, err, upgradeApplyUsage
	}
	request, err := parseUpgradePreviewArgs(args)
	return request, err, upgradeUsage
}

func maybeHandleUpgradeHelp(args []string, stdout io.Writer, machine machineOptions) (bool, error, string) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return true, fmt.Errorf("help does not accept additional arguments"), upgradeUsage
		}
		emitUpgradeUsage(stdout, machine, upgradeUsage)
		return true, nil, ""
	}
	if len(args) > 1 && args[0] == "apply" && isHelpToken(args[1]) {
		if len(args) != 2 {
			return true, fmt.Errorf("help does not accept additional arguments"), upgradeApplyUsage
		}
		emitUpgradeUsage(stdout, machine, upgradeApplyUsage)
		return true, nil, ""
	}
	return false, nil, ""
}

func emitUpgradeUsage(stdout io.Writer, machine machineOptions, usage string) {
	emitOutput(stdout, machine, appendMachineOptionLines([]line{{"result", "ok"}, {"command", "upgrade"}, {"usage", usage}}, machine), exitOK, failureClassNone)
}

func parseUpgradePreviewArgs(args []string) (upgradeRequest, error) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return upgradeRequest{}, fmt.Errorf("help does not accept additional arguments")
		}
		return upgradeRequest{}, nil
	}
	request := upgradeRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--target-version":
			return assignStringFlag(args, flag, &request.targetVersion)
		default:
			return flag.next, fmt.Errorf("unknown upgrade flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return upgradeRequest{}, err
	}
	if len(positionals) > 0 {
		return upgradeRequest{}, fmt.Errorf("unknown upgrade subcommand %q", positionals[0])
	}
	if err := validateUpgradeTargetVersion(request.targetVersion, false); err != nil {
		return upgradeRequest{}, err
	}
	return request, nil
}

func parseUpgradeApplyArgs(args []string) (upgradeRequest, error) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return upgradeRequest{}, fmt.Errorf("help does not accept additional arguments")
		}
		return upgradeRequest{}, nil
	}
	request := upgradeRequest{root: ".", apply: true}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--target-version":
			return assignStringFlag(args, flag, &request.targetVersion)
		default:
			return flag.next, fmt.Errorf("unknown upgrade apply flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return upgradeRequest{}, err
	}
	if len(positionals) > 0 {
		return upgradeRequest{}, fmt.Errorf("upgrade apply does not accept positional arguments")
	}
	if err := validateUpgradeTargetVersion(request.targetVersion, true); err != nil {
		return upgradeRequest{}, err
	}
	return request, nil
}

func validateUpgradeTargetVersion(target string, apply bool) error {
	target = strings.TrimSpace(target)
	if target == "" {
		if apply {
			return fmt.Errorf("upgrade apply requires --target-version")
		}
		return nil
	}
	if !semverLikePattern.MatchString(target) {
		return fmt.Errorf("--target-version %q must look like a semantic version", target)
	}
	return nil
}

func runUpgradePreview(project *cliProject, request upgradeRequest, machine machineOptions, stdout io.Writer) int {
	current := strings.TrimSpace(fmt.Sprint(project.loaded.RootConfig["runecontext_version"]))
	target := strings.TrimSpace(request.targetVersion)
	if target == "" {
		target = current
	}
	state := "current"
	plan := "no changes required"
	if target != current {
		state = "upgradeable"
		plan = fmt.Sprintf("set runecontext_version to %s", target)
	}
	output := []line{
		{"result", "ok"},
		{"command", "upgrade"},
		{"phase", "preview"},
		{"root", project.absRoot},
		{"selected_config_path", selectedConfigPath(project.loaded)},
		{"current_version", current},
		{"target_version", target},
		{"state", state},
		{"plan_action_1", plan},
		{"apply_required", boolString(state != "current")},
	}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runUpgradeApply(project *cliProject, request upgradeRequest, machine machineOptions, stdout, stderr io.Writer) int {
	current := strings.TrimSpace(fmt.Sprint(project.loaded.RootConfig["runecontext_version"]))
	target := strings.TrimSpace(request.targetVersion)
	if target == current {
		output := []line{{"result", "ok"}, {"command", "upgrade"}, {"phase", "apply"}, {"root", project.absRoot}, {"selected_config_path", selectedConfigPath(project.loaded)}, {"current_version", current}, {"target_version", target}, {"state", "current"}, {"changed", "false"}}
		emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
		return exitOK
	}
	configPath := selectedConfigPath(project.loaded)
	data, err := os.ReadFile(configPath)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	rewritten, err := rewriteRunecontextVersion(data, target)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	mode := configFileMode(configPath)
	if err := writeAtomicUpgradeConfig(configPath, rewritten, mode); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	output := []line{{"result", "ok"}, {"command", "upgrade"}, {"phase", "apply"}, {"root", project.absRoot}, {"selected_config_path", configPath}, {"previous_version", current}, {"target_version", target}, {"state", "upgradeable"}, {"changed", "true"}}
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}
