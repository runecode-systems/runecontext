package contracts

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func normalizeBundlePattern(aspect BundleAspect, raw string) (string, BundlePatternKind, error) {
	value, err := validateBundlePatternInput(raw)
	if err != nil {
		return "", "", err
	}
	cleaned, err := sanitizeBundlePatternPath(value)
	if err != nil {
		return "", "", err
	}
	rooted, err := rootBundlePatternToAspect(cleaned, aspect)
	if err != nil {
		return "", "", err
	}
	return classifyBundlePatternKind(rooted)
}

func validateBundlePatternInput(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("must not be empty")
	}
	value = strings.ReplaceAll(value, "\\", "/")
	if strings.HasPrefix(value, "/") || filepath.IsAbs(value) || isDriveQualifiedPath(value) {
		return "", fmt.Errorf("must not be absolute or drive-qualified")
	}
	return value, nil
}

func sanitizeBundlePatternPath(value string) (string, error) {
	for _, segment := range strings.Split(value, "/") {
		if segment == ".." {
			return "", fmt.Errorf("must not contain traversal segments")
		}
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("must not be empty")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("must not contain traversal segments")
	}
	return cleaned, nil
}

func rootBundlePatternToAspect(cleaned string, aspect BundleAspect) (string, error) {
	if cleaned == string(aspect) || strings.HasPrefix(cleaned, string(aspect)+"/") {
		if cleaned == string(aspect) {
			return "", fmt.Errorf("must reference a file path or glob beneath the aspect root")
		}
		return cleaned, nil
	}
	if bundlePatternUsesOtherAspect(cleaned, aspect) {
		return "", fmt.Errorf("must stay within the %q aspect", aspect)
	}
	return string(aspect) + "/" + cleaned, nil
}

func classifyBundlePatternKind(rooted string) (string, BundlePatternKind, error) {
	hasWildcard := false
	for _, segment := range strings.Split(rooted, "/") {
		if !strings.Contains(segment, "*") {
			continue
		}
		hasWildcard = true
		if segment != "*" && segment != "**" {
			return "", "", fmt.Errorf("wildcards must use whole-segment '*' or '**'")
		}
	}
	if hasWildcard {
		return rooted, BundlePatternKindGlob, nil
	}
	return rooted, BundlePatternKindExact, nil
}

func bundlePatternUsesOtherAspect(pattern string, aspect BundleAspect) bool {
	for _, candidate := range bundleAspects {
		if candidate == aspect {
			continue
		}
		prefix := string(candidate)
		if pattern == prefix || strings.HasPrefix(pattern, prefix+"/") {
			return true
		}
	}
	return false
}

func isDriveQualifiedPath(value string) bool {
	if len(value) < 2 {
		return false
	}
	return ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) && value[1] == ':'
}

func literalBundleAnchor(pattern string) string {
	segments := strings.Split(pattern, "/")
	anchor := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "*" || segment == "**" {
			break
		}
		anchor = append(anchor, segment)
	}
	if len(anchor) == 0 {
		return "."
	}
	return path.Clean(strings.Join(anchor, "/"))
}

func matchBundlePattern(pattern, candidate string) bool {
	matcher := newBundlePatternMatcher(pattern, candidate)
	return matcher.match(0, 0)
}

type bundlePatternMatcher struct {
	patternSegments   []string
	candidateSegments []string
	cache             map[bundlePatternState]bool
	seen              map[bundlePatternState]bool
}

type bundlePatternState struct{ i, j int }

func newBundlePatternMatcher(pattern, candidate string) *bundlePatternMatcher {
	return &bundlePatternMatcher{patternSegments: strings.Split(pattern, "/"), candidateSegments: strings.Split(candidate, "/"), cache: map[bundlePatternState]bool{}, seen: map[bundlePatternState]bool{}}
}

func (m *bundlePatternMatcher) match(i, j int) bool {
	state := bundlePatternState{i: i, j: j}
	if m.seen[state] {
		return m.cache[state]
	}
	m.seen[state] = true
	result := m.matchState(i, j)
	m.cache[state] = result
	return result
}

func (m *bundlePatternMatcher) matchState(i, j int) bool {
	if i == len(m.patternSegments) {
		return j == len(m.candidateSegments)
	}
	if m.patternSegments[i] == "**" {
		return m.matchDoubleStar(i, j)
	}
	if j >= len(m.candidateSegments) {
		return false
	}
	if m.patternSegments[i] == "*" || m.patternSegments[i] == m.candidateSegments[j] {
		return m.match(i+1, j+1)
	}
	return false
}

func (m *bundlePatternMatcher) matchDoubleStar(i, j int) bool {
	if i == len(m.patternSegments)-1 {
		return true
	}
	for next := j; next <= len(m.candidateSegments); next++ {
		if m.match(i+1, next) {
			return true
		}
	}
	return false
}
