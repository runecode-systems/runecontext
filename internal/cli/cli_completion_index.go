package cli

import (
	"sort"
	"strconv"
	"strings"
)

type completionIndex struct {
	binary            string
	allPaths          []string
	topCommands       []string
	subcommandsByPath map[string][]string
	flagsByPath       map[string][]FlagMetadata
	flagKinds         map[string]ValueKind
	enumFlags         map[string][]string
	positionalEnums   map[string][]CompletionPositionalEnumMetadata
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
		positionalEnums:   map[string][]CompletionPositionalEnumMetadata{},
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
			Value:      ValueSpec{Kind: flag.ValueKind, EnumValues: append([]string(nil), flag.EnumValues...)},
			Repeatable: flag.Repeatable,
		})
		key := completionFlagKey(flag.CommandPath, flag.Name)
		index.flagKinds[key] = flag.ValueKind
		if flag.ValueKind == ValueKindEnum {
			index.enumFlags[key] = append([]string(nil), flag.EnumValues...)
		}
	}
}

func collectCompletionPositionalEnums(index *completionIndex, metadata CompletionMetadata) {
	for _, positional := range metadata.PositionalEnums {
		index.positionalEnums[positional.CommandPath] = append(index.positionalEnums[positional.CommandPath], positional)
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
