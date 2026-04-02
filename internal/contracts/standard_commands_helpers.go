package contracts

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func validateWritableStandardCommand(v *Validator, loaded *LoadedProject) error {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return err
	}
	if loaded.Resolution == nil {
		return fmt.Errorf("loaded project resolution is required")
	}
	switch loaded.Resolution.SourceMode {
	case SourceModeEmbedded, SourceModePath:
		return nil
	default:
		return fmt.Errorf("standard write operations are only supported for embedded and local path sources")
	}
}

func normalizeStandardArtifactPath(path string) (string, error) {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return "", fmt.Errorf("standard path is required")
	}
	if hasParentTraversalSegment(trimmed) {
		return "", fmt.Errorf("standard path %q must not traverse outside standards/", path)
	}
	normalized := canonicalStandardPath(trimmed)
	if strings.HasPrefix(normalized, "standards/") {
		return validateCanonicalStandardPath(path, normalized)
	}
	normalized = filepath.ToSlash(filepath.Join("standards", normalized))
	if !strings.HasSuffix(normalized, ".md") {
		normalized += ".md"
	}
	return validateCanonicalStandardPath(path, normalized)
}

func canonicalStandardPath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "runecontext/")
	return normalized
}

func hasParentTraversalSegment(path string) bool {
	for _, segment := range strings.Split(filepath.ToSlash(path), "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func validateCanonicalStandardPath(inputPath, normalized string) (string, error) {
	if !strings.HasSuffix(normalized, ".md") {
		return "", fmt.Errorf("standard path %q must end with .md", inputPath)
	}
	if !isCanonicalStandardPathRef(normalized) {
		return "", fmt.Errorf("standard path %q must use canonical standards/<path>.md form", inputPath)
	}
	return normalized, nil
}

func normalizeStandardScopePaths(items []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		normalized, err := normalizeStandardScopePath(item)
		if err != nil {
			return nil, err
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out, nil
}

func normalizeStandardScopePath(value string) (string, error) {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return "", nil
	}
	if hasParentTraversalSegment(trimmed) {
		return "", fmt.Errorf("invalid --scope-path %q: path traversal outside standards is not allowed", value)
	}
	normalized := canonicalStandardPath(trimmed)
	if normalized == "." || normalized == "" || normalized == "runecontext" {
		return "", nil
	}
	if normalized != "standards" && !strings.HasPrefix(normalized, "standards/") {
		normalized = filepath.ToSlash(filepath.Join("standards", normalized))
	}
	if normalized != "standards" && !strings.HasPrefix(normalized, "standards/") {
		return "", fmt.Errorf("invalid --scope-path %q: normalized path must stay under standards/", value)
	}
	return normalized, nil
}

func normalizeStandardStatuses(items []StandardStatus) ([]StandardStatus, error) {
	seen := map[StandardStatus]struct{}{}
	out := make([]StandardStatus, 0, len(items))
	for _, status := range items {
		trimmed := StandardStatus(strings.TrimSpace(string(status)))
		if trimmed == "" {
			continue
		}
		if err := validateStandardStatus(trimmed); err != nil {
			return nil, err
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out, nil
}

func validateStandardStatus(status StandardStatus) error {
	switch status {
	case StandardStatusDraft, StandardStatusActive, StandardStatusDeprecated:
		return nil
	default:
		return fmt.Errorf("standard status must be one of draft, active, or deprecated")
	}
}

func validateStandardReplacementFields(status StandardStatus, replacedBy string) error {
	if replacedBy == "" {
		return nil
	}
	if status != StandardStatusDeprecated {
		return fmt.Errorf("--replaced-by requires --status deprecated")
	}
	if !isCanonicalStandardPathRef(replacedBy) {
		return fmt.Errorf("--replaced-by must use canonical standards/<path>.md form")
	}
	return nil
}

func normalizeOptionalStandardReplacement(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	return normalizeStandardArtifactPath(trimmed)
}

func standardIDFromPath(path string) string {
	if !strings.HasPrefix(path, "standards/") || !strings.HasSuffix(path, ".md") {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(path, "standards/"), ".md")
}

func renderStandardDocument(frontmatter standardFrontmatter, body string) ([]byte, error) {
	frontmatterData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, err
	}
	body = strings.TrimSpace(body)
	if body == "" {
		body = fmt.Sprintf("# %s\n\nDescribe the standard intent and requirements.", frontmatter.Title)
	}
	text := strings.Join([]string{"---", strings.TrimSuffix(string(frontmatterData), "\n"), "---", "", body, ""}, "\n")
	return []byte(text), nil
}

func dedupeStringsInOrder(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
