package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const generatedFlowsDir = "flows"

func run(root, output, toolID string) error {
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
	trimmedToolID := strings.TrimSpace(toolID)
	if trimmedToolID != "" {
		return runSingleTool(absRoot, absOutput, trimmedToolID, flows, tools)
	}
	if err := os.RemoveAll(absOutput); err != nil {
		return fmt.Errorf("reset output root %q: %w", absOutput, err)
	}
	if err := os.MkdirAll(absOutput, 0o755); err != nil {
		return fmt.Errorf("create output root %q: %w", absOutput, err)
	}
	for _, id := range sortedToolIDs(tools) {
		if err := renderTool(absRoot, absOutput, id, flows, tools[id]); err != nil {
			return err
		}
	}
	return nil
}

func runSingleTool(absRoot, absOutput, toolID string, flows []flowDefinition, tools map[string]toolDefinition) error {
	if _, ok := tools[toolID]; !ok {
		return fmt.Errorf("unknown tool %q", toolID)
	}
	if err := prepareSingleToolOutput(absOutput, toolID); err != nil {
		return err
	}
	return renderRequestedTool(absRoot, absOutput, toolID, flows, tools)
}

func prepareSingleToolOutput(absOutput, toolID string) error {
	toolDir := filepath.Join(absOutput, strings.TrimSpace(toolID))
	if err := ensurePathWithinRoot(absOutput, toolDir); err != nil {
		return err
	}
	if err := os.RemoveAll(toolDir); err != nil {
		return fmt.Errorf("reset adapter output %q: %w", toolDir, err)
	}
	return nil
}

func ensurePathWithinRoot(root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("resolve %q relative to %q: %w", path, root, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %q must remain under root %q", path, root)
	}
	return nil
}

func renderRequestedTool(absRoot, absOutput, toolID string, flows []flowDefinition, tools map[string]toolDefinition) error {
	toolID = strings.TrimSpace(toolID)
	tool, ok := tools[toolID]
	if !ok {
		return fmt.Errorf("unknown tool %q", toolID)
	}
	return renderTool(absRoot, absOutput, toolID, flows, tool)
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
	canonicalOutput, err := canonicalizePathAllowMissing(absOutput)
	if err != nil {
		return "", err
	}
	if err := validateOutputRoot(canonicalRoot, canonicalOutput); err != nil {
		return "", err
	}
	return canonicalOutput, nil
}

func canonicalizePathAllowMissing(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("path %q must be absolute", path)
	}
	current := cleanPath
	missing := make([]string, 0, 4)
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			return appendMissingSegments(resolved, missing), nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve output root symlinks %q: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("resolve output root symlinks %q: %w", cleanPath, err)
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func appendMissingSegments(base string, missing []string) string {
	resolved := base
	for i := len(missing) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, missing[i])
	}
	return filepath.Clean(resolved)
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
	if err := writeFlows(targetDir, tool, flows); err != nil {
		return err
	}
	return writeWorkflowContract(targetDir, toolID, tool, flows)
}

func generatedExcludes(flows []flowDefinition) map[string]struct{} {
	excludes := map[string]struct{}{
		"capabilities.yaml": {},
		"workflow.json":     {},
	}
	for _, flow := range flows {
		excludes[filepath.ToSlash(filepath.Join(generatedFlowsDir, flow.ID+".md"))] = struct{}{}
	}
	return excludes
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
