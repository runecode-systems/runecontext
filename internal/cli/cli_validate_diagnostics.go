package cli

import "strings"

func validateDiagnosticLines(prefix string, diagnostic emittedDiagnostic) []line {
	lines := []line{
		{prefix + "_severity", string(diagnostic.Severity)},
		{prefix + "_code", diagnostic.Code},
		{prefix + "_message", diagnostic.Message},
	}
	lines = appendOptionalDiagnosticLine(lines, prefix+"_path", diagnostic.Path)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_bundle", diagnostic.Bundle)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_aspect", diagnostic.Aspect)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_rule", diagnostic.Rule)
	lines = appendOptionalDiagnosticLine(lines, prefix+"_pattern", diagnostic.Pattern)
	if len(diagnostic.Matches) > 0 {
		lines = append(lines, line{prefix + "_matches", strings.Join(diagnostic.Matches, ",")})
	}
	return lines
}

func appendOptionalDiagnosticLine(lines []line, key, value string) []line {
	if value == "" {
		return lines
	}
	return append(lines, line{key, value})
}
