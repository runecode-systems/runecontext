package contracts

import (
	"path/filepath"
	"regexp"
	"strings"
)

var markdownFragmentPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type MarkdownArtifact struct {
	Path     string
	Body     string
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
	rel, body, err := prepareMarkdownArtifact(contentRoot, path, data, hasFrontmatter)
	if err != nil {
		return err
	}
	headings, err := extractMarkdownHeadingFragments(body)
	if err != nil {
		return &ValidationError{Path: rel, Message: err.Error()}
	}
	refs, err := extractMarkdownDeepRefs(body)
	if err != nil {
		return &ValidationError{Path: rel, Message: err.Error()}
	}
	index.MarkdownFiles[rel] = &MarkdownArtifact{Path: rel, Body: body, Headings: headings, Refs: refs}
	return nil
}

func prepareMarkdownArtifact(contentRoot, path string, data []byte, hasFrontmatter bool) (string, string, error) {
	rel, err := filepath.Rel(contentRoot, path)
	if err != nil {
		return "", "", err
	}
	body := strings.ReplaceAll(string(data), "\r\n", "\n")
	if hasFrontmatter {
		doc, err := parseFrontmatterMarkdown(path, data)
		if err != nil {
			return "", "", err
		}
		body = doc.Body
	}
	return filepath.ToSlash(rel), body, nil
}

func validateMarkdownDeepRefs(index *ProjectIndex) error {
	for _, path := range SortedKeys(index.MarkdownFiles) {
		if err := validateMarkdownArtifactRefs(path, index.MarkdownFiles[path], index); err != nil {
			return err
		}
	}
	return nil
}

func validateMarkdownArtifactRefs(path string, artifact *MarkdownArtifact, index *ProjectIndex) error {
	for _, ref := range artifact.Refs {
		if err := validateMarkdownDeepRefTarget(path, ref, index); err != nil {
			return err
		}
	}
	return nil
}

func validateMarkdownDeepRefTarget(path string, ref MarkdownDeepRef, index *ProjectIndex) error {
	if err := validateMarkdownDeepRefShape(ref); err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	target := index.MarkdownFiles[ref.Path]
	if target == nil {
		return &ValidationError{Path: path, Message: "markdown deep ref " + quote(ref.Raw) + " points to missing artifact " + quote(ref.Path)}
	}
	if _, ok := target.Headings[ref.Fragment]; !ok {
		return &ValidationError{Path: path, Message: "markdown deep ref " + quote(ref.Raw) + " points to missing heading fragment " + quote(ref.Fragment) + " in " + quote(ref.Path)}
	}
	return nil
}

func quote(value string) string {
	return `"` + value + `"`
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

func slugifyHeadingFragment(heading string) string {
	return slugifyASCII(heading, "section")
}
