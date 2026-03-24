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
