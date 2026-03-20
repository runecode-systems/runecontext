package contracts

import (
	"fmt"
	"strings"
)

func (v *Validator) ValidateProposalMarkdown(path string, data []byte) error {
	_, err := parseProposalMarkdown(path, data)
	return err
}

func parseProposalMarkdown(path string, data []byte) (*MarkdownDocument, error) {
	sections, err := parseLevel2Sections(path, data)
	if err != nil {
		return nil, err
	}
	expected := proposalSectionRequirements()
	if len(sections) < len(expected) {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("missing required section %q", expected[len(sections)].name)}
	}
	parsed, requiredNames, err := parseRequiredProposalSections(path, sections, expected)
	if err != nil {
		return nil, err
	}
	if err := parseExtraProposalSections(path, sections[len(expected):], parsed, requiredNames); err != nil {
		return nil, err
	}
	return &MarkdownDocument{Sections: parsed}, nil
}

type proposalSectionRequirement struct {
	name    string
	allowNA bool
}

func proposalSectionRequirements() []proposalSectionRequirement {
	return []proposalSectionRequirement{{"Summary", true}, {"Problem", true}, {"Proposed Change", false}, {"Why Now", true}, {"Assumptions", true}, {"Out of Scope", true}, {"Impact", true}}
}

func parseRequiredProposalSections(path string, sections []markdownSection, expected []proposalSectionRequirement) (map[string]string, map[string]struct{}, error) {
	parsed := map[string]string{}
	requiredNames := map[string]struct{}{}
	for _, section := range expected {
		requiredNames[section.name] = struct{}{}
	}
	for i, requirement := range expected {
		if err := validateProposalSection(path, sections[i], requirement); err != nil {
			return nil, nil, err
		}
		parsed[sections[i].Heading] = sections[i].Body
	}
	return parsed, requiredNames, nil
}

func validateProposalSection(path string, actual markdownSection, requirement proposalSectionRequirement) error {
	if actual.Heading != requirement.name {
		return &ValidationError{Path: path, Message: fmt.Sprintf("section %q appears where %q is required", actual.Heading, requirement.name)}
	}
	if actual.Body == "" {
		return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must contain content or explicit N/A", actual.Heading)}
	}
	if actual.Body == "N/A" && !requirement.allowNA {
		return &ValidationError{Path: path, Message: fmt.Sprintf("section %q may not be N/A", actual.Heading)}
	}
	return nil
}

func parseExtraProposalSections(path string, extras []markdownSection, parsed map[string]string, requiredNames map[string]struct{}) error {
	for _, extra := range extras {
		if _, ok := requiredNames[extra.Heading]; ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("duplicate required section %q", extra.Heading)}
		}
		if extra.Body == "" {
			return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must not be empty", extra.Heading)}
		}
		parsed[extra.Heading] = extra.Body
	}
	return nil
}

func markdownBodyWithoutFrontmatter(data []byte) (string, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.HasPrefix(text, "---\n") {
		doc, err := parseFrontmatterMarkdown("", data)
		if err != nil {
			return "", err
		}
		return doc.Body, nil
	}
	return text, nil
}

func parseLevel2Sections(path string, data []byte) ([]markdownSection, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	sections := make([]markdownSection, 0)
	state := markdownSectionParserState{}
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			state.flush(&sections)
			state.heading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		if state.heading == "" {
			if strings.TrimSpace(line) == "" {
				continue
			}
			return nil, &ValidationError{Path: path, Message: "unexpected content before first level-2 heading"}
		}
		state.body = append(state.body, line)
	}
	state.flush(&sections)
	if len(sections) == 0 {
		return nil, &ValidationError{Path: path, Message: "missing required level-2 sections"}
	}
	return sections, nil
}

type markdownSectionParserState struct {
	heading string
	body    []string
}

func (s *markdownSectionParserState) flush(sections *[]markdownSection) {
	if s.heading == "" {
		return
	}
	*sections = append(*sections, markdownSection{Heading: s.heading, Body: strings.TrimSpace(strings.Join(s.body, "\n"))})
	s.body = nil
}
