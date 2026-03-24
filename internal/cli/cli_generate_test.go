package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateUsageMissingSubcommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage error output, got %q", stderr.String())
	}
}

func TestRunGenerateHelpTokens(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "--help"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected help exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage=runectx generate") {
		t.Fatalf("expected generate usage output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunGenerateHelpRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "--help", "extra"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "help does not accept additional arguments") {
		t.Fatalf("expected help extra-arg error, got %q", stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
}

func TestRunGenerateIndexesWritesArtifacts(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "traceability", "valid-project"), projectRoot)
	contentRoot := filepath.Join(projectRoot, "runecontext")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "indexes", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], generateIndexesCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["changed_file_count"], "3"; got != want {
		t.Fatalf("expected changed_file_count %q, got %q", want, got)
	}
	for _, path := range []string{"manifest.yaml", "indexes/changes-by-status.yaml", "indexes/bundles.yaml"} {
		if _, err := os.Stat(filepath.Join(contentRoot, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected generated artifact %s: %v", path, err)
		}
	}
}

func TestRunGenerateIndexesExplainIncludesScope(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "traceability", "valid-project"), projectRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "indexes", "--path", projectRoot, "--explain"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["explain_scope"], "generated-indexes"; got != want {
		t.Fatalf("expected explain_scope %q, got %q", want, got)
	}
}

func TestRunGenerateIndexesRepairsMalformedGeneratedArtifacts(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "traceability", "valid-project"), projectRoot)
	indexPath := filepath.Join(projectRoot, "runecontext", "indexes", "changes-by-status.yaml")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatalf("mkdir indexes path: %v", err)
	}
	if err := os.WriteFile(indexPath, []byte("schema_version: 1\nstatuses: not-an-object\n"), 0o644); err != nil {
		t.Fatalf("write malformed generated index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "indexes", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	code = Run([]string{"validate", projectRoot}, &validateOut, &validateErr)
	if code != exitOK {
		t.Fatalf("expected validate success after regeneration, got %d (%s)", code, validateErr.String())
	}
}

func TestRunGenerateIndexesRejectsPathConflict(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "traceability", "valid-project"), projectRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"generate", "indexes", "--path", projectRoot, projectRoot}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot use both --path and a positional path argument") {
		t.Fatalf("expected --path conflict output, got %q", stderr.String())
	}
}
