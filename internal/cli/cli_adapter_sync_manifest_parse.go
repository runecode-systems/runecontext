package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func loadPreviousHostNativeFiles(absRoot, manifestPath string) ([]string, error) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	raw := parseManifestPathList(content, "host_native_files")
	validated := make([]string, 0, len(raw))
	for _, rel := range raw {
		normalized, pathErr := validateManifestHostNativePath(absRoot, rel)
		if pathErr != nil {
			return nil, pathErr
		}
		validated = append(validated, normalized)
	}
	return validated, nil
}

func parseManifestPathList(content []byte, section string) []string {
	sectionHeader := section + ":"
	lines := strings.Split(string(content), "\n")
	start := findManifestSectionStart(lines, sectionHeader)
	if start == -1 {
		return nil
	}
	values := collectManifestSectionList(lines[start+1:])
	sort.Strings(values)
	return values
}

func findManifestSectionStart(lines []string, sectionHeader string) int {
	for idx, raw := range lines {
		if strings.TrimSpace(raw) == sectionHeader {
			return idx
		}
	}
	return -1
}

func collectManifestSectionList(lines []string) []string {
	values := make([]string, 0)
	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(raw, " ") {
			break
		}
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func validateManifestHostNativePath(absRoot, rel string) (string, error) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return "", fmt.Errorf("adapter sync manifest host-native path must not be empty")
	}
	if strings.Contains(rel, "\\") {
		return "", fmt.Errorf("adapter sync manifest host-native path %q must use forward slashes", rel)
	}
	if strings.HasPrefix(rel, "/") {
		return "", fmt.Errorf("adapter sync manifest host-native path %q must be relative", rel)
	}
	cleaned := filepath.ToSlash(filepath.Clean(rel))
	if cleaned != rel {
		return "", fmt.Errorf("adapter sync manifest host-native path %q is not canonical", rel)
	}
	if !isHostNativePath(rel) {
		return "", fmt.Errorf("adapter sync manifest host-native path %q is outside supported host-native roots", rel)
	}
	absPath := filepath.Join(absRoot, filepath.FromSlash(rel))
	relToRoot, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", err
	}
	relToRoot = filepath.ToSlash(relToRoot)
	if relToRoot == ".." || strings.HasPrefix(relToRoot, "../") {
		return "", fmt.Errorf("adapter sync manifest host-native path %q escapes repository root", rel)
	}
	return rel, nil
}
