package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidateWarnsWhenStatusUsesExtensionsWithOptIn(t *testing.T) {
	projectRoot := prepareValidateExtensionsProject(t)
	appendExtensionSnippet(t, filepath.Join(projectRoot, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml"), "\nextensions:\n  dev.example.flag: true\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	assertOutputContains(t, stdout.String(), "diagnostic_count=1")
	assertOutputContains(t, stdout.String(), "diagnostic_1_code=extensions_present")
	assertOutputContains(t, stdout.String(), "diagnostic_1_severity=warning")
	if !strings.Contains(stdout.String(), "diagnostic_1_path=") || !strings.Contains(stdout.String(), "status.yaml") {
		t.Fatalf("expected warning path metadata, got %q", stdout.String())
	}
}

func TestRunValidateWarnsWhenBundleUsesExtensionsWithOptIn(t *testing.T) {
	projectRoot := prepareValidateExtensionsProject(t)
	appendExtensionSnippet(t, filepath.Join(projectRoot, "runecontext", "bundles", "auth-review.yaml"), "\nextensions:\n  dev.example.bundle:\n    owner: qa\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	assertOutputContains(t, stdout.String(), "diagnostic_count=1")
	assertOutputContains(t, stdout.String(), "diagnostic_1_code=extensions_present")
	assertOutputContains(t, stdout.String(), "diagnostic_1_bundle=auth-review")
	assertOutputContains(t, stdout.String(), "diagnostic_1_severity=warning")
}

func TestRunValidateExtensionsWarningsAppearInJSONOutput(t *testing.T) {
	projectRoot := prepareValidateExtensionsProject(t)
	appendExtensionSnippet(t, filepath.Join(projectRoot, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml"), "\nextensions:\n  dev.example.flag: true\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--json", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}

	var envelope struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal validate json output: %v", err)
	}
	if got := envelope.Data["diagnostic_count"]; got != "1" {
		t.Fatalf("expected diagnostic_count=1, got %q", got)
	}
	if got := envelope.Data["diagnostic_1_code"]; got != "extensions_present" {
		t.Fatalf("expected diagnostic_1_code=extensions_present, got %q", got)
	}
	if got := envelope.Data["diagnostic_1_severity"]; got != "warning" {
		t.Fatalf("expected diagnostic_1_severity=warning, got %q", got)
	}
}

func prepareValidateExtensionsProject(t *testing.T) string {
	t.Helper()
	projectRoot := t.TempDir()
	copyDirForCLI(t, filepath.Join(repoFixtureRoot(t, "traceability"), "valid-project"), projectRoot)
	configPath := filepath.Join(projectRoot, "runecontext.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read root config: %v", err)
	}
	updatedConfig := strings.TrimSpace(string(configData)) + "\nallow_extensions: true\n"
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	return projectRoot
}

func appendExtensionSnippet(t *testing.T, path string, snippet string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read extension target: %v", err)
	}
	updated := strings.TrimSpace(string(data)) + snippet
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("write extension target: %v", err)
	}
}

func assertOutputContains(t *testing.T, output string, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, output)
	}
}
