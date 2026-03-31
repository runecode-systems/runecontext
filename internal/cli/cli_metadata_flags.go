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
