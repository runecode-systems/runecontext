package cli

import (
	"fmt"
	"io"
	"strings"
)

type upgradeHumanOptions struct {
	color bool
}

func upgradeHumanOptionsForMachine(stdout io.Writer, machine machineOptions) upgradeHumanOptions {
	return upgradeHumanOptions{color: shouldUseStatusColor(stdout)}
}

func renderHumanUpgradePreview(plan upgradePlan, root, configPath string, options upgradeHumanOptions) string {
	var builder strings.Builder
	appendUpgradeHeader(&builder, "RuneContext Upgrade Preview", options)
	appendUpgradeSection(&builder, "Summary", []string{
		sanitizeStatusText(fmt.Sprintf("Project: %s", root)),
		sanitizeStatusText(fmt.Sprintf("Config: %s", configPath)),
		sanitizeStatusText(fmt.Sprintf("Current: %s", plan.CurrentVersion)),
		sanitizeStatusText(fmt.Sprintf("Target: %s", plan.TargetVersion)),
		sanitizeStatusText(fmt.Sprintf("State: %s", renderUpgradeStateBadge(string(plan.State), options.color))),
		sanitizeStatusText(fmt.Sprintf("Hops: %d", len(plan.UpgradeHops))),
		sanitizeStatusText(fmt.Sprintf("Network: %s", boolLabel(plan.NetworkAccess))),
	}, options)
	appendUpgradeSection(&builder, "Status", upgradePreviewStatusSummary(plan), options)
	appendUpgradeSection(&builder, "Plan", appendUpgradePlanLines(nil, plan), options)
	appendUpgradeSection(&builder, "Next", sanitizeUpgradeLines(plan.NextActions), options)
	appendUpgradeSection(&builder, "Conflicts", sanitizeUpgradeLines(plan.Conflicts), options)
	appendUpgradeSection(&builder, "Warnings", sanitizeUpgradeLines(plan.Warnings), options)
	return builder.String()
}

func renderHumanUpgradeApply(root, configPath, previous, current, target string, changed bool, mutations []string, networkAccess bool, options upgradeHumanOptions) string {
	var builder strings.Builder
	title := "RuneContext Upgrade Applied"
	if !changed {
		title = "RuneContext Upgrade No Changes"
	}
	appendUpgradeHeader(&builder, title, options)
	appendUpgradeSection(&builder, "Summary", []string{
		sanitizeStatusText(fmt.Sprintf("Project: %s", root)),
		sanitizeStatusText(fmt.Sprintf("Config: %s", configPath)),
		sanitizeStatusText(fmt.Sprintf("Previous: %s", previous)),
		sanitizeStatusText(fmt.Sprintf("Current: %s", current)),
		sanitizeStatusText(fmt.Sprintf("Target: %s", target)),
		sanitizeStatusText(fmt.Sprintf("Changed: %s", boolLabel(changed))),
		sanitizeStatusText(fmt.Sprintf("Network: %s", boolLabel(networkAccess))),
	}, options)
	appendUpgradeSection(&builder, "Changes", sanitizeUpgradeLines(mutations), options)
	return builder.String()
}

func renderHumanCLIUpgradePreview(plan cliUpgradePlan, options upgradeHumanOptions) string {
	var builder strings.Builder
	appendUpgradeHeader(&builder, "RuneContext CLI Update Preview", options)
	appendUpgradeSection(&builder, "Summary", []string{
		sanitizeStatusText(fmt.Sprintf("Installed: %s", plan.CurrentVersion)),
		sanitizeStatusText(fmt.Sprintf("Selected: %s", plan.SelectedRelease)),
		sanitizeStatusText(fmt.Sprintf("Target: %s", plan.TargetRelease)),
		sanitizeStatusText(fmt.Sprintf("State: %s", renderUpgradeStateBadge(string(plan.Availability), options.color))),
		sanitizeStatusText(fmt.Sprintf("Action: %s", plan.InstallAction)),
		sanitizeStatusText(fmt.Sprintf("Platform: %s", plan.Platform)),
		sanitizeStatusText(fmt.Sprintf("Network: %s", boolLabel(plan.NetworkAccess))),
	}, options)
	appendUpgradeSection(&builder, "Status", cliUpgradePreviewStatusSummary(plan), options)
	appendUpgradeSection(&builder, "Next", cliUpgradePreviewNextSteps(plan), options)
	appendUpgradeSection(&builder, "Guidance", sanitizeUpgradeLines(plan.FailureGuidance), options)
	return builder.String()
}

func renderHumanCLIUpgradeApply(plan cliUpgradePlan, options upgradeHumanOptions) string {
	var builder strings.Builder
	title := "RuneContext CLI Update Applied"
	if !plan.Changed {
		title = "RuneContext CLI Already Current"
	}
	appendUpgradeHeader(&builder, title, options)
	appendUpgradeSection(&builder, "Summary", []string{
		sanitizeStatusText(fmt.Sprintf("Previous: %s", plan.CurrentVersion)),
		sanitizeStatusText(fmt.Sprintf("Selected: %s", plan.SelectedRelease)),
		sanitizeStatusText(fmt.Sprintf("Target: %s", plan.TargetRelease)),
		sanitizeStatusText(fmt.Sprintf("State: %s", renderUpgradeStateBadge(string(plan.Availability), options.color))),
		sanitizeStatusText(fmt.Sprintf("Changed: %s", boolLabel(plan.Changed))),
		sanitizeStatusText(fmt.Sprintf("Binary: %s", plan.InstalledBinary)),
		sanitizeStatusText(fmt.Sprintf("Updated Path: %s", plan.UpdatedBinaryPath)),
		sanitizeStatusText(fmt.Sprintf("Platform: %s", plan.Platform)),
		sanitizeStatusText(fmt.Sprintf("Network: %s", boolLabel(plan.NetworkAccess))),
	}, options)
	appendUpgradeSection(&builder, "Status", cliUpgradePreviewStatusSummary(plan), options)
	appendUpgradeSection(&builder, "Guidance", sanitizeUpgradeLines(plan.FailureGuidance), options)
	return builder.String()
}

func appendUpgradeHeader(builder *strings.Builder, title string, options upgradeHumanOptions) {
	builder.WriteString(styleStatusText(title, ansiBold+ansiBlue, options.color))
	builder.WriteString("\n")
}

func appendUpgradeKeyValue(builder *strings.Builder, label, value string) {
	value = sanitizeStatusText(value)
	if strings.TrimSpace(value) == "" {
		return
	}
	builder.WriteString(styleStatusText(label+":", ansiBold, false))
	builder.WriteString(" ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func appendUpgradeSection(builder *strings.Builder, title string, lines []string, options upgradeHumanOptions) {
	if len(lines) == 0 {
		return
	}
	builder.WriteString("\n")
	builder.WriteString(styleStatusText(title, ansiBold, options.color))
	builder.WriteString("\n")
	for _, line := range lines {
		builder.WriteString("- ")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
}

func appendUpgradePlanLines(lines []string, plan upgradePlan) []string {
	setVersionActions := make([]string, 0, 1)
	for _, action := range sanitizeUpgradeLines(plan.PlanActions) {
		if strings.HasPrefix(strings.TrimSpace(action), "set runecontext_version to ") {
			setVersionActions = append(setVersionActions, action)
			continue
		}
		lines = append(lines, action)
	}
	for _, hop := range plan.UpgradeHops {
		lines = append(lines, sanitizeStatusText(fmt.Sprintf("hop %s -> %s", hop.From, hop.To)))
	}
	lines = append(lines, setVersionActions...)
	return lines
}

func sanitizeUpgradeLines(values []string) []string {
	lines := make([]string, 0, len(values))
	for _, value := range values {
		value = sanitizeStatusText(value)
		if strings.TrimSpace(value) == "" {
			continue
		}
		lines = append(lines, value)
	}
	return lines
}

func cliUpgradePreviewNextSteps(plan cliUpgradePlan) []string {
	if plan.Availability != cliUpgradeAvailabilityAvailable {
		return nil
	}
	return []string{sanitizeStatusText(fmt.Sprintf("Next: %s", cliUpgradeApplyCommand(plan)))}
}

func cliUpgradeApplyCommand(plan cliUpgradePlan) string {
	target := strings.TrimSpace(plan.TargetRelease)
	if target == "" {
		target = strings.TrimSpace(plan.RequestedTarget)
	}
	if target == "" {
		return "run `runectx upgrade cli apply`"
	}
	return fmt.Sprintf("run `runectx upgrade cli apply --target-version %s`", target)
}

func upgradePreviewStatusSummary(plan upgradePlan) []string {
	switch plan.State {
	case upgradeStateCurrent:
		return []string{"Project already matches the selected target version."}
	case upgradeStateUpgradeable:
		return []string{fmt.Sprintf("Project can be upgraded to %s.", plan.TargetVersion)}
	case upgradeStateMixedOrStaleTree:
		return []string{"Project version is current, but managed artifacts need refresh."}
	case upgradeStateConflicted:
		return []string{"Managed artifact conflicts must be resolved before apply."}
	case upgradeStateProjectNewerThanCLI:
		return []string{"Project requires a newer runectx binary."}
	default:
		return nil
	}
}

func cliUpgradePreviewStatusSummary(plan cliUpgradePlan) []string {
	switch plan.Availability {
	case cliUpgradeAvailabilityUpToDate:
		return []string{"Installed runectx is already current."}
	case cliUpgradeAvailabilityAvailable:
		return []string{"A newer runectx release is available."}
	case cliUpgradeAvailabilityUnknown:
		return []string{"Could not determine the latest release."}
	default:
		return nil
	}
}

func renderUpgradeStateBadge(state string, color bool) string {
	label := fmt.Sprintf("[%s]", emptyAsDash(state))
	switch strings.ToLower(strings.TrimSpace(state)) {
	case string(upgradeStateCurrent), string(cliUpgradeAvailabilityUpToDate):
		return styleStatusText(label, ansiGreen, color)
	case string(upgradeStateUpgradeable), string(upgradeStateMixedOrStaleTree), string(cliUpgradeAvailabilityAvailable):
		return styleStatusText(label, ansiYellow, color)
	case string(upgradeStateConflicted), string(upgradeStateUnsupportedProjectVersion), string(upgradeStateProjectNewerThanCLI):
		return styleStatusText(label, ansiRed, color)
	default:
		return styleStatusText(label, ansiDim, color)
	}
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
