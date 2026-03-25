package cli

func adapterCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "adapter",
		Path:  "adapter",
		Usage: adapterUsage,
		Flags: writeMachineFlags(),
		Subcommands: []CommandMetadata{
			{
				Name:  "sync",
				Path:  "adapter sync",
				Usage: adapterSyncUsage,
				Flags: writeCommandFlags(pathOnlyFlag()),
				Positionals: []PositionalMetadata{{
					Name:  "tool",
					Value: textValueWithSuggestionSpec(suggestionProviderAdapterNames),
				}},
			},
		},
	}
}
