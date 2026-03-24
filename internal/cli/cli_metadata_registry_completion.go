package cli

func completionCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "completion",
		Path:  "completion",
		Usage: completionUsage,
		Subcommands: []CommandMetadata{
			completionSuggestCommandMetadata(),
		},
		Positionals: []PositionalMetadata{
			{Name: "shell", Value: enumValueSpec("bash", "zsh", "fish")},
		},
	}
}

func completionSuggestCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "suggest",
		Path:  "completion suggest",
		Usage: completionSuggestUsage,
		Flags: []FlagMetadata{
			{Name: "--path", Value: textValueSpec()},
			{Name: "--prefix", Value: textValueSpec()},
		},
		Positionals: []PositionalMetadata{
			{Name: "provider", Value: enumValueSpec(suggestionProviderChangeIDs, suggestionProviderBundleIDs, suggestionProviderPromotionTargets, suggestionProviderAdapterNames)},
		},
	}
}
