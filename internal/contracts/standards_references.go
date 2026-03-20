package contracts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var plainStandardPathPattern = regexp.MustCompile("`(standards/[A-Za-z0-9][A-Za-z0-9._/-]*\\.md)`")

func validateStandardReferenceBodies(index *ProjectIndex) error {
	if err := validateChangeProposalStandardReferenceBodies(index); err != nil {
		return err
	}
	if err := validateSpecStandardReferenceBodies(index); err != nil {
		return err
	}
	return nil
}

func validateChangeProposalStandardReferenceBodies(index *ProjectIndex) error {
	for _, changeID := range SortedKeys(index.Changes) {
		record := index.Changes[changeID]
		proposalPath, err := changeArtifactRelativePath(index, record, "proposal.md")
		if err != nil {
			return err
		}
		artifact := index.MarkdownFiles[proposalPath]
		if artifact == nil {
			continue
		}
		if err := validateMarkdownStandardReferences(proposalPath, artifact, index); err != nil {
			return err
		}
	}
	return nil
}

func validateSpecStandardReferenceBodies(index *ProjectIndex) error {
	for _, path := range SortedKeys(index.Specs) {
		artifact := index.MarkdownFiles[path]
		if artifact == nil {
			continue
		}
		if err := validateMarkdownStandardReferences(path, artifact, index); err != nil {
			return err
		}
	}
	return nil
}

func validateMarkdownStandardReferences(path string, artifact *MarkdownArtifact, index *ProjectIndex) error {
	if err := validateMarkdownDeepStandardRefs(path, artifact, index); err != nil {
		return err
	}
	if err := validatePlainStandardPathRefs(path, artifact, index); err != nil {
		return err
	}
	return validateCopiedStandardBody(path, artifact, index)
}

func validateMarkdownDeepStandardRefs(path string, artifact *MarkdownArtifact, index *ProjectIndex) error {
	for _, ref := range artifact.Refs {
		if strings.HasPrefix(ref.Path, "standards/") {
			if _, ok := index.StandardPaths[ref.Path]; !ok {
				return &ValidationError{Path: path, Message: fmt.Sprintf("standard reference %q points to missing standard %q", ref.Raw, ref.Path)}
			}
		}
	}
	return nil
}

func validatePlainStandardPathRefs(path string, artifact *MarkdownArtifact, index *ProjectIndex) error {
	for _, refPath := range plainStandardPathRefs(artifact.Body) {
		if _, ok := index.StandardPaths[refPath.refPath]; !ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("standard path reference %q points to missing standard %q", refPath.raw, refPath.refPath)}
		}
	}
	return nil
}

type plainStandardPathRef struct {
	raw     string
	refPath string
}

func plainStandardPathRefs(body string) []plainStandardPathRef {
	refs := make([]plainStandardPathRef, 0)
	for _, segment := range markdownTextSegments(strings.ReplaceAll(body, "\r\n", "\n")) {
		if segment.fenced {
			continue
		}
		for _, match := range plainStandardPathPattern.FindAllStringSubmatch(segment.text, -1) {
			if len(match) >= 2 {
				refs = append(refs, plainStandardPathRef{raw: match[0], refPath: filepath.ToSlash(match[1])})
			}
		}
	}
	return refs
}

func validateCopiedStandardBody(path string, artifact *MarkdownArtifact, index *ProjectIndex) error {
	body := normalizeComparableMarkdownText(nonFencedComparableMarkdownText(artifact.Body))
	if body == "" {
		return nil
	}
	for _, standardPath := range SortedKeys(index.StandardPaths) {
		if copied, err := copiedStandardBodyError(path, body, standardPath, index); err != nil || copied {
			if err != nil {
				return err
			}
			return &ValidationError{Path: path, Message: fmt.Sprintf("markdown body appears to copy standard content from %q; reference the standard by path instead", standardPath)}
		}
	}
	return nil
}

func copiedStandardBodyError(path, body, standardPath string, index *ProjectIndex) (bool, error) {
	standardArtifact := index.MarkdownFiles[standardPath]
	if standardArtifact == nil {
		return false, nil
	}
	snippet := normalizeComparableMarkdownText(extractMarkdownComparableStandardBody(nonFencedComparableMarkdownText(standardArtifact.Body)))
	if snippet == "" {
		return false, nil
	}
	return strings.Contains(body, snippet), nil
}

func nonFencedComparableMarkdownText(body string) string {
	segments := markdownTextSegments(strings.ReplaceAll(body, "\r\n", "\n"))
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment.fenced {
			continue
		}
		parts = append(parts, segment.text)
	}
	return strings.Join(parts, "")
}

func extractMarkdownComparableStandardBody(body string) string {
	text := strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func normalizeComparableMarkdownText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	var builder strings.Builder
	lastSpace := true
	for _, r := range text {
		if unicode.IsSpace(r) {
			if lastSpace {
				continue
			}
			builder.WriteByte(' ')
			lastSpace = true
			continue
		}
		builder.WriteRune(unicode.ToLower(r))
		lastSpace = false
	}
	return strings.TrimSpace(builder.String())
}
