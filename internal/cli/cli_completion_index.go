package cli

import (
	"sort"
	"strconv"
	"strings"
)

type variadicPositionalSuggestion struct {
	StartPosition int
	Provider      string
}

type completionIndex struct {
	binary            string
	allPaths          []string
	topCommands       []string
	subcommandsByPath map[string][]string
	flagsByPath       map[string][]FlagMetadata
	flagKinds         map[string]ValueKind
	enumFlags         map[string][]string
	suggestionFlags   map[string]string
	positionalEnums   map[string][]CompletionPositionalEnumMetadata
	positionalSuggest map[string]string
	variadicSuggest   map[string]variadicPositionalSuggestion
}

func buildCompletionIndex(registry MetadataRegistry) completionIndex {
	metadata := CompletionMetadataFromRegistry(registry)
	index := newCompletionIndex(registry.Binary)
	collectCompletionCommands(&index, metadata)
	collectCompletionFlags(&index, metadata)
	collectCompletionPositionalEnums(&index, metadata)
	sortCompletionIndex(&index)
	return index
}

func newCompletionIndex(binary string) completionIndex {
	return completionIndex{
		binary:            binary,
		subcommandsByPath: map[string][]string{"": nil},
		flagsByPath:       map[string][]FlagMetadata{},
		flagKinds:         map[string]ValueKind{},
		enumFlags:         map[string][]string{},
		suggestionFlags:   map[string]string{},
		positionalEnums:   map[string][]CompletionPositionalEnumMetadata{},
		positionalSuggest: map[string]string{},
		variadicSuggest:   map[string]variadicPositionalSuggestion{},
	}
}

func collectCompletionCommands(index *completionIndex, metadata CompletionMetadata) {
	for _, command := range metadata.Commands {
		index.allPaths = append(index.allPaths, command.Path)
		index.subcommandsByPath[command.Path] = append([]string(nil), command.Subcommands...)
		if !strings.Contains(command.Path, " ") {
			index.topCommands = append(index.topCommands, command.Path)
		}
	}
}

func collectCompletionFlags(index *completionIndex, metadata CompletionMetadata) {
	for _, flag := range metadata.Flags {
		index.flagsByPath[flag.CommandPath] = append(index.flagsByPath[flag.CommandPath], FlagMetadata{
			Name:       flag.Name,
			Value:      ValueSpec{Kind: flag.ValueKind, EnumValues: append([]string(nil), flag.EnumValues...), SuggestionProvider: flag.SuggestionProvider},
			Repeatable: flag.Repeatable,
		})
		key := completionFlagKey(flag.CommandPath, flag.Name)
		index.flagKinds[key] = flag.ValueKind
		if flag.ValueKind == ValueKindEnum {
			index.enumFlags[key] = append([]string(nil), flag.EnumValues...)
		}
		if flag.SuggestionProvider != "" {
			index.suggestionFlags[key] = flag.SuggestionProvider
		}
	}
}

func collectCompletionPositionalEnums(index *completionIndex, metadata CompletionMetadata) {
	for _, positional := range metadata.PositionalEnums {
		index.positionalEnums[positional.CommandPath] = append(index.positionalEnums[positional.CommandPath], positional)
	}
	for _, positional := range metadata.PositionalSuggestions {
		key := positional.CommandPath + "|" + strconv.Itoa(positional.Position)
		index.positionalSuggest[key] = positional.SuggestionProvider
		if positional.Variadic {
			index.variadicSuggest[positional.CommandPath] = variadicPositionalSuggestion{
				StartPosition: positional.Position,
				Provider:      positional.SuggestionProvider,
			}
		}
	}
}

func sortCompletionIndex(index *completionIndex) {
	sort.Strings(index.allPaths)
	sort.Strings(index.topCommands)
	index.subcommandsByPath[""] = append([]string(nil), index.topCommands...)
	for path, flags := range index.flagsByPath {
		sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
		index.flagsByPath[path] = flags
	}
	for path, entries := range index.positionalEnums {
		sort.Slice(entries, func(i, j int) bool { return entries[i].Position < entries[j].Position })
		index.positionalEnums[path] = entries
	}
}

func completionFlagKey(path, flag string) string {
	return path + "|" + flag
}

func mapFromFlags(flagsByPath map[string][]FlagMetadata) map[string][]string {
	out := map[string][]string{}
	for path, flags := range flagsByPath {
		names := make([]string, 0, len(flags))
		for _, flag := range flags {
			names = append(names, flag.Name)
		}
		sort.Strings(names)
		out[path] = names
	}
	return out
}

func mapFromPositionalEnums(positionals map[string][]CompletionPositionalEnumMetadata) map[string][]string {
	out := map[string][]string{}
	for path, items := range positionals {
		for _, item := range items {
			key := path + "|" + strconv.Itoa(item.Position)
			out[key] = append([]string(nil), item.EnumValues...)
		}
	}
	return out
}

func sortedMapKeys[T any](input map[string]T) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
