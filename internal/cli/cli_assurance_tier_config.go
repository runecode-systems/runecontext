package cli

import (
	"bytes"
	"regexp"
)

var assuranceTierRegex = regexp.MustCompile(`(?m)^(\s*assurance_tier\s*:\s*)([^#\r\n]*)(\s*(?:#.*)?)(\r?\n|$)`)

func ensureAssuranceTierConfig(data []byte) ([]byte, bool) {
	lineEnding := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		lineEnding = "\r\n"
	}
	replaced := false
	replacer := func(match []byte) []byte {
		replaced = true
		submatches := assuranceTierRegex.FindSubmatch(match)
		if len(submatches) >= 5 {
			suffix := submatches[3]
			if len(suffix) > 0 && suffix[0] == '#' {
				suffix = append([]byte(" "), suffix...)
			}
			line := append([]byte{}, submatches[1]...)
			line = append(line, []byte("verified")...)
			line = append(line, suffix...)
			line = append(line, submatches[4]...)
			return line
		}
		return []byte("assurance_tier: verified")
	}
	updated := assuranceTierRegex.ReplaceAllFunc(data, replacer)
	if replaced {
		return updated, true
	}
	buffer := bytes.NewBuffer(updated)
	if len(updated) > 0 && !bytes.HasSuffix(updated, []byte("\n")) {
		buffer.WriteString(lineEnding)
	}
	buffer.WriteString("assurance_tier: verified")
	buffer.WriteString(lineEnding)
	return buffer.Bytes(), false
}
