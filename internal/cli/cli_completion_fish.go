package cli

import "strings"

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
		flags := index.flagsByPath[path]
		condition := fishConditionForPath(path)
		for _, flag := range flags {
			name := strings.TrimPrefix(flag.Name, "--")
			out.WriteString("complete -c ")
			out.WriteString(fishToken(index.binary))
			out.WriteString(" -n ")
			out.WriteString(fishSingleQuote(condition))
			out.WriteString(" -l ")
			out.WriteString(name)
			if flag.Value.Kind != ValueKindNone {
				out.WriteString(" -r")
			}
			out.WriteString("\n")
			if flag.Value.Kind == ValueKindEnum {
				out.WriteString("complete -c ")
				out.WriteString(fishToken(index.binary))
				out.WriteString(" -f -n ")
				out.WriteString(fishSingleQuote("(" + condition + "); and __runectx_prev_token_is " + flag.Name))
				out.WriteString(" -a ")
				out.WriteString(fishSingleQuote(strings.Join(flag.Value.EnumValues, " ")))
				out.WriteString("\n")
			}
		}
	}
}

func writeFishPositionalEnums(out *strings.Builder, index completionIndex) {
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
