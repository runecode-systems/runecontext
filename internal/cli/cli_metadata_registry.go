package cli

import "slices"

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
		Flags: readMachineFlags(),
		Subcommands: []CommandMetadata{
			{Name: "discover", Path: "standard discover", Usage: standardDiscoverUsage, Flags: readOnlyCommandFlags(standardDiscoverFlags())},
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
		{Name: "--verification-status", Value: enumValueSpec("pending", "passed", "failed", "skipped")},
		{Name: "--superseded-by", Value: textValueWithSuggestionSpec(suggestionProviderChangeIDs), Repeatable: true},
		{Name: "--closed-at", Value: textValueSpec()},
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

func noValueSpec() ValueSpec {
	return ValueSpec{Kind: ValueKindNone}
}

func textValueSpec() ValueSpec {
	return ValueSpec{Kind: ValueKindText}
}

func textValueWithSuggestionSpec(provider string) ValueSpec {
	return ValueSpec{Kind: ValueKindText, SuggestionProvider: provider}
}

func enumValueSpec(values ...string) ValueSpec {
	copyValues := append([]string(nil), values...)
	slices.Sort(copyValues)
	return ValueSpec{Kind: ValueKindEnum, EnumValues: copyValues}
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
