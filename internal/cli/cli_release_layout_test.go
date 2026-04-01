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
	if strings.Contains(string(data), "\"adapters\"") {
		t.Fatalf("expected release metadata to stop including adapters top-level directory")
	}
}

func TestLocateAdaptersRootFromReleaseStyleLayout(t *testing.T) {
	root := t.TempDir()
	schemaDir := filepath.Join(root, "schemas")
	repoAdaptersDir := filepath.Join(root, "adapters")
	stagedAdaptersDir := filepath.Join(root, "build", "generated", "adapters")
	seedReleaseStyleLayout(t, schemaDir, repoAdaptersDir)
	if err := os.MkdirAll(stagedAdaptersDir, 0o755); err != nil {
		t.Fatalf("mkdir staged adapters dir: %v", err)
	}
	seedAdapterPackForDiscovery(t, stagedAdaptersDir, "opencode")

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
	expectedCanonical, err := filepath.EvalSymlinks(stagedAdaptersDir)
	if err != nil {
		t.Fatalf("resolve expected adapters root symlinks: %v", err)
	}
	if gotCanonical != expectedCanonical {
		t.Fatalf("expected adapters root %q, got %q", expectedCanonical, gotCanonical)
	}
}

func seedAdapterPackForDiscovery(t *testing.T, adaptersDir, tool string) {
	t.Helper()
	toolDir := filepath.Join(adaptersDir, tool)
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir adapter pack dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "workflow.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write workflow contract: %v", err)
	}
}

func TestLocateAdaptersRootRejectsRepositoryAdaptersFallback(t *testing.T) {
	root := t.TempDir()
	schemaDir := filepath.Join(root, "schemas")
	repoAdaptersDir := filepath.Join(root, "adapters")
	seedReleaseStyleLayout(t, schemaDir, repoAdaptersDir)
	seedAdapterPackForDiscovery(t, repoAdaptersDir, "opencode")

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir test root: %v", err)
	}

	if err := os.Remove(filepath.Join(repoAdaptersDir, "opencode", "workflow.json")); err != nil {
		t.Fatalf("remove fallback workflow contract: %v", err)
	}

	if _, err := locateAdaptersRoot(); err == nil {
		t.Fatal("expected adapters root discovery to reject adapter roots without generated workflow contracts")
	}
}

func TestLocateSchemaRootFromInstalledShareLayout(t *testing.T) {
	installRoot := t.TempDir()
	schemaDir := filepath.Join(installRoot, "share", "runecontext", "schemas")
	adaptersDir := filepath.Join(installRoot, "share", "runecontext", "adapters")
	seedReleaseStyleLayout(t, schemaDir, adaptersDir)

	outsideRoot := t.TempDir()
	got, err := locateSchemaRootWithDeps(schemaRootDeps{
		getwd: func() (string, error) { return outsideRoot, nil },
		executable: func() (string, error) {
			return filepath.Join(installRoot, "bin", "runectx"), nil
		},
	})
	if err != nil {
		t.Fatalf("locate schema root: %v", err)
	}
	gotCanonical, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("resolve located schema root symlinks: %v", err)
	}
	expectedCanonical, err := filepath.EvalSymlinks(schemaDir)
	if err != nil {
		t.Fatalf("resolve expected schema root symlinks: %v", err)
	}
	if gotCanonical != expectedCanonical {
		t.Fatalf("expected schema root %q, got %q", expectedCanonical, gotCanonical)
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
		"capability-descriptor.schema.json",
		"change-status.schema.json",
		"context-pack.schema.json",
		"assurance-baseline.schema.json",
		"assurance-receipt.schema.json",
		"assurance-imported-history.schema.json",
	}
}
