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
	cleanRoot := filepath.Clean(absRoot)
	if !filepath.IsAbs(cleanRoot) {
		return "", fmt.Errorf("root %q must be absolute", absRoot)
	}
	canonicalRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil {
		return "", fmt.Errorf("resolve root symlinks %q: %w", cleanRoot, err)
	}
	var absOutput string
	if filepath.IsAbs(output) {
		absOutput = filepath.Clean(output)
	} else {
		absOutput = filepath.Clean(filepath.Join(canonicalRoot, output))
	}
	if err := validateOutputRoot(canonicalRoot, absOutput); err != nil {
		return "", err
	}
	return absOutput, nil
}

func validateOutputRoot(absRoot, absOutput string) error {
	if absOutput == string(filepath.Separator) {
		return fmt.Errorf("output root %q is not allowed", absOutput)
	}
	rel, err := filepath.Rel(absRoot, absOutput)
	if err != nil {
		return fmt.Errorf("resolve output root %q relative to %q: %w", absOutput, absRoot, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output root %q must stay under repository root %q", absOutput, absRoot)
	}
	if err := rejectSymlinkedOutputAncestors(absRoot, rel); err != nil {
		return err
	}
	return nil
}

func rejectSymlinkedOutputAncestors(absRoot, relOutput string) error {
	if relOutput == "." {
		return nil
	}
	current := absRoot
	for _, segment := range outputSegments(relOutput) {
		current = filepath.Join(current, segment)
		exists, symlinked, err := outputComponentState(current)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		if symlinked {
			return fmt.Errorf("output root component %q must not be a symlink", current)
		}
	}
	return nil
}

func outputSegments(relOutput string) []string {
	parts := strings.Split(relOutput, string(filepath.Separator))
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		segments = append(segments, part)
	}
	return segments
}

func outputComponentState(path string) (exists bool, symlinked bool, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("stat output component %q: %w", path, err)
	}
	return true, info.Mode()&os.ModeSymlink != 0, nil
}

func renderTool(root, output, toolID string, flows []flowDefinition, tool toolDefinition) error {
	sourceDir := filepath.Join(root, "adapters", "source", "packs", toolID)
	if info, err := os.Stat(sourceDir); err != nil || !info.IsDir() {
		return fmt.Errorf("adapter passthrough source directory missing for %q", toolID)
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
