package cli

import (
	"slices"
	"testing"
)

func TestCommandMetadataRegistryHasCompletionCommand(t *testing.T) {
	registry := CommandMetadataRegistry()
	if commandMetadataByPath(registry.Commands, "metadata") == nil {
		t.Fatalf("expected metadata command in registry")
	}
	var completion *CommandMetadata
	for i := range registry.Commands {
		if registry.Commands[i].Path == "completion" {
			completion = &registry.Commands[i]
			break
		}
	}
	if completion == nil {
		t.Fatalf("expected completion command in registry")
	}
	if completion.Usage != completionUsage {
		t.Fatalf("expected completion usage %q, got %q", completionUsage, completion.Usage)
	}
	if len(completion.Positionals) != 1 {
		t.Fatalf("expected one completion positional, got %d", len(completion.Positionals))
	}
	if got := completion.Positionals[0].Value.EnumValues; !slices.Equal(got, []string{"bash", "fish", "zsh"}) {
		t.Fatalf("expected shell enums [bash fish zsh], got %#v", got)
	}
	if len(completion.Subcommands) != 2 {
		t.Fatalf("expected two completion subcommands, got %#v", completion.Subcommands)
	}
	paths := []string{completion.Subcommands[0].Path, completion.Subcommands[1].Path}
	slices.Sort(paths)
	if !slices.Equal(paths, []string{"completion metadata", "completion suggest"}) {
		t.Fatalf("expected completion metadata/suggest subcommands, got %#v", paths)
	}
}

func TestCommandMetadataRegistryDefensiveCopy(t *testing.T) {
	first := CommandMetadataRegistry()
	first.Commands[0].Name = "mutated"
	if len(first.Commands) > 1 {
		first.Commands[1].Flags = nil
	}
	second := CommandMetadataRegistry()
	if second.Commands[0].Name == "mutated" {
		t.Fatalf("expected registry defensive copy for command name")
	}
	if len(second.Commands) > 1 && second.Commands[1].Flags == nil {
		t.Fatalf("expected registry defensive copy for nested flags")
	}
}

func TestCommandMetadataRegistryIncludesChangeUpdate(t *testing.T) {
	registry := CommandMetadataRegistry()
	change := commandMetadataByPath(registry.Commands, "change")
	if change == nil {
		t.Fatalf("expected change command in registry")
	}
	update := commandMetadataByPath(change.Subcommands, "change update")
	if update == nil {
		t.Fatalf("expected change update subcommand in registry")
	}
	if update.Usage != changeUpdateUsage {
		t.Fatalf("expected change update usage %q, got %q", changeUpdateUsage, update.Usage)
	}
	if len(update.Positionals) != 1 || update.Positionals[0].Name != "CHANGE_ID" {
		t.Fatalf("expected one CHANGE_ID positional for change update, got %#v", update.Positionals)
	}
	validateChangeUpdateFlags(t, update)
}

func TestCommandMetadataRegistryIncludesChangeAssessCommands(t *testing.T) {
	registry := CommandMetadataRegistry()
	change := commandMetadataByPath(registry.Commands, "change")
	if change == nil {
		t.Fatalf("expected change command in registry")
	}
	assessIntake := commandMetadataByPath(change.Subcommands, "change assess-intake")
	if assessIntake == nil {
		t.Fatalf("expected change assess-intake subcommand in registry")
	}
	if assessIntake.Usage != changeAssessIntakeUsage {
		t.Fatalf("expected change assess-intake usage %q, got %q", changeAssessIntakeUsage, assessIntake.Usage)
	}
	if len(assessIntake.Positionals) != 0 {
		t.Fatalf("expected no positionals for assess-intake, got %#v", assessIntake.Positionals)
	}
	assessDecomp := commandMetadataByPath(change.Subcommands, "change assess-decomposition")
	if assessDecomp == nil {
		t.Fatalf("expected change assess-decomposition subcommand in registry")
	}
	if assessDecomp.Usage != changeAssessDecompUsage {
		t.Fatalf("expected change assess-decomposition usage %q, got %q", changeAssessDecompUsage, assessDecomp.Usage)
	}
	if len(assessDecomp.Positionals) != 1 || assessDecomp.Positionals[0].Name != "CHANGE_ID" {
		t.Fatalf("expected one CHANGE_ID positional for assess-decomposition, got %#v", assessDecomp.Positionals)
	}
}

func TestCommandMetadataRegistryIncludesChangeDecompositionCommands(t *testing.T) {
	registry := CommandMetadataRegistry()
	change := commandMetadataByPath(registry.Commands, "change")
	if change == nil {
		t.Fatalf("expected change command in registry")
	}
	assertChangeDecompositionCommandMetadata(t, change, "change decomposition-plan", changeDecompPlanUsage)
	assertChangeDecompositionCommandMetadata(t, change, "change decomposition-apply", changeDecompApplyUsage)
}

func assertChangeDecompositionCommandMetadata(t *testing.T, change *CommandMetadata, path, usage string) {
	t.Helper()
	metadata := commandMetadataByPath(change.Subcommands, path)
	if metadata == nil {
		t.Fatalf("expected %s subcommand in registry", path)
	}
	if metadata.Usage != usage {
		t.Fatalf("expected %s usage %q, got %q", path, usage, metadata.Usage)
	}
	if len(metadata.Positionals) != 1 || metadata.Positionals[0].Name != "UMBRELLA_CHANGE_ID" {
		t.Fatalf("expected one UMBRELLA_CHANGE_ID positional for %s, got %#v", path, metadata.Positionals)
	}
	assertCommandHasFlags(t, metadata.Flags, path, []string{"--sub-change", "--depends-on"})
}

func assertCommandHasFlags(t *testing.T, flags []FlagMetadata, commandPath string, names []string) {
	t.Helper()
	for _, name := range names {
		if flagMetadataByName(flags, name) == nil {
			t.Fatalf("expected %s flag for %s", name, commandPath)
		}
	}
}

func validateChangeUpdateFlags(t *testing.T, update *CommandMetadata) {
	status := flagMetadataByName(update.Flags, "--status")
	if status == nil {
		t.Fatalf("expected --status flag for change update")
	}
	if !status.Required {
		t.Fatalf("expected --status to be required")
	}
	if got := status.Value.EnumValues; !slices.Equal(got, []string{"implemented", "planned", "verified"}) {
		t.Fatalf("expected status enums [implemented planned verified], got %#v", got)
	}
	verification := flagMetadataByName(update.Flags, "--verification-status")
	if verification == nil {
		t.Fatalf("expected --verification-status flag for change update")
	}
	if verification.Required {
		t.Fatalf("expected --verification-status to be optional")
	}
	if got := verification.Value.EnumValues; !slices.Equal(got, []string{"failed", "passed", "skipped"}) {
		t.Fatalf("expected verification_status enums [failed passed skipped], got %#v", got)
	}
	validateChangeUpdateRelationshipFlag(t, update, "--add-related-change")
	validateChangeUpdateRelationshipFlag(t, update, "--remove-related-change")
	recursive := flagMetadataByName(update.Flags, "--recursive")
	if recursive == nil {
		t.Fatalf("expected --recursive flag for change update")
	}
	if recursive.Value.Kind != ValueKindNone {
		t.Fatalf("expected --recursive to be a no-value flag")
	}
}

func validateChangeUpdateRelationshipFlag(t *testing.T, update *CommandMetadata, name string) {
	t.Helper()
	flag := flagMetadataByName(update.Flags, name)
	if flag == nil {
		t.Fatalf("expected %s flag for change update", name)
	}
	if flag.Required {
		t.Fatalf("expected %s to be optional", name)
	}
	if !flag.Repeatable {
		t.Fatalf("expected %s to be repeatable", name)
	}
}

func commandMetadataByPath(commands []CommandMetadata, path string) *CommandMetadata {
	for i := range commands {
		if commands[i].Path == path {
			return &commands[i]
		}
	}
	return nil
}

func flagMetadataByName(flags []FlagMetadata, name string) *FlagMetadata {
	for i := range flags {
		if flags[i].Name == name {
			return &flags[i]
		}
	}
	return nil
}
