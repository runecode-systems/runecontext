package cli

import (
	"sort"
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
	writeCaseEcho(&out, "_runectx_positional_enum", mapFromPositionalEnums(index.positionalEnums), "")
	writeBashRuntime(&out)
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

func writeBashRuntime(out *strings.Builder) {
	out.WriteString(bashRuntimeScript)
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

const bashRuntimeScript = `_runectx_complete() {
  local cur prev cmd token candidate idx positional subcommands flags enums
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev=""
  if (( COMP_CWORD > 0 )); then
    prev="${COMP_WORDS[COMP_CWORD-1]}"
  fi
  cmd=""
  positional=0
  idx=1
  while (( idx < COMP_CWORD )); do
    token="${COMP_WORDS[idx]}"
    if [[ "$token" == --* ]]; then
      if _runectx_flag_takes_value "$cmd" "$token"; then
        idx=$((idx+2))
      else
        idx=$((idx+1))
      fi
      continue
    fi
    candidate="$token"
    if [[ -n "$cmd" ]]; then candidate="$cmd $token"; fi
    if _runectx_has_command "$candidate"; then
      cmd="$candidate"
      positional=0
      idx=$((idx+1))
      continue
    fi
    positional=$((positional+1))
    idx=$((idx+1))
  done

  if [[ "$prev" == --* ]]; then
    enums=$(_runectx_enum_for_flag "$cmd|$prev")
    if [[ -n "$enums" ]]; then
      COMPREPLY=( $(compgen -W "$enums" -- "$cur") )
      return
    fi
  fi

  if [[ "$cur" == --* ]]; then
    flags=$(_runectx_flags_for "$cmd")
    COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
    return
  fi

  subcommands=$(_runectx_subcommands_for "$cmd")
  if [[ -n "$subcommands" && $positional -eq 0 ]]; then
    COMPREPLY=( $(compgen -W "$subcommands" -- "$cur") )
    return
  fi

  enums=$(_runectx_positional_enum "$cmd|$((positional+1))")
  if [[ -n "$enums" ]]; then
    COMPREPLY=( $(compgen -W "$enums" -- "$cur") )
    return
  fi

  flags=$(_runectx_flags_for "$cmd")
  COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
}

complete -F _runectx_complete runectx
`
