package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpgradeApplyTransactionalRollbackOnFailure(t *testing.T) {
	root := setupUpgradeStaleOpencodeTree(t, "v0.1.0-alpha.10")
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}
	originalSync := upgradeApplyAdapterSyncFn
	t.Cleanup(func() {
		upgradeApplyAdapterSyncFn = originalSync
	})
	upgradeApplyAdapterSyncFn = func(state adapterSyncState) error {
		return fmt.Errorf("forced upgrade transaction failure for %s", state.tool)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after failed apply: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected config rollback to restore original content")
	}
}

func TestRunUpgradeApplyRevalidatesAdapterConflictsInStage(t *testing.T) {
	root := setupUpgradeStaleOpencodeTree(t, "v0.1.0-alpha.10")
	probePath := filepath.Join(root, ".opencode", "skills", "runecontext-change-new.md")
	before, err := os.ReadFile(probePath)
	if err != nil {
		t.Fatalf("read managed probe file: %v", err)
	}
	installTestOnlyUpgradeHop(t, upgradeEdgeKey{From: "0.1.0-alpha.9", To: "0.1.0-alpha.10"}, testUpgradeHopMigration{
		applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
			conflictPath := filepath.Join(ctx.Root, ".opencode", "skills", "runecontext-change-new.md")
			return os.WriteFile(conflictPath, []byte("user owned conflict\n"), 0o644)
		},
		verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error { return nil },
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10", "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected staged conflict rejection, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "staged upgrade adapter conflicts detected") {
		t.Fatalf("expected staged conflict revalidation error, got %q", stderr.String())
	}
	data, err := os.ReadFile(probePath)
	if err != nil {
		t.Fatalf("read live managed probe file after failed apply: %v", err)
	}
	if string(data) != string(before) {
		t.Fatalf("expected conflicted staged write to be rolled back from live tree, err=%v", err)
	}
}

func setupUpgradeStaleOpencodeTree(t *testing.T, installedVersion string) string {
	t.Helper()
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, installedVersion)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")
	if code := Run([]string{"adapter", "sync", "--path", root, "opencode"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("expected adapter sync success")
	}
	staleRel := filepath.Join(root, ".opencode", "skills", "runecontext-stale-merge.md")
	if err := os.MkdirAll(filepath.Dir(staleRel), 0o755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:stale-merge -->\n"
	if err := os.WriteFile(staleRel, []byte(managed), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}
	return root
}
