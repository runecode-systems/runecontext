package contracts

import (
	"path/filepath"
	"strings"
)

func RewriteMarkdownReferenceTargets(data []byte, rewrites []MarkdownReferenceRewrite) ([]byte, int, error) {
	if len(rewrites) == 0 {
		return append([]byte(nil), data...), 0, nil
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	segments := markdownTextSegments(text)
	var out strings.Builder
	count := 0
	for _, segment := range segments {
		rewritten, rewritesApplied, err := rewriteMarkdownSegment(segment, rewrites)
		if err != nil {
			return nil, 0, err
		}
		out.WriteString(rewritten)
		count += rewritesApplied
	}
	return []byte(out.String()), count, nil
}

func rewriteMarkdownSegment(segment markdownSegment, rewrites []MarkdownReferenceRewrite) (string, int, error) {
	if segment.fenced {
		return segment.text, 0, nil
	}
	refs, err := extractMarkdownDeepRefs(segment.text)
	if err != nil {
		return "", 0, err
	}
	if len(refs) == 0 {
		return segment.text, 0, nil
	}
	return rewriteMarkdownReferences(segment.text, refs, rewrites), countMatchedMarkdownRewrites(refs, rewrites), nil
}

func rewriteMarkdownReferences(text string, refs []MarkdownDeepRef, rewrites []MarkdownReferenceRewrite) string {
	var out strings.Builder
	last := 0
	for _, ref := range refs {
		out.WriteString(text[last:ref.Start])
		updated, matched := applyMarkdownReferenceRewrite(ref, rewrites)
		if matched {
			out.WriteString(updated.Path)
			out.WriteByte('#')
			out.WriteString(updated.Fragment)
		} else {
			out.WriteString(ref.Raw)
		}
		last = ref.End
	}
	out.WriteString(text[last:])
	return out.String()
}

func countMatchedMarkdownRewrites(refs []MarkdownDeepRef, rewrites []MarkdownReferenceRewrite) int {
	count := 0
	for _, ref := range refs {
		if _, matched := applyMarkdownReferenceRewrite(ref, rewrites); matched {
			count++
		}
	}
	return count
}

func applyMarkdownReferenceRewrite(ref MarkdownDeepRef, rewrites []MarkdownReferenceRewrite) (MarkdownDeepRef, bool) {
	updated := ref
	for _, rewrite := range rewrites {
		if !matchesMarkdownRewrite(ref, rewrite) {
			continue
		}
		if rewrite.NewPath != "" {
			updated.Path = filepath.ToSlash(rewrite.NewPath)
		}
		if rewrite.NewFragment != "" {
			updated.Fragment = rewrite.NewFragment
		}
		return updated, true
	}
	return updated, false
}

func matchesMarkdownRewrite(ref MarkdownDeepRef, rewrite MarkdownReferenceRewrite) bool {
	if rewrite.OldPath != "" && ref.Path != filepath.ToSlash(rewrite.OldPath) {
		return false
	}
	if rewrite.OldFragment != "" && ref.Fragment != rewrite.OldFragment {
		return false
	}
	return true
}
