package contracts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestSourceResolutionEmbeddedGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "source-resolution", "embedded-project")

	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected embedded fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "embedded.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
	})
	if index.Resolution.MaterializedRoot() != filepath.Join(projectRoot, "runecontext") {
		t.Fatalf("expected embedded source to materialize from live tree")
	}
}

func TestSourceResolutionPathLocalAndRemoteCI(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "source-resolution", "path-project")

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected local path fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "path-local.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
	})
	if index.Resolution.Tree == nil || index.Resolution.Tree.SnapshotKind != "snapshot_copy" {
		t.Fatalf("expected path mode to use a snapshot-friendly local tree")
	}

	_, err = v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeRemoteCI,
	})
	if err == nil || !strings.Contains(err.Error(), "source.type=path is invalid in execution mode remote_ci") {
		t.Fatalf("expected remote/ci path resolution to fail, got %v", err)
	}
}

func TestSourceResolutionMonorepoNearestAncestor(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	monorepoRoot := fixturePath(t, "source-resolution", "monorepo")

	nestedStart := filepath.Join(monorepoRoot, "packages", "service", "internal")
	nestedIndex, err := v.ValidateProjectWithOptions(nestedStart, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryNearestAncestor,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected nested monorepo fixture to validate: %v", err)
	}
	defer nestedIndex.Close()
	assertResolutionMatchesGolden(t, nestedIndex.Resolution, fixturePath(t, "source-resolution", "golden", "monorepo-nested.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(monorepoRoot),
	})

	rootStart := filepath.Join(monorepoRoot, "packages", "worker")
	rootIndex, err := v.ValidateProjectWithOptions(rootStart, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryNearestAncestor,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected root monorepo fixture to validate: %v", err)
	}
	defer rootIndex.Close()
	assertResolutionMatchesGolden(t, rootIndex.Resolution, fixturePath(t, "source-resolution", "golden", "monorepo-root.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(monorepoRoot),
	})
}

func TestSourceResolutionGitPinnedCommitGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  commit: %s\n  subdir: runecontext\n", repoDir, commit))

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected git pinned fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "git-pinned.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
		"${COMMIT}":       commit,
	})
	if index.Resolution.Tree == nil || index.Resolution.Tree.SnapshotKind != "git_checkout" {
		t.Fatalf("expected git source to materialize via checkout")
	}
}

func TestSourceResolutionGitMutableRefRequiresOptInAndWarns(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  ref: main\n  allow_mutable_ref: true\n  subdir: runecontext\n", repoDir))

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected mutable-ref fixture to validate: %v", err)
	}
	defer index.Close()
	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "git-mutable-ref.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
		"${COMMIT}":       commit,
	})

	rejectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  ref: main\n  subdir: runecontext\n", repoDir))
	_, err = v.LoadProject(rejectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err == nil || !strings.Contains(err.Error(), "allow_mutable_ref") {
		t.Fatalf("expected missing mutable-ref opt-in to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsEmbeddedPathEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	outside := filepath.Join(filepath.Dir(projectRoot), "outside-runecontext")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside dir: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: embedded\n  path: ../outside-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "embedded source path") {
		t.Fatalf("expected embedded escape to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsUnsafeGitInputs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)

	t.Run("url starts with dash", func(t *testing.T) {
		projectRoot := writeRootConfigProject(t, "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: -bad-url\n  commit: "+commit+"\n  subdir: runecontext\n")
		_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err == nil || !strings.Contains(err.Error(), "git source url") {
			t.Fatalf("expected unsafe git url to fail, got %v", err)
		}
	})

	t.Run("ref starts with dash", func(t *testing.T) {
		projectRoot := writeRootConfigProject(t, "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: "+repoDir+"\n  ref: -main\n  allow_mutable_ref: true\n  subdir: runecontext\n")
		_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err == nil || !strings.Contains(err.Error(), "git ref") {
			t.Fatalf("expected unsafe git ref to fail, got %v", err)
		}
	})

	t.Run("ref contains dot dot", func(t *testing.T) {
		projectRoot := writeRootConfigProject(t, "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: "+repoDir+"\n  ref: feature..branch\n  allow_mutable_ref: true\n  subdir: runecontext\n")
		_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err == nil || !strings.Contains(err.Error(), "must not contain '..'") {
			t.Fatalf("expected dot-dot git ref to fail, got %v", err)
		}
	})

	t.Run("ref ends with slash", func(t *testing.T) {
		projectRoot := writeRootConfigProject(t, "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: "+repoDir+"\n  ref: feature/\n  allow_mutable_ref: true\n  subdir: runecontext\n")
		_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err == nil || !strings.Contains(err.Error(), "start or end with '/'") {
			t.Fatalf("expected trailing-slash git ref to fail, got %v", err)
		}
	})

	t.Run("subdir escapes repo", func(t *testing.T) {
		projectRoot := writeRootConfigProject(t, "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: "+repoDir+"\n  commit: "+commit+"\n  subdir: ../outside\n")
		_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err == nil || !strings.Contains(err.Error(), "git subdir") {
			t.Fatalf("expected escaping git subdir to fail, got %v", err)
		}
	})
}

func TestSourceResolutionRejectsPathSymlinkEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	outside := filepath.Join(projectRoot, "outside.txt")
	if err := os.MkdirAll(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution"), 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := tryCreateSymlink("../outside.txt", filepath.Join(localRoot, "escape-link")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "escapes declared local source tree") {
		t.Fatalf("expected path symlink escape to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsPathSymlinkCycle(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	if err := os.MkdirAll(localRoot, 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	if err := tryCreateSymlink(".", filepath.Join(localRoot, "loop")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "symlink cycle detected") {
		t.Fatalf("expected path symlink cycle to fail, got %v", err)
	}
}

func TestSourceResolutionSkipsDotGitDirectoryInSnapshots(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	gitDir := filepath.Join(localRoot, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write fake git head: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution"), 0o755); err != nil {
		t.Fatalf("mkdir changes dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution", "status.yaml"), []byte("schema_version: 1\nid: CHG-2026-001-a3f2-source-resolution\ntitle: Test snapshot exclusions\nstatus: proposed\ntype: feature\nsize: small\nverification_status: pending\ncontext_bundles: []\nrelated_specs: []\nrelated_decisions: []\nrelated_changes: []\ndepends_on: []\ninformed_by: []\nsupersedes: []\nsuperseded_by: []\ncreated_at: \"2026-03-17\"\nclosed_at: null\npromotion_assessment:\n  status: pending\n  suggested_targets: []\n"), 0o644); err != nil {
		t.Fatalf("write status file: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected path source with .git directory to resolve: %v", err)
	}
	defer loaded.Close()
	if _, err := os.Stat(filepath.Join(loaded.Resolution.MaterializedRoot(), ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot to exclude .git directory, got err=%v", err)
	}
}

func TestSourceResolutionRejectsOversizedPathSnapshot(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	if err := os.MkdirAll(localRoot, 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	data := strings.Repeat("a", int(localSnapshotLimits.MaxBytes)+1)
	if err := os.WriteFile(filepath.Join(localRoot, "large.txt"), []byte(data), 0o644); err != nil {
		t.Fatalf("write oversized file: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "maximum snapshot size") {
		t.Fatalf("expected oversized snapshot to fail, got %v", err)
	}
}

func assertResolutionMatchesGolden(t *testing.T, resolution *SourceResolution, goldenPath string, replacements map[string]string) {
	t.Helper()
	if resolution == nil {
		t.Fatal("expected resolution metadata")
	}
	expected := normalizeResolutionValue(t, mustParseYAML(t, replacePlaceholders(string(readFixture(t, goldenPath)), replacements)))
	actual := normalizeResolutionValue(t, comparableResolution(resolution))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("resolution metadata mismatch\nexpected: %#v\nactual:   %#v", expected, actual)
	}
}

func comparableResolution(resolution *SourceResolution) map[string]any {
	result := map[string]any{
		"selected_config_path": filepath.ToSlash(resolution.SelectedConfigPath),
		"project_root":         filepath.ToSlash(resolution.ProjectRoot),
		"source_root":          filepath.ToSlash(resolution.SourceRoot),
		"source_mode":          string(resolution.SourceMode),
		"source_ref":           filepath.ToSlash(resolution.SourceRef),
		"verification_posture": string(resolution.VerificationPosture),
	}
	if resolution.ResolvedCommit != "" {
		result["resolved_commit"] = resolution.ResolvedCommit
	}
	diagnostics := make([]any, 0, len(resolution.Diagnostics))
	for _, diagnostic := range resolution.Diagnostics {
		diagnostics = append(diagnostics, map[string]any{
			"severity": string(diagnostic.Severity),
			"code":     diagnostic.Code,
			"message":  diagnostic.Message,
		})
	}
	result["diagnostics"] = diagnostics
	return result
}

func normalizeResolutionValue(t *testing.T, value any) any {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeResolutionValue(t, item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeResolutionValue(t, item)
		}
		return result
	case string:
		return filepath.ToSlash(typed)
	default:
		return typed
	}
}

func mustParseYAML(t *testing.T, text string) any {
	t.Helper()
	value, err := parseYAML([]byte(text))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	return value
}

func replacePlaceholders(text string, replacements map[string]string) string {
	for oldValue, newValue := range replacements {
		text = strings.ReplaceAll(text, oldValue, newValue)
	}
	return text
}

func writeRootConfigProject(t *testing.T, config string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	return root
}

func createGitSourceRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()
	runGitTest(t, repoDir, "init", "--initial-branch=main")
	templateRoot := fixturePath(t, "source-resolution", "templates", "minimal-runecontext")
	copyDirForTest(t, templateRoot, filepath.Join(repoDir, "runecontext"))
	runGitTest(t, repoDir, "add", ".")
	runGitTest(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	commit := strings.TrimSpace(gitOutputForTest(t, repoDir, "rev-parse", "HEAD"))
	return repoDir, commit
}

func TestSourceResolutionGitPinnedCommitWorksFromAdvertisedRefs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  commit: %s\n  subdir: runecontext\n", repoDir, commit))

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected pinned commit to resolve from advertised refs: %v", err)
	}
	defer loaded.Close()
	if loaded.Resolution == nil || loaded.Resolution.ResolvedCommit != commit {
		t.Fatalf("expected resolved commit %q, got %#v", commit, loaded.Resolution)
	}
}

func copyDirForTest(t *testing.T, srcRoot, dstRoot string) {
	t.Helper()
	if err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstRoot, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture directory: %v", err)
	}
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = sanitizedGitEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func gitOutputForTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = sanitizedGitEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

func tryCreateSymlink(target, path string) error {
	if err := os.Symlink(target, path); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			return fmt.Errorf("symlink tests skipped: %w", err)
		}
		return fmt.Errorf("create symlink: %w", err)
	}
	return nil
}
