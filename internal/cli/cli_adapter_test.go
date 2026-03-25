package cli

import (
	"bytes"
	"fmt"
	"os"
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
	if !strings.Contains(stderr.String(), "mutation does not support symlinked targets") {
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
