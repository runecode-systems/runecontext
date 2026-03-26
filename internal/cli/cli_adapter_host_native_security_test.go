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

	staleRel := ".opencode/skills/runecontext-delete-race.md"

	stalePath := filepath.Join(projectRoot, filepath.FromSlash(staleRel))
	if err := os.MkdirAll(filepath.Dir(stalePath), 0o755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:delete-race -->\n"
	managed = "---\ndescription: stale managed marker\n---\n" + managed
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

func TestRunAdapterSyncIgnoresUnrelatedFilesUnderHostNativeRoots(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	unrelated := filepath.Join(projectRoot, ".opencode", "skills", "notes.md")
	if err := os.MkdirAll(filepath.Dir(unrelated), 0o755); err != nil {
		t.Fatalf("mkdir unrelated file dir: %v", err)
	}
	if err := os.WriteFile(unrelated, []byte("user note\n"), 0o644); err != nil {
		t.Fatalf("write unrelated file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected sync success with unrelated file present, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("expected unrelated file preserved, got %v", err)
	}
}

func TestRunAdapterSyncIgnoresManagedFileFromDifferentTool(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	foreign := filepath.Join(projectRoot, ".opencode", "skills", "runecontext-foreign.md")
	if err := os.MkdirAll(filepath.Dir(foreign), 0o755); err != nil {
		t.Fatalf("mkdir foreign file dir: %v", err)
	}
	content := "---\nname: runecontext-foreign\ndescription: foreign\n---\n<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: claude-code -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:foreign -->\n"
	if err := os.WriteFile(foreign, []byte(content), 0o644); err != nil {
		t.Fatalf("write foreign managed file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "sync", "--path", projectRoot, "opencode"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected sync success with foreign-tool managed file, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatalf("expected foreign-tool managed file preserved, got %v", err)
	}
}

func TestApplyAdapterSyncDeleteRejectsCrossToolSwap(t *testing.T) {
	projectRoot := t.TempDir()
	runAdapterSyncAndParse(t, projectRoot, "opencode")

	staleRel := ".opencode/skills/runecontext-delete-cross-tool.md"
	stalePath := filepath.Join(projectRoot, filepath.FromSlash(staleRel))
	if err := os.MkdirAll(filepath.Dir(stalePath), 0o755); err != nil {
		t.Fatalf("mkdir stale dir: %v", err)
	}
	opencodeManaged := "---\ndescription: opencode stale marker\n---\n<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:delete-cross-tool -->\n"
	if err := os.WriteFile(stalePath, []byte(opencodeManaged), 0o644); err != nil {
		t.Fatalf("write opencode stale file: %v", err)
	}

	state, err := buildAdapterSyncState(adapterRequest{root: projectRoot, explicitRoot: true, tool: "opencode"})
	if err != nil {
		t.Fatalf("build state: %v", err)
	}
	claudeReplacement := "---\nname: runecontext-delete-cross-tool\ndescription: claude replacement\n---\n<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: claude-code -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:delete-cross-tool -->\n"
	if err := os.WriteFile(stalePath, []byte(claudeReplacement), 0o644); err != nil {
		t.Fatalf("write claude replacement stale file: %v", err)
	}
	if err := applyAdapterSync(state); err == nil {
		t.Fatalf("expected cross-tool delete ownership rejection")
	}
}

func TestParseHostNativeOwnershipHeaderAcceptsCRLF(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"description: windows newlines",
		"---",
		"<!-- runecontext-managed-artifact: host-native-v1 -->",
		"<!-- runecontext-tool: opencode -->",
		"<!-- runecontext-kind: flow_asset -->",
		"<!-- runecontext-id: runecontext:change-new -->",
	}, "\r\n") + "\r\n"
	if _, ok := parseHostNativeOwnershipHeader([]byte(content)); !ok {
		t.Fatalf("expected CRLF ownership header parsing to succeed")
	}
}

func TestParseHostNativeOwnershipHeaderAcceptsCROnly(t *testing.T) {
	content := strings.Join([]string{
		"---",
		"description: classic mac newlines",
		"---",
		"<!-- runecontext-managed-artifact: host-native-v1 -->",
		"<!-- runecontext-tool: opencode -->",
		"<!-- runecontext-kind: flow_asset -->",
		"<!-- runecontext-id: runecontext:change-new -->",
	}, "\r") + "\r"
	if _, ok := parseHostNativeOwnershipHeader([]byte(content)); !ok {
		t.Fatalf("expected CR-only ownership header parsing to succeed")
	}
}
