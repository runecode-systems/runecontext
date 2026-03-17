package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidateSuccess(t *testing.T) {
	root := fixtureRoot(t, "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stdout.String() == "" {
		t.Fatalf("expected success output, got empty stdout")
	}
	if !strings.Contains(stdout.String(), "result=ok") {
		t.Fatalf("expected success result line, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "command=validate") {
		t.Fatalf("expected command line, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "root=") {
		t.Fatalf("expected success output, got %q", stdout.String())
	}
}

func TestRunValidateFailure(t *testing.T) {
	root := fixtureRoot(t, "reject-change-missing-related-spec")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=invalid") {
		t.Fatalf("expected invalid result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_path=") {
		t.Fatalf("expected error path output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=") {
		t.Fatalf("expected validation failure output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsInvalidProposal(t *testing.T) {
	root := fixtureRoot(t, "reject-proposal-invalid")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_path=") || !strings.Contains(stderr.String(), "proposal.md") {
		t.Fatalf("expected proposal path in output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsInvalidBundle(t *testing.T) {
	root := fixtureRoot(t, "reject-bundle-invalid")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_path=") || !strings.Contains(stderr.String(), "bundles") {
		t.Fatalf("expected bundle path in output, got %q", stderr.String())
	}
}

func TestRunValidateUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "a", "b"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage=runectx validate [path]") {
		t.Fatalf("expected usage output, got %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=unknown command") {
		t.Fatalf("expected unknown command output, got %q", stderr.String())
	}
}

func TestRunNoCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=missing command") {
		t.Fatalf("expected missing command output, got %q", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "fixtures", "traceability", name)
}
