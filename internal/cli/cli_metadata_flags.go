package cli

func changeNewFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--title", Value: textValueSpec(), Required: true},
		{Name: "--type", Value: enumValueSpec("project", "feature", "bug", "standard", "chore"), Required: true},
		{Name: "--size", Value: enumValueSpec("small", "medium", "large")},
		{Name: "--bundle", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Repeatable: true},
		{Name: "--shape", Value: enumValueSpec("minimum", "full")},
		{Name: "--description", Value: textValueSpec()},
		{Name: "--path", Value: textValueSpec()},
	}
}

func changeShapeFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--design", Value: textValueSpec()},
		{Name: "--verification", Value: textValueSpec()},
		{Name: "--task", Value: textValueSpec(), Repeatable: true},
		{Name: "--reference", Value: textValueSpec(), Repeatable: true},
		{Name: "--path", Value: textValueSpec()},
	}
}

func changeCloseFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--verification-status", Value: enumValueSpec("passed", "failed", "skipped")},
		{Name: "--superseded-by", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs), Repeatable: true},
		{Name: "--closed-at", Value: textValueSpec()},
		{Name: "--recursive", Value: noValueSpec()},
		{Name: "--path", Value: textValueSpec()},
	}
}

func changeUpdateFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--status", Value: enumValueSpec("planned", "implemented", "verified"), Required: true},
		{Name: "--verification-status", Value: enumValueSpec("passed", "failed", "skipped")},
		{Name: "--add-related-change", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs), Repeatable: true},
		{Name: "--remove-related-change", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs), Repeatable: true},
		{Name: "--recursive", Value: noValueSpec()},
		{Name: "--path", Value: textValueSpec()},
	}
}

func changeAssessIntakeFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--title", Value: textValueSpec(), Required: true},
		{Name: "--type", Value: enumValueSpec("project", "feature", "bug", "standard", "chore"), Required: true},
		{Name: "--size", Value: enumValueSpec("small", "medium", "large")},
		{Name: "--bundle", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Repeatable: true},
		{Name: "--description", Value: textValueSpec()},
		{Name: "--path", Value: textValueSpec()},
	}
}

func changeDecompositionFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--sub-change", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs), Repeatable: true, Required: true},
		{Name: "--depends-on", Value: textValueSpec(), Repeatable: true},
		{Name: "--path", Value: textValueSpec()},
	}
}

func validateFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--ssh-allowed-signers", Value: textValueSpec()},
		{Name: "--path", Value: textValueSpec()},
	}
}

func initFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--mode", Value: enumValueSpec("embedded", "linked")},
		{Name: "--seed-bundle", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs)},
		{Name: "--path", Value: textValueSpec()},
	}
}

func promoteFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--accept", Value: noValueSpec()},
		{Name: "--complete", Value: noValueSpec()},
		{Name: "--target", Value: textValueWithSuggestionSpec(suggestionProviderPromotionTargets), Repeatable: true},
		{Name: "--path", Value: textValueSpec()},
	}
}

func standardDiscoverFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec()},
		{Name: "--change", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)},
		{Name: "--scope-path", Value: textValueSpec(), Repeatable: true},
		{Name: "--focus", Value: textValueSpec()},
		{Name: "--confirm-handoff", Value: noValueSpec()},
		{Name: "--target", Value: textValueWithSuggestionSpec(suggestionProviderPromotionTargets)},
	}
}

func standardListFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec()},
		{Name: "--scope-path", Value: textValueSpec(), Repeatable: true},
		{Name: "--focus", Value: textValueSpec()},
		{Name: "--status", Value: enumValueSpec("draft", "active", "deprecated"), Repeatable: true},
	}
}

func standardCreateFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec(), Required: true},
		{Name: "--id", Value: textValueSpec()},
		{Name: "--title", Value: textValueSpec(), Required: true},
		{Name: "--status", Value: enumValueSpec("draft", "active", "deprecated")},
		{Name: "--replaced-by", Value: textValueSpec()},
		{Name: "--alias", Value: textValueSpec(), Repeatable: true},
		{Name: "--suggested-context-bundle", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Repeatable: true},
		{Name: "--body", Value: textValueSpec()},
		{Name: "--project-path", Value: textValueSpec()},
	}
}

func standardUpdateFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--path", Value: textValueSpec(), Required: true},
		{Name: "--title", Value: textValueSpec()},
		{Name: "--status", Value: enumValueSpec("draft", "active", "deprecated")},
		{Name: "--replaced-by", Value: textValueSpec()},
		{Name: "--clear-replaced-by", Value: noValueSpec()},
		{Name: "--replace-aliases", Value: noValueSpec()},
		{Name: "--alias", Value: textValueSpec(), Repeatable: true},
		{Name: "--replace-suggested-context-bundles", Value: noValueSpec()},
		{Name: "--suggested-context-bundle", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Repeatable: true},
		{Name: "--project-path", Value: textValueSpec()},
	}
}

func pathOnlyFlag() []FlagMetadata {
	return []FlagMetadata{{Name: "--path", Value: textValueSpec()}}
}

func readOnlyCommandFlags(extra []FlagMetadata) []FlagMetadata {
	return appendFlags(readMachineFlags(), extra)
}

func writeCommandFlags(extra []FlagMetadata) []FlagMetadata {
	return appendFlags(writeMachineFlags(), extra)
}

func readMachineFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--json", Value: noValueSpec()},
		{Name: "--non-interactive", Value: noValueSpec()},
		{Name: "--explain", Value: noValueSpec()},
	}
}

func writeMachineFlags() []FlagMetadata {
	return []FlagMetadata{
		{Name: "--json", Value: noValueSpec()},
		{Name: "--non-interactive", Value: noValueSpec()},
		{Name: "--dry-run", Value: noValueSpec()},
		{Name: "--explain", Value: noValueSpec()},
	}
}

func appendFlags(base, extra []FlagMetadata) []FlagMetadata {
	flags := make([]FlagMetadata, 0, len(base)+len(extra))
	flags = append(flags, base...)
	flags = append(flags, extra...)
	return flags
}
