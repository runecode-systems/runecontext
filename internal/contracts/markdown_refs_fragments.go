package contracts

import (
	"path/filepath"
	"strings"
	"unicode"
)

func isLineNumberFragment(fragment string) bool {
	if fragment == "" {
		return false
	}
	if allDigits(fragment) {
		return true
	}
	if !strings.HasPrefix(fragment, "L") && !strings.HasPrefix(fragment, "l") {
		return false
	}
	body := fragment[1:]
	if allDigits(body) {
		return true
	}
	return isLineRangeFragment(body)
}

func isLineRangeFragment(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return allDigits(parts[0]) && allDigits(parts[1])
}

func markdownTextSegments(text string) []markdownSegment {
	segments := make([]markdownSegment, 0)
	lines := strings.SplitAfter(text, "\n")
	offset := 0
	currentStart := 0
	currentFence := false
	fence := markdownFence{}
	flush := func(end int) {
		if end <= currentStart {
			return
		}
		segments = append(segments, markdownSegment{text: text[currentStart:end], offset: currentStart, fenced: currentFence})
		currentStart = end
	}
	for _, line := range lines {
		marker, ok := markdownFenceMarker(strings.TrimSpace(line))
		if !ok {
			offset += len(line)
			continue
		}
		if !currentFence {
			flush(offset)
			currentFence = true
			fence = marker
			offset += len(line)
			continue
		}
		if marker.char == fence.char && marker.length >= fence.length {
			offset += len(line)
			flush(offset)
			currentFence = false
			fence = markdownFence{}
			continue
		}
		offset += len(line)
	}
	flush(len(text))
	return segments
}

type markdownSegment struct {
	text   string
	offset int
	fenced bool
}

type markdownFence struct {
	char   byte
	length int
}

func markdownFenceMarker(trimmed string) (markdownFence, bool) {
	trimmed = strings.TrimLeft(trimmed, "> ")
	if len(trimmed) < 3 {
		return markdownFence{}, false
	}
	first := trimmed[0]
	if first != '`' && first != '~' {
		return markdownFence{}, false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == first {
		length++
	}
	if length < 3 {
		return markdownFence{}, false
	}
	return markdownFence{char: first, length: length}, true
}

func isMarkdownPathChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '/' || r == '-'
}

func isMarkdownFragmentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || unicode.IsDigit(r) || r == '-' || r == '_'
}

func isLikelyExternalMarkdownURL(text string, pathStart int) bool {
	if pathStart < 3 {
		return false
	}
	prefix := text[:pathStart]
	if lastSpace := strings.LastIndexAny(prefix, " \t\n\r<(["); lastSpace >= 0 {
		prefix = prefix[lastSpace+1:]
	}
	return strings.Contains(prefix, "://")
}

func isIndexedMarkdownDeepRefCandidatePath(path string) bool {
	trimmed := strings.TrimPrefix(filepath.ToSlash(path), "/")
	for {
		switch {
		case strings.HasPrefix(trimmed, "./"):
			trimmed = strings.TrimPrefix(trimmed, "./")
		case strings.HasPrefix(trimmed, "../"):
			trimmed = strings.TrimPrefix(trimmed, "../")
		default:
			for _, root := range []string{"changes/", "specs/", "decisions/", "standards/"} {
				if strings.HasPrefix(trimmed, root) {
					return true
				}
			}
			return false
		}
	}
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
