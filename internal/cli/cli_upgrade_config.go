package cli

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func rewriteRunecontextVersion(data []byte, target string) ([]byte, error) {
	raw := string(data)
	newLine := "\n"
	if strings.Contains(raw, "\r\n") {
		newLine = "\r\n"
	}
	hasTrailingNewline := strings.HasSuffix(raw, newLine)
	lines := splitLinesPreserveStyle(raw, newLine)
	for i, line := range lines {
		updated, ok := rewriteRunecontextVersionLine(line, target)
		if !ok {
			continue
		}
		lines[i] = updated
		rendered := strings.Join(lines, newLine)
		if hasTrailingNewline {
			rendered += newLine
		}
		return []byte(rendered), nil
	}
	if err := validateRunecontextVersionKeyPresent(data); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("runecontext.yaml contains runecontext_version but it is not a rewriteable scalar value")
}

func splitLinesPreserveStyle(raw, newLine string) []string {
	if strings.HasSuffix(raw, newLine) {
		trimmed := strings.TrimSuffix(raw, newLine)
		if trimmed == "" {
			return []string{""}
		}
		return strings.Split(trimmed, newLine)
	}
	return strings.Split(raw, newLine)
}

func rewriteRunecontextVersionLine(line, target string) (string, bool) {
	commentIndex := indexOfCommentStart(line)
	valueSection := line
	trailingComment := ""
	if commentIndex >= 0 {
		valueSection = line[:commentIndex]
		trailingComment = line[commentIndex:]
	}
	beforeColon, afterColon, ok := parseRunecontextVersionValueSection(valueSection)
	if !ok {
		return "", false
	}
	beforeValue, currentValue, afterValue, ok := splitScalarValue(afterColon)
	if !ok || strings.TrimSpace(currentValue) == "" {
		return "", false
	}
	replacement := target
	if quote := wrappingQuote(currentValue); quote != "" {
		replacement = quote + target + quote
	}
	rewritten := beforeColon + ":" + beforeValue + replacement + afterValue
	if trailingComment != "" {
		rewritten += trailingComment
	}
	return rewritten, true
}

func parseRunecontextVersionValueSection(valueSection string) (string, string, bool) {
	colonIndex := strings.Index(valueSection, ":")
	if colonIndex < 0 {
		return "", "", false
	}
	beforeColon := valueSection[:colonIndex]
	if strings.TrimSpace(beforeColon) != "runecontext_version" {
		return "", "", false
	}
	afterColon := valueSection[colonIndex+1:]
	return beforeColon, afterColon, true
}

func splitScalarValue(afterColon string) (string, string, string, bool) {
	trimmedLeft := strings.TrimLeft(afterColon, " \t")
	beforeValue := afterColon[:len(afterColon)-len(trimmedLeft)]
	if trimmedLeft == "" {
		return "", "", "", false
	}
	if quote := trimmedLeft[:1]; quote == "\"" || quote == "'" {
		if end := strings.Index(trimmedLeft[1:], quote); end >= 0 {
			value := trimmedLeft[:end+2]
			afterValue := trimmedLeft[end+2:]
			return beforeValue, value, afterValue, true
		}
		return "", "", "", false
	}
	valueEnd := firstWhitespaceIndex(trimmedLeft)
	if valueEnd < 0 {
		return beforeValue, trimmedLeft, "", true
	}
	value := trimmedLeft[:valueEnd]
	afterValue := trimmedLeft[valueEnd:]
	return beforeValue, value, afterValue, true
}

func firstWhitespaceIndex(value string) int {
	for i := 0; i < len(value); i++ {
		if value[i] == ' ' || value[i] == '\t' {
			return i
		}
	}
	return -1
}

func wrappingQuote(value string) string {
	if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return "\""
	}
	if len(value) >= 2 && strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return "'"
	}
	return ""
}

func indexOfCommentStart(line string) int {
	for i := 1; i < len(line); i++ {
		if line[i] == '#' && (line[i-1] == ' ' || line[i-1] == '\t') {
			return i
		}
	}
	return -1
}

func validateRunecontextVersionKeyPresent(data []byte) error {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("parse runecontext.yaml: %w", err)
	}
	if len(document.Content) == 0 {
		return fmt.Errorf("runecontext.yaml is missing root document")
	}
	mapping := document.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("runecontext.yaml root must be a mapping")
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		if key.Kind == yaml.ScalarNode && key.Value == "runecontext_version" {
			return nil
		}
	}
	return fmt.Errorf("runecontext.yaml is missing runecontext_version")
}

func configFileMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0o644
	}
	return info.Mode().Perm()
}

func writeAtomicUpgradeConfig(path string, data []byte, mode os.FileMode) error {
	return writeAtomicFile(path, data, mode)
}
