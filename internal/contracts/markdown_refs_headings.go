package contracts

import (
	"fmt"
	"strings"
)

func extractMarkdownHeadingFragments(body string) (map[string]string, error) {
	headings := map[string]string{}
	counts := map[string]int{}
	for _, heading := range markdownHeadings(body) {
		fragment := allocateHeadingFragment(heading, headings, counts)
		headings[fragment] = heading
	}
	return headings, nil
}

func markdownHeadings(body string) []string {
	results := make([]string, 0)
	for _, segment := range markdownTextSegments(strings.ReplaceAll(body, "\r\n", "\n")) {
		if segment.fenced {
			continue
		}
		for _, line := range strings.Split(segment.text, "\n") {
			if heading, ok := parseMarkdownHeading(line); ok {
				results = append(results, heading)
			}
		}
	}
	return results
}

func parseMarkdownHeading(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	level := markdownHeadingLevel(trimmed)
	if level == 0 || len(trimmed) <= level || trimmed[level] != ' ' {
		return "", false
	}
	heading := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(trimmed[level:]), "#"))
	if heading == "" {
		return "", false
	}
	return heading, true
}

func markdownHeadingLevel(trimmed string) int {
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	return level
}

func allocateHeadingFragment(heading string, headings map[string]string, counts map[string]int) string {
	base := slugifyHeadingFragment(heading)
	if _, exists := headings[base]; !exists {
		counts[base] = 0
		return base
	}
	next := counts[base] + 1
	for {
		fragment := fmt.Sprintf("%s-%d", base, next)
		if _, exists := headings[fragment]; !exists {
			counts[base] = next
			return fragment
		}
		next++
	}
}
