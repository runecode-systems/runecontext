package cli

import (
	"strings"
	"testing"
)

func TestRunUpgradePreviewHumanOutputShowsNextStep(t *testing.T) {
	plan := upgradePlan{
		State:          upgradeStateUpgradeable,
		CurrentVersion: "0.1.0-alpha.9",
		TargetVersion:  "0.1.0-alpha.10",
		PlanActions:    []string{"set runecontext_version to 0.1.0-alpha.10"},
		NextActions:    []string{"run `runectx upgrade apply`"},
	}
	out := renderHumanUpgradePreview(plan, "/tmp/project", "/tmp/project/runecontext.yaml", upgradeHumanOptions{color: false})
	if !strings.Contains(out, "RuneContext Upgrade Preview") {
		t.Fatalf("expected human preview title, got %q", out)
	}
	if !strings.Contains(out, "Project can be upgraded to 0.1.0-alpha.10.") {
		t.Fatalf("expected upgradeable summary text, got %q", out)
	}
	if !strings.Contains(out, "Next") || !strings.Contains(out, "run `runectx upgrade apply`") {
		t.Fatalf("expected apply next-step guidance, got %q", out)
	}
	if strings.Contains(out, "target_version=") {
		t.Fatalf("expected human output instead of machine key/value output, got %q", out)
	}
}

func TestRunUpgradeCLIPreviewHumanOutputShowsApplyNextStep(t *testing.T) {
	plan := cliUpgradePlan{
		Availability:    cliUpgradeAvailabilityAvailable,
		CurrentVersion:  "0.1.0-alpha.8",
		SelectedRelease: "0.1.0-alpha.9",
		TargetRelease:   "0.1.0-alpha.9",
		InstallAction:   "install_selected_release",
		Platform:        "linux/amd64",
	}
	out := renderHumanCLIUpgradePreview(plan, upgradeHumanOptions{color: false})
	if !strings.Contains(out, "RuneContext CLI Update Preview") {
		t.Fatalf("expected human CLI preview title, got %q", out)
	}
	if !strings.Contains(out, "A newer runectx release is available.") {
		t.Fatalf("expected availability summary, got %q", out)
	}
	if !strings.Contains(out, "run `runectx upgrade cli apply --target-version 0.1.0-alpha.9`") {
		t.Fatalf("expected CLI apply next-step guidance, got %q", out)
	}
	if strings.Contains(out, "availability_state=") {
		t.Fatalf("expected human output instead of machine key/value output, got %q", out)
	}
}

func TestRunUpgradeCLIApplyHumanOutputShowsAlreadyCurrentSummary(t *testing.T) {
	plan := cliUpgradePlan{
		Availability:    cliUpgradeAvailabilityUpToDate,
		CurrentVersion:  "0.1.0-alpha.8",
		SelectedRelease: "0.1.0-alpha.8",
		TargetRelease:   "0.1.0-alpha.8",
		InstalledBinary: "runectx",
		Platform:        "linux/amd64",
		Changed:         false,
	}
	out := renderHumanCLIUpgradeApply(plan, upgradeHumanOptions{color: false})
	if !strings.Contains(out, "RuneContext CLI Already Current") {
		t.Fatalf("expected already-current title, got %q", out)
	}
	if !strings.Contains(out, "Changed: no") || !strings.Contains(out, "Installed runectx is already current.") {
		t.Fatalf("expected changed summary, got %q", out)
	}
}

func TestRunUpgradePreviewHumanOutputDoesNotDuplicateHopActions(t *testing.T) {
	plan := upgradePlan{
		State:          upgradeStateUpgradeable,
		CurrentVersion: "0.1.0-alpha.12",
		TargetVersion:  "0.1.0-alpha.13",
		UpgradeHops:    []upgradeHop{{From: "0.1.0-alpha.12", To: "0.1.0-alpha.13"}},
		PlanActions: []string{
			"migrate runecontext_version 0.1.0-alpha.12 -> 0.1.0-alpha.13",
			"set runecontext_version to 0.1.0-alpha.13",
		},
	}
	out := renderHumanUpgradePreview(plan, "/tmp/project", "/tmp/project/runecontext.yaml", upgradeHumanOptions{color: false})
	if got := strings.Count(out, "migrate runecontext_version 0.1.0-alpha.12 -> 0.1.0-alpha.13"); got != 1 {
		t.Fatalf("expected migration line once, got %d in output %q", got, out)
	}
	if !strings.Contains(out, "hop 0.1.0-alpha.12 -> 0.1.0-alpha.13") {
		t.Fatalf("expected hop line in output, got %q", out)
	}
}
