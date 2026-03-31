package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsCompatibleProjectVersionForInstalled(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		project   string
		installed string
		want      bool
	}{
		{name: "alpha5 on alpha10", project: "0.1.0-alpha.5", installed: "0.1.0-alpha.10", want: true},
		{name: "alpha8 on alpha9", project: "0.1.0-alpha.8", installed: "0.1.0-alpha.9", want: true},
		{name: "alpha8 on alpha7", project: "0.1.0-alpha.8", installed: "0.1.0-alpha.7", want: false},
		{name: "alpha4 below supported", project: "0.1.0-alpha.4", installed: "0.1.0-alpha.10", want: false},
		{name: "alpha9 above supported", project: "0.1.0-alpha.9", installed: "0.1.0-alpha.10", want: false},
		{name: "non alpha installed", project: "0.1.0-alpha.8", installed: "0.1.0", want: false},
		{name: "non alpha project", project: "1.2.3", installed: "0.1.0-alpha.10", want: false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isCompatibleProjectVersionForInstalled(tc.project, tc.installed); got != tc.want {
				t.Fatalf("isCompatibleProjectVersionForInstalled(%q, %q) = %t, want %t", tc.project, tc.installed, got, tc.want)
			}
		})
	}
}

func writeEmbeddedProjectVersion(t *testing.T, root, version string) {
	t.Helper()
	config := "schema_version: 1\nrunecontext_version: " + version + "\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir content root: %v", err)
	}
}

func TestRunUpgradePreviewSupportsAlphaFiveWithoutRegisteredEdge(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.9")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.5")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "current"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["target_version"], "0.1.0-alpha.5"; got != want {
		t.Fatalf("expected target_version %q, got %q", want, got)
	}
	if got, want := fields["plan_action_1"], "no changes required"; got != want {
		t.Fatalf("expected no-op plan action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewAlphaFiveTargetAlphaNineVersionBumpOnly(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.9")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.5")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "0"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["plan_action_1"], "set runecontext_version to 0.1.0-alpha.9"; got != want {
		t.Fatalf("expected version-bump-only plan action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewDirectEdgeIncludesSingleHopContract(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.9")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["hop_count"], "1"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["hop_1_from"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected hop_1_from %q, got %q", want, got)
	}
	if got, want := fields["hop_1_to"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected hop_1_to %q, got %q", want, got)
	}
	if got, want := fields["hop_action_1"], "migrate runecontext_version 0.1.0-alpha.8 -> 0.1.0-alpha.9"; got != want {
		t.Fatalf("expected hop_action_1 %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewMultiHopUsesDeterministicOrderedPath(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "1"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["hop_1_from"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected hop_1_from %q, got %q", want, got)
	}
	if got, want := fields["hop_1_to"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected hop_1_to %q, got %q", want, got)
	}
	if got, want := fields["plan_action_2"], "set runecontext_version to 0.1.0-alpha.10"; got != want {
		t.Fatalf("expected final version rewrite plan action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewAliasTargetLatestUsesPlannedHops(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "latest"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["target_version"], "0.1.0-alpha.10"; got != want {
		t.Fatalf("expected target_version %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "true"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "1"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewCurrentVersionNoOpReportsZeroHops(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.10")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "current"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "0"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["plan_action_1"], "no changes required"; got != want {
		t.Fatalf("expected no-op action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewAlphaFiveToAlphaTenAllowsVersionBumpWithRequiredHopsOnly(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.5")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "0"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["plan_action_1"], "set runecontext_version to 0.1.0-alpha.10"; got != want {
		t.Fatalf("expected version-bump-only plan action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewAlphaNineTargetAlphaTenVersionBumpOnly(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["hop_count"], "0"; got != want {
		t.Fatalf("expected hop_count %q, got %q", want, got)
	}
	if got, want := fields["plan_action_1"], "set runecontext_version to 0.1.0-alpha.10"; got != want {
		t.Fatalf("expected version rewrite action %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewFailsClosedWhenProjectIsNewerThanCLI(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.9")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.10")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "project_newer_than_cli"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got := fields["next_action_1"]; !strings.Contains(got, "upgrade runectx") {
		t.Fatalf("expected next action to guide CLI upgrade, got %q", got)
	}
}

func TestRunUpgradePreviewStableProjectNewerThanCLI(t *testing.T) {
	setRunecontextVersionForTests(t, "v1.2.3")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "1.2.4")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "project_newer_than_cli"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got := fields["next_action_1"]; !strings.Contains(got, "upgrade runectx") {
		t.Fatalf("expected CLI upgrade guidance, got %q", got)
	}
}

func TestRunUpgradePreviewStableProjectFromDifferentMinorIsUnsupported(t *testing.T) {
	setRunecontextVersionForTests(t, "v1.2.3")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "1.3.0")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "unsupported_project_version"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewRejectsDowngradeTarget(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.8"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "project_newer_than_cli"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got := fields["next_action_1"]; !strings.Contains(got, "upgrade runectx") {
		t.Fatalf("expected CLI upgrade guidance, got %q", got)
	}
}

func TestCompareKnownRunecontextVersions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		left       string
		right      string
		wantCmp    int
		wantParsed bool
	}{
		{name: "equal", left: "0.1.0-alpha.9", right: "0.1.0-alpha.9", wantCmp: 0, wantParsed: true},
		{name: "older", left: "0.1.0-alpha.8", right: "0.1.0-alpha.9", wantCmp: -1, wantParsed: true},
		{name: "newer", left: "0.1.0-alpha.10", right: "0.1.0-alpha.9", wantCmp: 1, wantParsed: true},
		{name: "stable newer", left: "1.2.4", right: "1.2.3", wantCmp: 1, wantParsed: true},
		{name: "stable older", left: "1.2.3", right: "1.2.4", wantCmp: -1, wantParsed: true},
		{name: "stable equal", left: "1.2.3", right: "1.2.3", wantCmp: 0, wantParsed: true},
		{name: "pre-release lower than stable", left: "1.2.3-alpha.1", right: "1.2.3", wantCmp: -1, wantParsed: true},
		{name: "pre-release numeric ordering", left: "1.2.3-alpha.2", right: "1.2.3-alpha.10", wantCmp: -1, wantParsed: true},
		{name: "unknown", left: "not-a-version", right: "0.1.0-alpha.9", wantCmp: 0, wantParsed: false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotCmp, gotParsed := compareKnownRunecontextVersions(tc.left, tc.right)
			if gotCmp != tc.wantCmp || gotParsed != tc.wantParsed {
				t.Fatalf("compareKnownRunecontextVersions(%q, %q) = (%d, %t), want (%d, %t)", tc.left, tc.right, gotCmp, gotParsed, tc.wantCmp, tc.wantParsed)
			}
		})
	}
}

func TestIsComparableVersionLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		current   string
		installed string
		want      bool
	}{
		{name: "same alpha line", current: "0.1.0-alpha.11", installed: "0.1.0-alpha.10", want: true},
		{name: "same stable line", current: "1.2.4", installed: "1.2.3", want: true},
		{name: "different minor", current: "1.3.0", installed: "1.2.3", want: false},
		{name: "different major", current: "2.0.0", installed: "1.2.3", want: false},
		{name: "invalid current", current: "invalid", installed: "1.2.3", want: false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isComparableVersionLine(tc.current, tc.installed); got != tc.want {
				t.Fatalf("isComparableVersionLine(%q, %q) = %t, want %t", tc.current, tc.installed, got, tc.want)
			}
		})
	}
}

func TestRunUpgradeApplyAlphaNineToAlphaTenRewrites(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["current_version"], "0.1.0-alpha.10"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
	if got, want := fields["changed"], "true"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}
}
