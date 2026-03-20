package contracts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func buildContextPackInventories(contentRoot string, resolution *BundleResolution) (ContextPackAspectSet, ContextPackExcludedAspectSet, []contextPackFileDigest, error) {
	selected := newContextPackAspectSet()
	excluded := newContextPackExcludedAspectSet()
	digests := make([]contextPackFileDigest, 0)
	if resolution == nil {
		return selected, excluded, digests, fmt.Errorf("bundle resolution is required")
	}
	for _, aspect := range bundleAspects {
		aspectResolution := resolution.Aspects[aspect]
		selectedItems, selectedDigests, err := buildContextPackSelectedFiles(contentRoot, aspectResolution.Selected)
		if err != nil {
			return ContextPackAspectSet{}, ContextPackExcludedAspectSet{}, nil, err
		}
		excludedItems := buildContextPackExcludedFiles(aspectResolution.Excluded)
		assignContextPackSelectedAspect(&selected, aspect, selectedItems)
		assignContextPackExcludedAspect(&excluded, aspect, excludedItems)
		digests = append(digests, selectedDigests...)
	}
	return selected, excluded, digests, nil
}

func newContextPackAspectSet() ContextPackAspectSet {
	return ContextPackAspectSet{
		Project:   []ContextPackSelectedFile{},
		Standards: []ContextPackSelectedFile{},
		Specs:     []ContextPackSelectedFile{},
		Decisions: []ContextPackSelectedFile{},
	}
}

func newContextPackExcludedAspectSet() ContextPackExcludedAspectSet {
	return ContextPackExcludedAspectSet{
		Project:   []ContextPackExcludedFile{},
		Standards: []ContextPackExcludedFile{},
		Specs:     []ContextPackExcludedFile{},
		Decisions: []ContextPackExcludedFile{},
	}
}

func assignContextPackSelectedAspect(target *ContextPackAspectSet, aspect BundleAspect, items []ContextPackSelectedFile) {
	switch aspect {
	case BundleAspectProject:
		target.Project = items
	case BundleAspectStandards:
		target.Standards = items
	case BundleAspectSpecs:
		target.Specs = items
	case BundleAspectDecisions:
		target.Decisions = items
	}
}

func assignContextPackExcludedAspect(target *ContextPackExcludedAspectSet, aspect BundleAspect, items []ContextPackExcludedFile) {
	switch aspect {
	case BundleAspectProject:
		target.Project = items
	case BundleAspectStandards:
		target.Standards = items
	case BundleAspectSpecs:
		target.Specs = items
	case BundleAspectDecisions:
		target.Decisions = items
	}
}

func buildContextPackSelectedFiles(contentRoot string, entries []BundleInventoryEntry) ([]ContextPackSelectedFile, []contextPackFileDigest, error) {
	result := make([]ContextPackSelectedFile, 0, len(entries))
	digests := make([]contextPackFileDigest, 0, len(entries))
	for _, entry := range entries {
		if len(entry.MatchedBy) == 0 {
			return nil, nil, fmt.Errorf("selected context-pack file %q is missing selector provenance", entry.Path)
		}
		digest, err := digestContextPackFile(contentRoot, entry.Path)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, ContextPackSelectedFile{
			Path:       entry.Path,
			SHA256:     digest.SHA256,
			SelectedBy: contextPackRuleReferences(entry.MatchedBy),
		})
		digests = append(digests, digest)
	}
	return result, digests, nil
}

func buildContextPackExcludedFiles(entries []BundleInventoryEntry) []ContextPackExcludedFile {
	result := make([]ContextPackExcludedFile, 0, len(entries))
	for _, entry := range entries {
		result = append(result, ContextPackExcludedFile{
			Path:     entry.Path,
			LastRule: contextPackRuleReference(entry.FinalRule),
		})
	}
	return result
}

func contextPackRuleReferences(items []BundleRuleReference) []ContextPackRuleReference {
	result := make([]ContextPackRuleReference, len(items))
	for i, item := range items {
		result[i] = contextPackRuleReference(item)
	}
	return result
}

func contextPackRuleReference(item BundleRuleReference) ContextPackRuleReference {
	return ContextPackRuleReference{
		Bundle:  item.Bundle,
		Aspect:  item.Aspect,
		Rule:    item.Rule,
		Pattern: item.Pattern,
		Kind:    item.Kind,
	}
}

func digestContextPackFile(contentRoot, relativePath string) (contextPackFileDigest, error) {
	fullPath := filepath.Join(contentRoot, filepath.FromSlash(relativePath))
	data, err := readContextPackProjectFile(contentRoot, fullPath)
	if err != nil {
		return contextPackFileDigest{}, fmt.Errorf("hash context-pack file %q: %w", relativePath, err)
	}
	data = normalizeContextPackFileContent(data)
	sum := sha256.Sum256(data)
	return contextPackFileDigest{Path: relativePath, SHA256: fmt.Sprintf("%x", sum[:]), ReferencedBytes: int64(len(data))}, nil
}

func isPortableLocalSourceRef(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || hasDisallowedLocalSourcePrefix(trimmed) || hasDisallowedLocalSourceSeparator(trimmed) || hasDriveQualifiedPrefix(trimmed) {
		return false
	}
	return !hasTraversalSegments(trimmed)
}

func hasDisallowedLocalSourcePrefix(value string) bool {
	return filepath.IsAbs(value) || strings.HasPrefix(value, "/") || strings.HasPrefix(value, `\\`) || strings.HasPrefix(value, "//")
}

func hasDisallowedLocalSourceSeparator(value string) bool {
	return strings.Contains(value, `\`)
}

func hasDriveQualifiedPrefix(value string) bool {
	if len(value) < 2 || value[1] != ':' {
		return false
	}
	prefix := value[0]
	return (prefix >= 'A' && prefix <= 'Z') || (prefix >= 'a' && prefix <= 'z')
}

func hasTraversalSegments(value string) bool {
	cleaned := path.Clean(value)
	return cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(value, "/./") || strings.Contains(value, "/../") || strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") || strings.HasSuffix(value, "/.") || strings.HasSuffix(value, "/..")
}

func normalizeContextPackFileContent(data []byte) []byte {
	if !looksLikePortableText(data) || !bytes.Contains(data, []byte{'\r'}) {
		return data
	}
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(normalized, []byte{'\r'}, []byte{'\n'})
}

func looksLikePortableText(data []byte) bool {
	return utf8.Valid(data) && !bytes.Contains(data, []byte{0})
}
