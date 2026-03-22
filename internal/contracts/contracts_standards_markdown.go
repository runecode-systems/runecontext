package contracts

import (
	"fmt"
	"strings"
)

func (v *Validator) ValidateStandardsMarkdown(path string, data []byte) error {
	_, err := parseStandardsMarkdown(path, data)
	return err
}

func parseStandardsMarkdown(path string, data []byte) (*MarkdownDocument, error) {
	sections, err := parseLevel2Sections(path, data)
	if err != nil {
		return nil, err
	}
	if err := validateStandardsSectionSequence(path, sections); err != nil {
		return nil, err
	}
	parsed, refs, refsBySection, err := collectStandardsMarkdownSections(path, sections)
	if err != nil {
		return nil, err
	}
	return &MarkdownDocument{Sections: parsed, Refs: uniqueSortedStrings(refs), RefsBySection: refsBySection}, nil
}

func validateStandardsSectionSequence(path string, sections []markdownSection) error {
	if len(sections) == 0 || sections[0].Heading != "Applicable Standards" {
		return &ValidationError{Path: path, Message: "missing required section \"Applicable Standards\""}
	}
	return validateStandardsSectionOrdering(path, sections)
}

func validateStandardsSectionOrdering(path string, sections []markdownSection) error {
	canonicalOrder := map[string]int{"Applicable Standards": 0, "Standards Added Since Last Refresh": 1, "Standards Considered But Excluded": 2, "Resolution Notes": 3}
	seen := map[string]struct{}{}
	lastCanonical := -1
	customStarted := false
	for _, section := range sections {
		var err error
		lastCanonical, customStarted, err = validateStandardsSectionOrderState(path, section, canonicalOrder, seen, lastCanonical, customStarted)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateStandardsSectionOrderState(path string, section markdownSection, canonicalOrder map[string]int, seen map[string]struct{}, lastCanonical int, customStarted bool) (int, bool, error) {
	if _, dup := seen[section.Heading]; dup {
		return lastCanonical, customStarted, &ValidationError{Path: path, Message: fmt.Sprintf("duplicate section %q", section.Heading)}
	}
	seen[section.Heading] = struct{}{}
	if section.Body == "" {
		return lastCanonical, customStarted, &ValidationError{Path: path, Message: fmt.Sprintf("section %q must not be empty", section.Heading)}
	}
	order, isCanonical := canonicalOrder[section.Heading]
	if !isCanonical {
		return lastCanonical, true, nil
	}
	if customStarted {
		return lastCanonical, customStarted, &ValidationError{Path: path, Message: fmt.Sprintf("canonical section %q cannot appear after custom sections", section.Heading)}
	}
	if order < lastCanonical {
		return lastCanonical, customStarted, &ValidationError{Path: path, Message: fmt.Sprintf("section %q appears out of order", section.Heading)}
	}
	return order, customStarted, nil
}

func collectStandardsMarkdownSections(path string, sections []markdownSection) (map[string]string, []string, map[string][]string, error) {
	parsed := map[string]string{}
	refs := make([]string, 0)
	refsBySection := map[string][]string{}
	for _, section := range sections {
		parsed[section.Heading] = section.Body
	}
	for _, heading := range []string{"Applicable Standards", "Standards Added Since Last Refresh", "Standards Considered But Excluded"} {
		body, ok := parsed[heading]
		if !ok {
			continue
		}
		if err := validateStandardsSectionReferences(path, heading, body); err != nil {
			return nil, nil, nil, err
		}
		sectionRefs := extractStandardsSectionReferences(body)
		refs = append(refs, sectionRefs...)
		refsBySection[heading] = append([]string(nil), sectionRefs...)
	}
	return parsed, refs, refsBySection, nil
}

func validateStandardsSectionReferences(path, heading, body string) error {
	if allowsEmptyApplicableStandardsBody(heading, body) {
		return nil
	}
	for _, item := range parseStandardsSectionItems(body) {
		if item.err != nil {
			return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list standards as bullet path references instead of copied body text", heading)}
		}
		refs := extractStandardsLikeBacktickedRefs(item.text)
		if len(refs) != 1 {
			return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list exactly one backticked standard path per bullet", heading)}
		}
		if !isCanonicalStandardPathRef(refs[0]) {
			return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list standards as backticked RuneContext-root-relative paths", heading)}
		}
	}
	return nil
}

func allowsEmptyApplicableStandardsBody(heading, body string) bool {
	if heading != "Applicable Standards" {
		return false
	}
	return strings.TrimSpace(body) == "N/A"
}

type standardsSectionItem struct {
	text string
	err  error
}

func parseStandardsSectionItems(body string) []standardsSectionItem {
	items := make([]standardsSectionItem, 0)
	inBullet := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			items = append(items, standardsSectionItem{text: strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))})
			inBullet = true
			continue
		}
		if inBullet && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			continue
		}
		items = append(items, standardsSectionItem{err: fmt.Errorf("invalid")})
	}
	return items
}

func extractStandardsSectionReferences(body string) []string {
	refs := make([]string, 0)
	for _, item := range parseStandardsSectionItems(body) {
		if item.err != nil {
			continue
		}
		lineRefs := extractStandardsLikeBacktickedRefs(item.text)
		if len(lineRefs) == 1 && isCanonicalStandardPathRef(lineRefs[0]) {
			refs = append(refs, lineRefs[0])
		}
	}
	return refs
}

func extractStandardsLikeBacktickedRefs(value string) []string {
	all := extractBacktickedPaths(value)
	refs := make([]string, 0, len(all))
	for _, ref := range all {
		if strings.HasPrefix(ref, "standards/") {
			refs = append(refs, ref)
		}
	}
	return refs
}

func extractBacktickedPaths(value string) []string {
	refs := make([]string, 0)
	for i := 0; i < len(value); i++ {
		if value[i] != '`' {
			continue
		}
		end := strings.Index(value[i+1:], "`")
		if end < 0 {
			break
		}
		refs = append(refs, value[i+1:i+1+end])
		i += end + 1
	}
	return refs
}

func isCanonicalStandardPathRef(ref string) bool {
	if strings.Contains(ref, "#") {
		return false
	}
	if !strings.HasPrefix(ref, "standards/") || !strings.HasSuffix(ref, ".md") {
		return false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(ref, "standards/"), ".md")
	if id == "" {
		return false
	}
	if strings.Contains(ref, "//") || strings.Contains(ref, "../") || strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "/") {
		return false
	}
	return artifactIDPattern.MatchString(id)
}
