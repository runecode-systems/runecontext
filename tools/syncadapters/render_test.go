package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOutputRejectsUnsafeTargets(t *testing.T) {
	root := t.TempDir()
	for _, output := range []string{string(filepath.Separator), root, filepath.Join(root, "..", "outside")} {
		output := output
		t.Run(output, func(t *testing.T) {
			if _, err := resolveOutput(root, output); err == nil {
				t.Fatalf("expected resolveOutput to reject %q", output)
			}
		})
	}
}

func TestResolveOutputAllowsRepositoryDescendants(t *testing.T) {
	root := t.TempDir()
	assertRepositoryDescendantOutput(t, root, "build/generated/adapters")
	assertRepositoryDescendantOutput(t, root, filepath.Join(root, "build", "generated", "adapters"))
}

func TestResolveOutputRejectsSymlinkedOutputAncestor(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "build")); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlinked output ancestor: %v", err)
	}
	if _, err := resolveOutput(root, "build/generated/adapters"); err == nil {
		t.Fatal("expected resolveOutput to reject symlinked output ancestor")
	}
}

func assertRepositoryDescendantOutput(t *testing.T, root, output string) {
	t.Helper()
	t.Run(output, func(t *testing.T) {
		got, err := resolveOutput(root, output)
		if err != nil {
			t.Fatalf("resolveOutput returned error: %v", err)
		}
		rel, err := filepath.Rel(root, got)
		if err != nil {
			t.Fatalf("compute relative path: %v", err)
		}
		if rel == "." || rel == ".." || rel == "" {
			t.Fatalf("expected repository descendant output, got %q", got)
		}
	})
}

func TestRunGeneratesStructuredFlowContracts(t *testing.T) {
	root := repoRootForSyncAdaptersTests(t)
	output := filepath.Join(root, "build", "generated", "adapters-test", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(root, "build", "generated", "adapters-test"))
	})
	if err := run(root, output, ""); err != nil {
		t.Fatalf("run syncadapters: %v", err)
	}

	assertWorkflowContractGenerated(t, output)
	assertFlowMarkdownSections(t, output)
}

func TestRunGeneratesSingleRequestedTool(t *testing.T) {
	root := repoRootForSyncAdaptersTests(t)
	output := filepath.Join(root, "build", "generated", "adapters-test", t.Name())
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(root, "build", "generated", "adapters-test"))
	})
	if err := run(root, output, "opencode"); err != nil {
		t.Fatalf("run syncadapters for one tool: %v", err)
	}
	if _, err := os.Stat(filepath.Join(output, "opencode", "workflow.json")); err != nil {
		t.Fatalf("expected opencode workflow contract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(output, "codex")); !os.IsNotExist(err) {
		t.Fatalf("expected only requested tool output, codex err=%v", err)
	}
}

func assertWorkflowContractGenerated(t *testing.T, output string) {
	t.Helper()
	workflowPath := filepath.Join(output, "opencode", "workflow.json")
	raw, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read workflow contract: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode workflow contract: %v", err)
	}
	if payload["schema_version"] != "1" {
		t.Fatalf("expected schema_version 1, got %#v", payload["schema_version"])
	}
	if payload["adapter"] != "opencode" {
		t.Fatalf("expected adapter opencode, got %#v", payload["adapter"])
	}
}

func assertFlowMarkdownSections(t *testing.T, output string) {
	t.Helper()
	flowPath := filepath.Join(output, "opencode", "flows", "change-new.md")
	content, err := os.ReadFile(flowPath)
	if err != nil {
		t.Fatalf("read generated flow markdown: %v", err)
	}
	text := string(content)
	for _, section := range []string{
		"## Required Outcome",
		"## Guardrails",
		"## Inputs To Gather",
		"## Decision Rules",
		"## Workflow Steps",
		"## Stop Condition",
		"## Recommended Next Commands",
		"## Examples",
	} {
		if !strings.Contains(text, section) {
			t.Fatalf("expected section %q in generated flow markdown", section)
		}
	}
}

func repoRootForSyncAdaptersTests(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		next := filepath.Dir(wd)
		if next == wd {
			t.Fatal("could not locate repository root")
		}
		wd = next
	}
}
