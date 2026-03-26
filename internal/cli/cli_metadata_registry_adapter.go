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
			{
				Name:  "render-host-native",
				Path:  "adapter render-host-native",
				Usage: adapterRenderUsage,
				Flags: writeCommandFlags([]FlagMetadata{{Name: "--role", Value: enumValueSpec(hostNativeKindFlowAsset, hostNativeKindDiscoverabilityShim)}}),
				Positionals: []PositionalMetadata{
					{Name: "tool", Value: textValueWithSuggestionSpec(suggestionProviderAdapterNamesShellInjection)},
					{Name: "operation", Value: enumValueSpec("change-new", "change-shape", "index", "promote", "standard-discover")},
				},
			},
		},
	}
}
