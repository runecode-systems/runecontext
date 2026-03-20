package contracts

func cloneBundleResolution(resolution *BundleResolution) *BundleResolution {
	if resolution == nil {
		return nil
	}
	clone := &BundleResolution{ID: resolution.ID, Linearization: append([]string(nil), resolution.Linearization...), Aspects: make(map[BundleAspect]BundleAspectResolution, len(resolution.Aspects)), Diagnostics: cloneBundleDiagnostics(resolution.Diagnostics)}
	for aspect, aspectResolution := range resolution.Aspects {
		clone.Aspects[aspect] = BundleAspectResolution{Rules: cloneBundleRuleEvaluations(aspectResolution.Rules), Selected: cloneBundleInventoryEntries(aspectResolution.Selected), Excluded: cloneBundleInventoryEntries(aspectResolution.Excluded), Matchable: append([]string(nil), aspectResolution.Matchable...)}
	}
	return clone
}

func cloneBundleRuleEvaluations(items []BundleRuleEvaluation) []BundleRuleEvaluation {
	result := make([]BundleRuleEvaluation, len(items))
	for i, item := range items {
		result[i] = BundleRuleEvaluation{Bundle: item.Bundle, Aspect: item.Aspect, Rule: item.Rule, Pattern: item.Pattern, PatternKind: item.PatternKind, Matches: append([]string(nil), item.Matches...), Diagnostics: cloneBundleDiagnostics(item.Diagnostics)}
	}
	return result
}

func cloneBundleInventoryEntries(items []BundleInventoryEntry) []BundleInventoryEntry {
	result := make([]BundleInventoryEntry, len(items))
	for i, item := range items {
		result[i] = BundleInventoryEntry{Path: item.Path, MatchedBy: append([]BundleRuleReference(nil), item.MatchedBy...), FinalRule: item.FinalRule}
	}
	return result
}

func cloneBundleDiagnostics(items []BundleDiagnostic) []BundleDiagnostic {
	result := make([]BundleDiagnostic, len(items))
	for i, item := range items {
		result[i] = BundleDiagnostic{Severity: item.Severity, Code: item.Code, Message: item.Message, Bundle: item.Bundle, Aspect: item.Aspect, Rule: item.Rule, Pattern: item.Pattern, Matches: append([]string(nil), item.Matches...)}
	}
	return result
}
