package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunAdapterUsageAndHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"adapter"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage="+adapterUsage) {
		t.Fatalf("expected adapter usage output, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"adapter", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+adapterUsage) {
		t.Fatalf("expected adapter help usage, got %q", stdout.String())
	}
}

func TestRunAdapterSyncHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"adapter", "sync", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+adapterSyncUsage) {
		t.Fatalf("expected adapter sync help usage, got %q", stdout.String())
	}
}

func TestRunAdapterSyncDryRunIsReadOnly(t *testing.T) {
	projectRoot := t.TempDir()
	hostNativeRoot := filepath.Join(projectRoot, ".opencode")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--dry-run", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], adapterSyncCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "false"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
	if _, err := os.Stat(hostNativeRoot); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run to avoid host-native writes, got err=%v", err)
	}
}

func TestRunAdapterSyncNeverPerformsImplicitUpgrade(t *testing.T) {
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], adapterSyncCommand; got != want {
		t.Fatalf("expected adapter sync command %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "false"; got != want {
		t.Fatalf("expected local-only adapter sync network_access=%q, got %q", want, got)
	}
	if strings.Contains(stdout.String(), "upgrade") {
		t.Fatalf("expected no implicit upgrade plan fields in adapter sync output, got %q", stdout.String())
	}
}

func TestRunAdapterSyncAppliesHostNativeFiles(t *testing.T) {
	projectRoot := t.TempDir()
	userConfigPath := createUserOwnedConfig(t, projectRoot)

	fields := runAdapterSyncAndParse(t, projectRoot, "opencode")
	if got, want := fields["mutation_performed"], "true"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got := fields["changed_file_count"]; got == "0" {
		t.Fatalf("expected changed files on first sync, got %#v", fields)
	}

	hostNativeSkillPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	if _, err := os.Stat(hostNativeSkillPath); err != nil {
		t.Fatalf("expected host-native skill file to exist: %v", err)
	}
	assertNoAdapterTrackingTree(t, projectRoot)
	assertAdapterSyncBoundaries(t, userConfigPath, projectRoot)

	fields = assertAdapterSyncNoOpPreservesMtime(t, projectRoot, hostNativeSkillPath, "opencode")
	if got, want := fields["changed_file_count"], "0"; got != want {
		t.Fatalf("expected idempotent sync changed_file_count %q, got %q", want, got)
	}
}

func TestRunAdapterSyncWritesHostNativeArtifactsByTool(t *testing.T) {
	projectRoot := t.TempDir()

	assertOpenCodeHostNativeArtifacts(t, projectRoot)
	assertClaudeCodeHostNativeArtifacts(t, projectRoot)
	assertCodexHostNativeArtifacts(t, projectRoot)
}

func assertOpenCodeHostNativeArtifacts(t *testing.T, projectRoot string) {
	t.Helper()
	opencode := runAdapterSyncAndParse(t, projectRoot, "opencode")
	if got, want := opencode["host_native_file_count"], "16"; got != want {
		t.Fatalf("expected opencode host_native_file_count %q, got %q", want, got)
	}
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md"), "runectx adapter render-host-native --role flow_asset opencode change-new")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-assess-intake.md"), "runectx adapter render-host-native --role flow_asset opencode change-assess-intake")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-assess-intake.md"), "runectx adapter render-host-native --role discoverability_shim opencode change-assess-intake")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-assess-decomposition.md"), "runectx adapter render-host-native --role flow_asset opencode change-assess-decomposition")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-assess-decomposition.md"), "runectx adapter render-host-native --role discoverability_shim opencode change-assess-decomposition")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-decomposition-plan.md"), "runectx adapter render-host-native --role flow_asset opencode change-decomposition-plan")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-decomposition-plan.md"), "runectx adapter render-host-native --role discoverability_shim opencode change-decomposition-plan")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-decomposition-apply.md"), "runectx adapter render-host-native --role flow_asset opencode change-decomposition-apply")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-decomposition-apply.md"), "runectx adapter render-host-native --role discoverability_shim opencode change-decomposition-apply")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-new.md"), "runectx adapter render-host-native --role discoverability_shim opencode change-new")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-new.md"), "description: Create a new RuneContext change")
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md"))
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-new.md"))
}

func assertClaudeCodeHostNativeArtifacts(t *testing.T, projectRoot string) {
	t.Helper()
	claude := runAdapterSyncAndParse(t, projectRoot, "claude-code")
	if got, want := claude["host_native_file_count"], "9"; got != want {
		t.Fatalf("expected claude host_native_file_count %q, got %q", want, got)
	}
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-new.md"), "runectx adapter render-host-native --role flow_asset claude-code change-new")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-assess-intake.md"), "runectx adapter render-host-native --role flow_asset claude-code change-assess-intake")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-assess-decomposition.md"), "runectx adapter render-host-native --role flow_asset claude-code change-assess-decomposition")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-decomposition-plan.md"), "runectx adapter render-host-native --role flow_asset claude-code change-decomposition-plan")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-decomposition-apply.md"), "runectx adapter render-host-native --role flow_asset claude-code change-decomposition-apply")
	assertShellInjectionCallPresent(t, filepath.Join(projectRoot, ".claude", "commands", "runecontext.md"), "runectx adapter render-host-native --role discoverability_shim claude-code index")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-new.md"), "name: runecontext-change-new")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-new.md"), "description: Create a new RuneContext change")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".claude", "commands", "runecontext.md"), "name: runecontext")
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-new.md"))
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".claude", "commands", "runecontext.md"))
}

func assertCodexHostNativeArtifacts(t *testing.T, projectRoot string) {
	t.Helper()
	codex := runAdapterSyncAndParse(t, projectRoot, "codex")
	if got, want := codex["host_native_file_count"], "8"; got != want {
		t.Fatalf("expected codex host_native_file_count %q, got %q", want, got)
	}
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-new.md"), "name: runecontext-change-new")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-assess-intake.md"), "name: runecontext-change-assess-intake")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-assess-decomposition.md"), "name: runecontext-change-assess-decomposition")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-decomposition-plan.md"), "name: runecontext-change-decomposition-plan")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-decomposition-apply.md"), "name: runecontext-change-decomposition-apply")
	assertFrontmatterContains(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-new.md"), "description: Create a new RuneContext change")
	assertNoShellInjectionCall(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-new.md"))
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".agents", "skills", "runecontext-change-new.md"))
}

func TestRunAdapterSyncHostNativeConflictFailsClosed(t *testing.T) {
	projectRoot := t.TempDir()
	conflictPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	if err := os.MkdirAll(filepath.Dir(conflictPath), 0o755); err != nil {
		t.Fatalf("mkdir host-native conflict parent: %v", err)
	}
	if err := os.WriteFile(conflictPath, []byte("user owned\n"), 0o644); err != nil {
		t.Fatalf("write host-native conflict file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "existing file is not RuneContext-managed") {
		t.Fatalf("expected host-native ownership conflict, got %q", stderr.String())
	}
	errorFields := parseCLIKeyValueOutput(t, stderr.String())
	if got := errorFields["error_message"]; strings.Contains(got, filepath.ToSlash(projectRoot)) {
		t.Fatalf("expected repo-relative conflict path in error_message, got %q", got)
	}
}

func TestRunAdapterSyncRemovesStaleHostNativeFiles(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")
	staleRel := ".opencode/skills/runecontext-stale.md"

	staleHostNative := filepath.Join(projectRoot, filepath.FromSlash(staleRel))
	if err := os.MkdirAll(filepath.Dir(staleHostNative), 0o755); err != nil {
		t.Fatalf("mkdir stale host-native dir: %v", err)
	}
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:stale -->\n"
	if err := os.WriteFile(staleHostNative, []byte(managed), 0o644); err != nil {
		t.Fatalf("write stale host-native file: %v", err)
	}

	fields := runAdapterSyncAndParse(t, projectRoot, "opencode")
	if got := fields["changed_file_count"]; got == "0" {
		t.Fatalf("expected stale host-native cleanup mutation, got %#v", fields)
	}
	if _, err := os.Stat(staleHostNative); !os.IsNotExist(err) {
		t.Fatalf("expected stale host-native artifact removal, got err=%v", err)
	}
}

func assertAdapterSyncNoOpPreservesMtime(t *testing.T, projectRoot, managedReadmePath, tool string) map[string]string {
	t.Helper()
	beforeInfo, err := os.Stat(managedReadmePath)
	if err != nil {
		t.Fatalf("stat managed README before re-sync: %v", err)
	}
	time.Sleep(2 * time.Second)
	fields := runAdapterSyncAndParse(t, projectRoot, tool)
	afterInfo, err := os.Stat(managedReadmePath)
	if err != nil {
		t.Fatalf("stat managed README after re-sync: %v", err)
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Fatalf("expected no-op sync to preserve file mtime, before=%s after=%s", beforeInfo.ModTime().UTC().Format(time.RFC3339Nano), afterInfo.ModTime().UTC().Format(time.RFC3339Nano))
	}
	return fields
}

func runAdapterSyncAndParse(t *testing.T, projectRoot, tool string) map[string]string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, tool}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())
}

func TestRunAdapterSyncWritesExpectedHostNativeFilePermissions(t *testing.T) {
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	adaptersRoot := filepath.Join(root, "adapters", "source", "packs")
	t.Chdir(root)
	assertAdapterSyncWritesExpectedFilePermissions(t, adaptersRoot)
}

func assertAdapterSyncWritesExpectedFilePermissions(t *testing.T, adaptersRoot string) {
	t.Helper()

	sourcePath := filepath.Join(adaptersRoot, "opencode", "automation", "validate_after_authoritative_edit.sh")
	originalMode := statMode(t, sourcePath)
	t.Cleanup(func() {
		if err := os.Chmod(sourcePath, originalMode); err != nil {
			t.Fatalf("restore source mode: %v", err)
		}
	})
	if err := os.Chmod(sourcePath, 0o755); err != nil {
		t.Fatalf("chmod source executable: %v", err)
	}
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}

	syncedPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	syncedMode := statMode(t, syncedPath)
	syncedData, err := os.ReadFile(syncedPath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if !strings.Contains(string(syncedData), "adapter render-host-native") {
		t.Fatalf("expected synced host-native file to contain render-host-native call")
	}
	if runtime.GOOS != "windows" && syncedMode.Perm() != 0o644 {
		t.Fatalf("expected synced host-native file permissions 0644, got %#o", syncedMode.Perm())
	}
}

func TestRunAdapterSyncSyncedSkillContainsRenderCall(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	_ = runAdapterSyncAndParse(t, projectRoot, "opencode")

	skillPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read synced host-native skill: %v", err)
	}
	if !strings.Contains(string(skillData), "adapter render-host-native") {
		t.Fatalf("expected render-host-native mapping in host-native skill, got %q", string(skillData))
	}
}

func TestRunAdapterSyncRejectsSymlinkedHostNativeTarget(t *testing.T) {
	projectRoot := t.TempDir()
	symlinkTarget := filepath.Join(projectRoot, "outside-change-new.md")
	if err := os.WriteFile(symlinkTarget, []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	hostNativePath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	if err := os.MkdirAll(filepath.Dir(hostNativePath), 0o755); err != nil {
		t.Fatalf("mkdir host-native dir: %v", err)
	}
	if err := os.Symlink(symlinkTarget, hostNativePath); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create host-native symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), ".opencode/skills/runecontext-change-new.md") {
		t.Fatalf("expected host-native symlink target rejection, got %q", stderr.String())
	}
}

func TestRunAdapterSyncRejectsSymlinkedHostNativeAncestor(t *testing.T) {
	projectRoot := t.TempDir()
	outside := t.TempDir()
	symlinkRoot := filepath.Join(projectRoot, ".opencode")
	if err := os.Symlink(outside, symlinkRoot); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create dot runecontext symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), ".opencode") {
		t.Fatalf("expected host-native ancestor symlink rejection, got %q", stderr.String())
	}
}

func TestRunAdapterSyncUnknownToolFails(t *testing.T) {
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "missing-tool"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "adapter \"missing-tool\" not found") {
		t.Fatalf("expected unknown adapter output, got %q", stderr.String())
	}
}

func TestRunAdapterSyncGenericToolUnsupported(t *testing.T) {
	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "generic"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "does not define repo-local host-native artifacts") {
		t.Fatalf("expected generic unsupported host-native error, got %q", stderr.String())
	}
}

func TestValidateAfterAuthoritativeEditScriptBoundaries(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skipf("bash not available: %v", err)
	}
	test := prepareValidateHookBoundaryTest(t)

	t.Run("runs validate for authoritative paths", func(t *testing.T) {
		called, err := runValidateHookScript(test, "runecontext/changes/CHG-2026-001-a3f2-auth-gateway/status.yaml")
		if err != nil {
			t.Fatalf("run validate hook: %v", err)
		}
		if !strings.Contains(called, "validate --path ") {
			t.Fatalf("expected validate invocation, got %q", called)
		}
	})

	t.Run("skips unrelated paths", func(t *testing.T) {
		_, err := runValidateHookScript(test, "pkg/app/main.go")
		if err != nil {
			t.Fatalf("run validate hook: %v", err)
		}
		if _, err := os.Stat(test.calledPath); !os.IsNotExist(err) {
			t.Fatalf("expected no validate call for unrelated paths, got err=%v", err)
		}
	})
}
