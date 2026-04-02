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

func TestResolveOutputAllowsAbsoluteSymlinkAliasUnderRepositoryRoot(t *testing.T) {
	realRoot := t.TempDir()
	aliasParent := t.TempDir()
	aliasRoot := filepath.Join(aliasParent, "repo-alias")
	if err := os.Symlink(realRoot, aliasRoot); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create repository alias symlink: %v", err)
	}

	output := filepath.Join(aliasRoot, "build", "generated", "adapters")
	got, err := resolveOutput(realRoot, output)
	if err != nil {
		t.Fatalf("resolveOutput returned error: %v", err)
	}
	expected, err := canonicalizePathAllowMissing(filepath.Join(realRoot, "build", "generated", "adapters"))
	if err != nil {
		t.Fatalf("canonicalize expected output path: %v", err)
	}
	if got != expected {
		t.Fatalf("expected canonical output %q, got %q", expected, got)
	}
}

func assertRepositoryDescendantOutput(t *testing.T, root, output string) {
	t.Helper()
	t.Run(output, func(t *testing.T) {
		got, err := resolveOutput(root, output)
		if err != nil {
			t.Fatalf("resolveOutput returned error: %v", err)
		}
		canonicalRoot, err := canonicalizePathAllowMissing(root)
		if err != nil {
			t.Fatalf("canonicalize root: %v", err)
		}
		rel, err := filepath.Rel(canonicalRoot, got)
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

func TestRunRejectsSymlinkedPassthroughFile(t *testing.T) {
	root := t.TempDir()
	seedSingleToolFixtureRoot(t, root)
	packDir := filepath.Join(root, "adapters", "source", "packs", "opencode")
	outside := filepath.Join(root, "outside.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	symlinkPath := filepath.Join(packDir, "README.md")
	if err := os.Symlink(outside, symlinkPath); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlinked passthrough file: %v", err)
	}

	output := filepath.Join(root, "build", "generated", "adapters")
	err := run(root, output, "opencode")
	if err == nil {
		t.Fatal("expected symlinked passthrough file to be rejected")
	}
	if !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("expected symlink rejection error, got %v", err)
	}
}

func seedSingleToolFixtureRoot(t *testing.T, root string) {
	t.Helper()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n")
	writeFile(t, filepath.Join(root, "adapters", "source", "shared", "flows", "change-new.json"), singleToolFlowDefinitionJSON)
	writeFile(t, filepath.Join(root, "adapters", "source", "tools", "opencode.json"), singleToolDefinitionJSON)
	writeFile(t, filepath.Join(root, "adapters", "source", "packs", "opencode", "flows", "conversational-parity.md"), "ok\n")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

const singleToolFlowDefinitionJSON = `{"id":"change-new","command_path":"change new","description":"desc","usage":"runectx change new --title T --type feature","required_outcome":"create change","guardrails":["g"],"inputs_to_gather":["i"],"decision_rules":["d"],"workflow_steps":["w"],"stop_condition":"stop","recommended_next_commands":["runectx change shape <change-id>"],"examples":[{"scenario":"s","user_prompt":"u","assistant_response":"a"}]}`

const singleToolDefinitionJSON = `{"display_name":"OpenCode","flow_intro":"Intro","capabilities":{"prompts":"supported","shell_access":"supported","hooks":"supported","dynamic_suggestions":"supported","structured_output":"supported"}}`

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
