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
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.9"

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

func TestRunUpgradePreviewAlphaFiveTargetAlphaNineStillRejected(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.9"

	root := t.TempDir()
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.5")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "unsupported_project_version"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got := fields["plan_action_1"]; !strings.Contains(got, "no registered upgrader edge") {
		t.Fatalf("expected missing-edge plan action, got %q", got)
	}
}

func TestRunUpgradePreviewAlphaNineTargetAlphaTenSupportedEdge(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.10"

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
	if got, want := fields["plan_action_1"], "set runecontext_version to 0.1.0-alpha.10"; got != want {
		t.Fatalf("expected transition plan action %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyAlphaNineToAlphaTenRewrites(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.10"

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
