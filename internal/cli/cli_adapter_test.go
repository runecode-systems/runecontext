package cli

import (
	"bytes"
	"fmt"
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
	managedRoot := filepath.Join(projectRoot, ".runecontext", "adapters", "generic", "managed")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--dry-run", "--path", projectRoot, "generic"}, &stdout, &stderr)
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
	if _, err := os.Stat(managedRoot); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run to avoid managed-root writes, got err=%v", err)
	}
}

func TestRunAdapterSyncAppliesManagedFilesAndManifest(t *testing.T) {
	projectRoot := t.TempDir()
	userConfigPath := createUserOwnedConfig(t, projectRoot)

	fields := runAdapterSyncAndParse(t, projectRoot, "opencode")
	if got, want := fields["mutation_performed"], "true"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got := fields["changed_file_count"]; got == "0" {
		t.Fatalf("expected changed files on first sync, got %#v", fields)
	}

	managedReadmePath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "managed", "README.md")
	if _, err := os.Stat(managedReadmePath); err != nil {
		t.Fatalf("expected managed adapter README to exist: %v", err)
	}
	assertAdapterManifestConvenience(t, projectRoot)
	assertAdapterSyncBoundaries(t, userConfigPath, projectRoot)

	fields = assertAdapterSyncNoOpPreservesMtime(t, projectRoot, managedReadmePath, "opencode")
	if got, want := fields["changed_file_count"], "0"; got != want {
		t.Fatalf("expected idempotent sync changed_file_count %q, got %q", want, got)
	}
}

func TestRunAdapterSyncWritesHostNativeArtifactsByTool(t *testing.T) {
	projectRoot := t.TempDir()

	opencode := runAdapterSyncAndParse(t, projectRoot, "opencode")
	if got, want := opencode["host_native_file_count"], "8"; got != want {
		t.Fatalf("expected opencode host_native_file_count %q, got %q", want, got)
	}
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md"))
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".opencode", "commands", "runecontext-change-new.md"))

	claude := runAdapterSyncAndParse(t, projectRoot, "claude-code")
	if got, want := claude["host_native_file_count"], "5"; got != want {
		t.Fatalf("expected claude host_native_file_count %q, got %q", want, got)
	}
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".claude", "skills", "runecontext-change-new.md"))
	assertManagedArtifactMarker(t, filepath.Join(projectRoot, ".claude", "commands", "runecontext.md"))

	codex := runAdapterSyncAndParse(t, projectRoot, "codex")
	if got, want := codex["host_native_file_count"], "4"; got != want {
		t.Fatalf("expected codex host_native_file_count %q, got %q", want, got)
	}
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

	manifestPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "sync-manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	staleRel := ".opencode/skills/runecontext-stale.md"
	updated := strings.Replace(string(manifestData), "host_native_files:\n", "host_native_files:\n  - "+staleRel+"\n", 1)
	if err := os.WriteFile(manifestPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write updated manifest: %v", err)
	}

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

func TestRunAdapterSyncPreservesExecutableBitFromAdapterSource(t *testing.T) {
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	adaptersRoot := filepath.Join(root, "adapters")
	t.Chdir(root)
	assertAdapterSyncPreservesExecutableBitFromSource(t, adaptersRoot)
}

func assertAdapterSyncPreservesExecutableBitFromSource(t *testing.T, adaptersRoot string) {
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
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read source file: %v", err)
	}

	projectRoot := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}

	syncedPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "managed", "automation", "validate_after_authoritative_edit.sh")
	syncedMode := statMode(t, syncedPath)
	syncedData, err := os.ReadFile(syncedPath)
	if err != nil {
		t.Fatalf("read synced file: %v", err)
	}
	if !bytes.Equal(sourceData, syncedData) {
		t.Fatalf("expected synced file content to match source")
	}
	if runtime.GOOS != "windows" && syncedMode.Perm() != 0o755 {
		t.Fatalf("expected synced executable permissions 0755, got %s", fmt.Sprintf("%#o", syncedMode.Perm()))
	}
}

func TestRunAdapterSyncSyncedHookRunsDirectly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("direct executable hook test is not supported on windows")
	}

	projectRoot := prepareCLIWorkflowProject(t)
	_ = runAdapterSyncAndParse(t, projectRoot, "opencode")

	scriptPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "managed", "automation", "validate_after_authoritative_edit.sh")
	if mode := statMode(t, scriptPath).Perm(); mode&0o111 == 0 {
		t.Fatalf("expected synced hook to be executable, got mode %s", fmt.Sprintf("%#o", mode))
	}

	fakeBin := t.TempDir()
	calledPath := filepath.Join(projectRoot, "validate-called")
	writeFakeRunectxExecutable(t, filepath.Join(fakeBin, "runectx"))

	cmd := exec.Command(scriptPath, "runecontext/changes/CHG-2026-001-a3f2-auth-gateway/status.yaml")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"PATH="+fakeBin+":"+os.Getenv("PATH"),
		"RUNECTX_ARGS_OUT="+calledPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run synced hook directly: %v\n%s", err, string(out))
	}

	called, err := os.ReadFile(calledPath)
	if err != nil {
		t.Fatalf("read fake runectx invocation: %v", err)
	}
	if !strings.Contains(string(called), "validate --path "+projectRoot) {
		t.Fatalf("expected validate invocation with project root, got %q", string(called))
	}
}

func TestRunAdapterSyncRejectsSymlinkedManagedTarget(t *testing.T) {
	projectRoot := t.TempDir()
	symlinkTarget := filepath.Join(projectRoot, "outside-readme.md")
	if err := os.WriteFile(symlinkTarget, []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	managedReadmePath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "managed", "README.md")
	if err := os.MkdirAll(filepath.Dir(managedReadmePath), 0o755); err != nil {
		t.Fatalf("mkdir managed dir: %v", err)
	}
	if err := os.Symlink(symlinkTarget, managedReadmePath); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create managed symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "adapter sync rejects symlinked path") {
		t.Fatalf("expected symlink-target rejection, got %q", stderr.String())
	}
}

func TestRunAdapterSyncRejectsSymlinkedAncestor(t *testing.T) {
	projectRoot := t.TempDir()
	outside := t.TempDir()
	symlinkRoot := filepath.Join(projectRoot, ".runecontext")
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
	if !strings.Contains(stderr.String(), "adapter sync rejects symlinked path") {
		t.Fatalf("expected ancestor symlink rejection, got %q", stderr.String())
	}
}

func TestRunAdapterSyncRejectsSymlinkedSourceFile(t *testing.T) {
	t.Skip("helper test no longer needed")
}

func TestCopyManagedFileRejectsSymlinkSource(t *testing.T) {
	srcRoot := t.TempDir()
	managedRoot := t.TempDir()
	filePath := filepath.Join(srcRoot, "script.sh")
	if err := os.WriteFile(filePath, []byte("echo hi"), 0o755); err != nil {
		t.Fatalf("write script file: %v", err)
	}
	symlinkPath := filepath.Join(srcRoot, "link.sh")
	if err := os.Symlink(filePath, symlinkPath); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlink: %v", err)
	}
	if err := copyManagedFile(srcRoot, managedRoot, "link.sh"); err == nil {
		t.Fatalf("expected error copying symlink source")
	}
}

func statMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode()
}

func createUserOwnedConfig(t *testing.T, projectRoot string) string {
	t.Helper()
	userConfigPath := filepath.Join(projectRoot, ".opencode", "config.yml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("mkdir user config dir: %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("user_owned: true\n"), 0o644); err != nil {
		t.Fatalf("write user config file: %v", err)
	}
	return userConfigPath
}

func assertUserOwnedConfigPreserved(t *testing.T, userConfigPath string) {
	t.Helper()
	userData, err := os.ReadFile(userConfigPath)
	if err != nil {
		t.Fatalf("read user config after sync: %v", err)
	}
	if string(userData) != "user_owned: true\n" {
		t.Fatalf("expected user config boundary to be preserved, got %q", string(userData))
	}
}

func assertAdapterSyncBoundaries(t *testing.T, userConfigPath, projectRoot string) {
	t.Helper()
	assertUserOwnedConfigPreserved(t, userConfigPath)
	if _, err := os.Stat(filepath.Join(projectRoot, "adapters")); !os.IsNotExist(err) {
		t.Fatalf("expected sync to avoid user-owned adapter source tree writes, got err=%v", err)
	}
}

func assertAdapterManifestConvenience(t *testing.T, projectRoot string) {
	t.Helper()
	manifestPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "sync-manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected adapter manifest to exist: %v", err)
	}
	if !strings.Contains(string(manifestData), "manifest_kind: convenience_metadata") {
		t.Fatalf("expected convenience manifest marker, got %q", string(manifestData))
	}
	if !strings.Contains(string(manifestData), "host_native_files:") {
		t.Fatalf("expected host-native manifest section, got %q", string(manifestData))
	}
	if !strings.Contains(string(manifestData), "host_native_discoverability_shims:") {
		t.Fatalf("expected discoverability shim section, got %q", string(manifestData))
	}
}

func assertManagedArtifactMarker(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read managed artifact %s: %v", path, err)
	}
	if _, ok := parseHostNativeOwnershipHeader(data); !ok {
		t.Fatalf("expected ownership marker in %s", path)
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

func TestRunAdapterSyncDeletesStaleManagedFiles(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "generic")
	stalePath := filepath.Join(projectRoot, ".runecontext", "adapters", "generic", "managed", "stale.txt")
	if err := os.WriteFile(stalePath, []byte("stale\n"), 0o644); err != nil {
		t.Fatalf("write stale managed file: %v", err)
	}
	fields := runAdapterSyncAndParse(t, projectRoot, "generic")
	if got := fields["changed_file_count"]; got == "0" {
		t.Fatalf("expected stale-file cleanup mutation, got %#v", fields)
	}
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("expected stale managed file removal, got err=%v", err)
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

type validateHookBoundaryTest struct {
	scriptPath  string
	projectRoot string
	fakeBin     string
	calledPath  string
}

func prepareValidateHookBoundaryTest(t *testing.T) validateHookBoundaryTest {
	t.Helper()
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	test := validateHookBoundaryTest{
		scriptPath:  filepath.Join(repoRoot, "adapters", "opencode", "automation", "validate_after_authoritative_edit.sh"),
		projectRoot: prepareCLIWorkflowProject(t),
		fakeBin:     t.TempDir(),
	}
	test.calledPath = filepath.Join(test.projectRoot, "validate-called")
	writeFakeRunectxExecutable(t, filepath.Join(test.fakeBin, "runectx"))
	return test
}

func writeFakeRunectxExecutable(t *testing.T, path string) {
	t.Helper()
	stub := "#!/usr/bin/env bash\nset -euo pipefail\nprintf '%s\\n' \"$*\" > \"$RUNECTX_ARGS_OUT\"\n"
	if err := os.WriteFile(path, []byte(stub), 0o755); err != nil {
		t.Fatalf("write fake runectx: %v", err)
	}
}

func runValidateHookScript(test validateHookBoundaryTest, changedPath string) (string, error) {
	_ = os.Remove(test.calledPath)
	cmd := exec.Command("bash", test.scriptPath, changedPath)
	cmd.Dir = test.projectRoot
	cmd.Env = append(os.Environ(),
		"PATH="+test.fakeBin+":"+os.Getenv("PATH"),
		"RUNECTX_ARGS_OUT="+test.calledPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("script failed: %w\n%s", err, string(out))
	}
	called, err := os.ReadFile(test.calledPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(called), nil
}
