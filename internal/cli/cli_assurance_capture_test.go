package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunAssuranceCaptureRequiresVerifiedTier(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "capture", "context-pack", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "assurance_tier must be verified") {
		t.Fatalf("expected verified-tier error, got %q", stderr.String())
	}
}

func TestRunAssuranceCaptureContextPackWritesReceipt(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	runAssuranceEnableVerified(t, projectRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "capture", "context-pack", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected capture success, got %d (%s)", code, stderr.String())
	}
	receiptPath := assertCaptureOutput(t, stdout.String())
	receiptAbsolute := filepath.Join(projectRoot, filepath.FromSlash(receiptPath))
	if _, err := os.Stat(receiptAbsolute); err != nil {
		t.Fatalf("expected receipt file at %s: %v", receiptAbsolute, err)
	}
	assertProjectValidAfterCapture(t, projectRoot)
}

func runAssuranceEnableVerified(t *testing.T, projectRoot string) {
	t.Helper()

	var enableOut bytes.Buffer
	var enableErr bytes.Buffer
	if code := Run([]string{"assurance", "enable", "verified", "--path", projectRoot}, &enableOut, &enableErr); code != exitOK {
		t.Fatalf("assurance enable failed: %d (%s)", code, enableErr.String())
	}
}

func assertCaptureOutput(t *testing.T, output string) string {
	t.Helper()

	fields := parseCLIKeyValueOutput(t, output)
	if got, want := fields["command"], "assurance capture"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	receiptPath := fields["receipt_path"]
	if receiptPath == "" {
		t.Fatalf("expected receipt_path in output, got %#v", fields)
	}
	if got := fields["changed_file_count"]; got != "1" {
		t.Fatalf("expected one changed file, got %q", got)
	}
	if got := fields["changed_file_1_action"]; got != "created" {
		t.Fatalf("expected created mutation, got %q", got)
	}
	return receiptPath
}

func assertProjectValidAfterCapture(t *testing.T, projectRoot string) {
	t.Helper()

	validator := contracts.NewValidator(schemaRoot(t))
	loaded, err := validator.LoadProject(projectRoot, contracts.ResolveOptions{ConfigDiscovery: contracts.ConfigDiscoveryExplicitRoot, ExecutionMode: contracts.ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project after capture: %v", err)
	}
	defer loaded.Close()
	if _, err := validator.ValidateLoadedProject(loaded); err != nil {
		t.Fatalf("validate project after capture: %v", err)
	}
}

func schemaRoot(t *testing.T) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "schemas")
}
