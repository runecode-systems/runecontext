package contracts

import (
	"errors"
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

func TestValidateLoadedProjectRejectsGeneratedIndexEscapeViaSymlink(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	copyDirForTest(t, fixturePath(t, "traceability", "valid-project"), projectRoot)
	contentRoot := filepath.Join(projectRoot, "runecontext")
	outsideDir := t.TempDir()
	outsideManifest := filepath.Join(outsideDir, "manifest.yaml")
	if err := os.WriteFile(outsideManifest, []byte("schema_version: 1\nentries: []\n"), 0o644); err != nil {
		t.Fatalf("write outside manifest: %v", err)
	}
	if err := os.Symlink(outsideManifest, filepath.Join(contentRoot, "manifest.yaml")); err != nil {
		if errors.Is(err, os.ErrPermission) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create manifest symlink: %v", err)
	}

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load project: %v", err)
	}
	defer loaded.Close()

	_, err = v.ValidateLoadedProject(loaded)
	if err == nil {
		t.Fatal("expected generated index escape validation error")
	}
	if !strings.Contains(err.Error(), "escapes the selected project subtree") {
		t.Fatalf("expected subtree escape error, got %v", err)
	}
}
