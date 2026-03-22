package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitDryRun(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "project")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root, "--dry-run"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected exit OK, got %d/stderr=%q", code, stderr.String())
	}

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("expected project root to remain absent on dry-run, found %v", err)
	}

	fields := parseCLIKeyValueOutput(t, stdout.String())
	if fields["command"] != "init" {
		t.Fatalf("expected init command output, got %#v", fields)
	}
	if strings.TrimSpace(fields["plan_action_1"]) == "" {
		t.Fatalf("expected plan_action entries, got %#v", fields)
	}
	if fields["plan_action_1"] != "ensure directory "+absRoot {
		t.Fatalf("unexpected plan action, got %q", fields["plan_action_1"])
	}
}

func TestRunInitCreatesEmbeddedProject(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "embedded-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d/stderr=%q", code, stderr.String())
	}

	configPath := filepath.Join(root, "runecontext.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "source:\n  type: embedded") {
		t.Fatalf("expected embedded source, got %s", string(data))
	}
	for _, sub := range []string{"bundles", "changes"} {
		if stat, err := os.Stat(filepath.Join(root, sub)); err != nil || !stat.IsDir() {
			t.Fatalf("expected %s directory, got %v %v", sub, stat, err)
		}
	}
}

func TestRunInitLinkedModeAndSeedBundle(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "linked-project")
	bundleName := "base"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root, "--mode", "linked", "--seed-bundle", bundleName}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d/stderr=%q", code, stderr.String())
	}

	configData, err := os.ReadFile(filepath.Join(root, "runecontext.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(configData), "source:\n  type: path") {
		t.Fatalf("expected linked path source, got %s", string(configData))
	}

	bundlePath := filepath.Join(root, "bundles", bundleName+".yaml")
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	if !strings.Contains(string(bundleData), "includes: {}") {
		t.Fatalf("expected includes map, got %s", string(bundleData))
	}
}

func TestRunInitMachineOptionsReported(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "machine-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root, "--non-interactive"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d/stderr=%q", code, stderr.String())
	}

	fields := parseCLIKeyValueOutput(t, stdout.String())
	if fields["non_interactive"] != "true" {
		t.Fatalf("expected non_interactive true, got %#v", fields)
	}
	if fields["dry_run"] != "false" {
		t.Fatalf("expected dry_run false, got %#v", fields)
	}
}

func TestRunInitRejectsInvalidSeedBundleName(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "bad-seed")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root, "--seed-bundle", ".."}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage error exit code, got %d/stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--seed-bundle name must not contain path separators") {
		t.Fatalf("expected seed bundle validation, got %q", stderr.String())
	}
}

func TestRunInitJSONEnvelope(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "json-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"init", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success code, got %d/stderr=%q", code, stderr.String())
	}

	var envelope machineEnvelope
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &envelope); err != nil {
		t.Fatalf("unmarshal json output: %v", err)
	}
	if envelope.Command != "init" {
		t.Fatalf("expected init command envelope, got %+v", envelope)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}
	if envelope.Data["root"] != absRoot {
		t.Fatalf("unexpected root in json output: %s", envelope.Data["root"])
	}
}
