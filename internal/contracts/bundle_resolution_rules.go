package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func (c *BundleCatalog) evaluateExactRule(rule bundleRule) ([]string, []BundleDiagnostic, error) {
	aspectRoot, err := canonicalContainedRoot(filepath.Join(c.Root, string(rule.Aspect)))
	if err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: err.Error()}
	}
	logicalPath := filepath.Join(c.Root, filepath.FromSlash(rule.Pattern))
	if err := ensureExactBundlePathExists(logicalPath, rule); err != nil {
		return []string{}, missingExactRuleDiagnostics(rule, err), nil
	}
	if err := validateResolvedBundlePath(logicalPath, c.Root, aspectRoot); err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q %v", rule.Pattern, err)}
	}
	if err := validateExactBundleFile(logicalPath, rule); err != nil {
		return nil, nil, err
	}
	return []string{rule.Pattern}, nil, nil
}

func ensureExactBundlePathExists(logicalPath string, rule bundleRule) error {
	if _, err := os.Lstat(logicalPath); err != nil {
		return err
	}
	return nil
}

func missingExactRuleDiagnostics(rule bundleRule, err error) []BundleDiagnostic {
	if !os.IsNotExist(err) {
		return nil
	}
	return []BundleDiagnostic{{Severity: DiagnosticSeverityWarning, Code: "missing_exact_path", Message: fmt.Sprintf("exact %s rule did not match an existing file", rule.Kind), Bundle: rule.Bundle, Aspect: rule.Aspect, Rule: rule.Kind, Pattern: rule.Pattern}}
}

func validateExactBundleFile(logicalPath string, rule bundleRule) error {
	info, err := os.Stat(logicalPath)
	if err != nil {
		return &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q could not be evaluated: %v", rule.Pattern, err)}
	}
	if info.IsDir() {
		return &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q resolves to a directory; exact bundle rules must reference files", rule.Pattern)}
	}
	if !info.Mode().IsRegular() {
		return &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q resolves to a non-regular file", rule.Pattern)}
	}
	return nil
}

func (c *BundleCatalog) evaluateGlobRule(rule bundleRule) ([]string, []BundleDiagnostic, error) {
	anchorPath, diagnostics, err := c.globRuleAnchor(rule)
	if err != nil || diagnostics != nil {
		return []string{}, diagnostics, err
	}
	aspectRoot, err := canonicalContainedRoot(filepath.Join(c.Root, string(rule.Aspect)))
	if err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: err.Error()}
	}
	matches, err := c.collectGlobMatches(rule, aspectRoot, anchorPath)
	if err != nil {
		return nil, nil, err
	}
	return matches, emptyGlobDiagnostics(rule, matches), nil
}

func (c *BundleCatalog) globRuleAnchor(rule bundleRule) (string, []BundleDiagnostic, error) {
	anchorPath := filepath.Join(c.Root, filepath.FromSlash(literalBundleAnchor(rule.Pattern)))
	if _, err := os.Lstat(anchorPath); err != nil {
		if os.IsNotExist(err) {
			return anchorPath, []BundleDiagnostic{{Severity: DiagnosticSeverityInfo, Code: "empty_glob_match", Message: fmt.Sprintf("glob %s rule matched no files", rule.Kind), Bundle: rule.Bundle, Aspect: rule.Aspect, Rule: rule.Kind, Pattern: rule.Pattern}}, nil
		}
		return "", nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle glob %q could not be evaluated: %v", rule.Pattern, err)}
	}
	return anchorPath, nil, nil
}

func (c *BundleCatalog) collectGlobMatches(rule bundleRule, aspectRoot, anchorPath string) ([]string, error) {
	matches := make([]string, 0)
	seen := map[string]struct{}{}
	err := walkBundleFiles(c.Root, aspectRoot, anchorPath, map[string]struct{}{}, &bundleWalkState{}, func(logicalPath string) error {
		rel := runeContextRelativePath(c.Root, logicalPath)
		if !matchBundlePattern(rule.Pattern, rel) {
			return nil
		}
		if _, ok := seen[rel]; ok {
			return nil
		}
		seen[rel] = struct{}{}
		matches = append(matches, rel)
		return nil
	})
	if err != nil {
		return nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle glob %q %v", rule.Pattern, err)}
	}
	sort.Strings(matches)
	return matches, nil
}

func emptyGlobDiagnostics(rule bundleRule, matches []string) []BundleDiagnostic {
	if len(matches) > 0 {
		return nil
	}
	return []BundleDiagnostic{{Severity: DiagnosticSeverityInfo, Code: "empty_glob_match", Message: fmt.Sprintf("glob %s rule matched no files", rule.Kind), Bundle: rule.Bundle, Aspect: rule.Aspect, Rule: rule.Kind, Pattern: rule.Pattern}}
}

func extractBundleRules(sourcePath, bundleID string, kind BundleRuleKind, raw any) (map[BundleAspect][]bundleRule, error) {
	result := map[BundleAspect][]bundleRule{}
	if raw == nil {
		return result, nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %s rules must decode to an object", kind)}
	}
	for key, value := range obj {
		aspectRules, err := extractAspectBundleRules(sourcePath, bundleID, kind, key, value)
		if err != nil {
			return nil, err
		}
		result[BundleAspect(key)] = aspectRules
	}
	return result, nil
}

func extractAspectBundleRules(sourcePath, bundleID string, kind BundleRuleKind, key string, value any) ([]bundleRule, error) {
	aspect := BundleAspect(key)
	if !isKnownBundleAspect(aspect) {
		return nil, &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q uses unknown aspect %q", bundleID, key)}
	}
	items, ok := value.([]any)
	if !ok {
		return nil, &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q aspect %q %ss must decode to an array", bundleID, key, kind)}
	}
	rules := make([]bundleRule, 0, len(items))
	for i, item := range items {
		rule, err := normalizeBundleRule(sourcePath, bundleID, kind, aspect, item, i)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func normalizeBundleRule(sourcePath, bundleID string, kind BundleRuleKind, aspect BundleAspect, item any, index int) (bundleRule, error) {
	normalized, patternKind, err := normalizeBundlePattern(aspect, fmt.Sprint(item))
	if err != nil {
		return bundleRule{}, &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q %s rule %q is invalid: %v", bundleID, kind, item, err)}
	}
	return bundleRule{Bundle: bundleID, Aspect: aspect, Kind: kind, Pattern: normalized, RawPattern: fmt.Sprint(item), PatternKind: patternKind, SourcePath: sourcePath, Index: index}, nil
}

func isKnownBundleAspect(aspect BundleAspect) bool {
	for _, known := range bundleAspects {
		if aspect == known {
			return true
		}
	}
	return false
}
