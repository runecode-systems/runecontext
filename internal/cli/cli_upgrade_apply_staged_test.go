package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRunUpgradeApplyMultiHopSuccess(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["previous_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected previous_version %q, got %q", want, got)
	}
	if got, want := fields["current_version"], "0.1.0-alpha.10"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
	if got, want := fields["changed"], "true"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyRollbackOnPerHopValidationFailure(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}

	originalRegistry := upgradeApplyMigrationRegistryFn
	t.Cleanup(func() { upgradeApplyMigrationRegistryFn = originalRegistry })
	upgradeApplyMigrationRegistryFn = func() upgradeApplyMigrationRegistry {
		registry := defaultUpgradeApplyMigrationRegistry()
		registry.hopSpecific[upgradeEdgeKey{From: "0.1.0-alpha.8", To: "0.1.0-alpha.9"}] = testUpgradeHopMigration{
			applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
				return os.WriteFile(ctx.ConfigPath, []byte("schema_version: [broken\n"), 0o644)
			},
			verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error { return nil },
		}
		return registry
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after failed apply: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected failed per-hop validation to leave live config unchanged")
	}
}

func TestRunUpgradeApplyRollbackOnPerHopVerifyFailure(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}

	originalRegistry := upgradeApplyMigrationRegistryFn
	t.Cleanup(func() { upgradeApplyMigrationRegistryFn = originalRegistry })
	upgradeApplyMigrationRegistryFn = func() upgradeApplyMigrationRegistry {
		registry := defaultUpgradeApplyMigrationRegistry()
		registry.hopSpecific[upgradeEdgeKey{From: "0.1.0-alpha.8", To: "0.1.0-alpha.9"}] = testUpgradeHopMigration{
			applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
				return rewriteStageRunecontextVersion(ctx.ConfigPath, hop.To)
			},
			verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
				return fmt.Errorf("forced hop verify failure")
			},
		}
		return registry
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after failed apply: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected failed per-hop verify to leave live config unchanged")
	}
}

func TestRunUpgradeApplyRefreshesManagedArtifactsFromStagedFinalTree(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	originalRegistry := upgradeApplyMigrationRegistryFn
	t.Cleanup(func() { upgradeApplyMigrationRegistryFn = originalRegistry })
	upgradeApplyMigrationRegistryFn = func() upgradeApplyMigrationRegistry {
		registry := defaultUpgradeApplyMigrationRegistry()
		registry.hopSpecific[upgradeEdgeKey{From: "0.1.0-alpha.9", To: "0.1.0-alpha.10"}] = testUpgradeHopMigration{
			applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
				if err := rewriteStageRunecontextVersion(ctx.ConfigPath, hop.To); err != nil {
					return err
				}
				staleRel := ".opencode/skills/runecontext-stale-merge.md"
				staleAbs := filepath.Join(ctx.Root, filepath.FromSlash(staleRel))
				if err := os.MkdirAll(filepath.Dir(staleAbs), 0o755); err != nil {
					return err
				}
				managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:stale-merge -->\n"
				return os.WriteFile(staleAbs, []byte(managed), 0o644)
			},
			verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error { return nil },
		}
		return registry
	}

	staleAbs := filepath.Join(root, ".opencode", "skills", "runecontext-stale-merge.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(staleAbs); !os.IsNotExist(err) {
		t.Fatalf("expected staged-only stale managed host-native file to be removed during final refresh, err=%v", err)
	}
}

func TestStagedConfigPathRejectsOutsideRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	stageRoot := filepath.Join(t.TempDir(), "stage")
	outside := filepath.Join(filepath.Dir(root), "outside.yaml")

	_, err := stagedConfigPath(root, stageRoot, outside)
	if err == nil {
		t.Fatalf("expected stagedConfigPath to reject config path outside root")
	}
}

func TestApplyStageCommitReplacesDirectoryWithFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "foo"), 0o755); err != nil {
		t.Fatalf("mkdir live dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "foo", "bar.txt"), []byte("legacy"), 0o644); err != nil {
		t.Fatalf("write live file: %v", err)
	}

	stageRoot := filepath.Join(t.TempDir(), "stage")
	if err := os.MkdirAll(stageRoot, 0o755); err != nil {
		t.Fatalf("mkdir stage root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stageRoot, "foo"), []byte("replacement"), 0o644); err != nil {
		t.Fatalf("write stage replacement file: %v", err)
	}

	stage := stagedUpgradeTree{
		stageRoot:    stageRoot,
		changedFiles: []string{filepath.Join(root, "foo")},
		deletedFiles: []string{filepath.Join(root, "foo", "bar.txt")},
	}
	if err := applyStageCommit(root, stage); err != nil {
		t.Fatalf("applyStageCommit: %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "foo"))
	if err != nil {
		t.Fatalf("stat replacement path: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("expected foo to be a file after commit")
	}
}

func TestApplyStageCommitReplacesFileWithDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "foo"), []byte("legacy-file"), 0o644); err != nil {
		t.Fatalf("write live file: %v", err)
	}

	stageRoot := filepath.Join(t.TempDir(), "stage")
	if err := os.MkdirAll(filepath.Join(stageRoot, "foo"), 0o755); err != nil {
		t.Fatalf("mkdir stage dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stageRoot, "foo", "bar.txt"), []byte("replacement-child"), 0o644); err != nil {
		t.Fatalf("write stage child file: %v", err)
	}

	stage := stagedUpgradeTree{
		stageRoot:    stageRoot,
		changedFiles: []string{filepath.Join(root, "foo", "bar.txt")},
		deletedFiles: []string{filepath.Join(root, "foo")},
	}
	if err := applyStageCommit(root, stage); err != nil {
		t.Fatalf("applyStageCommit: %v", err)
	}

	info, err := os.Stat(filepath.Join(root, "foo"))
	if err != nil {
		t.Fatalf("stat replacement dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected foo to be a directory after commit")
	}
	if _, err := os.Stat(filepath.Join(root, "foo", "bar.txt")); err != nil {
		t.Fatalf("expected replacement child file: %v", err)
	}
}

type testUpgradeHopMigration struct {
	applyFn  func(ctx upgradeMigrationContext, hop upgradeHop) error
	verifyFn func(ctx upgradeMigrationContext, hop upgradeHop) error
}

func (m testUpgradeHopMigration) Apply(ctx upgradeMigrationContext, hop upgradeHop) error {
	if m.applyFn == nil {
		return nil
	}
	return m.applyFn(ctx, hop)
}

func (m testUpgradeHopMigration) Verify(ctx upgradeMigrationContext, hop upgradeHop) error {
	if m.verifyFn == nil {
		return nil
	}
	return m.verifyFn(ctx, hop)
}
