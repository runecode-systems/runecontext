package cli

import "strings"

func shellSingleQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func fishSingleQuote(value string) string {
	return shellSingleQuote(value)
}

func fishConditionForPath(path string) string {
	if path == "" {
		return "true"
	}
	parts := strings.Fields(path)
	if len(parts) == 0 {
		return "true"
	}
	return "__fish_seen_subcommand_from " + strings.Join(parts, " ")
}
