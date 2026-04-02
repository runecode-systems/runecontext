package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAdapterRenderHostNativeHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "--help"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+adapterRenderUsage) {
		t.Fatalf("expected adapter render-host-native help usage, got %q", stdout.String())
	}
}

func TestRunAdapterRenderHostNativeOutputsMinimalMarkdown(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "opencode", "change-new"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	text := stdout.String()
	for _, token := range []string{
		"canonical_flow_source:",
		"canonical_workflow_contract:",
		"adapter_role:",
		"operation_identifier:",
		"command_path:",
		"usage:",
		"required_outcome:",
		"guardrails:",
		"workflow_steps:",
		"stop_condition:",
		"recommended_next_commands:",
		"interaction_rule:",
		hostNativeNoQuestionRule,
	} {
		if !strings.Contains(text, token) {
			t.Fatalf("expected token %q in render output, got %q", token, text)
		}
	}
	if strings.Contains(text, "```") {
		t.Fatalf("expected machine-oriented output without code fences, got %q", text)
	}
}

func TestRunAdapterRenderHostNativeSupportsAssessmentOperations(t *testing.T) {
	for _, operation := range []string{"change-assess-intake", "change-assess-decomposition", "change-decomposition-plan", "change-decomposition-apply"} {
		t.Run(operation, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := Run([]string{"adapter", "render-host-native", "opencode", operation}, &stdout, &stderr)
			if code != exitOK {
				t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
			}
			text := stdout.String()
			if !strings.Contains(text, "operation_identifier: `runecontext:"+operation+"`") {
				t.Fatalf("expected operation identifier for %s, got %q", operation, text)
			}
			expectedPath := expectedCommandPathForOperation(operation)
			if !strings.Contains(text, "command_path: `"+expectedPath+"`") {
				t.Fatalf("expected command path for %s, got %q", operation, text)
			}
		})
	}
}

func expectedCommandPathForOperation(operation string) string {
	switch operation {
	case "change-assess-intake":
		return "change assess-intake"
	case "change-assess-decomposition":
		return "change assess-decomposition"
	case "change-decomposition-plan":
		return "change decomposition-plan"
	case "change-decomposition-apply":
		return "change decomposition-apply"
	default:
		return strings.ReplaceAll(operation, "-", " ")
	}
}

func TestRunAdapterRenderHostNativeIndexForClaude(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "--role", "discoverability-shim", "claude-code", "index"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	text := stdout.String()
	if !strings.Contains(text, "operation_identifier: `runecontext:index`") {
		t.Fatalf("expected index operation identifier, got %q", text)
	}
	if !strings.Contains(text, "operation: `runecontext:change-new`") {
		t.Fatalf("expected indexed change-new operation, got %q", text)
	}
	if !strings.Contains(text, "operation: `runecontext:change-assess-intake`") {
		t.Fatalf("expected indexed change-assess-intake operation, got %q", text)
	}
	if !strings.Contains(text, "operation: `runecontext:change-assess-decomposition`") {
		t.Fatalf("expected indexed change-assess-decomposition operation, got %q", text)
	}
	if !strings.Contains(text, "operation: `runecontext:change-decomposition-plan`") {
		t.Fatalf("expected indexed change-decomposition-plan operation, got %q", text)
	}
	if !strings.Contains(text, "operation: `runecontext:change-decomposition-apply`") {
		t.Fatalf("expected indexed change-decomposition-apply operation, got %q", text)
	}
	if !strings.Contains(text, "interaction_rule: "+hostNativeNoQuestionRule) {
		t.Fatalf("expected no-question interaction rule in index output, got %q", text)
	}
}

func TestRunAdapterRenderHostNativeSupportsCodexRender(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "codex", "change-new"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "operation_identifier: `runecontext:change-new`") {
		t.Fatalf("expected codex flow output, got %q", stdout.String())
	}
}

func TestRunAdapterRenderHostNativeJSONOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "--json", "opencode", "change-new"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	text := stdout.String()
	for _, token := range []string{"\"command\":\"adapter_render_host_native\"", "\"adapter\":\"opencode\"", "\"operation\":\"change-new\"", "\"body\":\"- canonical_flow_source:", "canonical_workflow_contract"} {
		if !strings.Contains(text, token) {
			t.Fatalf("expected token %q in JSON output, got %q", token, text)
		}
	}
}

func TestRunAdapterRenderHostNativeRegeneratesMissingRequestedPack(t *testing.T) {
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	adaptersRoot := filepath.Join(root, "build", "generated", "adapters")
	t.Cleanup(func() {
		regen := exec.Command("go", "run", "./tools/syncadapters", "--root", root, "--output", "build/generated/adapters", "--tool", "opencode")
		regen.Dir = root
		if output, err := regen.CombinedOutput(); err != nil {
			t.Fatalf("restore staged opencode pack: %v\n%s", err, string(output))
		}
	})
	if err := os.RemoveAll(filepath.Join(adaptersRoot, "opencode")); err != nil {
		t.Fatalf("remove staged opencode pack: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir repo root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "opencode", "change-new"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected render-host-native to regenerate missing pack, got %d (%s)", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(adaptersRoot, "opencode", "workflow.json")); err != nil {
		t.Fatalf("expected regenerated opencode workflow contract: %v", err)
	}
	if !strings.Contains(stdout.String(), "operation_identifier: `runecontext:change-new`") {
		t.Fatalf("expected rendered change-new output, got %q", stdout.String())
	}
}

func TestRunAdapterRenderHostNativeUsesInstalledShareCanonicalPaths(t *testing.T) {
	root := t.TempDir()
	runtimeRoot := filepath.Join(root, "share", "runecontext")
	schemaDir := filepath.Join(runtimeRoot, "schemas")
	adaptersDir := filepath.Join(runtimeRoot, "adapters")
	seedReleaseStyleLayout(t, schemaDir, adaptersDir)
	seedAdapterPackForDiscovery(t, adaptersDir, "opencode")

	workflow := `{"schema_version":"1","adapter":"opencode","display_name":"OpenCode","flow_intro":"Intro","flows":[{"id":"change-new","command_path":"change new","description":"Create a new RuneContext change","usage":"runectx change new --title TITLE --type TYPE","required_outcome":"Create a new change directory with a stable change id and initialized change artifacts.","guardrails":["Map every generated suggestion directly to the documented runectx command and flags."],"inputs_to_gather":["change title","change type"],"decision_rules":["If required title or type is missing, ask for those values before proposing command execution."],"workflow_steps":["Collect missing required and optional command inputs."],"stop_condition":"Stop immediately after reporting change creation output and do not chain additional workflow commands.","recommended_next_commands":["runectx change shape <change-id>"],"examples":[{"scenario":"Create feature","user_prompt":"Create change","assistant_response":"I will run runectx change new..."}]}]}`
	if err := os.WriteFile(filepath.Join(adaptersDir, "opencode", "workflow.json"), []byte(workflow), 0o644); err != nil {
		t.Fatalf("write workflow contract: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	if err := os.Chdir(runtimeRoot); err != nil {
		t.Fatalf("chdir runtime root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "opencode", "change-new"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	text := stdout.String()
	if !strings.Contains(text, "canonical_flow_source: `share/runecontext/adapters/opencode/flows/change-new.md`") {
		t.Fatalf("expected share-layout canonical flow source, got %q", text)
	}
	if !strings.Contains(text, "canonical_workflow_contract: `share/runecontext/adapters/opencode/workflow.json`") {
		t.Fatalf("expected share-layout canonical workflow contract, got %q", text)
	}
}
