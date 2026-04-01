package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const generatedFlowsDir = "flows"

func run(root, output string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	absOutput, err := resolveOutput(absRoot, output)
	if err != nil {
		return err
	}
	flows, err := loadFlowDefinitions(absRoot)
	if err != nil {
		return err
	}
	tools, err := loadToolDefinitions(absRoot)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(absOutput); err != nil {
		return fmt.Errorf("reset output root %q: %w", absOutput, err)
	}
	if err := os.MkdirAll(absOutput, 0o755); err != nil {
		return fmt.Errorf("create output root %q: %w", absOutput, err)
	}
	for _, toolID := range sortedToolIDs(tools) {
		if err := renderTool(absRoot, absOutput, toolID, flows, tools[toolID]); err != nil {
			return err
		}
	}
	return nil
}

func resolveOutput(absRoot, output string) (string, error) {
	if strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("output root must be non-empty")
	}
	if filepath.IsAbs(output) {
		return filepath.Clean(output), nil
	}
	return filepath.Clean(filepath.Join(absRoot, output)), nil
}

func renderTool(root, output, toolID string, flows []flowDefinition, tool toolDefinition) error {
	sourceDir := filepath.Join(root, "adapters", toolID)
	if info, err := os.Stat(sourceDir); err != nil || !info.IsDir() {
		return fmt.Errorf("adapter source directory missing for %q", toolID)
	}
	targetDir := filepath.Join(output, toolID)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create adapter output %q: %w", targetDir, err)
	}
	excludes := generatedExcludes(flows)
	if err := copyPassthroughTree(sourceDir, targetDir, excludes); err != nil {
		return fmt.Errorf("copy passthrough tree for %q: %w", toolID, err)
	}
	if err := writeCapabilities(targetDir, toolID, tool); err != nil {
		return err
	}
	return writeFlows(targetDir, tool, flows)
}

func generatedExcludes(flows []flowDefinition) map[string]struct{} {
	excludes := map[string]struct{}{
		"capabilities.yaml": {},
	}
	for _, flow := range flows {
		excludes[filepath.ToSlash(filepath.Join(generatedFlowsDir, flow.ID+".md"))] = struct{}{}
	}
	return excludes
}

func copyPassthroughTree(sourceRoot, targetRoot string, excludes map[string]struct{}) error {
	return filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if _, excluded := excludes[rel]; excluded {
			return nil
		}
		targetPath := filepath.Join(targetRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return copyFileWithMode(path, targetPath)
	})
}

func copyFileWithMode(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, info.Mode().Perm())
}

func writeCapabilities(targetDir, toolID string, tool toolDefinition) error {
	path := filepath.Join(targetDir, "capabilities.yaml")
	var lines []string
	lines = append(lines,
		"schema_version: 1",
		"adapter: "+toolID,
		"capabilities:",
	)
	for _, key := range capabilityOrder() {
		lines = append(lines, "  "+key+": "+tool.Capabilities[key])
	}
	lines = append(lines, "fallbacks:")
	for _, key := range capabilityOrder() {
		lines = append(lines, "  "+key+": \""+capabilityFallback(key, tool.Capabilities[key])+"\"")
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func writeFlows(targetDir string, tool toolDefinition, flows []flowDefinition) error {
	flowsDir := filepath.Join(targetDir, generatedFlowsDir)
	if err := os.MkdirAll(flowsDir, 0o755); err != nil {
		return fmt.Errorf("create flows dir: %w", err)
	}
	for _, flow := range flows {
		if err := writeFlowFile(flowsDir, tool, flow); err != nil {
			return err
		}
	}
	return nil
}

func writeFlowFile(flowsDir string, tool toolDefinition, flow flowDefinition) error {
	path := filepath.Join(flowsDir, flow.ID+".md")
	lines := []string{
		"# " + tool.DisplayName + " Flow: " + flow.CommandPath,
		"",
		tool.FlowIntro,
		"",
		"## Intent",
		"",
		"- " + flow.Description,
		"",
		"## Command Mapping",
		"",
		"```sh",
		flow.Usage,
		"```",
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("write flow %q: %w", flow.ID, err)
	}
	return nil
}
