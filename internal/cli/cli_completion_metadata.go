package cli

import (
	"fmt"
	"sort"
	"strings"
)

// CompletionMetadata is machine-readable completion metadata derived from the CLI registry.
type CompletionMetadata struct {
	Binary                string
	Commands              []CompletionCommandMetadata
	Flags                 []CompletionFlagMetadata
	PositionalEnums       []CompletionPositionalEnumMetadata
	PositionalSuggestions []CompletionPositionalSuggestionMetadata
}

// CompletionCommandMetadata records a command path and direct subcommands.
type CompletionCommandMetadata struct {
	Path        string
	Subcommands []string
}

// CompletionFlagMetadata records one flag with optional enum values.
type CompletionFlagMetadata struct {
	CommandPath        string
	Name               string
	ValueKind          ValueKind
	EnumValues         []string
	SuggestionProvider string
	Repeatable         bool
}

// CompletionPositionalEnumMetadata records enum completions for positional slots.
type CompletionPositionalEnumMetadata struct {
	CommandPath string
	Position    int
	EnumValues  []string
}

// CompletionPositionalSuggestionMetadata records dynamic suggestion provider for positional slots.
type CompletionPositionalSuggestionMetadata struct {
	CommandPath        string
	Position           int
	SuggestionProvider string
	Variadic           bool
}

// CompletionMetadataRegistry returns completion metadata derived from the command registry.
func CompletionMetadataRegistry() CompletionMetadata {
	registry := CommandMetadataRegistry()
	return CompletionMetadataFromRegistry(registry)
}

// CompletionMetadataFromRegistry returns completion metadata derived from the provided registry.
func CompletionMetadataFromRegistry(registry MetadataRegistry) CompletionMetadata {
	builder := completionMetadataBuilder{binary: registry.Binary}
	builder.walkCommands("", registry.Commands)
	return builder.build()
}

type completionMetadataBuilder struct {
	binary                string
	commands              []CompletionCommandMetadata
	flags                 []CompletionFlagMetadata
	positionals           []CompletionPositionalEnumMetadata
	positionalSuggestions []CompletionPositionalSuggestionMetadata
}

func (builder *completionMetadataBuilder) walkCommands(parentPath string, commands []CommandMetadata) {
	for _, command := range commands {
		path := commandPath(parentPath, command.Name)
		builder.commands = append(builder.commands, CompletionCommandMetadata{Path: path, Subcommands: commandNames(command.Subcommands)})
		builder.appendFlags(path, command.Flags)
		builder.appendPositionalEnums(path, command.Positionals)
		builder.walkCommands(path, command.Subcommands)
	}
}

func (builder *completionMetadataBuilder) appendFlags(path string, flags []FlagMetadata) {
	for _, flag := range flags {
		item := CompletionFlagMetadata{
			CommandPath:        path,
			Name:               flag.Name,
			ValueKind:          flag.Value.Kind,
			EnumValues:         append([]string(nil), flag.Value.EnumValues...),
			SuggestionProvider: flag.Value.SuggestionProvider,
			Repeatable:         flag.Repeatable,
		}
		builder.flags = append(builder.flags, item)
	}
}

func (builder *completionMetadataBuilder) appendPositionalEnums(path string, positionals []PositionalMetadata) {
	position := 0
	for _, positional := range positionals {
		position++
		if positional.Value.SuggestionProvider != "" {
			builder.positionalSuggestions = append(builder.positionalSuggestions, CompletionPositionalSuggestionMetadata{
				CommandPath:        path,
				Position:           position,
				SuggestionProvider: positional.Value.SuggestionProvider,
				Variadic:           positional.Variadic,
			})
		}
		if positional.Value.Kind != ValueKindEnum {
			if positional.Variadic {
				break
			}
			continue
		}
		builder.positionals = append(builder.positionals, CompletionPositionalEnumMetadata{
			CommandPath: path,
			Position:    position,
			EnumValues:  append([]string(nil), positional.Value.EnumValues...),
		})
		if positional.Variadic {
			break
		}
	}
}

func (builder completionMetadataBuilder) build() CompletionMetadata {
	commands := append([]CompletionCommandMetadata(nil), builder.commands...)
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })
	flags := append([]CompletionFlagMetadata(nil), builder.flags...)
	sort.Slice(flags, func(i, j int) bool {
		left := fmt.Sprintf("%s|%s", flags[i].CommandPath, flags[i].Name)
		right := fmt.Sprintf("%s|%s", flags[j].CommandPath, flags[j].Name)
		return left < right
	})
	positionals := append([]CompletionPositionalEnumMetadata(nil), builder.positionals...)
	sort.Slice(positionals, func(i, j int) bool {
		left := fmt.Sprintf("%s|%03d", positionals[i].CommandPath, positionals[i].Position)
		right := fmt.Sprintf("%s|%03d", positionals[j].CommandPath, positionals[j].Position)
		return left < right
	})
	positionalsSuggestions := append([]CompletionPositionalSuggestionMetadata(nil), builder.positionalSuggestions...)
	sort.Slice(positionalsSuggestions, func(i, j int) bool {
		left := fmt.Sprintf("%s|%03d", positionalsSuggestions[i].CommandPath, positionalsSuggestions[i].Position)
		right := fmt.Sprintf("%s|%03d", positionalsSuggestions[j].CommandPath, positionalsSuggestions[j].Position)
		return left < right
	})
	return CompletionMetadata{Binary: builder.binary, Commands: commands, Flags: flags, PositionalEnums: positionals, PositionalSuggestions: positionalsSuggestions}
}

func commandNames(commands []CommandMetadata) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name)
	}
	sort.Strings(names)
	return names
}

func commandPath(parentPath, name string) string {
	if parentPath == "" {
		return name
	}
	return strings.TrimSpace(parentPath + " " + name)
}
