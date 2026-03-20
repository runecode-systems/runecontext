package contracts

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
)

func writeGeneratedYAML(path string, doc any) error {
	data, err := renderGeneratedYAML(doc)
	if err != nil {
		return err
	}
	return writeFileAtomically(path, data, 0o644)
}

func renderGeneratedYAML(doc any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeYAMLDocument(&buf, doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sortedBundleIDs(index *ProjectIndex) []string {
	if index == nil || index.Bundles == nil {
		return []string{}
	}
	return SortedKeys(index.Bundles.bundles)
}

func resolvedBundleParents(resolution *BundleResolution, bundleID string) []string {
	if resolution == nil {
		return []string{}
	}
	parents := make([]string, 0, len(resolution.Linearization))
	for _, id := range resolution.Linearization {
		if id == bundleID {
			continue
		}
		parents = append(parents, id)
	}
	return parents
}

func generatedBundleAspectPatterns(bundle *bundleDefinition, aspect BundleAspect) GeneratedBundleAspectPatterns {
	result := GeneratedBundleAspectPatterns{Includes: []GeneratedBundlePattern{}, Excludes: []GeneratedBundlePattern{}}
	if bundle == nil {
		return result
	}
	for _, rule := range bundle.Includes[aspect] {
		result.Includes = append(result.Includes, GeneratedBundlePattern{Pattern: rule.Pattern, Kind: rule.PatternKind})
	}
	for _, rule := range bundle.Excludes[aspect] {
		result.Excludes = append(result.Excludes, GeneratedBundlePattern{Pattern: rule.Pattern, Kind: rule.PatternKind})
	}
	return result
}

func generatedRelativeArtifactPath(root, targetPath string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("generated artifacts require a content root")
	}
	if strings.TrimSpace(targetPath) == "" {
		return "", fmt.Errorf("generated artifacts require a target path")
	}
	rel, err := filepath.Rel(root, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve relative path for %q: %w", targetPath, err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "", fmt.Errorf("resolve relative path for %q: empty relative output", targetPath)
	}
	if strings.HasPrefix(rel, "../") || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") {
		return "", fmt.Errorf("path %q escapes RuneContext content root", targetPath)
	}
	return rel, nil
}
