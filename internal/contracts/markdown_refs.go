package contracts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var markdownFragmentPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type MarkdownArtifact struct {
	Path     string
	Headings map[string]string
	Refs     []MarkdownDeepRef
}

type MarkdownDeepRef struct {
	Raw      string
	Path     string
	Fragment string
	Start    int
	End      int
}

type MarkdownReferenceRewrite struct {
	OldPath     string
	NewPath     string
	OldFragment string
	NewFragment string
}

func indexMarkdownArtifact(index *ProjectIndex, contentRoot, path string, data []byte, hasFrontmatter bool) error {
	rel, err := filepath.Rel(contentRoot, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	body := strings.ReplaceAll(string(data), "\r\n", "\n")
	if hasFrontmatter {
		doc, err := parseFrontmatterMarkdown(path, data)
		if err != nil {
			return err
		}
		body = doc.Body
	}
	headings, err := extractMarkdownHeadingFragments(body)
	if err != nil {
		return &ValidationError{Path: rel, Message: err.Error()}
	}
	refs, err := extractMarkdownDeepRefs(body)
	if err != nil {
		return &ValidationError{Path: rel, Message: err.Error()}
	}
	index.MarkdownFiles[rel] = &MarkdownArtifact{
		Path:     rel,
		Headings: headings,
		Refs:     refs,
	}
	return nil
}

func validateMarkdownDeepRefs(index *ProjectIndex) error {
	for _, path := range SortedKeys(index.MarkdownFiles) {
		artifact := index.MarkdownFiles[path]
		for _, ref := range artifact.Refs {
			if err := validateMarkdownDeepRefShape(ref); err != nil {
				return &ValidationError{Path: path, Message: err.Error()}
			}
			target := index.MarkdownFiles[ref.Path]
			if target == nil {
				return &ValidationError{Path: path, Message: fmt.Sprintf("markdown deep ref %q points to missing artifact %q", ref.Raw, ref.Path)}
			}
			if _, ok := target.Headings[ref.Fragment]; !ok {
				return &ValidationError{Path: path, Message: fmt.Sprintf("markdown deep ref %q points to missing heading fragment %q in %q", ref.Raw, ref.Fragment, ref.Path)}
			}
		}
	}
	return nil
}

func RewriteMarkdownReferenceTargets(data []byte, rewrites []MarkdownReferenceRewrite) ([]byte, int, error) {
	if len(rewrites) == 0 {
		return append([]byte(nil), data...), 0, nil
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	segments := markdownTextSegments(text)
	var out strings.Builder
	count := 0
	for _, segment := range segments {
		if segment.fenced {
			out.WriteString(segment.text)
			continue
		}
		refs, err := extractMarkdownDeepRefs(segment.text)
		if err != nil {
			return nil, 0, err
		}
		if len(refs) == 0 {
			out.WriteString(segment.text)
			continue
		}
		last := 0
		for _, ref := range refs {
			out.WriteString(segment.text[last:ref.Start])
			updated := ref
			matched := false
			for _, rewrite := range rewrites {
				if rewrite.OldPath != "" && ref.Path != filepath.ToSlash(rewrite.OldPath) {
					continue
				}
				if rewrite.OldFragment != "" && ref.Fragment != rewrite.OldFragment {
					continue
				}
				if rewrite.NewPath != "" {
					updated.Path = filepath.ToSlash(rewrite.NewPath)
				}
				if rewrite.NewFragment != "" {
					updated.Fragment = rewrite.NewFragment
				}
				matched = true
				break
			}
			if matched {
				count++
				out.WriteString(updated.Path)
				out.WriteByte('#')
				out.WriteString(updated.Fragment)
			} else {
				out.WriteString(ref.Raw)
			}
			last = ref.End
		}
		out.WriteString(segment.text[last:])
	}
	return []byte(out.String()), count, nil
}

func extractMarkdownHeadingFragments(body string) (map[string]string, error) {
	headings := map[string]string{}
	counts := map[string]int{}
	for _, segment := range markdownTextSegments(strings.ReplaceAll(body, "\r\n", "\n")) {
		if segment.fenced {
			continue
		}
		for _, line := range strings.Split(segment.text, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "#") {
				continue
			}
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			if level == 0 || level > 6 || len(trimmed) <= level || trimmed[level] != ' ' {
				continue
			}
			heading := strings.TrimSpace(trimmed[level:])
			heading = strings.TrimSpace(strings.TrimRight(heading, "#"))
			if heading == "" {
				continue
			}
			base := slugifyHeadingFragment(heading)
			fragment := base
			if _, exists := headings[fragment]; exists {
				next := counts[base] + 2
				for {
					fragment = fmt.Sprintf("%s-%d", base, next)
					if _, exists := headings[fragment]; !exists {
						break
					}
					next++
				}
				counts[base] = next - 1
			} else if counts[base] == 0 {
				counts[base] = 1
			}
			if fragment == base {
				counts[base] = 1
			}
			headings[fragment] = heading
		}
	}
	if len(headings) == 0 {
		return headings, nil
	}
	return headings, nil
}

func extractMarkdownDeepRefs(body string) ([]MarkdownDeepRef, error) {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	segments := markdownTextSegments(body)
	refs := make([]MarkdownDeepRef, 0)
	for _, segment := range segments {
		if segment.fenced {
			continue
		}
		segmentRefs, err := extractMarkdownDeepRefsFromText(segment.text, segment.offset)
		if err != nil {
			return nil, err
		}
		refs = append(refs, segmentRefs...)
	}
	return refs, nil
}

func extractMarkdownDeepRefsFromText(text string, baseOffset int) ([]MarkdownDeepRef, error) {
	refs := make([]MarkdownDeepRef, 0)
	for i := 0; i < len(text); i++ {
		if text[i] != '#' {
			continue
		}
		fragmentEnd := i + 1
		for fragmentEnd < len(text) && isMarkdownFragmentChar(rune(text[fragmentEnd])) {
			fragmentEnd++
		}
		if fragmentEnd == i+1 {
			continue
		}
		pathStart := i - 1
		for pathStart >= 0 && isMarkdownPathChar(rune(text[pathStart])) {
			pathStart--
		}
		pathStart++
		if pathStart >= i {
			continue
		}
		candidatePath := text[pathStart:i]
		if !strings.HasSuffix(candidatePath, ".md") {
			continue
		}
		if pathStart > 0 && isMarkdownPathChar(rune(text[pathStart-1])) {
			continue
		}
		ref := MarkdownDeepRef{
			Raw:      text[pathStart:fragmentEnd],
			Path:     filepath.ToSlash(candidatePath),
			Fragment: text[i+1 : fragmentEnd],
			Start:    baseOffset + pathStart,
			End:      baseOffset + fragmentEnd,
		}
		if err := validateMarkdownDeepRefShape(ref); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
		i = fragmentEnd - 1
	}
	return refs, nil
}

func validateMarkdownDeepRefShape(ref MarkdownDeepRef) error {
	if strings.HasPrefix(ref.Path, "/") {
		return fmt.Errorf("markdown deep ref %q must not use an absolute path", ref.Raw)
	}
	if strings.HasPrefix(ref.Path, "./") || strings.HasPrefix(ref.Path, "../") {
		return fmt.Errorf("markdown deep ref %q must use a RuneContext-root-relative path", ref.Raw)
	}
	if strings.Contains(ref.Path, "../") || strings.Contains(ref.Path, "/../") || strings.HasSuffix(ref.Path, "/..") || ref.Path == ".." {
		return fmt.Errorf("markdown deep ref %q must not use traversal segments", ref.Raw)
	}
	if strings.Contains(ref.Path, "//") {
		return fmt.Errorf("markdown deep ref %q must not contain empty path segments", ref.Raw)
	}
	if isLineNumberFragment(ref.Fragment) {
		return fmt.Errorf("markdown deep ref %q must use a heading fragment, not a line-number fragment", ref.Raw)
	}
	if !markdownFragmentPattern.MatchString(ref.Fragment) {
		return fmt.Errorf("markdown deep ref %q must use a lowercase heading fragment slug", ref.Raw)
	}
	return nil
}

func slugifyHeadingFragment(heading string) string {
	heading = strings.ToLower(strings.TrimSpace(heading))
	var b strings.Builder
	lastDash := false
	for _, r := range heading {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || unicode.IsSpace(r) || r == '_' || r == '/':
			if b.Len() == 0 || lastDash {
				continue
			}
			b.WriteByte('-')
			lastDash = true
		}
	}
	fragment := strings.Trim(b.String(), "-")
	if fragment == "" {
		return "section"
	}
	return fragment
}

func isLineNumberFragment(fragment string) bool {
	if fragment == "" {
		return false
	}
	if len(fragment) > 1 && (fragment[0] == 'L' || fragment[0] == 'l') {
		allDigits := true
		for _, r := range fragment[1:] {
			if !unicode.IsDigit(r) {
				allDigits = false
				break
			}
		}
		if allDigits {
			return true
		}
	}
	if strings.HasPrefix(fragment, "L") || strings.HasPrefix(fragment, "l") {
		parts := strings.Split(fragment[1:], "-")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			allDigits := true
			for _, part := range parts {
				for _, r := range part {
					if !unicode.IsDigit(r) {
						allDigits = false
						break
					}
				}
			}
			if allDigits {
				return true
			}
		}
	}
	if allDigits(fragment) {
		return true
	}
	return false
}

type markdownSegment struct {
	text   string
	offset int
	fenced bool
}

func markdownTextSegments(text string) []markdownSegment {
	segments := make([]markdownSegment, 0)
	lines := strings.SplitAfter(text, "\n")
	offset := 0
	currentStart := 0
	currentFence := false
	fenceMarker := ""
	flush := func(end int) {
		if end <= currentStart {
			return
		}
		segments = append(segments, markdownSegment{text: text[currentStart:end], offset: currentStart, fenced: currentFence})
		currentStart = end
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		marker := markdownFenceMarker(trimmed)
		if marker == "" {
			offset += len(line)
			continue
		}
		if !currentFence {
			flush(offset)
			currentFence = true
			fenceMarker = marker
			offset += len(line)
			continue
		}
		if marker == fenceMarker {
			offset += len(line)
			flush(offset)
			currentFence = false
			fenceMarker = ""
			continue
		}
		offset += len(line)
	}
	flush(len(text))
	return segments
}

func markdownFenceMarker(trimmed string) string {
	if strings.HasPrefix(trimmed, "```") {
		return "```"
	}
	if strings.HasPrefix(trimmed, "~~~") {
		return "~~~"
	}
	return ""
}

func isMarkdownPathChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '/' || r == '-'
}

func isMarkdownFragmentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
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
