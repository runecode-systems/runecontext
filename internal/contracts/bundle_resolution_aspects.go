package contracts

func (c *BundleCatalog) resolveAspect(aspect BundleAspect, ordered []*bundleDefinition) (BundleAspectResolution, []BundleDiagnostic, error) {
	orderedRules := orderedAspectRules(aspect, ordered)
	result := BundleAspectResolution{Rules: make([]BundleRuleEvaluation, 0, len(orderedRules)), Selected: []BundleInventoryEntry{}, Excluded: []BundleInventoryEntry{}}
	states := map[string]*bundlePathState{}
	diagnostics := make([]BundleDiagnostic, 0)
	for _, rule := range orderedRules {
		evaluation, err := c.evaluateRule(rule)
		if err != nil {
			return BundleAspectResolution{}, nil, err
		}
		result.Rules = append(result.Rules, evaluation)
		diagnostics = append(diagnostics, evaluation.Diagnostics...)
		applyBundleEvaluation(states, evaluation)
	}
	finalizeAspectInventory(&result, states)
	return result, diagnostics, nil
}

type bundlePathState struct {
	matchedBy []BundleRuleReference
	finalRule BundleRuleReference
}

func orderedAspectRules(aspect BundleAspect, ordered []*bundleDefinition) []bundleRule {
	orderedRules := make([]bundleRule, 0)
	for _, bundle := range ordered {
		orderedRules = append(orderedRules, bundle.Includes[aspect]...)
		orderedRules = append(orderedRules, bundle.Excludes[aspect]...)
	}
	return orderedRules
}

func applyBundleEvaluation(states map[string]*bundlePathState, evaluation BundleRuleEvaluation) {
	ref := BundleRuleReference{Bundle: evaluation.Bundle, Aspect: evaluation.Aspect, Rule: evaluation.Rule, Pattern: evaluation.Pattern, Kind: evaluation.PatternKind}
	for _, matchedPath := range evaluation.Matches {
		state := states[matchedPath]
		if state == nil {
			state = &bundlePathState{}
			states[matchedPath] = state
		}
		state.matchedBy = append(state.matchedBy, ref)
		state.finalRule = ref
	}
}

func finalizeAspectInventory(result *BundleAspectResolution, states map[string]*bundlePathState) {
	for _, matchedPath := range SortedKeys(states) {
		entry := BundleInventoryEntry{Path: matchedPath, MatchedBy: append([]BundleRuleReference(nil), states[matchedPath].matchedBy...), FinalRule: states[matchedPath].finalRule}
		if states[matchedPath].finalRule.Rule == BundleRuleKindInclude {
			result.Selected = append(result.Selected, entry)
		} else {
			result.Excluded = append(result.Excluded, entry)
		}
	}
}

func (c *BundleCatalog) evaluateRule(rule bundleRule) (BundleRuleEvaluation, error) {
	evaluation := BundleRuleEvaluation{Bundle: rule.Bundle, Aspect: rule.Aspect, Rule: rule.Kind, Pattern: rule.Pattern, PatternKind: rule.PatternKind, Matches: []string{}}
	matches, diagnostics, err := c.evaluateRuleMatches(rule)
	if err != nil {
		return BundleRuleEvaluation{}, err
	}
	evaluation.Matches = matches
	evaluation.Diagnostics = diagnostics
	return evaluation, nil
}

func (c *BundleCatalog) evaluateRuleMatches(rule bundleRule) ([]string, []BundleDiagnostic, error) {
	if rule.PatternKind == BundlePatternKindExact {
		return c.evaluateExactRule(rule)
	}
	return c.evaluateGlobRule(rule)
}
