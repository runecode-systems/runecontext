package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunUpgradePreviewDoesNotCreateUnsyncedHostNativeArtifacts(t *testing.T) {
	root := t.TempDir()
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["plan_action_count"], "1"; got != want {
		t.Fatalf("expected only version-bump action for unsynced tools, got %q", got)
	}
	if got, want := fields["plan_action_1"], "set runecontext_version to 0.1.0-alpha.10"; got != want {
		t.Fatalf("expected version-bump-only action %q, got %q", want, got)
	}
}

func TestRunUpgradeApplyDoesNotCreateUnsyncedHostNativeArtifacts(t *testing.T) {
	root := t.TempDir()
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")
	writeEmbeddedProjectVersion(t, root, "0.1.0-alpha.9")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".opencode")); !os.IsNotExist(err) {
		t.Fatalf("expected unsynced opencode tree to remain absent, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents")); !os.IsNotExist(err) {
		t.Fatalf("expected unsynced codex tree to remain absent, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "commands", "runecontext.md")); !os.IsNotExist(err) {
		t.Fatalf("expected unsynced claude command shim to remain absent, err=%v", err)
	}
	if content, err := os.ReadFile(filepath.Join(root, "runecontext.yaml")); err != nil {
		t.Fatalf("read upgraded config: %v", err)
	} else if !bytes.Contains(content, []byte("runecontext_version: 0.1.0-alpha.10")) {
		t.Fatalf("expected config version bump to alpha.10, got %q", string(content))
	}
}

func TestRunUpgradePreviewRefreshesOnlyPreviouslySyncedToolArtifacts(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")
	root := createEmbeddedProjectForUpgradeTests(t)
	if code := Run([]string{"adapter", "sync", "--path", root, "opencode"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("expected adapter sync success")
	}
	if err := os.Remove(filepath.Join(root, ".opencode", "commands", "runecontext-change-new.md")); err != nil {
		t.Fatalf("remove synced opencode command shim: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected preview success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["state"], "mixed_or_stale_tree"; got != want {
		t.Fatalf("expected stale-tree state %q, got %q", want, got)
	}
	if !hasPlanActionValue(fields, "refresh host-native opencode artifact: created .opencode/commands/runecontext-change-new.md") {
		t.Fatalf("expected synced opencode create action somewhere in plan actions, got %#v", fields)
	}
	for key, value := range fields {
		if len(key) >= len("plan_action_") && key[:len("plan_action_")] == "plan_action_" && (valueHasUnsyncedToolPrefix(value, "claude-code") || valueHasUnsyncedToolPrefix(value, "codex")) {
			t.Fatalf("expected unsynced tools to stay untouched, got %q for %s", value, key)
		}
	}
}

func hasPlanActionValue(fields map[string]string, want string) bool {
	for key, value := range fields {
		if len(key) >= len("plan_action_") && key[:len("plan_action_")] == "plan_action_" && value == want {
			return true
		}
	}
	return false
}

func valueHasUnsyncedToolPrefix(value, tool string) bool {
	return bytes.Contains([]byte(value), []byte("refresh host-native "+tool+" artifact"))
}
