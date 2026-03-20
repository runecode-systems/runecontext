package contracts

import "strings"

func (v *Validator) ParseSpec(path string, data []byte) (*FrontmatterDocument, error) {
	return v.parseTypedFrontmatter(path, data, "spec.schema.json", "specs")
}

func (v *Validator) ParseDecision(path string, data []byte) (*FrontmatterDocument, error) {
	return v.parseTypedFrontmatter(path, data, "decision.schema.json", "decisions")
}

func (v *Validator) ParseStandard(path string, data []byte) (*FrontmatterDocument, error) {
	return v.parseTypedFrontmatter(path, data, "standard.schema.json", "standards")
}

func (v *Validator) parseTypedFrontmatter(path string, data []byte, schemaName, root string) (*FrontmatterDocument, error) {
	doc, err := parseFrontmatterMarkdown(path, data)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateValue(schemaName, path, doc.Frontmatter); err != nil {
		return nil, err
	}
	if err := validatePathMatchedID(path, root, doc.Frontmatter["id"]); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseFrontmatterMarkdown(path string, data []byte) (*FrontmatterDocument, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, &ValidationError{Path: path, Message: "missing YAML frontmatter opening delimiter"}
	}
	frontmatterText, body, ok := splitFrontmatter(strings.TrimPrefix(text, "---\n"))
	if !ok {
		return nil, &ValidationError{Path: path, Message: "missing YAML frontmatter closing delimiter"}
	}
	frontmatterMap, err := parseFrontmatterMap(path, frontmatterText)
	if err != nil {
		return nil, err
	}
	return &FrontmatterDocument{Frontmatter: frontmatterMap, Body: body}, nil
}

func parseFrontmatterMap(path, frontmatterText string) (map[string]any, error) {
	frontmatterBytes := []byte(frontmatterText + "\n")
	if err := rejectRestrictedYAMLFeatures(frontmatterBytes); err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	frontmatter, err := parseYAML(frontmatterBytes)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	frontmatterMap, ok := frontmatter.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: "frontmatter must decode to a mapping"}
	}
	return frontmatterMap, nil
}

func splitFrontmatter(remaining string) (string, string, bool) {
	lines := strings.Split(remaining, "\n")
	for i, line := range lines {
		if line == "---" {
			return strings.Join(lines[:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", "", false
}
