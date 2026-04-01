package cli

import (
	"bytes"
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
