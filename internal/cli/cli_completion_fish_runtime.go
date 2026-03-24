package cli

const fishPositionalRuntimeHelpers = `function __runectx_count_positionals_for_path
  set -l path_parts $argv
  if test (count $path_parts) -eq 0
    return 1
  end
  set -l tokens (commandline -opc)
  set -l path_start 0
  for i in (seq 1 (count $tokens))
    set -l matched 1
    for j in (seq 1 (count $path_parts))
      set -l k (math $i + $j - 1)
      if test $k -gt (count $tokens)
        set matched 0
        break
      end
      if test "$tokens[$k]" != "$path_parts[$j]"
        set matched 0
        break
      end
    end
    if test $matched -eq 1
      set path_start (math $i + (count $path_parts))
    end
  end
  if test $path_start -eq 0
    return 1
  end
  set -l positionals 0
  set -l expect_value 0
  for i in (seq $path_start (count $tokens))
    set -l token $tokens[$i]
    if test $expect_value -eq 1
      set expect_value 0
      continue
    end
    if string match -q -- '--*' $token
      if string match -q -- '*=*' $token
        continue
      end
      if __runectx_flag_name_takes_value $token
        set expect_value 1
      end
      continue
    end
    set positionals (math $positionals + 1)
  end
  echo $positionals
  return 0
end

function __runectx_positional_index_at_least
  set -l min $argv[1]
  set -e argv[1]
  set -l count (__runectx_count_positionals_for_path $argv)
  if test -z "$count"
    return 1
  end
  test (math $count + 1) -ge $min
end

`
