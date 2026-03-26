package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseMetadataIncludesAdaptersDirectory(t *testing.T) {
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	metadataPath := filepath.Join(root, "nix", "release", "metadata.nix")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("read release metadata: %v", err)
	}
	if !strings.Contains(string(data), "\"adapters\"") {
		t.Fatalf("expected release metadata to include adapters top-level directory")
	}
}

func TestLocateAdaptersRootFromReleaseStyleLayout(t *testing.T) {
	root := t.TempDir()
	schemaDir := filepath.Join(root, "schemas")
	adaptersDir := filepath.Join(root, "adapters")
	seedReleaseStyleLayout(t, schemaDir, adaptersDir)

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir test root: %v", err)
	}

	got, err := locateAdaptersRoot()
	if err != nil {
		t.Fatalf("locate adapters root: %v", err)
	}
	gotCanonical, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("resolve located adapters root symlinks: %v", err)
	}
	expectedCanonical, err := filepath.EvalSymlinks(adaptersDir)
	if err != nil {
		t.Fatalf("resolve expected adapters root symlinks: %v", err)
	}
	if gotCanonical != expectedCanonical {
		t.Fatalf("expected adapters root %q, got %q", expectedCanonical, gotCanonical)
	}
}

func seedReleaseStyleLayout(t *testing.T, schemaDir, adaptersDir string) {
	t.Helper()
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		t.Fatalf("mkdir schema dir: %v", err)
	}
	if err := os.MkdirAll(adaptersDir, 0o755); err != nil {
		t.Fatalf("mkdir adapters dir: %v", err)
	}
	for _, name := range requiredSchemaNames() {
		if err := os.WriteFile(filepath.Join(schemaDir, name), []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("write schema %s: %v", name, err)
		}
	}
}

func requiredSchemaNames() []string {
	return []string{
		"runecontext.schema.json",
		"bundle.schema.json",
		"change-status.schema.json",
		"context-pack.schema.json",
		"assurance-baseline.schema.json",
		"assurance-receipt.schema.json",
		"assurance-imported-history.schema.json",
	}
}
