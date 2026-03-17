package contracts

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const maxBundleInheritanceDepth = 8

var bundleTraversalLimits = snapshotLimits{
	MaxFiles: 10000,
	MaxBytes: 0,
	MaxDepth: 64,
}

type BundleAspect string

const (
	BundleAspectProject   BundleAspect = "project"
	BundleAspectStandards BundleAspect = "standards"
	BundleAspectSpecs     BundleAspect = "specs"
	BundleAspectDecisions BundleAspect = "decisions"
)

var bundleAspects = []BundleAspect{
	BundleAspectProject,
	BundleAspectStandards,
	BundleAspectSpecs,
	BundleAspectDecisions,
}

type BundleRuleKind string

const (
	BundleRuleKindInclude BundleRuleKind = "include"
	BundleRuleKindExclude BundleRuleKind = "exclude"
)

type BundlePatternKind string

const (
	BundlePatternKindExact BundlePatternKind = "exact"
	BundlePatternKindGlob  BundlePatternKind = "glob"
)

type BundleDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
	Bundle   string             `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Aspect   BundleAspect       `json:"aspect,omitempty" yaml:"aspect,omitempty"`
	Rule     BundleRuleKind     `json:"rule,omitempty" yaml:"rule,omitempty"`
	Pattern  string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Matches  []string           `json:"matches,omitempty" yaml:"matches,omitempty"`
}

type BundleRuleReference struct {
	Bundle  string            `json:"bundle" yaml:"bundle"`
	Aspect  BundleAspect      `json:"aspect" yaml:"aspect"`
	Rule    BundleRuleKind    `json:"rule" yaml:"rule"`
	Pattern string            `json:"pattern" yaml:"pattern"`
	Kind    BundlePatternKind `json:"kind" yaml:"kind"`
}

type BundleInventoryEntry struct {
	Path      string                `json:"path" yaml:"path"`
	MatchedBy []BundleRuleReference `json:"matched_by" yaml:"matched_by"`
	FinalRule BundleRuleReference   `json:"final_rule" yaml:"final_rule"`
}

type BundleRuleEvaluation struct {
	Bundle      string             `json:"bundle" yaml:"bundle"`
	Aspect      BundleAspect       `json:"aspect" yaml:"aspect"`
	Rule        BundleRuleKind     `json:"rule" yaml:"rule"`
	Pattern     string             `json:"pattern" yaml:"pattern"`
	PatternKind BundlePatternKind  `json:"pattern_kind" yaml:"pattern_kind"`
	Matches     []string           `json:"matches" yaml:"matches"`
	Diagnostics []BundleDiagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

type BundleAspectResolution struct {
	Rules     []BundleRuleEvaluation `json:"rules" yaml:"rules"`
	Selected  []BundleInventoryEntry `json:"selected" yaml:"selected"`
	Excluded  []BundleInventoryEntry `json:"excluded" yaml:"excluded"`
	Matchable []string               `json:"matchable,omitempty" yaml:"matchable,omitempty"`
}

type BundleResolution struct {
	ID            string                                  `json:"id" yaml:"id"`
	Linearization []string                                `json:"linearization" yaml:"linearization"`
	Aspects       map[BundleAspect]BundleAspectResolution `json:"aspects" yaml:"aspects"`
	Diagnostics   []BundleDiagnostic                      `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

type bundleRule struct {
	Bundle      string
	Aspect      BundleAspect
	Kind        BundleRuleKind
	Pattern     string
	RawPattern  string
	PatternKind BundlePatternKind
	SourcePath  string
	Index       int
}

type bundleDefinition struct {
	ID       string
	Path     string
	Extends  []string
	Includes map[BundleAspect][]bundleRule
	Excludes map[BundleAspect][]bundleRule
}

type BundleCatalog struct {
	Root        string
	bundles     map[string]*bundleDefinition
	resolutions map[string]*BundleResolution
}

type bundleWalkState struct {
	files int
	depth int
}

func loadBundleCatalog(v *Validator, rootConfigPath string, rootData []byte, contentRoot string) (*BundleCatalog, error) {
	resolvedRoot, err := filepath.EvalSymlinks(contentRoot)
	if err != nil {
		return nil, &ValidationError{Path: contentRoot, Message: err.Error()}
	}
	catalog := &BundleCatalog{
		Root:        filepath.Clean(resolvedRoot),
		bundles:     map[string]*bundleDefinition{},
		resolutions: map[string]*BundleResolution{},
	}
	bundlePaths := make([]string, 0)
	bundlesRoot := filepath.Join(contentRoot, "bundles")
	if err := walkProjectFiles(bundlesRoot, func(path string) error {
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		bundlePaths = append(bundlePaths, path)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(bundlePaths)
	for _, bundlePath := range bundlePaths {
		data, err := readProjectFile(bundlesRoot, bundlePath)
		if err != nil {
			return nil, err
		}
		if err := v.ValidateYAMLFile("bundle.schema.json", bundlePath, data); err != nil {
			return nil, err
		}
		if err := v.ValidateExtensionOptIn(rootConfigPath, rootData, bundlePath, data); err != nil {
			return nil, err
		}
		bundle, err := parseBundleDefinition(bundlePath, data)
		if err != nil {
			return nil, err
		}
		if existing, ok := catalog.bundles[bundle.ID]; ok {
			return nil, &ValidationError{Path: bundlePath, Message: fmt.Sprintf("bundle id %q is duplicated (already declared in %s)", bundle.ID, existing.Path)}
		}
		catalog.bundles[bundle.ID] = bundle
	}
	for _, id := range SortedKeys(catalog.bundles) {
		bundle := catalog.bundles[id]
		for _, parentID := range bundle.Extends {
			if _, ok := catalog.bundles[parentID]; !ok {
				return nil, &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle %q extends unknown parent %q", bundle.ID, parentID)}
			}
		}
	}
	for _, id := range SortedKeys(catalog.bundles) {
		if _, err := catalog.Resolve(id); err != nil {
			return nil, err
		}
	}
	return catalog, nil
}

func (c *BundleCatalog) Resolve(id string) (*BundleResolution, error) {
	if c == nil {
		return nil, fmt.Errorf("bundle catalog is unavailable")
	}
	if resolution, ok := c.resolutions[id]; ok {
		return cloneBundleResolution(resolution), nil
	}
	ordered, err := c.linearize(id)
	if err != nil {
		return nil, err
	}
	resolution := &BundleResolution{
		ID:            id,
		Linearization: make([]string, 0, len(ordered)),
		Aspects:       map[BundleAspect]BundleAspectResolution{},
	}
	for _, bundle := range ordered {
		resolution.Linearization = append(resolution.Linearization, bundle.ID)
	}
	for _, aspect := range bundleAspects {
		aspectResolution, diagnostics, err := c.resolveAspect(aspect, ordered)
		if err != nil {
			return nil, err
		}
		resolution.Aspects[aspect] = aspectResolution
		resolution.Diagnostics = append(resolution.Diagnostics, diagnostics...)
	}
	c.resolutions[id] = resolution
	return cloneBundleResolution(resolution), nil
}

func (c *BundleCatalog) linearize(id string) ([]*bundleDefinition, error) {
	ordered := make([]*bundleDefinition, 0)
	emitted := map[string]struct{}{}
	stack := make([]string, 0)
	stackIndex := map[string]int{}
	var visit func(string, int) error
	visit = func(bundleID string, depth int) error {
		bundle, ok := c.bundles[bundleID]
		if !ok {
			return &ValidationError{Path: filepath.Join(c.Root, "bundles"), Message: fmt.Sprintf("unknown bundle %q", bundleID)}
		}
		if _, ok := emitted[bundleID]; ok {
			return nil
		}
		if depth > maxBundleInheritanceDepth {
			return &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle inheritance depth exceeds maximum of %d", maxBundleInheritanceDepth)}
		}
		if idx, ok := stackIndex[bundleID]; ok {
			cycle := append(append([]string{}, stack[idx:]...), bundleID)
			return &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle inheritance cycle detected: %s", strings.Join(cycle, " -> "))}
		}
		stackIndex[bundleID] = len(stack)
		stack = append(stack, bundleID)
		for _, parentID := range bundle.Extends {
			if err := visit(parentID, depth+1); err != nil {
				return err
			}
		}
		delete(stackIndex, bundleID)
		stack = stack[:len(stack)-1]
		emitted[bundleID] = struct{}{}
		ordered = append(ordered, bundle)
		return nil
	}
	if err := visit(id, 1); err != nil {
		return nil, err
	}
	return ordered, nil
}

func (c *BundleCatalog) resolveAspect(aspect BundleAspect, ordered []*bundleDefinition) (BundleAspectResolution, []BundleDiagnostic, error) {
	orderedRules := make([]bundleRule, 0)
	for _, bundle := range ordered {
		orderedRules = append(orderedRules, bundle.Includes[aspect]...)
		orderedRules = append(orderedRules, bundle.Excludes[aspect]...)
	}
	result := BundleAspectResolution{
		Rules:    make([]BundleRuleEvaluation, 0, len(orderedRules)),
		Selected: []BundleInventoryEntry{},
		Excluded: []BundleInventoryEntry{},
	}
	type pathState struct {
		matchedBy []BundleRuleReference
		finalRule BundleRuleReference
	}
	states := map[string]*pathState{}
	diagnostics := make([]BundleDiagnostic, 0)
	for _, rule := range orderedRules {
		evaluation, err := c.evaluateRule(rule)
		if err != nil {
			return BundleAspectResolution{}, nil, err
		}
		result.Rules = append(result.Rules, evaluation)
		diagnostics = append(diagnostics, evaluation.Diagnostics...)
		ref := BundleRuleReference{
			Bundle:  evaluation.Bundle,
			Aspect:  evaluation.Aspect,
			Rule:    evaluation.Rule,
			Pattern: evaluation.Pattern,
			Kind:    evaluation.PatternKind,
		}
		for _, matchedPath := range evaluation.Matches {
			state := states[matchedPath]
			if state == nil {
				state = &pathState{}
				states[matchedPath] = state
			}
			state.matchedBy = append(state.matchedBy, ref)
			state.finalRule = ref
		}
	}
	paths := SortedKeys(states)
	for _, matchedPath := range paths {
		state := states[matchedPath]
		entry := BundleInventoryEntry{
			Path:      matchedPath,
			MatchedBy: append([]BundleRuleReference(nil), state.matchedBy...),
			FinalRule: state.finalRule,
		}
		if state.finalRule.Rule == BundleRuleKindInclude {
			result.Selected = append(result.Selected, entry)
		} else {
			result.Excluded = append(result.Excluded, entry)
		}
	}
	return result, diagnostics, nil
}

func (c *BundleCatalog) evaluateRule(rule bundleRule) (BundleRuleEvaluation, error) {
	evaluation := BundleRuleEvaluation{
		Bundle:      rule.Bundle,
		Aspect:      rule.Aspect,
		Rule:        rule.Kind,
		Pattern:     rule.Pattern,
		PatternKind: rule.PatternKind,
		Matches:     []string{},
	}
	var (
		matches     []string
		diagnostics []BundleDiagnostic
		err         error
	)
	if rule.PatternKind == BundlePatternKindExact {
		matches, diagnostics, err = c.evaluateExactRule(rule)
	} else {
		matches, diagnostics, err = c.evaluateGlobRule(rule)
	}
	if err != nil {
		return BundleRuleEvaluation{}, err
	}
	evaluation.Matches = matches
	evaluation.Diagnostics = diagnostics
	return evaluation, nil
}

func (c *BundleCatalog) evaluateExactRule(rule bundleRule) ([]string, []BundleDiagnostic, error) {
	aspectRoot, err := canonicalContainedRoot(filepath.Join(c.Root, string(rule.Aspect)))
	if err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: err.Error()}
	}
	logicalPath := filepath.Join(c.Root, filepath.FromSlash(rule.Pattern))
	if _, err := os.Lstat(logicalPath); err != nil {
		if os.IsNotExist(err) {
			return []string{}, []BundleDiagnostic{{
				Severity: DiagnosticSeverityWarning,
				Code:     "missing_exact_path",
				Message:  fmt.Sprintf("exact %s rule did not match an existing file", rule.Kind),
				Bundle:   rule.Bundle,
				Aspect:   rule.Aspect,
				Rule:     rule.Kind,
				Pattern:  rule.Pattern,
			}}, nil
		}
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q could not be evaluated: %v", rule.Pattern, err)}
	}
	if err := validateResolvedBundlePath(logicalPath, c.Root, aspectRoot); err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q %v", rule.Pattern, err)}
	}
	info, err := os.Stat(logicalPath)
	if err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q could not be evaluated: %v", rule.Pattern, err)}
	}
	if info.IsDir() {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q resolves to a directory; exact bundle rules must reference files", rule.Pattern)}
	}
	if !info.Mode().IsRegular() {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle rule %q resolves to a non-regular file", rule.Pattern)}
	}
	return []string{rule.Pattern}, nil, nil
}

func (c *BundleCatalog) evaluateGlobRule(rule bundleRule) ([]string, []BundleDiagnostic, error) {
	anchor := literalBundleAnchor(rule.Pattern)
	anchorPath := filepath.Join(c.Root, filepath.FromSlash(anchor))
	if _, err := os.Lstat(anchorPath); err != nil {
		if os.IsNotExist(err) {
			return []string{}, []BundleDiagnostic{{
				Severity: DiagnosticSeverityInfo,
				Code:     "empty_glob_match",
				Message:  fmt.Sprintf("glob %s rule matched no files", rule.Kind),
				Bundle:   rule.Bundle,
				Aspect:   rule.Aspect,
				Rule:     rule.Kind,
				Pattern:  rule.Pattern,
			}}, nil
		}
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle glob %q could not be evaluated: %v", rule.Pattern, err)}
	}
	matches := make([]string, 0)
	seen := map[string]struct{}{}
	aspectRoot, err := canonicalContainedRoot(filepath.Join(c.Root, string(rule.Aspect)))
	if err != nil {
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: err.Error()}
	}
	err = walkBundleFiles(c.Root, aspectRoot, anchorPath, map[string]struct{}{}, &bundleWalkState{}, func(logicalPath string) error {
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
		return nil, nil, &ValidationError{Path: rule.SourcePath, Message: fmt.Sprintf("bundle glob %q %v", rule.Pattern, err)}
	}
	sort.Strings(matches)
	diagnostics := []BundleDiagnostic{}
	if len(matches) == 0 {
		diagnostics = append(diagnostics, BundleDiagnostic{
			Severity: DiagnosticSeverityInfo,
			Code:     "empty_glob_match",
			Message:  fmt.Sprintf("glob %s rule matched no files", rule.Kind),
			Bundle:   rule.Bundle,
			Aspect:   rule.Aspect,
			Rule:     rule.Kind,
			Pattern:  rule.Pattern,
		})
	}
	return matches, diagnostics, nil
}

func parseBundleDefinition(path string, data []byte) (*bundleDefinition, error) {
	parsed, err := parseYAML(data)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	obj, err := expectObject(path, parsed, "bundle")
	if err != nil {
		return nil, err
	}
	bundle := &bundleDefinition{
		ID:       fmt.Sprint(obj["id"]),
		Path:     path,
		Extends:  extractStringList(obj["extends"]),
		Includes: map[BundleAspect][]bundleRule{},
		Excludes: map[BundleAspect][]bundleRule{},
	}
	includes, err := extractBundleRules(path, bundle.ID, BundleRuleKindInclude, obj["includes"])
	if err != nil {
		return nil, err
	}
	excludes, err := extractBundleRules(path, bundle.ID, BundleRuleKindExclude, obj["excludes"])
	if err != nil {
		return nil, err
	}
	bundle.Includes = includes
	bundle.Excludes = excludes
	return bundle, nil
}

func extractStringList(raw any) []string {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprint(item))
	}
	return result
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
			normalized, patternKind, err := normalizeBundlePattern(aspect, fmt.Sprint(item))
			if err != nil {
				return nil, &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q %s rule %q is invalid: %v", bundleID, kind, item, err)}
			}
			rules = append(rules, bundleRule{
				Bundle:      bundleID,
				Aspect:      aspect,
				Kind:        kind,
				Pattern:     normalized,
				RawPattern:  fmt.Sprint(item),
				PatternKind: patternKind,
				SourcePath:  sourcePath,
				Index:       i,
			})
		}
		result[aspect] = rules
	}
	return result, nil
}

func isKnownBundleAspect(aspect BundleAspect) bool {
	for _, known := range bundleAspects {
		if aspect == known {
			return true
		}
	}
	return false
}

func normalizeBundlePattern(aspect BundleAspect, raw string) (string, BundlePatternKind, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", "", fmt.Errorf("must not be empty")
	}
	value = strings.ReplaceAll(value, "\\", "/")
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "\\") || filepath.IsAbs(value) || isDriveQualifiedPath(value) {
		return "", "", fmt.Errorf("must not be absolute or drive-qualified")
	}
	segments := strings.Split(value, "/")
	for _, segment := range segments {
		if segment == ".." {
			return "", "", fmt.Errorf("must not contain traversal segments")
		}
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == "" {
		return "", "", fmt.Errorf("must not be empty")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", "", fmt.Errorf("must not contain traversal segments")
	}
	rooted := cleaned
	if rooted == string(aspect) || strings.HasPrefix(rooted, string(aspect)+"/") {
		// already rooted
	} else if bundlePatternUsesOtherAspect(rooted, aspect) {
		return "", "", fmt.Errorf("must stay within the %q aspect", aspect)
	} else {
		rooted = string(aspect) + "/" + rooted
	}
	if rooted == string(aspect) {
		return "", "", fmt.Errorf("must reference a file path or glob beneath the aspect root")
	}
	hasWildcard := false
	for _, segment := range strings.Split(rooted, "/") {
		if strings.Contains(segment, "*") {
			hasWildcard = true
			if segment != "*" && segment != "**" {
				return "", "", fmt.Errorf("wildcards must use whole-segment '*' or '**'")
			}
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
	patternSegments := strings.Split(pattern, "/")
	candidateSegments := strings.Split(candidate, "/")
	type state struct{ i, j int }
	cache := map[state]bool{}
	seen := map[state]bool{}
	var match func(int, int) bool
	match = func(i, j int) bool {
		key := state{i: i, j: j}
		if seen[key] {
			return cache[key]
		}
		seen[key] = true
		var result bool
		switch {
		case i == len(patternSegments):
			result = j == len(candidateSegments)
		case patternSegments[i] == "**":
			if i == len(patternSegments)-1 {
				result = true
				break
			}
			for next := j; next <= len(candidateSegments); next++ {
				if match(i+1, next) {
					result = true
					break
				}
			}
		case j >= len(candidateSegments):
			result = false
		case patternSegments[i] == "*" || patternSegments[i] == candidateSegments[j]:
			result = match(i+1, j+1)
		default:
			result = false
		}
		cache[key] = result
		return result
	}
	return match(0, 0)
}

func walkBundleFiles(contentRoot, aspectRoot, logicalPath string, active map[string]struct{}, state *bundleWalkState, visit func(string) error) error {
	if state == nil {
		state = &bundleWalkState{}
	}
	state.depth++
	defer func() { state.depth-- }()
	if state.depth > bundleTraversalLimits.MaxDepth {
		return fmt.Errorf("bundle traversal exceeds maximum depth of %d", bundleTraversalLimits.MaxDepth)
	}
	resolvedPath, err := filepath.EvalSymlinks(logicalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !isWithinRoot(contentRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes the RuneContext root", resolvedPath)
	}
	if !isWithinRoot(aspectRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes the selected aspect root", resolvedPath)
	}
	info, err := os.Stat(logicalPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		resolvedKey := filepath.Clean(resolvedPath)
		if _, ok := active[resolvedKey]; ok {
			return fmt.Errorf("encountered a symlink cycle at %q", logicalPath)
		}
		active[resolvedKey] = struct{}{}
		defer delete(active, resolvedKey)
		entries, err := os.ReadDir(logicalPath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := walkBundleFiles(contentRoot, aspectRoot, filepath.Join(logicalPath, entry.Name()), active, state, visit); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("resolved path %q is not a regular file", resolvedPath)
	}
	state.files++
	if state.files > bundleTraversalLimits.MaxFiles {
		return fmt.Errorf("bundle traversal exceeds maximum file count of %d", bundleTraversalLimits.MaxFiles)
	}
	return visit(logicalPath)
}

func validateResolvedBundlePath(logicalPath, contentRoot, aspectRoot string) error {
	canonicalAspectRoot, err := canonicalContainedRoot(aspectRoot)
	if err != nil {
		return err
	}
	resolvedPath, err := filepath.EvalSymlinks(logicalPath)
	if err != nil {
		return err
	}
	if !isWithinRoot(contentRoot, resolvedPath) {
		return fmt.Errorf("resolves to %q, which escapes the RuneContext root", resolvedPath)
	}
	if !isWithinRoot(canonicalAspectRoot, resolvedPath) {
		return fmt.Errorf("resolves to %q, which escapes the selected aspect root", resolvedPath)
	}
	return nil
}

func canonicalContainedRoot(root string) (string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Clean(root), nil
		}
		return "", fmt.Errorf("resolve root %q: %w", root, err)
	}
	return filepath.Clean(resolvedRoot), nil
}

func cloneBundleResolution(resolution *BundleResolution) *BundleResolution {
	if resolution == nil {
		return nil
	}
	clone := &BundleResolution{
		ID:            resolution.ID,
		Linearization: append([]string(nil), resolution.Linearization...),
		Aspects:       make(map[BundleAspect]BundleAspectResolution, len(resolution.Aspects)),
		Diagnostics:   cloneBundleDiagnostics(resolution.Diagnostics),
	}
	for aspect, aspectResolution := range resolution.Aspects {
		clone.Aspects[aspect] = BundleAspectResolution{
			Rules:     cloneBundleRuleEvaluations(aspectResolution.Rules),
			Selected:  cloneBundleInventoryEntries(aspectResolution.Selected),
			Excluded:  cloneBundleInventoryEntries(aspectResolution.Excluded),
			Matchable: append([]string(nil), aspectResolution.Matchable...),
		}
	}
	return clone
}

func cloneBundleRuleEvaluations(items []BundleRuleEvaluation) []BundleRuleEvaluation {
	result := make([]BundleRuleEvaluation, len(items))
	for i, item := range items {
		result[i] = BundleRuleEvaluation{
			Bundle:      item.Bundle,
			Aspect:      item.Aspect,
			Rule:        item.Rule,
			Pattern:     item.Pattern,
			PatternKind: item.PatternKind,
			Matches:     append([]string(nil), item.Matches...),
			Diagnostics: cloneBundleDiagnostics(item.Diagnostics),
		}
	}
	return result
}

func cloneBundleInventoryEntries(items []BundleInventoryEntry) []BundleInventoryEntry {
	result := make([]BundleInventoryEntry, len(items))
	for i, item := range items {
		result[i] = BundleInventoryEntry{
			Path:      item.Path,
			MatchedBy: append([]BundleRuleReference(nil), item.MatchedBy...),
			FinalRule: item.FinalRule,
		}
	}
	return result
}

func cloneBundleDiagnostics(items []BundleDiagnostic) []BundleDiagnostic {
	result := make([]BundleDiagnostic, len(items))
	for i, item := range items {
		result[i] = BundleDiagnostic{
			Severity: item.Severity,
			Code:     item.Code,
			Message:  item.Message,
			Bundle:   item.Bundle,
			Aspect:   item.Aspect,
			Rule:     item.Rule,
			Pattern:  item.Pattern,
			Matches:  append([]string(nil), item.Matches...),
		}
	}
	return result
}
