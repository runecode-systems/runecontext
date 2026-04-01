package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStandardListEmitsStructuredResults(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"standard", "list", "--path", projectRoot, "--scope-path", "security", "--focus", "review", "--status", "active"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], standardListCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["standard_count"], "1"; got != want {
		t.Fatalf("expected standard_count %q, got %q", want, got)
	}
	if got, want := fields["standard_1_path"], "standards/security/review.md"; got != want {
		t.Fatalf("expected first standard path %q, got %q", want, got)
	}
}

func TestRunStandardCreateAndUpdateMutateSafely(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var createOut bytes.Buffer
	var createErr bytes.Buffer

	createCode := Run([]string{"standard", "create", "--project-path", projectRoot, "--path", "custom/authoring", "--title", "Authoring Standard", "--status", "draft"}, &createOut, &createErr)
	if createCode != exitOK {
		t.Fatalf("expected create success, got %d (%s)", createCode, createErr.String())
	}
	createFields := parseCLIKeyValueOutput(t, createOut.String())
	if got, want := createFields["command"], standardCreateCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := createFields["mutation_performed"], "true"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	createdPath := filepath.Join(projectRoot, "runecontext", "standards", "custom", "authoring.md")
	if _, err := os.Stat(createdPath); err != nil {
		t.Fatalf("expected created standard file, got %v", err)
	}

	var updateOut bytes.Buffer
	var updateErr bytes.Buffer
	updateCode := Run([]string{"standard", "update", "--project-path", projectRoot, "--path", "standards/custom/authoring.md", "--title", "Authoring Standard v2", "--status", "active", "--replace-aliases", "--alias", "custom/authoring-v1"}, &updateOut, &updateErr)
	if updateCode != exitOK {
		t.Fatalf("expected update success, got %d (%s)", updateCode, updateErr.String())
	}
	updateFields := parseCLIKeyValueOutput(t, updateOut.String())
	if got, want := updateFields["command"], standardUpdateCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	text := strings.ReplaceAll(string(mustReadBytesForCLI(t, createdPath)), "\r\n", "\n")
	if !strings.Contains(text, "title: Authoring Standard v2") {
		t.Fatalf("expected updated title in standard file, got:\n%s", text)
	}
	if !strings.Contains(text, "- custom/authoring-v1") {
		t.Fatalf("expected updated alias in standard file, got:\n%s", text)
	}
}

func TestRunStandardCreateDryRunDoesNotWrite(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"standard", "--dry-run", "create", "--project-path", projectRoot, "--path", "custom/dry-run", "--title", "Dry Run Standard"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected dry-run success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["dry_run"], "true"; got != want {
		t.Fatalf("expected dry_run %q, got %q", want, got)
	}
	dryRunPath := filepath.Join(projectRoot, "runecontext", "standards", "custom", "dry-run.md")
	if _, err := os.Stat(dryRunPath); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run standard file to be absent, got err=%v", err)
	}
}

func mustReadBytesForCLI(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
