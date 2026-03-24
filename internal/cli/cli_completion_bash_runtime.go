package cli

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
  local cur prev cmd token token_flag token_inline prev_flag prev_has_inline cur_flag cur_value candidate idx positional subcommands flags enums provider root_path variadic_start
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
    candidate="$subcommands"
    enums=$(_runectx_positional_enum "$cmd|1")
    if [[ -n "$enums" ]]; then
      candidate="$candidate $enums"
    fi
    COMPREPLY=( $(compgen -W "$candidate" -- "$cur") )
    return
  fi

  enums=$(_runectx_positional_enum "$cmd|$((positional+1))")
  if [[ -n "$enums" ]]; then
    COMPREPLY=( $(compgen -W "$enums" -- "$cur") )
    return
  fi

  provider=$(_runectx_suggest_provider_for_positional "$cmd|$((positional+1))")
  if [[ -z "$provider" ]]; then
    provider=$(_runectx_variadic_suggest_provider_for_command "$cmd")
    if [[ -n "$provider" ]]; then
      variadic_start=$(_runectx_variadic_suggest_start_for_command "$cmd")
      if [[ -z "$variadic_start" || $((positional+1)) -lt "$variadic_start" ]]; then
        provider=""
      fi
    fi
  fi
  if [[ -n "$provider" ]]; then
    COMPREPLY=( $(_runectx_dynamic_suggest "$provider" "$cur" "$root_path") )
    return
  fi

  flags=$(_runectx_flags_for "$cmd")
  COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
}
`
