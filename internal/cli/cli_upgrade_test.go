package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunUpgradePreviewOnReferenceFixture(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["phase"], "preview"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["apply_required"], "true"; got != want {
		t.Fatalf("expected apply_required %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewSupportsStateClassificationAndAliases(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.9"

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "latest"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "upgradeable"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "true"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"upgrade", "--path", root, "--target-version", "installed"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields = parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["network_access"], "false"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
}

func TestRunUpgradePreviewPathSourceIsExternallyManaged(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(filepath.Dir(root), "external-runecontext")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatalf("mkdir external source: %v", err)
	}
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: path\n  path: ../external-runecontext\n"
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir content root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "conflicted"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
	if !strings.Contains(fields["plan_action_1"], "externally managed") {
		t.Fatalf("expected externally managed plan action, got %#v", fields)
	}
	if !strings.Contains(fields["next_action_1"], "external-runecontext") {
		t.Fatalf("expected next action to include path source guidance, got %#v", fields)
	}
}

func TestRunUpgradePreviewGitSourceOnlyMutatesConfigAndNotLinkedTree(t *testing.T) {
	repoDir := t.TempDir()
	runGitForCLI(t, repoDir, "init", "--initial-branch=main")
	if err := os.MkdirAll(filepath.Join(repoDir, "runecontext", "changes"), 0o755); err != nil {
		t.Fatalf("mkdir linked tree: %v", err)
	}
	linkedProbe := filepath.Join(repoDir, "runecontext", "changes", "probe.txt")
	if err := os.WriteFile(linkedProbe, []byte("linked tree content\n"), 0o644); err != nil {
		t.Fatalf("write linked probe: %v", err)
	}
	runGitForCLI(t, repoDir, "add", ".")
	runGitForCLI(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "seed")
	commit := strings.TrimSpace(gitOutputForCLI(t, repoDir, "rev-parse", "HEAD"))

	projectRoot := t.TempDir()
	config := fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  commit: %s\n  subdir: runecontext\n", repoDir, commit)
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", projectRoot, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got := fields["changed_1"]; got != "updated runecontext.yaml" {
		t.Fatalf("expected only root config mutation to be reported first, got %q", got)
	}
	probe, err := os.ReadFile(linkedProbe)
	if err != nil {
		t.Fatalf("read linked probe: %v", err)
	}
	if string(probe) != "linked tree content\n" {
		t.Fatalf("expected linked source tree to remain untouched, got %q", string(probe))
	}
}

func TestRunUpgradeApplyFailsClosedOnHostNativeOwnershipConflict(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	conflictPath := filepath.Join(root, ".opencode", "skills", "runecontext-change-new.md")
	if err := os.MkdirAll(filepath.Dir(conflictPath), 0o755); err != nil {
		t.Fatalf("mkdir conflict dir: %v", err)
	}
	if err := os.WriteFile(conflictPath, []byte("user-owned skill content\n"), 0o644); err != nil {
		t.Fatalf("write conflict file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "state\\=conflicted") {
		t.Fatalf("expected conflicted-state apply rejection, got %q", stderr.String())
	}
}

func TestRunUpgradePreviewAndApplyDetectStaleManagedHostNativeTree(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	if code := Run([]string{"adapter", "sync", "--path", root, "opencode"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("expected adapter sync success")
	}
	staleRel := ".opencode/skills/runecontext-stale-merge.md"
	staleAbs := filepath.Join(root, filepath.FromSlash(staleRel))
	if err := os.MkdirAll(filepath.Dir(staleAbs), 0o755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:stale-merge -->\n"
	if err := os.WriteFile(staleAbs, []byte(managed), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected upgrade preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "mixed_or_stale_tree"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"upgrade", "apply", "--path", root, "--target-version", "current"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected upgrade apply success for stale tree, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(staleAbs); !os.IsNotExist(err) {
		t.Fatalf("expected stale managed host-native file to be removed, err=%v", err)
	}
}

func TestRunUpgradeApplyTransactionalRollbackOnFailure(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	configPath := filepath.Join(root, "runecontext.yaml")
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before apply: %v", err)
	}

	original := upgradeApplyAdapterSyncFn
	t.Cleanup(func() { upgradeApplyAdapterSyncFn = original })
	upgradeApplyAdapterSyncFn = func(state adapterSyncState) error {
		return fmt.Errorf("forced upgrade transaction failure for %s", state.tool)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
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

func TestRunUpgradeApplyIdempotentRerun(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected first apply success, got %d (%s)", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected second apply success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["changed"], "false"; got != want {
		t.Fatalf("expected idempotent changed=%q, got %q", want, got)
	}
}

func TestRunUpgradePreviewUnsupportedProjectVersion(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.9"

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte("schema_version: 1\nrunecontext_version: 9.9.9\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir content root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["state"], "unsupported_project_version"; got != want {
		t.Fatalf("expected state %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyRewritesTargetVersion(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["phase"], "apply"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["previous_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected previous_version %q, got %q", want, got)
	}
	if got, want := fields["current_version"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
	if got, want := fields["changed"], "true"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"validate", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected upgraded project to validate, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "result=ok") {
		t.Fatalf("expected validate success output, got %q", stdout.String())
	}
}

func TestRunUpgradeApplyRequiresTargetVersion(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_message=upgrade apply requires --target-version") {
		t.Fatalf("expected missing target version error, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyNoOpUsesStableOutputFields(t *testing.T) {
	root := repoFixtureRoot(t, "reference-projects", "embedded")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.8"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["previous_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected previous_version %q, got %q", want, got)
	}
	if got, want := fields["current_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
	if got, want := fields["target_version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected target_version %q, got %q", want, got)
	}
	if got, want := fields["changed"], "false"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyRejectsUnregisteredSemverTransition(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "1.2.3-rc.1+build.5"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unsupported_project_version") {
		t.Fatalf("expected unsupported version rejection, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyHandlesValidYAMLSpacing(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\nrunecontext_version : 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(configPath, []byte(config), 0o640); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.Chmod(configPath, 0o640); err != nil {
		t.Fatalf("chmod config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if !strings.Contains(string(updated), "runecontext_version : 0.1.0-alpha.9") {
		t.Fatalf("expected updated runecontext version, got %q", string(updated))
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat updated config: %v", err)
	}
	if runtime.GOOS != "windows" {
		if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
			t.Fatalf("expected config mode %o, got %o", want, got)
		}
	}
}

func TestRunUpgradeApplyPreservesCommentsAndCRLF(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\r\nrunecontext_version: 0.1.0-alpha.8 # pinned here\r\nassurance_tier: plain\r\nsource:\r\n  type: embedded\r\n  path: runecontext\r\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	content := string(updated)
	if !strings.Contains(content, "runecontext_version: 0.1.0-alpha.9 # pinned here") {
		t.Fatalf("expected preserved inline comment, got %q", content)
	}
	if !strings.Contains(content, "\r\n") {
		t.Fatalf("expected CRLF line endings to remain, got %q", content)
	}
	if !strings.HasSuffix(content, "\r\n") {
		t.Fatalf("expected trailing CRLF newline to remain, got %q", content)
	}
}

func TestRunUpgradeApplyPreservesTrailingLFNewline(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.8\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--target-version", "0.1.0-alpha.9"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	updated, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	content := string(updated)
	if !strings.HasSuffix(content, "\n") {
		t.Fatalf("expected trailing LF newline to remain, got %q", content)
	}
}

func TestValidateRunecontextVersionKeyPresentReturnsMissingKeyError(t *testing.T) {
	err := validateRunecontextVersionKeyPresent([]byte("schema_version: 1\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"))
	if err == nil {
		t.Fatalf("expected missing runecontext_version error")
	}
	if !strings.Contains(err.Error(), "runecontext.yaml is missing runecontext_version") {
		t.Fatalf("expected missing runecontext_version error, got %v", err)
	}
}

func TestRewriteRunecontextVersionReportsUnrewriteableScalar(t *testing.T) {
	config := "schema_version: 1\nrunecontext_version:\n  major: 1\n"
	if _, err := rewriteRunecontextVersion([]byte(config), "0.1.0-alpha.9"); err == nil {
		t.Fatalf("expected rewrite error for non-scalar runecontext_version")
	} else if !strings.Contains(err.Error(), "not a rewriteable scalar") {
		t.Fatalf("expected scalar warning, got %v", err)
	}
}

func TestBuildUpgradeReadinessFromIndexIgnoresMissingOptionalAdapterPacks(t *testing.T) {
	root := repoFixtureRoot(t, "bundle-resolution", "valid-project")
	v := contracts.NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate project: %v", err)
	}
	defer index.Close()

	original := collectUpgradeAdapterPlansFn
	t.Cleanup(func() { collectUpgradeAdapterPlansFn = original })
	collectUpgradeAdapterPlansFn = func(absRoot string, includeCreate bool) (map[string]adapterSyncState, []string, []string, error) {
		return nil, nil, nil, fmt.Errorf("could not locate installed adapter packs")
	}

	plan, err := buildUpgradeReadinessFromIndex(root, index)
	if err != nil {
		t.Fatalf("expected optional adapter pack errors to be non-fatal, got %v", err)
	}
	if len(plan.Warnings) == 0 {
		t.Fatalf("expected warning for missing optional adapter pack")
	}
}
