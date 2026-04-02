package cli

// ValueKind describes how a flag or positional argument accepts values.
type ValueKind string

const (
	ValueKindNone ValueKind = "none"
	ValueKindText ValueKind = "text"
	ValueKindEnum ValueKind = "enum"
)

// ValueSpec describes accepted value metadata for completion/documentation.
type ValueSpec struct {
	Kind               ValueKind
	EnumValues         []string
	SuggestionProvider string
}

// FlagMetadata describes one CLI flag.
type FlagMetadata struct {
	Name       string
	Value      ValueSpec
	Repeatable bool
	Required   bool
}

// PositionalMetadata describes one positional argument.
type PositionalMetadata struct {
	Name     string
	Optional bool
	Variadic bool
	Value    ValueSpec
}

// CommandMetadata describes one command/subcommand in the CLI tree.
type CommandMetadata struct {
	Name        string
	Path        string
	Usage       string
	Flags       []FlagMetadata
	Positionals []PositionalMetadata
	Subcommands []CommandMetadata
}

// MetadataRegistry is the canonical typed CLI command registry.
type MetadataRegistry struct {
	Binary   string
	Commands []CommandMetadata
}

// CommandMetadataRegistry returns a copy of the canonical command metadata.
func CommandMetadataRegistry() MetadataRegistry {
	registry := cliMetadataRegistry
	registry.Commands = copyCommands(registry.Commands)
	return registry
}

var cliMetadataRegistry = MetadataRegistry{
	Binary:   "runectx",
	Commands: rootCommandsMetadata(),
}

func rootCommandsMetadata() []CommandMetadata {
	commands := []CommandMetadata{
		{Name: "help", Path: "help", Usage: "runectx help"},
		{Name: "validate", Path: "validate", Usage: validateUsage, Flags: readOnlyCommandFlags(validateFlags())},
		{Name: "status", Path: "status", Usage: statusUsage, Flags: readOnlyCommandFlags(pathOnlyFlag())},
		changeCommandMetadata(),
		generateCommandMetadata(),
		bundleCommandMetadata(),
		{Name: "doctor", Path: "doctor", Usage: doctorUsage, Flags: readOnlyCommandFlags(pathOnlyFlag())},
		{Name: "init", Path: "init", Usage: initUsage, Flags: writeCommandFlags(initFlags())},
		{Name: "promote", Path: "promote", Usage: promoteUsage, Flags: writeCommandFlags(promoteFlags()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
		standardCommandMetadata(),
		assuranceCommandMetadata(),
		adapterCommandMetadata(),
		completionCommandMetadata(),
		{Name: "metadata", Path: "metadata", Usage: metadataUsage},
		upgradeCommandMetadata(),
	}
	return append(commands, versionRootCommandsMetadata()...)
}

func changeCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "change",
		Path:  "change",
		Usage: changeUsage,
		Flags: writeMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "new", Path: "change new", Usage: changeNewUsage, Flags: writeCommandFlags(changeNewFlags())},
			{Name: "shape", Path: "change shape", Usage: changeShapeUsage, Flags: writeCommandFlags(changeShapeFlags()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "close", Path: "change close", Usage: changeCloseUsage, Flags: writeCommandFlags(changeCloseFlags()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "reallocate", Path: "change reallocate", Usage: changeReallocateUsage, Flags: writeCommandFlags(pathOnlyFlag()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "update", Path: "change update", Usage: changeUpdateUsage, Flags: writeCommandFlags(changeUpdateFlags()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "assess-intake", Path: "change assess-intake", Usage: changeAssessIntakeUsage, Flags: readOnlyCommandFlags(changeAssessIntakeFlags())},
			{Name: "assess-decomposition", Path: "change assess-decomposition", Usage: changeAssessDecompUsage, Flags: readOnlyCommandFlags(pathOnlyFlag()), Positionals: []PositionalMetadata{{Name: "CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "decomposition-plan", Path: "change decomposition-plan", Usage: changeDecompPlanUsage, Flags: readOnlyCommandFlags(changeDecompositionFlags()), Positionals: []PositionalMetadata{{Name: "UMBRELLA_CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
			{Name: "decomposition-apply", Path: "change decomposition-apply", Usage: changeDecompApplyUsage, Flags: writeCommandFlags(changeDecompositionFlags()), Positionals: []PositionalMetadata{{Name: "UMBRELLA_CHANGE_ID", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs)}}},
		},
	}
}

func generateCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "generate",
		Path:  "generate",
		Usage: generateUsage,
		Flags: readMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "indexes", Path: "generate indexes", Usage: generateIndexesUsage, Flags: readOnlyCommandFlags(pathOnlyFlag())},
		},
	}
}

func bundleCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "bundle",
		Path:  "bundle",
		Usage: bundleUsage,
		Flags: readMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "resolve", Path: "bundle resolve", Usage: bundleResolveUsage, Flags: readOnlyCommandFlags(pathOnlyFlag()), Positionals: []PositionalMetadata{{Name: "bundle-id", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Variadic: true}}},
		},
	}
}

func standardCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "standard",
		Path:  "standard",
		Usage: standardUsage,
		Flags: writeMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "discover", Path: "standard discover", Usage: standardDiscoverUsage, Flags: readOnlyCommandFlags(standardDiscoverFlags())},
			{Name: "list", Path: "standard list", Usage: standardListUsage, Flags: readOnlyCommandFlags(standardListFlags())},
			{Name: "create", Path: "standard create", Usage: standardCreateUsage, Flags: writeCommandFlags(standardCreateFlags())},
			{Name: "update", Path: "standard update", Usage: standardUpdateUsage, Flags: writeCommandFlags(standardUpdateFlags())},
		},
	}
}

func assuranceCommandMetadata() CommandMetadata {
	return CommandMetadata{
		Name:  "assurance",
		Path:  "assurance",
		Usage: assuranceCommandUsage,
		Flags: writeMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "enable", Path: "assurance enable", Usage: assuranceEnableUsage, Flags: writeCommandFlags(pathOnlyFlag()), Positionals: []PositionalMetadata{{Name: "mode", Value: enumValueSpec("verified")}}},
			{Name: "backfill", Path: "assurance backfill", Usage: assuranceBackfillUsage, Flags: writeCommandFlags(pathOnlyFlag())},
			{Name: "capture", Path: "assurance capture", Usage: assuranceCaptureUsage, Flags: writeCommandFlags(pathOnlyFlag()), Positionals: []PositionalMetadata{{Name: "subject", Value: enumValueSpec("context-pack")}, {Name: "bundle-id", Value: textValueWithSuggestionSpec(suggestionProviderBundleIDs), Variadic: true}}},
		},
	}
}

func copyCommands(items []CommandMetadata) []CommandMetadata {
	out := make([]CommandMetadata, 0, len(items))
	for _, item := range items {
		copied := item
		copied.Flags = append([]FlagMetadata(nil), item.Flags...)
		copied.Positionals = append([]PositionalMetadata(nil), item.Positionals...)
		copied.Subcommands = copyCommands(item.Subcommands)
		for i := range copied.Flags {
			copied.Flags[i].Value.EnumValues = append([]string(nil), copied.Flags[i].Value.EnumValues...)
		}
		for i := range copied.Positionals {
			copied.Positionals[i].Value.EnumValues = append([]string(nil), copied.Positionals[i].Value.EnumValues...)
		}
		out = append(out, copied)
	}
	return out
}
