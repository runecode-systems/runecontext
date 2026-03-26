package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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
	if got, want := fields["previous_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected previous_version %q, got %q", want, got)
	}
	if got, want := fields["current_version"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
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

func TestRunUpgradeApplyNoOpUsesStableOutputFields(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.8"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["previous_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected previous_version %q, got %q", want, got)
	}
	if got, want := fields["current_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
	if got, want := fields["target_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected target_version %q, got %q", want, got)
	}
	if got, want := fields["changed"], "false"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
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
	if !strings.Contains(string(updated), "runecontext_version : 0.1.0-alpha.9") {
		t.Fatalf("expected updated runecontext version, got %q", string(updated))
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat updated config: %v", err)
	}
	if runtime.GOOS != "windows" {
		if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
			t.Fatalf("expected config mode %o, got %o", want, got)
		}
	}
}

func TestRunUpgradeApplyPreservesCommentsAndCRLF(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\r\nrunecontext_version: 0.1.0-alpha.8 # pinned here\r\nassurance_tier: plain\r\nsource:\r\n  type: embedded\r\n  path: runecontext\r\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
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
	content := string(updated)
	if !strings.Contains(content, "runecontext_version: 0.1.0-alpha.9 # pinned here") {
		t.Fatalf("expected preserved inline comment, got %q", content)
	}
	if !strings.Contains(content, "\r\n") {
		t.Fatalf("expected CRLF line endings to remain, got %q", content)
	}
	if !strings.HasSuffix(content, "\r\n") {
		t.Fatalf("expected trailing CRLF newline to remain, got %q", content)
	}
}

func TestRunUpgradeApplyPreservesTrailingLFNewline(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
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
	content := string(updated)
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("expected trailing LF newline to remain, got %q", content)
	}
}
