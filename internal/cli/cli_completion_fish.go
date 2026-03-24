package cli

import (
	"sort"
	"strconv"
	"strings"
)

func buildFishCompletionScript(index completionIndex) string {
	var out strings.Builder
	writeFishHeader(&out)
	for _, command := range index.topCommands {
		out.WriteString("complete -c ")
		out.WriteString(fishToken(index.binary))
		out.WriteString(" -f -n '__fish_use_subcommand' -a ")
		out.WriteString(fishSingleQuote(command))
		out.WriteString("\n")
	}
	writeFishSubcommands(&out, index)
	writeFishFlags(&out, index)
	writeFishPositionalEnums(&out, index)
	return out.String()
}

func writeFishHeader(out *strings.Builder) {
	out.WriteString("# RuneContext completion for fish\n")
	out.WriteString("function __runectx_prev_token_is\n")
	out.WriteString("  set -l tokens (commandline -opc)\n")
	out.WriteString("  if test (count $tokens) -lt 2\n")
	out.WriteString("    return 1\n")
	out.WriteString("  end\n")
	out.WriteString("  test $tokens[-1] = $argv[1]\n")
	out.WriteString("end\n\n")
	out.WriteString("function __runectx_root_path\n")
	out.WriteString("  set -l tokens (commandline -opc)\n")
	out.WriteString("  for i in (seq 1 (count $tokens))\n")
	out.WriteString("    set -l token $tokens[$i]\n")
	out.WriteString("    if test \"$token\" = \"--path\"\n")
	out.WriteString("      set -l next_index (math $i + 1)\n")
	out.WriteString("      if test $next_index -le (count $tokens)\n")
	out.WriteString("        echo $tokens[$next_index]\n")
	out.WriteString("        return 0\n")
	out.WriteString("      end\n")
	out.WriteString("    end\n")
	out.WriteString("    if string match -q -- '--path=*' $token\n")
	out.WriteString("      echo (string replace -- '--path=' '' $token)\n")
	out.WriteString("      return 0\n")
	out.WriteString("    end\n")
	out.WriteString("  end\n")
	out.WriteString("  return 1\n")
	out.WriteString("end\n\n")
	out.WriteString("function __runectx_dynamic_suggest\n")
	out.WriteString("  set -l provider $argv[1]\n")
	out.WriteString("  set -l prefix (commandline -ct)\n")
	out.WriteString("  set -l root (__runectx_root_path)\n")
	out.WriteString("  if test -n \"$root\"\n")
	out.WriteString("    command (commandline -poc)[1] completion suggest --path \"${root}\" --prefix \"${prefix}\" $provider 2>/dev/null\n")
	out.WriteString("    return\n")
	out.WriteString("  end\n")
	out.WriteString("  command (commandline -poc)[1] completion suggest --prefix \"${prefix}\" $provider 2>/dev/null\n")
	out.WriteString("end\n\n")
}

func writeFishFlagValueFunction(out *strings.Builder, index completionIndex) {
	flags := map[string]struct{}{}
	for key, kind := range index.flagKinds {
		if kind == ValueKindNone {
			continue
		}
		_, flag, ok := strings.Cut(key, "|")
		if !ok || flag == "" {
			continue
		}
		flags[flag] = struct{}{}
	}
	names := make([]string, 0, len(flags))
	for flag := range flags {
		names = append(names, flag)
	}
	sort.Strings(names)
	out.WriteString("function __runectx_flag_name_takes_value\n")
	out.WriteString("  switch $argv[1]\n")
	for _, name := range names {
		out.WriteString("  case ")
		out.WriteString(name)
		out.WriteString("\n")
		out.WriteString("    return 0\n")
	}
	out.WriteString("  end\n")
	out.WriteString("  return 1\n")
	out.WriteString("end\n\n")
	out.WriteString(fishPositionalRuntimeHelpers)
}

func writeFishSubcommands(out *strings.Builder, index completionIndex) {
	for _, path := range sortedMapKeys(index.subcommandsByPath) {
		subcommands := index.subcommandsByPath[path]
		if path == "" || len(subcommands) == 0 {
			continue
		}
		condition := fishConditionForPath(path)
		for _, command := range subcommands {
			out.WriteString("complete -c ")
			out.WriteString(fishToken(index.binary))
			out.WriteString(" -f -n ")
			out.WriteString(fishSingleQuote(condition))
			out.WriteString(" -a ")
			out.WriteString(fishSingleQuote(command))
			out.WriteString("\n")
		}
	}
}

func writeFishFlags(out *strings.Builder, index completionIndex) {
	for _, path := range sortedMapKeys(index.flagsByPath) {
		writeFishFlagsForPath(out, index, path)
	}
}

func writeFishFlagsForPath(out *strings.Builder, index completionIndex, path string) {
	condition := fishConditionForPath(path)
	for _, flag := range index.flagsByPath[path] {
		writeFishBaseFlagCompletion(out, index.binary, condition, flag)
		writeFishEnumFlagCompletion(out, index.binary, condition, flag)
		writeFishDynamicFlagCompletion(out, index.binary, condition, path, flag, index.suggestionFlags)
	}
}

func writeFishBaseFlagCompletion(out *strings.Builder, binary, condition string, flag FlagMetadata) {
	name := strings.TrimPrefix(flag.Name, "--")
	out.WriteString("complete -c ")
	out.WriteString(fishToken(binary))
	out.WriteString(" -n ")
	out.WriteString(fishSingleQuote(condition))
	out.WriteString(" -l ")
	out.WriteString(name)
	if flag.Value.Kind != ValueKindNone {
		out.WriteString(" -r")
	}
	out.WriteString("\n")
}

func writeFishEnumFlagCompletion(out *strings.Builder, binary, condition string, flag FlagMetadata) {
	if flag.Value.Kind != ValueKindEnum {
		return
	}
	out.WriteString("complete -c ")
	out.WriteString(fishToken(binary))
	out.WriteString(" -f -n ")
	out.WriteString(fishSingleQuote("(" + condition + "); and __runectx_prev_token_is " + flag.Name))
	out.WriteString(" -a ")
	out.WriteString(fishSingleQuote(strings.Join(flag.Value.EnumValues, " ")))
	out.WriteString("\n")
}

func writeFishDynamicFlagCompletion(out *strings.Builder, binary, condition, path string, flag FlagMetadata, providers map[string]string) {
	provider := providers[completionFlagKey(path, flag.Name)]
	if provider == "" {
		return
	}
	out.WriteString("complete -c ")
	out.WriteString(fishToken(binary))
	out.WriteString(" -f -n ")
	out.WriteString(fishSingleQuote("(" + condition + "); and __runectx_prev_token_is " + flag.Name))
	out.WriteString(" -a ")
	out.WriteString(fishSingleQuote("(__runectx_dynamic_suggest " + provider + ")"))
	out.WriteString("\n")
}

func writeFishPositionalEnums(out *strings.Builder, index completionIndex) {
	writeFishFlagValueFunction(out, index)
	writeFishPositionalEnumEntries(out, index)
	writeFishDynamicPositionalEntries(out, index)
}

func writeFishPositionalEnumEntries(out *strings.Builder, index completionIndex) {
	for _, path := range sortedMapKeys(index.positionalEnums) {
		items := index.positionalEnums[path]
		if len(items) == 0 {
			continue
		}
		for _, item := range items {
			if item.Position != 1 {
				continue
			}
			condition := fishConditionForPath(path)
			out.WriteString("complete -c ")
			out.WriteString(fishToken(index.binary))
			out.WriteString(" -f -n ")
			out.WriteString(fishSingleQuote(condition))
			out.WriteString(" -a ")
			out.WriteString(fishSingleQuote(strings.Join(item.EnumValues, " ")))
			out.WriteString("\n")
		}
	}
}

func writeFishDynamicPositionalEntries(out *strings.Builder, index completionIndex) {
	writeFishDynamicPositionalEntriesByPosition(out, index)
	writeFishDynamicVariadicPositionalEntries(out, index)
}

func writeFishDynamicPositionalEntriesByPosition(out *strings.Builder, index completionIndex) {
	for _, key := range sortedMapKeys(index.positionalSuggest) {
		provider := index.positionalSuggest[key]
		path, position, ok := parsePositionalSuggestionKey(key)
		if !ok || provider == "" {
			continue
		}
		if _, hasVariadic := index.variadicSuggest[path]; hasVariadic {
			continue
		}
		condition := fishConditionForPath(path)
		if position > 1 {
			condition += "; and __runectx_positional_index_at_least " + strconv.Itoa(position) + " " + path
		}
		out.WriteString("complete -c ")
		out.WriteString(fishToken(index.binary))
		out.WriteString(" -f -n ")
		out.WriteString(fishSingleQuote(condition))
		out.WriteString(" -a ")
		out.WriteString(fishSingleQuote("(__runectx_dynamic_suggest " + provider + ")"))
		out.WriteString("\n")
	}
}

func writeFishDynamicVariadicPositionalEntries(out *strings.Builder, index completionIndex) {
	for _, path := range sortedMapKeys(index.variadicSuggest) {
		v := index.variadicSuggest[path]
		if v.Provider == "" {
			continue
		}
		condition := fishConditionForPath(path) + "; and __runectx_positional_index_at_least " + strconv.Itoa(v.StartPosition) + " " + path
		out.WriteString("complete -c ")
		out.WriteString(fishToken(index.binary))
		out.WriteString(" -f -n ")
		out.WriteString(fishSingleQuote(condition))
		out.WriteString(" -a ")
		out.WriteString(fishSingleQuote("(__runectx_dynamic_suggest " + v.Provider + ")"))
		out.WriteString("\n")
	}
}

func parsePositionalSuggestionKey(key string) (string, int, bool) {
	path, rawPosition, ok := strings.Cut(key, "|")
	if !ok || rawPosition == "" {
		return "", 0, false
	}
	position := 0
	for _, ch := range rawPosition {
		if ch < '0' || ch > '9' {
			return "", 0, false
		}
		position = position*10 + int(ch-'0')
	}
	if position <= 0 {
		return "", 0, false
	}
	return path, position, true
}
