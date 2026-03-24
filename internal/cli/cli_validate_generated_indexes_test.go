package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidateFailsOnMalformedGeneratedIndexesWhenPresent(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "traceability", "valid-project"), projectRoot)
	indexPath := filepath.Join(projectRoot, "runecontext", "indexes", "changes-by-status.yaml")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatalf("mkdir indexes path: %v", err)
	}
	if err := os.WriteFile(indexPath, []byte("schema_version: 1\nkind: changes_by_status\nchanges: not-an-array\n"), 0o644); err != nil {
		t.Fatalf("write malformed generated index: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "changes-by-status.yaml") {
		t.Fatalf("expected generated index path in error output, got %q", stderr.String())
	}
}
