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
	writeCaseSingleValue(&out, "_runectx_suggest_provider_for_flag", index.suggestionFlags, "")
	writeCaseEcho(&out, "_runectx_positional_enum", mapFromPositionalEnums(index.positionalEnums), "")
	writeCaseSingleValue(&out, "_runectx_suggest_provider_for_positional", index.positionalSuggest, "")
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

const bashRuntimeScript = `_runectx_dynamic_suggest() {
  local provider prefix root cli
  provider="$1"
  prefix="$2"
  root="$3"
  cli="${COMP_WORDS[0]}"
  if [[ -z "$provider" ]]; then
    return
  fi
  if [[ -n "$root" ]]; then
    "$cli" completion suggest --path "$root" --prefix "$prefix" "$provider" 2>/dev/null
    return
  fi
  "$cli" completion suggest --prefix "$prefix" "$provider" 2>/dev/null
}

_runectx_complete() {
  local cur prev cmd token token_flag token_inline prev_flag prev_has_inline cur_flag cur_value candidate idx positional subcommands flags enums provider root_path
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev=""
  if (( COMP_CWORD > 0 )); then
    prev="${COMP_WORDS[COMP_CWORD-1]}"
  fi
  cmd=""
  positional=0
  root_path=""
  idx=1
  while (( idx < COMP_CWORD )); do
    token="${COMP_WORDS[idx]}"
    token_flag="$token"
    token_inline=""
    if [[ "$token" == *=* ]]; then
      token_flag="${token%%=*}"
      token_inline="${token#*=}"
    fi
    if [[ "$token" == --* ]]; then
      if _runectx_flag_takes_value "$cmd" "$token_flag"; then
        if [[ "$token_flag" == "--path" ]]; then
          if [[ -n "$token_inline" ]]; then
            root_path="$token_inline"
          elif (( idx + 1 < COMP_CWORD )); then
            root_path="${COMP_WORDS[idx+1]}"
          fi
        fi
        if [[ -n "$token_inline" ]]; then
          idx=$((idx+1))
        else
          idx=$((idx+2))
        fi
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

  prev_flag="$prev"
  prev_has_inline=0
  if [[ "$prev" == *=* ]]; then
    prev_flag="${prev%%=*}"
    prev_has_inline=1
  fi

  if [[ "$prev_flag" == --* && $prev_has_inline -eq 0 ]]; then
    enums=$(_runectx_enum_for_flag "$cmd|$prev_flag")
    if [[ -n "$enums" ]]; then
      COMPREPLY=( $(compgen -W "$enums" -- "$cur") )
      return
    fi
    provider=$(_runectx_suggest_provider_for_flag "$cmd|$prev_flag")
    if [[ -n "$provider" ]]; then
      COMPREPLY=( $(_runectx_dynamic_suggest "$provider" "$cur" "$root_path") )
      return
    fi
  fi

  cur_flag="$cur"
  if [[ "$cur" == *=* ]]; then
    cur_flag="${cur%%=*}"
  fi
  if [[ "$cur_flag" == --* ]]; then
    if [[ "$cur" == *=* ]]; then
      enums=$(_runectx_enum_for_flag "$cmd|$cur_flag")
      if [[ -n "$enums" ]]; then
        cur_value="${cur#*=}"
        COMPREPLY=( $(compgen -W "$enums" -- "$cur_value") )
        for idx in "${!COMPREPLY[@]}"; do
          COMPREPLY[$idx]="$cur_flag=${COMPREPLY[$idx]}"
        done
        return
      fi
      provider=$(_runectx_suggest_provider_for_flag "$cmd|$cur_flag")
      if [[ -n "$provider" ]]; then
        cur_value="${cur#*=}"
        COMPREPLY=( $(_runectx_dynamic_suggest "$provider" "$cur_value" "$root_path") )
        for idx in "${!COMPREPLY[@]}"; do
          COMPREPLY[$idx]="$cur_flag=${COMPREPLY[$idx]}"
        done
        return
      fi
    fi
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

  provider=$(_runectx_suggest_provider_for_positional "$cmd|$((positional+1))")
  if [[ -n "$provider" ]]; then
    COMPREPLY=( $(_runectx_dynamic_suggest "$provider" "$cur" "$root_path") )
    return
  fi

  flags=$(_runectx_flags_for "$cmd")
  COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
}
`
