package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func discoverConfig(path string, mode ConfigDiscoveryMode) (string, string, error) {
	start, err := resolveConfigSearchStart(path)
	if err != nil {
		return "", "", err
	}
	switch mode {
	case ConfigDiscoveryNearestAncestor:
		return discoverNearestConfig(start)
	case ConfigDiscoveryExplicitRoot:
		return discoverExplicitConfig(start)
	default:
		return "", "", &ValidationError{Path: start, Message: fmt.Sprintf("unsupported config discovery mode %q", mode)}
	}
}

func resolveConfigSearchStart(path string) (string, error) {
	start, err := filepath.Abs(path)
	if err != nil {
		return "", &ValidationError{Path: path, Message: err.Error()}
	}
	info, err := os.Stat(start)
	if err == nil && !info.IsDir() {
		return filepath.Dir(start), nil
	}
	return start, nil
}

func discoverNearestConfig(start string) (string, string, error) {
	current := start
	for {
		candidate := filepath.Join(current, "runecontext.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Clean(candidate), filepath.Clean(current), nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", "", &ValidationError{Path: start, Message: "no runecontext.yaml found in current directory or ancestors"}
		}
		current = next
	}
}

func discoverExplicitConfig(start string) (string, string, error) {
	candidate := filepath.Join(start, "runecontext.yaml")
	if _, err := os.Stat(candidate); err != nil {
		return "", "", &ValidationError{Path: candidate, Message: err.Error()}
	}
	return filepath.Clean(candidate), filepath.Clean(start), nil
}

func resolveEmbeddedSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared, err := normalizeContainedRelativePath(rawPath)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %v", err)}
	}
	return validateEmbeddedSourceRoot(configPath, projectRoot, rawPath, declared)
}

func validateEmbeddedSourceRoot(configPath, projectRoot, rawPath, declared string) (string, string, error) {
	absRoot := filepath.Clean(filepath.Join(projectRoot, filepath.FromSlash(declared)))
	resolvedProjectRoot, resolvedRoot, err := canonicalizePaths(projectRoot, absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !isWithinRoot(resolvedProjectRoot, resolvedRoot) {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %q escapes the selected project root", rawPath)}
	}
	if err := validateResolvedDirectory(configPath, absRoot); err != nil {
		return "", "", err
	}
	return declared, absRoot, nil
}

func resolveDeclaredLocalSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared := cleanSourceRootValue(rawPath)
	absRoot := rawPath
	if !filepath.IsAbs(rawPath) {
		absRoot = filepath.Join(projectRoot, rawPath)
	}
	absRoot = filepath.Clean(absRoot)
	if err := validateResolvedDirectory(configPath, absRoot); err != nil {
		return "", "", err
	}
	return declared, absRoot, nil
}

func validateResolvedDirectory(configPath, absRoot string) error {
	info, err := os.Stat(absRoot)
	if err != nil {
		return &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		return &ValidationError{Path: configPath, Message: fmt.Sprintf("resolved source root %q is not a directory", absRoot)}
	}
	return nil
}

func normalizeContainedRelativePath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("must not be empty")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("must not be absolute")
	}
	cleaned := filepath.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("must not escape its containing root")
	}
	return filepath.ToSlash(cleaned), nil
}

func cleanSourceRootValue(value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func canonicalizePaths(root, target string) (string, string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve root %q: %w", root, err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", "", fmt.Errorf("resolve path %q: %w", target, err)
	}
	return resolvedRoot, resolvedTarget, nil
}
