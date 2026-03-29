package cli

import (
	"slices"
	"testing"
)

func TestCommandMetadataRegistryHasCompletionCommand(t *testing.T) {
	registry := CommandMetadataRegistry()
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
