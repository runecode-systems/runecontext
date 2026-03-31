package cli

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

func runUpgradeCLI(args []string, machine machineOptions, stdout, stderr io.Writer) int {
	if handled, code := runUpgradeCLIHelpIfRequested(args, machine, stdout, stderr); handled {
		return code
	}
	request, usageErr, usage := parseUpgradeCLIRequest(args)
	if usageErr != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("upgrade", usage, usageErr), machine), exitUsage, failureClassUsage)
		return exitUsage
	}
	if usage == "" {
		return exitOK
	}
	plan, err := buildCLIUpgradePlan(request)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", "", err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	if request.apply {
		return runUpgradeCLIApply(plan, machine, stdout, stderr)
	}
	emitCLIUpgradePreviewResult(stdout, machine, plan)
	return exitOK
}

func runUpgradeCLIApply(plan cliUpgradePlan, machine machineOptions, stdout, stderr io.Writer) int {
	if plan.Availability != cliUpgradeAvailabilityAvailable {
		plan.Changed = false
		emitCLIUpgradeApplyResult(stdout, machine, plan)
		return exitOK
	}
	applied, err := applyCLIUpgradePlanFn.Apply(plan)
	if err != nil {
		emitOutput(stderr, machine, appendMachineOptionLines(buildCommandInvalidLines("upgrade", "", err), machine), exitInvalid, failureClassInvalid)
		return exitInvalid
	}
	emitCLIUpgradeApplyResult(stdout, machine, applied)
	return exitOK
}

func emitCLIUpgradePreviewResult(stdout io.Writer, machine machineOptions, plan cliUpgradePlan) {
	if !machine.jsonOutput {
		_, _ = io.WriteString(stdout, renderHumanCLIUpgradePreview(plan, upgradeHumanOptionsForMachine(stdout, machine)))
		return
	}
	emitOutput(stdout, machine, appendMachineOptionLines(buildCLIUpgradePreviewOutput(plan), machine), exitOK, failureClassNone)
}

func emitCLIUpgradeApplyResult(stdout io.Writer, machine machineOptions, plan cliUpgradePlan) {
	if !machine.jsonOutput {
		_, _ = io.WriteString(stdout, renderHumanCLIUpgradeApply(plan, upgradeHumanOptionsForMachine(stdout, machine)))
		return
	}
	emitOutput(stdout, machine, appendMachineOptionLines(buildCLIUpgradeApplyOutput(plan), machine), exitOK, failureClassNone)
}

func runUpgradeCLIHelpIfRequested(args []string, machine machineOptions, stdout, stderr io.Writer) (bool, int) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("upgrade", upgradeCLIUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return true, exitUsage
		}
		emitUpgradeUsage(stdout, machine, upgradeCLIUsage)
		return true, exitOK
	}
	if len(args) > 1 && args[0] == "apply" && isHelpToken(args[1]) {
		if len(args) != 2 {
			emitOutput(stderr, machine, appendMachineOptionLines(buildCommandUsageErrorLines("upgrade", upgradeCLIApplyUsage, fmt.Errorf("help does not accept additional arguments")), machine), exitUsage, failureClassUsage)
			return true, exitUsage
		}
		emitUpgradeUsage(stdout, machine, upgradeCLIApplyUsage)
		return true, exitOK
	}
	return false, exitOK
}

func parseUpgradeCLIRequest(args []string) (cliUpgradeRequest, error, string) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return cliUpgradeRequest{}, fmt.Errorf("help does not accept additional arguments"), upgradeCLIUsage
		}
		return cliUpgradeRequest{}, nil, ""
	}
	if len(args) > 0 && args[0] == "apply" {
		request, err := parseUpgradeCLIApplyArgs(args[1:])
		return request, err, upgradeCLIApplyUsage
	}
	request, err := parseUpgradeCLIPreviewArgs(args)
	return request, err, upgradeCLIUsage
}

func parseUpgradeCLIPreviewArgs(args []string) (cliUpgradeRequest, error) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return cliUpgradeRequest{}, fmt.Errorf("help does not accept additional arguments")
		}
		return cliUpgradeRequest{}, nil
	}
	request := cliUpgradeRequest{}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--target-version":
			return assignStringFlag(args, flag, &request.targetVersion)
		default:
			return flag.next, fmt.Errorf("unknown upgrade cli flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return cliUpgradeRequest{}, err
	}
	if len(positionals) > 0 {
		return cliUpgradeRequest{}, fmt.Errorf("unknown upgrade cli subcommand %q", positionals[0])
	}
	if err := validateUpgradeTargetVersion(request.targetVersion); err != nil {
		return cliUpgradeRequest{}, err
	}
	return request, nil
}

func parseUpgradeCLIApplyArgs(args []string) (cliUpgradeRequest, error) {
	if len(args) > 0 && isHelpToken(args[0]) {
		if len(args) != 1 {
			return cliUpgradeRequest{}, fmt.Errorf("help does not accept additional arguments")
		}
		return cliUpgradeRequest{}, nil
	}
	request := cliUpgradeRequest{apply: true}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--target-version":
			return assignStringFlag(args, flag, &request.targetVersion)
		default:
			return flag.next, fmt.Errorf("unknown upgrade cli apply flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return cliUpgradeRequest{}, err
	}
	if len(positionals) > 0 {
		return cliUpgradeRequest{}, fmt.Errorf("upgrade cli apply does not accept positional arguments")
	}
	if strings.TrimSpace(request.targetVersion) == "" {
		request.targetVersion = "current"
	}
	if err := validateUpgradeTargetVersion(request.targetVersion); err != nil {
		return cliUpgradeRequest{}, err
	}
	return request, nil
}

func buildCLIUpgradePlan(request cliUpgradeRequest) (cliUpgradePlan, error) {
	current := normalizedRunecontextVersion()
	selected, network, err := resolveCLIUpgradeTargetRelease(current, request.targetVersion)
	if err != nil {
		return cliUpgradePlan{}, err
	}
	plan := cliUpgradePlan{
		Availability:    cliUpgradeAvailabilityUpToDate,
		CurrentVersion:  current,
		SelectedRelease: selected,
		TargetRelease:   selected,
		RequestedTarget: strings.TrimSpace(request.targetVersion),
		NetworkAccess:   network,
		Mutating:        request.apply,
		PlannedAction:   "no_install_needed",
		ReleaseSource:   "explicit_cli_upgrade",
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		InstallAction:   "none",
		InstalledBinary: "runectx",
	}
	if selected == "" {
		plan.Availability = cliUpgradeAvailabilityUnknown
		plan.PlannedAction = "unable_to_resolve_target_release"
		plan.InstallAction = "none"
		plan.FailureGuidance = []string{"rerun `runectx upgrade cli --target-version latest` when release metadata is available"}
		return plan, nil
	}
	if comparison, ok := compareKnownRunecontextVersions(current, selected); ok && comparison < 0 {
		plan.Availability = cliUpgradeAvailabilityAvailable
		plan.PlannedAction = "download_and_install"
		plan.InstallAction = "install_selected_release"
		plan.FailureGuidance = []string{"rerun `runectx upgrade cli` to preview the selected release", cliUpgradeApplyCommand(plan)}
		return plan, nil
	}
	plan.FailureGuidance = []string{"run `runectx version` to confirm the installed release"}
	return plan, nil
}

func resolveCLIUpgradeTargetRelease(currentVersion, requested string) (string, bool, error) {
	target := strings.TrimSpace(requested)
	if target == "" || strings.EqualFold(target, "current") || strings.EqualFold(target, "installed") {
		return currentVersion, false, nil
	}
	if strings.EqualFold(target, "latest") {
		release, err := resolveLatestCLIReleaseFn.ResolveLatestRelease(currentVersion)
		if err != nil {
			return "", true, err
		}
		resolved := strings.TrimSpace(release)
		if !semverLikePattern.MatchString(resolved) {
			return "", true, fmt.Errorf("resolved latest release %q must look like a semantic version", resolved)
		}
		return resolved, true, nil
	}
	if !semverLikePattern.MatchString(target) {
		return "", false, fmt.Errorf("--target-version %q must look like a semantic version", target)
	}
	return target, false, nil
}
