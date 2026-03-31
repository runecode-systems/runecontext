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
		"adapter_role:",
		"operation_identifier:",
		"command_path:",
		"usage:",
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
	for _, operation := range []string{"change-assess-intake", "change-assess-decomposition"} {
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
			expectedPath := "change assess-" + strings.TrimPrefix(operation, "change-assess-")
			if !strings.Contains(text, "command_path: `"+expectedPath+"`") {
				t.Fatalf("expected command path for %s, got %q", operation, text)
			}
		})
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
}

func TestRunAdapterRenderHostNativeRejectsUnsupportedTool(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"adapter", "render-host-native", "codex", "change-new"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "does not support shell-output injection") {
		t.Fatalf("expected unsupported shell injection error, got %q", stderr.String())
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
	for _, token := range []string{"\"command\":\"adapter_render_host_native\"", "\"adapter\":\"opencode\"", "\"operation\":\"change-new\"", "\"body\":\"- canonical_flow_source:"} {
		if !strings.Contains(text, token) {
			t.Fatalf("expected token %q in JSON output, got %q", token, text)
		}
	}
}
