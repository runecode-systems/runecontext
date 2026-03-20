package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteGeneratedIndexesWritesStandardPaths(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate copied fixture project: %v", err)
	}
	defer index.Close()

	if err := index.WriteGeneratedIndexes(); err != nil {
		t.Fatalf("write generated indexes: %v", err)
	}

	paths := []struct {
		path   string
		schema string
	}{
		{path: filepath.Join(index.ContentRoot, "manifest.yaml"), schema: "manifest.schema.json"},
		{path: filepath.Join(index.ContentRoot, "indexes", "changes-by-status.yaml"), schema: "changes-by-status-index.schema.json"},
		{path: filepath.Join(index.ContentRoot, "indexes", "bundles.yaml"), schema: "bundles-index.schema.json"},
	}
	for _, item := range paths {
		data, err := os.ReadFile(item.path)
		if err != nil {
			t.Fatalf("read generated file %s: %v", item.path, err)
		}
		if err := v.ValidateYAMLFile(item.schema, item.path, data); err != nil {
			t.Fatalf("expected generated file to satisfy %s: %v", item.schema, err)
		}
	}
}

func TestGeneratedRelativeArtifactPathRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.yaml")
	_, err := generatedRelativeArtifactPath(root, outside)
	if err == nil || !strings.Contains(err.Error(), "escapes RuneContext content root") {
		t.Fatalf("expected escape rejection, got %v", err)
	}
}

func TestGeneratedRelativeArtifactPathReturnsSlashCanonicalRelativePaths(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "changes", "CHG-2026-125-a1b2-canonical", "status.yaml")
	rel, err := generatedRelativeArtifactPath(root, target)
	if err != nil {
		t.Fatalf("generated relative path: %v", err)
	}
	want := "changes/CHG-2026-125-a1b2-canonical/status.yaml"
	if rel != want {
		t.Fatalf("expected %s, got %s", want, rel)
	}
}
