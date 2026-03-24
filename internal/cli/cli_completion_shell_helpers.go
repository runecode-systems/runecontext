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

func shellToken(value string) string {
	if value == "" {
		return shellSingleQuote(value)
	}
	for _, r := range value {
		if !isShellBarewordRune(r) {
			return shellSingleQuote(value)
		}
	}
	return value
}

func fishToken(value string) string {
	return shellToken(value)
}

func isShellBarewordRune(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' || r == '-' || r == '.'
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
