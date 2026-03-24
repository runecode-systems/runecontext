package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLoadedProjectRejectsMalformedGeneratedIndexesWhenPresent(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := copyTraceabilityFixtureProject(t, "valid-project")
	malformedPath := filepath.Join(projectRoot, "runecontext", "indexes", "changes-by-status.yaml")
	if err := os.MkdirAll(filepath.Dir(malformedPath), 0o755); err != nil {
		t.Fatalf("mkdir indexes path: %v", err)
	}
	if err := os.WriteFile(malformedPath, []byte("schema_version: 1\nkind: changes_by_status\nchanges: not-an-array\n"), 0o644); err != nil {
		t.Fatalf("write malformed generated index: %v", err)
	}

	_, err := v.ValidateProject(projectRoot)
	if err == nil || !strings.Contains(err.Error(), "changes-by-status.yaml") {
		t.Fatalf("expected malformed generated index validation failure, got %v", err)
	}
}
