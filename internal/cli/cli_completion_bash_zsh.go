package cli

import (
	"sort"
	"strconv"
	"strings"
)

func buildBashCompletionScript(index completionIndex) string {
	var out strings.Builder
	writeBashHeader(&out)
	for _, path := range index.allPaths {
		out.WriteString("    ")
		out.WriteString(shellSingleQuote(path))
		out.WriteString(") return 0 ;;\n")
	}
	writeBashHasCommandFooter(&out)

	writeCaseEcho(&out, "_runectx_subcommands_for", index.subcommandsByPath, "")
	writeCaseEcho(&out, "_runectx_flags_for", mapFromFlags(index.flagsByPath), "")
	writeBashFlagValueFunction(&out, index.flagKinds)
	writeCaseEcho(&out, "_runectx_enum_for_flag", index.enumFlags, "")
	writeCaseSingleValue(&out, "_runectx_suggest_provider_for_flag", index.suggestionFlags, "")
	writeCaseEcho(&out, "_runectx_positional_enum", mapFromPositionalEnums(index.positionalEnums), "")
	writeCaseSingleValue(&out, "_runectx_suggest_provider_for_positional", index.positionalSuggest, "")
	writeCaseSingleValue(&out, "_runectx_variadic_suggest_provider_for_command", mapFromVariadicProviders(index.variadicSuggest), "")
	writeCaseSingleValue(&out, "_runectx_variadic_suggest_start_for_command", mapFromVariadicStarts(index.variadicSuggest), "")
	writeBashRuntime(&out, index.binary)
	return out.String()
}

func buildZshCompletionScript(index completionIndex) string {
	return "# RuneContext completion for zsh\n" +
		"autoload -U +X bashcompinit && bashcompinit\n\n" +
		buildBashCompletionScript(index)
}

func writeBashHeader(out *strings.Builder) {
	out.WriteString("# RuneContext completion for bash\n")
	out.WriteString("_runectx_has_command() {\n")
	out.WriteString("  case \"$1\" in\n")
}

func writeBashHasCommandFooter(out *strings.Builder) {
	out.WriteString("    *) return 1 ;;\n")
	out.WriteString("  esac\n")
	out.WriteString("}\n\n")
}

func writeBashFlagValueFunction(out *strings.Builder, flagKinds map[string]ValueKind) {
	out.WriteString("_runectx_flag_takes_value() {\n")
	out.WriteString("  case \"$1|$2\" in\n")
	keys := make([]string, 0, len(flagKinds))
	for key := range flagKinds {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if flagKinds[key] == ValueKindNone {
			continue
		}
		out.WriteString("    ")
		out.WriteString(shellSingleQuote(key))
		out.WriteString(") return 0 ;;\n")
	}
	out.WriteString("    *) return 1 ;;\n")
	out.WriteString("  esac\n")
	out.WriteString("}\n\n")
}

func writeBashRuntime(out *strings.Builder, binary string) {
	out.WriteString(bashRuntimeScript)
	out.WriteString("\ncomplete -F _runectx_complete ")
	out.WriteString(shellToken(binary))
	out.WriteString("\n")
}

func writeCaseEcho(out *strings.Builder, functionName string, values map[string][]string, defaultValue string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out.WriteString(functionName)
	out.WriteString("() {\n")
	out.WriteString("  case \"$1\" in\n")
	for _, key := range keys {
		joined := strings.Join(values[key], " ")
		out.WriteString("    ")
		out.WriteString(shellSingleQuote(key))
		out.WriteString(") echo ")
		out.WriteString(shellSingleQuote(joined))
		out.WriteString(" ;;\n")
	}
	out.WriteString("    *) echo ")
	out.WriteString(shellSingleQuote(defaultValue))
	out.WriteString(" ;;\n")
	out.WriteString("  esac\n")
	out.WriteString("}\n\n")
}

func writeCaseSingleValue(out *strings.Builder, functionName string, values map[string]string, defaultValue string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out.WriteString(functionName)
	out.WriteString("() {\n")
	out.WriteString("  case \"$1\" in\n")
	for _, key := range keys {
		out.WriteString("    ")
		out.WriteString(shellSingleQuote(key))
		out.WriteString(") echo ")
		out.WriteString(shellSingleQuote(values[key]))
		out.WriteString(" ;;\n")
	}
	out.WriteString("    *) echo ")
	out.WriteString(shellSingleQuote(defaultValue))
	out.WriteString(" ;;\n")
	out.WriteString("  esac\n")
	out.WriteString("}\n\n")
}

func mapFromVariadicProviders(items map[string]variadicPositionalSuggestion) map[string]string {
	out := map[string]string{}
	for path, item := range items {
		out[path] = item.Provider
	}
	return out
}

func mapFromVariadicStarts(items map[string]variadicPositionalSuggestion) map[string]string {
	out := map[string]string{}
	for path, item := range items {
		out[path] = strconv.Itoa(item.StartPosition)
	}
	return out
}
