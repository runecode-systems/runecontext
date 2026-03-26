package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpgradePreviewOnReferenceFixture(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["phase"], "preview"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["apply_required"], "true"; got != want {
		t.Fatalf("expected apply_required %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyRewritesTargetVersion(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["phase"], "apply"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["changed"], "true"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"validate", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected upgraded project to validate, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "result=ok") {
		t.Fatalf("expected validate success output, got %q", stdout.String())
	}
}

func TestRunUpgradeApplyRequiresTargetVersion(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_message=upgrade apply requires --target-version") {
		t.Fatalf("expected missing target version error, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyAcceptsSemverWithPrereleaseAndBuildMetadata(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "1.2.3-rc.1+build.5"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["target_version"], "1.2.3-rc.1+build.5"; got != want {
		t.Fatalf("expected target_version %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyHandlesValidYAMLSpacing(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\nrunecontext_version : 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(configPath, []byte(config), 0o640); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.Chmod(configPath, 0o640); err != nil {
		t.Fatalf("chmod config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if !strings.Contains(string(updated), "runecontext_version: 0.1.0-alpha.9") {
		t.Fatalf("expected updated runecontext version, got %q", string(updated))
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat updated config: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("expected config mode %o, got %o", want, got)
	}
}
