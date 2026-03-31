package cli

import (
	"fmt"
	"io"
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
	if len(remaining) > 0 && remaining[0] == "cli" {
		return runUpgradeCLI(remaining[1:], machine, stdout, stderr)
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
	return runUpgradePreview(project, request, machine, stdout, stderr)
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
	if err := validateUpgradeTargetVersion(request.targetVersion); err != nil {
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
	if err := validateUpgradeTargetVersion(request.targetVersion); err != nil {
		return upgradeRequest{}, err
	}
	return request, nil
}

func validateUpgradeTargetVersion(target string) error {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil
	}
	if isUpgradeTargetAlias(target) {
		return nil
	}
	if !semverLikePattern.MatchString(target) {
		return fmt.Errorf("--target-version %q must look like a semantic version", target)
	}
	return nil
}

func runUpgradePreview(project *cliProject, request upgradeRequest, machine machineOptions, stdout, stderr io.Writer) int {
	plan, err := buildUpgradePlan(project, request.targetVersion)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if !machine.jsonOutput {
		_, _ = io.WriteString(stdout, renderHumanUpgradePreview(plan, project.absRoot, selectedConfigPath(project.loaded), upgradeHumanOptionsForMachine(stdout, machine)))
		return exitOK
	}
	applyRequired := plan.State == upgradeStateUpgradeable || plan.State == upgradeStateMixedOrStaleTree
	output := []line{
		{"result", "ok"},
		{"command", "upgrade"},
		{"phase", "preview"},
		{"root", project.absRoot},
		{"selected_config_path", selectedConfigPath(project.loaded)},
		{"current_version", plan.CurrentVersion},
		{"target_version", plan.TargetVersion},
		{"state", string(plan.State)},
		{"hop_count", fmt.Sprintf("%d", len(plan.UpgradeHops))},
		{"network_access", boolString(plan.NetworkAccess)},
		{"plan_action_count", fmt.Sprintf("%d", len(plan.PlanActions))},
		{"apply_required", boolString(applyRequired)},
	}
	for i, hop := range plan.UpgradeHops {
		index := i + 1
		output = append(output,
			line{fmt.Sprintf("hop_%d_from", index), hop.From},
			line{fmt.Sprintf("hop_%d_to", index), hop.To},
		)
	}
	output = appendStringItems(output, "hop_action", plan.HopActions)
	output = appendStringItems(output, "plan_action", plan.PlanActions)
	output = appendStringItems(output, "next_action", plan.NextActions)
	output = appendStringItems(output, "conflict", plan.Conflicts)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func runUpgradeApply(project *cliProject, request upgradeRequest, machine machineOptions, stdout, stderr io.Writer) int {
	plan, err := buildUpgradePlan(project, request.targetVersion)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	configPath := selectedConfigPath(project.loaded)
	if plan.State == upgradeStateCurrent {
		if !machine.jsonOutput {
			_, _ = io.WriteString(stdout, renderHumanUpgradeApply(project.absRoot, configPath, plan.CurrentVersion, plan.CurrentVersion, plan.TargetVersion, false, nil, plan.NetworkAccess, upgradeHumanOptionsForMachine(stdout, machine)))
			return exitOK
		}
		output := upgradeApplyOutput(project.absRoot, configPath, plan.CurrentVersion, plan.CurrentVersion, plan.TargetVersion, false)
		emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
		return exitOK
	}
	if plan.State != upgradeStateUpgradeable && plan.State != upgradeStateMixedOrStaleTree {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, fmt.Errorf("upgrade apply is not allowed while state=%s", plan.State)), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if err := applyUpgradePlan(project, plan); err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", project.absRoot, err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if !machine.jsonOutput {
		_, _ = io.WriteString(stdout, renderHumanUpgradeApply(project.absRoot, configPath, plan.CurrentVersion, plan.TargetVersion, plan.TargetVersion, true, plan.ApplyMutations, plan.NetworkAccess, upgradeHumanOptionsForMachine(stdout, machine)))
		return exitOK
	}
	output := upgradeApplyOutput(project.absRoot, configPath, plan.CurrentVersion, plan.TargetVersion, plan.TargetVersion, true)
	output = append(output, line{"network_access", boolString(plan.NetworkAccess)})
	output = appendStringItems(output, "changed", plan.ApplyMutations)
	emitOutput(stdout, machine, appendMachineOptionLines(output, machine), exitOK, failureClassNone)
	return exitOK
}

func upgradeApplyOutput(root, configPath, previous, current, target string, changed bool) []line {
	return []line{
		{"result", "ok"},
		{"command", "upgrade"},
		{"phase", "apply"},
		{"root", root},
		{"selected_config_path", configPath},
		{"previous_version", previous},
		{"current_version", current},
		{"target_version", target},
		{"state", "current"},
		{"changed", boolString(changed)},
	}
}

func isUpgradeTargetAlias(value string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	switch trimmed {
	case "latest", "installed", "current":
		return true
	default:
		return false
	}
}
