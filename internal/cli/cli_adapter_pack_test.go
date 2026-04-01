package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdapterPackDocsExist(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}

	adaptersRoot := filepath.Join(repoRoot, "adapters")
	if info, err := os.Stat(adaptersRoot); err != nil {
		t.Fatalf("adapters root: %v", err)
	} else if !info.IsDir() {
		t.Fatalf("adapters root is not a directory: %s", adaptersRoot)
	}

	entries, err := os.ReadDir(adaptersRoot)
	if err != nil {
		t.Fatalf("list adapters: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "source" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			runAdapterDocChecks(t, adaptersRoot, name)
		})
	}
}

func runAdapterDocChecks(t *testing.T, adaptersRoot, name string) {
	t.Helper()
	base := filepath.Join(adaptersRoot, name)
	info, err := os.Stat(base)
	if err != nil {
		t.Fatalf("adapter directory not found: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("adapter path is not a directory: %s", base)
	}

	checkAdapterReadme(t, base, name)
	checkAdapterCapabilities(t, base, name)
	checkAdapterParityDoc(t, base, name)
	checkAdapterFlowPlaybooks(t, base, name)
}

func checkAdapterReadme(t *testing.T, base, name string) {
	t.Helper()
	readme := filepath.Join(base, "README.md")
	content, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("README missing for adapter %s: %v", name, err)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		t.Fatalf("README.md for adapter %s is empty", name)
	}
	text := string(content)
	for _, required := range []string{"## Scope", "## Capability Declaration"} {
		if !strings.Contains(text, required) {
			t.Fatalf("README.md for adapter %s missing section %q", name, required)
		}
	}
}

func checkAdapterCapabilities(t *testing.T, base, name string) {
	t.Helper()
	path := filepath.Join(base, "capabilities.yaml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("capabilities.yaml missing for adapter %s: %v", name, err)
	}
	text := string(content)
	for _, required := range []string{"schema_version:", "adapter:", "capabilities:", "prompts:", "shell_access:", "hooks:", "dynamic_suggestions:", "structured_output:"} {
		if !strings.Contains(text, required) {
			t.Fatalf("capabilities.yaml for adapter %s missing %q", name, required)
		}
	}
}

func checkAdapterParityDoc(t *testing.T, base, name string) {
	t.Helper()
	parityRel := filepath.Join("flows", "conversational-parity.md")
	parityPath := filepath.Join(base, parityRel)
	info, err := os.Stat(parityPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("conversational parity doc missing for adapter %s", name)
		}
		t.Fatalf("conversational parity doc error for adapter %s: %v", name, err)
	}
	if info.IsDir() {
		t.Fatalf("conversational parity path for adapter %s is a directory", name)
	}

	content, err := os.ReadFile(parityPath)
	if err != nil {
		t.Fatalf("read conversational parity doc for adapter %s: %v", name, err)
	}
	if len(bytes.TrimSpace(content)) == 0 {
		t.Fatalf("conversational parity doc for adapter %s is empty", name)
	}

	assertParitySections(t, string(content), parityRel)
}

func assertParitySections(t *testing.T, text, path string) {
	t.Helper()
	assertRequiredSections(t, text, path)
	assertHostCapabilitiesSection(t, text, path)
	assertFlowMappings(t, text, path)
	assertStandardDiscoverMappings(t, text, path)
}

func assertRequiredSections(t *testing.T, text, path string) {
	t.Helper()
	sections := []string{"## Mapping Rule", "## Flow Mappings", "## Candidate Data Rule", "## Reviewability", "## Host Capabilities"}
	for _, sec := range sections {
		if !strings.Contains(text, sec) {
			t.Fatalf("%s missing section %q", path, sec)
		}
	}
}

func assertHostCapabilitiesSection(t *testing.T, text, path string) {
	t.Helper()
	if hostCount := strings.Count(text, "## Host Capabilities"); hostCount != 1 {
		t.Fatalf("expected exactly one Host Capabilities section in %s, got %d", path, hostCount)
	}
	bullets := []string{"- Prompts:", "- Shell access:", "- Hooks:", "- Dynamic suggestions:", "- Structured output:"}
	for _, bullet := range bullets {
		if !strings.Contains(text, bullet) {
			t.Fatalf("%s missing Host Capabilities detail %q", path, bullet)
		}
	}
}

func assertFlowMappings(t *testing.T, text, path string) {
	t.Helper()
	for _, flow := range []string{"change new", "change assess-intake", "change assess-decomposition", "change decomposition-plan", "change decomposition-apply", "change shape", "standard discover", "promote"} {
		if !strings.Contains(text, flow) {
			t.Fatalf("%s missing conversational flow mapping for %q", path, flow)
		}
	}
}

func assertStandardDiscoverMappings(t *testing.T, text, path string) {
	t.Helper()
	for _, requiredFlag := range []string{"--scope-path", "--focus"} {
		if !strings.Contains(text, requiredFlag) {
			t.Fatalf("%s missing standard discover mapping for %q", path, requiredFlag)
		}
	}
}

func checkAdapterFlowPlaybooks(t *testing.T, base, name string) {
	t.Helper()
	required := []string{"change-new.md", "change-assess-intake.md", "change-assess-decomposition.md", "change-decomposition-plan.md", "change-decomposition-apply.md", "change-shape.md", "standard-discover.md", "promote.md"}
	for _, file := range required {
		path := filepath.Join(base, "flows", file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("adapter flow playbook missing for %s (%s): %v", name, file, err)
		}
		if len(bytes.TrimSpace(content)) == 0 {
			t.Fatalf("adapter flow playbook empty for %s (%s)", name, file)
		}
		if !strings.Contains(string(content), "runectx ") {
			t.Fatalf("adapter flow playbook for %s (%s) missing explicit runectx mapping", name, file)
		}
	}
}
