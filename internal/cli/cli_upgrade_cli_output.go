package cli

import "strings"

func buildCLIUpgradePreviewOutput(plan cliUpgradePlan) []line {
	output := []line{
		{"result", "ok"},
		{"command", "upgrade"},
		{"phase", "preview"},
		{"scope", "cli"},
		{"availability_state", string(plan.Availability)},
		{"current_release", plan.CurrentVersion},
		{"selected_release", plan.SelectedRelease},
		{"target_release", plan.TargetRelease},
		{"planned_install_action", plan.PlannedAction},
		{"install_action", plan.InstallAction},
		{"network_access", boolString(plan.NetworkAccess)},
		{"release_source", plan.ReleaseSource},
		{"platform", plan.Platform},
	}
	return appendStringItems(output, "failure_guidance", plan.FailureGuidance)
}

func buildCLIUpgradeApplyOutput(plan cliUpgradePlan) []line {
	output := []line{
		{"result", "ok"},
		{"command", "upgrade"},
		{"phase", "apply"},
		{"scope", "cli"},
		{"availability_state", string(plan.Availability)},
		{"previous_release", plan.CurrentVersion},
		{"selected_release", plan.SelectedRelease},
		{"target_release", plan.TargetRelease},
		{"planned_install_action", plan.PlannedAction},
		{"install_action", plan.InstallAction},
		{"network_access", boolString(plan.NetworkAccess)},
		{"release_source", plan.ReleaseSource},
		{"platform", plan.Platform},
		{"installed_binary", plan.InstalledBinary},
		{"changed", boolString(plan.Changed)},
	}
	if strings.TrimSpace(plan.UpdatedBinaryPath) != "" {
		output = append(output, line{"updated_binary_path", plan.UpdatedBinaryPath})
	}
	return appendStringItems(output, "failure_guidance", plan.FailureGuidance)
}
