package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunUpgradeApplyMultiHopSuccess(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
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

func TestRunUpgradeApplyAdvancesStageToIntervalHopFromBeforeMigration(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.13")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.10")

	installIntervalPlannerAndMigrationForAdvance(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.13", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["current_version"], "0.1.0-alpha.13"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
}

func installIntervalPlannerAndMigrationForAdvance(t *testing.T) {
	t.Helper()
	originalRegistry := upgradeApplyMigrationRegistryFn
	originalPlannerRegistry := upgradePlannerRegistryFn
	t.Cleanup(func() {
		upgradeApplyMigrationRegistryFn = originalRegistry
		upgradePlannerRegistryFn = originalPlannerRegistry
	})

	upgradePlannerRegistryFn = func() upgradePlannerRegistry {
		registry := defaultUpgradePlannerRegistry()
		registry.registerEdge("0.1.0-alpha.12", "0.1.0-alpha.13")
		return registry
	}

	upgradeApplyMigrationRegistryFn = func() upgradeApplyMigrationRegistry {
		registry := defaultUpgradeApplyMigrationRegistry()
		registry.hopSpecific[upgradeEdgeKey{From: "0.1.0-alpha.12", To: "0.1.0-alpha.13"}] = testUpgradeHopMigration{
			applyFn:  verifyAdvanceBeforeMigrationApply,
			verifyFn: verifyMigrationResultVersion,
		}
		return registry
	}
}

func verifyAdvanceBeforeMigrationApply(ctx upgradeMigrationContext, hop upgradeHop) error {
	if got, want := readRunecontextVersionFromConfig(ctx.ConfigPath), "0.1.0-alpha.12"; got != want {
		return fmt.Errorf("expected staged version %s before migration apply, got %s", want, got)
	}
	return rewriteStageRunecontextVersion(ctx.ConfigPath, hop.To)
}

func verifyMigrationResultVersion(ctx upgradeMigrationContext, hop upgradeHop) error {
	if got, want := readRunecontextVersionFromConfig(ctx.ConfigPath), hop.To; got != want {
		return fmt.Errorf("expected staged version %s after migration, got %s", want, got)
	}
	return nil
}

func TestRunUpgradeApplyRollbackOnPerHopValidationFailure(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}

	installTestOnlyUpgradeHop(t, upgradeEdgeKey{From: "0.1.0-alpha.9", To: "0.1.0-alpha.10"}, testUpgradeHopMigration{
		applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
			return os.WriteFile(ctx.ConfigPath, []byte("schema_version: [broken\n"), 0o644)
		},
		verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error { return nil },
	})

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
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")
	configPath, before := readUpgradeConfigBeforeApply(t, root)

	installTestOnlyUpgradeHop(t, upgradeEdgeKey{From: "0.1.0-alpha.9", To: "0.1.0-alpha.10"}, testUpgradeHopMigration{
		applyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
			return rewriteStageRunecontextVersion(ctx.ConfigPath, hop.To)
		},
		verifyFn: func(ctx upgradeMigrationContext, hop upgradeHop) error {
			return fmt.Errorf("forced hop verify failure")
		},
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.10"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	after := readUpgradeConfigAfterApply(t, configPath)
	if string(before) != string(after) {
		t.Fatalf("expected failed per-hop verify to leave live config unchanged")
	}
}

func readUpgradeConfigBeforeApply(t *testing.T, root string) (string, []byte) {
	t.Helper()
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}
	return configPath, before
}

func readUpgradeConfigAfterApply(t *testing.T, configPath string) []byte {
	t.Helper()
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after failed apply: %v", err)
	}
	return after
}

func installTestOnlyUpgradeHop(t *testing.T, edge upgradeEdgeKey, migration testUpgradeHopMigration) {
	t.Helper()
	originalRegistry := upgradeApplyMigrationRegistryFn
	originalPlannerRegistry := upgradePlannerRegistryFn
	t.Cleanup(func() {
		upgradeApplyMigrationRegistryFn = originalRegistry
		upgradePlannerRegistryFn = originalPlannerRegistry
	})
	upgradePlannerRegistryFn = func() upgradePlannerRegistry {
		registry := defaultUpgradePlannerRegistry()
		registry.registerEdge(edge.From, edge.To)
		return registry
	}
	upgradeApplyMigrationRegistryFn = func() upgradeApplyMigrationRegistry {
		registry := defaultUpgradeApplyMigrationRegistry()
		registry.hopSpecific[edge] = migration
		return registry
	}
}

func TestRunUpgradeApplyRefreshesManagedArtifactsFromStagedFinalTree(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.8")

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

func TestRunUpgradeApplyMigratesAssuranceLayoutToCanonicalPath(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.13")
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.12")
	setVerifiedTierForUpgradeTest(t, root)
	legacyBackfill := "imported-git-history-abcdef1234567890abcdef1234567890abcdef12.json"
	writeLegacyAssuranceFixtureForMigration(t, root, legacyBackfill)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.13", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "assurance")); !os.IsNotExist(err) {
		t.Fatalf("expected legacy assurance root removed after migration, err=%v", err)
	}
	baselineData, err := os.ReadFile(filepath.Join(root, "runecontext", "assurance", "baseline.yaml"))
	if err != nil {
		t.Fatalf("read migrated baseline: %v", err)
	}
	if !strings.Contains(string(baselineData), "runecontext/assurance/backfill/") {
		t.Fatalf("expected imported_evidence path rewritten, got %q", string(baselineData))
	}
}

func TestRunUpgradeApplyRollsBackOnAssuranceLayoutConflict(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.13")
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.12")
	setVerifiedTierForUpgradeTest(t, root)
	writeLegacyAssuranceFixtureForMigration(t, root, "imported-git-history-abcdef1234567890abcdef1234567890abcdef12.json")
	canonicalBaseline := filepath.Join(root, "runecontext", "assurance", "baseline.yaml")
	if err := os.MkdirAll(filepath.Dir(canonicalBaseline), 0o755); err != nil {
		t.Fatalf("mkdir canonical assurance dir: %v", err)
	}
	if err := os.WriteFile(canonicalBaseline, []byte("conflict\n"), 0o644); err != nil {
		t.Fatalf("write canonical baseline conflict: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.13"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit for migration conflict, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "assurance", "baseline.yaml")); err != nil {
		t.Fatalf("expected live legacy baseline preserved after conflict, err=%v", err)
	}
}

func setVerifiedTierForUpgradeTest(t *testing.T, root string) {
	t.Helper()
	configPath := filepath.Join(root, "runecontext.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	updated := strings.Replace(string(data), "assurance_tier: plain", "assurance_tier: verified", 1)
	if err := os.WriteFile(configPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeLegacyAssuranceFixtureForMigration(t *testing.T, root, historyFile string) {
	t.Helper()
	legacyBaseline := filepath.Join(root, "assurance", "baseline.yaml")
	if err := os.MkdirAll(filepath.Join(root, "assurance", "backfill"), 0o755); err != nil {
		t.Fatalf("mkdir legacy assurance backfill: %v", err)
	}
	historyJSON := legacyImportedHistoryFixtureJSON()
	if err := os.WriteFile(filepath.Join(root, "assurance", "backfill", historyFile), []byte(historyJSON), 0o644); err != nil {
		t.Fatalf("write legacy backfill file: %v", err)
	}
	baseline := strings.Join([]string{
		"schema_version: 1",
		"kind: baseline",
		"subject_id: project-root",
		"created_at: 1710000000",
		"canonicalization: runecontext-canonical-json-v1",
		"value:",
		"  adoption_commit: abcdef1234567890abcdef1234567890abcdef12",
		"  source_posture: embedded",
		"  imported_evidence:",
		"    - path: assurance/backfill/" + historyFile,
		"      provenance: imported_git_history",
		"",
	}, "\n")
	if err := os.WriteFile(legacyBaseline, []byte(baseline), 0o644); err != nil {
		t.Fatalf("write legacy baseline: %v", err)
	}
}

func legacyImportedHistoryFixtureJSON() string {
	return strings.Join([]string{
		"{",
		"  \"schema_version\": 1,",
		"  \"kind\": \"history\",",
		"  \"provenance\": \"imported_git_history\",",
		"  \"generated_at\": 1710000000,",
		"  \"adoption_commit\": \"abcdef1234567890abcdef1234567890abcdef12\",",
		"  \"commits\": [",
		"    {",
		"      \"commit\": \"abcdef1234567890abcdef1234567890abcdef12\",",
		"      \"committed_at\": 1710000000,",
		"      \"author_name\": \"RuneContext\",",
		"      \"author_email\": \"tests@example.com\",",
		"      \"subject\": \"seed\"",
		"    }",
		"  ]",
		"}",
		"",
	}, "\n")
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
