package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func extractMarkdownDeepRefsFromText(text string, baseOffset int) ([]MarkdownDeepRef, error) {
	refs := make([]MarkdownDeepRef, 0)
	for i := 0; i < len(text); i++ {
		ref, next, ok, err := nextMarkdownDeepRef(text, i, baseOffset)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		refs = append(refs, ref)
		i = next - 1
	}
	return refs, nil
}

func nextMarkdownDeepRef(text string, hashIndex, baseOffset int) (MarkdownDeepRef, int, bool, error) {
	if text[hashIndex] != '#' {
		return MarkdownDeepRef{}, hashIndex + 1, false, nil
	}
	fragmentEnd := utf8TokenEnd(text, hashIndex+1, isMarkdownFragmentChar)
	if fragmentEnd == hashIndex+1 {
		return MarkdownDeepRef{}, fragmentEnd, false, nil
	}
	pathStart := utf8TokenStart(text, hashIndex, isMarkdownPathChar)
	if pathStart >= hashIndex {
		return MarkdownDeepRef{}, fragmentEnd, false, nil
	}
	ref, ok := markdownDeepRefCandidate(text, pathStart, hashIndex, fragmentEnd, baseOffset)
	if !ok {
		return MarkdownDeepRef{}, fragmentEnd, false, nil
	}
	if isLikelyExternalMarkdownURL(text, pathStart) {
		return MarkdownDeepRef{}, fragmentEnd, false, nil
	}
	if err := validateMarkdownDeepRefShape(ref); err != nil {
		return MarkdownDeepRef{}, fragmentEnd, false, err
	}
	return ref, fragmentEnd, true, nil
}

func markdownDeepRefCandidate(text string, pathStart, hashIndex, fragmentEnd, baseOffset int) (MarkdownDeepRef, bool) {
	candidatePath := text[pathStart:hashIndex]
	if !isValidMarkdownRefPathCandidate(candidatePath, text, pathStart) {
		return MarkdownDeepRef{}, false
	}
	return MarkdownDeepRef{
		Raw:      text[pathStart:fragmentEnd],
		Path:     filepath.ToSlash(candidatePath),
		Fragment: text[hashIndex+1 : fragmentEnd],
		Start:    baseOffset + pathStart,
		End:      baseOffset + fragmentEnd,
	}, true
}

func isValidMarkdownRefPathCandidate(candidatePath, text string, pathStart int) bool {
	if !isIndexedMarkdownDeepRefCandidatePath(candidatePath) || !strings.HasSuffix(candidatePath, ".md") {
		return false
	}
	if strings.Contains(candidatePath, "://") || strings.HasPrefix(candidatePath, "//") {
		return false
	}
	if pathStart >= 2 && text[pathStart-2:pathStart] == "//" {
		return false
	}
	return true
}

func utf8TokenEnd(text string, start int, allow func(rune) bool) int {
	pos := start
	for pos < len(text) {
		r, size := utf8.DecodeRuneInString(text[pos:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if !allow(r) {
			break
		}
		pos += size
	}
	return pos
}

func utf8TokenStart(text string, end int, allow func(rune) bool) int {
	pos := end
	for pos > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:pos])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if !allow(r) {
			break
		}
		pos -= size
	}
	return pos
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
