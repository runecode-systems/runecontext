package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdapterSyncHostNativeSpoofedMarkerFailsClosed(t *testing.T) {
	projectRoot := t.TempDir()
	conflictPath := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	if err := os.MkdirAll(filepath.Dir(conflictPath), 0o755); err != nil {
		t.Fatalf("mkdir host-native conflict parent: %v", err)
	}
	content := "# user file\nrunecontext-managed-artifact: host-native-v1\n"
	if err := os.WriteFile(conflictPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write spoofed marker file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "existing file is not RuneContext-managed") {
		t.Fatalf("expected spoofed-marker conflict, got %q", stderr.String())
	}
}

func TestRunAdapterSyncRejectsInvalidManifestHostNativePath(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	manifestPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "sync-manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	updated := strings.Replace(string(manifestData), "host_native_files:\n", "host_native_files:\n  - ../outside.md\n", 1)
	if err := os.WriteFile(manifestPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write updated manifest: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "outside supported host-native roots") {
		t.Fatalf("expected invalid manifest host-native path error, got %q", stderr.String())
	}
}

func TestApplyAdapterSyncRechecksHostNativeOwnershipBeforeWrite(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	path := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-change-new.md")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read host-native file: %v", err)
	}
	modified := strings.Replace(string(original), "RuneContext Skill", "RuneContext Skill (local)", 1)
	if modified == string(original) {
		t.Fatalf("expected local rewrite marker in host-native file")
	}
	if err := os.WriteFile(path, []byte(modified), 0o644); err != nil {
		t.Fatalf("write locally modified host-native file: %v", err)
	}

	state, err := buildAdapterSyncState(adapterRequest{root: projectRoot, explicitRoot: true, tool: "opencode"})
	if err != nil {
		t.Fatalf("build state: %v", err)
	}
	if err := os.WriteFile(path, []byte("user owned\n"), 0o644); err != nil {
		t.Fatalf("write user-owned replacement: %v", err)
	}
	if err := applyAdapterSync(state); err == nil {
		t.Fatalf("expected ownership recheck failure during apply")
	}
}

func TestApplyAdapterSyncRechecksHostNativeOwnershipBeforeDelete(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	manifestPath := filepath.Join(projectRoot, ".runecontext", "adapters", "opencode", "sync-manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	staleRel := ".opencode/skills/runecontext-delete-race.md"
	updated := strings.Replace(string(manifestData), "host_native_files:\n", "host_native_files:\n  - "+staleRel+"\n", 1)
	if err := os.WriteFile(manifestPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write updated manifest: %v", err)
	}

	stalePath := filepath.Join(projectRoot, filepath.FromSlash(staleRel))
	if err := os.MkdirAll(filepath.Dir(stalePath), 0o755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:delete-race -->\n"
	if err := os.WriteFile(stalePath, []byte(managed), 0o644); err != nil {
		t.Fatalf("write managed stale file: %v", err)
	}

	state, err := buildAdapterSyncState(adapterRequest{root: projectRoot, explicitRoot: true, tool: "opencode"})
	if err != nil {
		t.Fatalf("build state: %v", err)
	}
	if err := os.WriteFile(stalePath, []byte("user owned\n"), 0o644); err != nil {
		t.Fatalf("write user-owned stale file: %v", err)
	}
	if err := applyAdapterSync(state); err == nil {
		t.Fatalf("expected ownership recheck failure for delete during apply")
	}
}
