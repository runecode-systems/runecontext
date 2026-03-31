package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpgradePreviewRejectsSymlinkedManagedScanTarget(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), projectRoot)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	if code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("expected adapter sync success")
	}
	outside := filepath.Join(projectRoot, "outside-managed.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	managedPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	if err := os.Remove(managedPath); err != nil {
		t.Fatalf("remove managed file: %v", err)
	}
	if err := os.Symlink(outside, managedPath); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create managed scan symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", projectRoot, "--target-version", "current", "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "managed host-native scan rejects symlinked path") {
		t.Fatalf("expected managed scan symlink rejection, got %q", stderr.String())
	}
}
