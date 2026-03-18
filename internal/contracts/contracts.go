package contracts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

var (
	changeIDPattern   = regexp.MustCompile(`^CHG-\d{4}-\d{3}-[a-z0-9]{4,6}-[a-z0-9]+(-[a-z0-9]+)*$`)
	artifactIDPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?(?:/[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?)*$`)
	bundleIDPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

type Validator struct {
	schemaRoot string
	cacheMu    sync.RWMutex
	cache      map[string]*jsonschema.Schema
}

type ValidationError struct {
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

type MarkdownDocument struct {
	Sections      map[string]string
	Refs          []string
	RefsBySection map[string][]string
}

type FrontmatterDocument struct {
	Frontmatter map[string]any
	Body        string
}

type ProjectIndex struct {
	RootConfigPath string
	ContentRoot    string
	Resolution     *SourceResolution
	Bundles        *BundleCatalog
	Diagnostics    []ValidationDiagnostic
	ChangeIDs      map[string]struct{}
	Changes        map[string]*ChangeRecord
	MarkdownFiles  map[string]*MarkdownArtifact
	StandardPaths  map[string]struct{}
	Standards      map[string]*StandardRecord
	SpecPaths      map[string]struct{}
	Specs          map[string]*SpecRecord
	DecisionPaths  map[string]struct{}
	Decisions      map[string]*DecisionRecord
	StatusFiles    map[string]StatusFileRecord
}

type StatusFileRecord struct {
	Data map[string]any
	Raw  []byte
}

type markdownSection struct {
	Heading string
	Body    string
}

func NewValidator(schemaRoot string) *Validator {
	return &Validator{
		schemaRoot: schemaRoot,
		cache:      map[string]*jsonschema.Schema{},
	}
}

func (p *ProjectIndex) Close() error {
	if p == nil || p.Resolution == nil {
		return nil
	}
	return p.Resolution.Close()
}

func (p *ProjectIndex) ResolveBundle(id string) (*BundleResolution, error) {
	if p == nil || p.Bundles == nil {
		return nil, fmt.Errorf("bundle catalog is unavailable")
	}
	return p.Bundles.Resolve(id)
}

func (v *Validator) ValidateYAMLFile(schemaName, path string, data []byte) error {
	if err := rejectRestrictedYAMLFeatures(data); err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	parsed, err := parseYAML(data)
	if err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	return v.ValidateValue(schemaName, path, parsed)
}

func (v *Validator) ValidateValue(schemaName, path string, value any) error {
	schema, err := v.loadSchema(schemaName)
	if err != nil {
		return err
	}
	if err := schema.Validate(value); err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	return nil
}

func (v *Validator) ValidateExtensionOptIn(rootConfigPath string, rootData []byte, artifactPath string, artifactData []byte) error {
	rootValue, err := parseYAML(rootData)
	if err != nil {
		return &ValidationError{Path: rootConfigPath, Message: err.Error()}
	}
	rootMap, ok := rootValue.(map[string]any)
	if !ok {
		return &ValidationError{Path: rootConfigPath, Message: "root config must decode to a mapping"}
	}
	artifactValue, err := parseYAML(artifactData)
	if err != nil {
		return &ValidationError{Path: artifactPath, Message: err.Error()}
	}
	artifactMap, ok := artifactValue.(map[string]any)
	if !ok {
		return &ValidationError{Path: artifactPath, Message: "artifact must decode to a mapping"}
	}
	if _, hasExtensions := artifactMap["extensions"]; hasExtensions {
		allow, _ := rootMap["allow_extensions"].(bool)
		if !allow {
			return &ValidationError{Path: artifactPath, Message: "extensions require `allow_extensions: true` in runecontext.yaml"}
		}
	}
	return nil
}

func (v *Validator) ValidateProposalMarkdown(path string, data []byte) error {
	_, err := parseProposalMarkdown(path, data)
	return err
}

func (v *Validator) ValidateStandardsMarkdown(path string, data []byte) error {
	_, err := parseStandardsMarkdown(path, data)
	return err
}

func (v *Validator) ParseSpec(path string, data []byte) (*FrontmatterDocument, error) {
	doc, err := parseFrontmatterMarkdown(path, data)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateValue("spec.schema.json", path, doc.Frontmatter); err != nil {
		return nil, err
	}
	if err := validatePathMatchedID(path, "specs", doc.Frontmatter["id"]); err != nil {
		return nil, err
	}
	return doc, nil
}

func (v *Validator) ParseDecision(path string, data []byte) (*FrontmatterDocument, error) {
	doc, err := parseFrontmatterMarkdown(path, data)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateValue("decision.schema.json", path, doc.Frontmatter); err != nil {
		return nil, err
	}
	if err := validatePathMatchedID(path, "decisions", doc.Frontmatter["id"]); err != nil {
		return nil, err
	}
	return doc, nil
}

func (v *Validator) ParseStandard(path string, data []byte) (*FrontmatterDocument, error) {
	doc, err := parseFrontmatterMarkdown(path, data)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateValue("standard.schema.json", path, doc.Frontmatter); err != nil {
		return nil, err
	}
	if err := validatePathMatchedID(path, "standards", doc.Frontmatter["id"]); err != nil {
		return nil, err
	}
	return doc, nil
}

func (v *Validator) ValidateTraceabilityProject(root string) error {
	index, err := v.ValidateProject(root)
	if err != nil {
		return err
	}
	defer index.Close()
	return nil
}

func (v *Validator) ValidateProject(root string) (*ProjectIndex, error) {
	return v.ValidateProjectWithOptions(root, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
}

func resolveContentRoot(projectRoot string, rootData []byte) (*SourceResolution, error) {
	resolution, err := resolveSourceFromConfig(filepath.Join(projectRoot, "runecontext.yaml"), projectRoot, rootData, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		return nil, err
	}
	return resolution, nil
}

func parseProposalMarkdown(path string, data []byte) (*MarkdownDocument, error) {
	sections, err := parseLevel2Sections(path, data)
	if err != nil {
		return nil, err
	}
	expected := []struct {
		name    string
		allowNA bool
	}{
		{name: "Summary", allowNA: true},
		{name: "Problem", allowNA: true},
		{name: "Proposed Change", allowNA: false},
		{name: "Why Now", allowNA: true},
		{name: "Assumptions", allowNA: true},
		{name: "Out of Scope", allowNA: true},
		{name: "Impact", allowNA: true},
	}
	if len(sections) < len(expected) {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("missing required section %q", expected[len(sections)].name)}
	}
	parsed := map[string]string{}
	requiredNames := map[string]struct{}{}
	for _, section := range expected {
		requiredNames[section.name] = struct{}{}
	}
	for i, section := range expected {
		actual := sections[i]
		if actual.Heading != section.name {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q appears where %q is required", actual.Heading, section.name)}
		}
		if actual.Body == "" {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q must contain content or explicit N/A", actual.Heading)}
		}
		if actual.Body == "N/A" && !section.allowNA {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q may not be N/A", actual.Heading)}
		}
		parsed[actual.Heading] = actual.Body
	}
	for _, extra := range sections[len(expected):] {
		if _, ok := requiredNames[extra.Heading]; ok {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("duplicate required section %q", extra.Heading)}
		}
		if extra.Body == "" {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q must not be empty", extra.Heading)}
		}
		parsed[extra.Heading] = extra.Body
	}
	return &MarkdownDocument{Sections: parsed}, nil
}

func parseStandardsMarkdown(path string, data []byte) (*MarkdownDocument, error) {
	sections, err := parseLevel2Sections(path, data)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 || sections[0].Heading != "Applicable Standards" {
		return nil, &ValidationError{Path: path, Message: "missing required section \"Applicable Standards\""}
	}
	canonicalOrder := map[string]int{
		"Applicable Standards":               0,
		"Standards Added Since Last Refresh": 1,
		"Standards Considered But Excluded":  2,
		"Resolution Notes":                   3,
	}
	seen := map[string]struct{}{}
	parsed := map[string]string{}
	refs := make([]string, 0)
	refsBySection := map[string][]string{}
	lastCanonical := -1
	customStarted := false
	for _, section := range sections {
		if _, dup := seen[section.Heading]; dup {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("duplicate section %q", section.Heading)}
		}
		seen[section.Heading] = struct{}{}
		if section.Body == "" {
			return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q must not be empty", section.Heading)}
		}
		if order, ok := canonicalOrder[section.Heading]; ok {
			if customStarted {
				return nil, &ValidationError{Path: path, Message: fmt.Sprintf("canonical section %q cannot appear after custom sections", section.Heading)}
			}
			if order < lastCanonical {
				return nil, &ValidationError{Path: path, Message: fmt.Sprintf("section %q appears out of order", section.Heading)}
			}
			lastCanonical = order
		} else {
			customStarted = true
		}
		parsed[section.Heading] = section.Body
	}
	for _, heading := range []string{"Applicable Standards", "Standards Added Since Last Refresh", "Standards Considered But Excluded"} {
		body, ok := parsed[heading]
		if !ok {
			continue
		}
		if err := validateStandardsSectionReferences(path, heading, body); err != nil {
			return nil, err
		}
		sectionRefs := extractStandardsSectionReferences(body)
		refs = append(refs, sectionRefs...)
		refsBySection[heading] = append([]string(nil), sectionRefs...)
	}
	return &MarkdownDocument{Sections: parsed, Refs: uniqueSortedStrings(refs), RefsBySection: refsBySection}, nil
}

func validateStandardsSectionReferences(path, heading, body string) error {
	lines := strings.Split(body, "\n")
	inBullet := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			refs := extractStandardsLikeBacktickedRefs(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			if len(refs) != 1 {
				return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list exactly one backticked standard path per bullet", heading)}
			}
			ref := refs[0]
			if !isCanonicalStandardPathRef(ref) {
				return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list standards as backticked RuneContext-root-relative paths", heading)}
			}
			inBullet = true
			continue
		}
		if inBullet && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			continue
		}
		return &ValidationError{Path: path, Message: fmt.Sprintf("section %q must list standards as bullet path references instead of copied body text", heading)}
	}
	return nil
}

func extractStandardsSectionReferences(body string) []string {
	refs := make([]string, 0)
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		lineRefs := extractStandardsLikeBacktickedRefs(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
		if len(lineRefs) != 1 || !isCanonicalStandardPathRef(lineRefs[0]) {
			continue
		}
		refs = append(refs, lineRefs[0])
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
	currentHeading := ""
	currentBody := make([]string, 0)

	flush := func() {
		if currentHeading == "" {
			return
		}
		sections = append(sections, markdownSection{
			Heading: currentHeading,
			Body:    strings.TrimSpace(strings.Join(currentBody, "\n")),
		})
		currentBody = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			flush()
			currentHeading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		if currentHeading == "" {
			if strings.TrimSpace(line) == "" {
				continue
			}
			return nil, &ValidationError{Path: path, Message: "unexpected content before first level-2 heading"}
		}
		currentBody = append(currentBody, line)
	}
	flush()
	if len(sections) == 0 {
		return nil, &ValidationError{Path: path, Message: "missing required level-2 sections"}
	}
	return sections, nil
}

func parseFrontmatterMarkdown(path string, data []byte) (*FrontmatterDocument, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, &ValidationError{Path: path, Message: "missing YAML frontmatter opening delimiter"}
	}
	remaining := strings.TrimPrefix(text, "---\n")
	frontmatterText, body, ok := splitFrontmatter(remaining)
	if !ok {
		return nil, &ValidationError{Path: path, Message: "missing YAML frontmatter closing delimiter"}
	}
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
	return &FrontmatterDocument{Frontmatter: frontmatterMap, Body: body}, nil
}

func splitFrontmatter(remaining string) (string, string, bool) {
	lines := strings.Split(remaining, "\n")
	for i, line := range lines {
		if line != "---" {
			continue
		}
		frontmatter := strings.Join(lines[:i], "\n")
		body := strings.Join(lines[i+1:], "\n")
		return frontmatter, body, true
	}
	return "", "", false
}

func parseYAML(data []byte) (any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return nil, err
	}
	if node.Kind == 0 {
		return nil, fmt.Errorf("empty YAML document")
	}
	if err := ensureNoDuplicateKeys(&node); err != nil {
		return nil, err
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return nil, err
	}
	return normalizeYAMLValue(value), nil
}

func ensureNoDuplicateKeys(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		seen := map[string]struct{}{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			if _, ok := seen[key]; ok {
				return fmt.Errorf("duplicate YAML key %q", key)
			}
			seen[key] = struct{}{}
		}
	}
	for _, child := range node.Content {
		if err := ensureNoDuplicateKeys(child); err != nil {
			return err
		}
	}
	return nil
}

func rejectRestrictedYAMLFeatures(data []byte) error {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return err
	}
	var walk func(*yaml.Node) error
	walk = func(n *yaml.Node) error {
		if n.Anchor != "" || n.Kind == yaml.AliasNode {
			return fmt.Errorf("YAML anchors and aliases are not allowed")
		}
		if n.Style&yaml.TaggedStyle != 0 {
			return fmt.Errorf("YAML tags are not allowed")
		}
		if isNonEmptyFlowCollection(n) {
			return fmt.Errorf("YAML flow-style collections are not allowed")
		}
		if n.Style&yaml.LiteralStyle != 0 || n.Style&yaml.FoldedStyle != 0 {
			return fmt.Errorf("YAML multiline strings are not allowed")
		}
		for _, child := range n.Content {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(&node)
}

func isNonEmptyFlowCollection(node *yaml.Node) bool {
	if node.Style&yaml.FlowStyle == 0 {
		return false
	}
	if node.Kind != yaml.SequenceNode && node.Kind != yaml.MappingNode {
		return false
	}
	return len(node.Content) > 0
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[any]any:
		result := make(map[string]any, len(typed))
		for k, v := range typed {
			result[fmt.Sprint(k)] = normalizeYAMLValue(v)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		for k, v := range typed {
			result[k] = normalizeYAMLValue(v)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeYAMLValue(item)
		}
		return result
	default:
		return typed
	}
}

func (v *Validator) loadSchema(name string) (*jsonschema.Schema, error) {
	v.cacheMu.RLock()
	if schema, ok := v.cache[name]; ok {
		v.cacheMu.RUnlock()
		return schema, nil
	}
	v.cacheMu.RUnlock()
	fullPath := filepath.Join(v.schemaRoot, name)
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	var doc any
	schemaData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, &ValidationError{Path: fullPath, Message: err.Error()}
	}
	if err := yaml.Unmarshal(schemaData, &doc); err != nil {
		return nil, &ValidationError{Path: fullPath, Message: err.Error()}
	}
	if doc == nil {
		return nil, &ValidationError{Path: fullPath, Message: "schema file is empty"}
	}
	if err := compiler.AddResource(fullPath, normalizeYAMLValue(doc)); err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(fullPath)
	if err != nil {
		return nil, err
	}
	v.cacheMu.Lock()
	defer v.cacheMu.Unlock()
	if cached, ok := v.cache[name]; ok {
		return cached, nil
	}
	v.cache[name] = schema
	return schema, nil
}

func validatePathMatchedID(path, root string, rawID any) error {
	id := fmt.Sprint(rawID)
	artifactRoot, err := findNearestArtifactRoot(path, root)
	if err != nil {
		return &ValidationError{Path: path, Message: fmt.Sprintf("path does not live under %s/", root)}
	}
	rel, err := filepath.Rel(artifactRoot, filepath.Clean(path))
	if err != nil {
		return &ValidationError{Path: path, Message: err.Error()}
	}
	rel = strings.TrimSuffix(filepath.ToSlash(rel), ".md")
	if rel != id {
		return &ValidationError{Path: path, Message: fmt.Sprintf("frontmatter id %q must match path-relative stem %q", id, rel)}
	}
	return nil
}

func findNearestArtifactRoot(path, root string) (string, error) {
	current := filepath.Clean(filepath.Dir(path))
	for {
		if filepath.Base(current) == root {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", os.ErrNotExist
		}
		current = next
	}
}

func buildProjectIndex(v *Validator, contentRoot string) (*ProjectIndex, error) {
	index := &ProjectIndex{
		ChangeIDs:     map[string]struct{}{},
		Changes:       map[string]*ChangeRecord{},
		MarkdownFiles: map[string]*MarkdownArtifact{},
		StandardPaths: map[string]struct{}{},
		Standards:     map[string]*StandardRecord{},
		SpecPaths:     map[string]struct{}{},
		Specs:         map[string]*SpecRecord{},
		DecisionPaths: map[string]struct{}{},
		Decisions:     map[string]*DecisionRecord{},
		StatusFiles:   map[string]StatusFileRecord{},
	}
	if err := walkChangeDirectories(filepath.Join(contentRoot, "changes"), func(changeDir string) error {
		statusPath := filepath.Join(changeDir, "status.yaml")
		statusData, err := readProjectFile(changeDir, statusPath)
		if err != nil {
			if os.IsNotExist(err) {
				return &ValidationError{Path: statusPath, Message: "missing required file"}
			}
			return err
		}
		if err := v.ValidateYAMLFile("change-status.schema.json", statusPath, statusData); err != nil {
			return err
		}
		parsed, err := parseYAML(statusData)
		if err != nil {
			return err
		}
		obj, err := expectObject(statusPath, parsed, "status file")
		if err != nil {
			return err
		}
		record, err := buildChangeRecord(changeDir, statusPath, obj)
		if err != nil {
			return err
		}
		index.ChangeIDs[record.ID] = struct{}{}
		index.Changes[record.ID] = record
		index.StatusFiles[statusPath] = StatusFileRecord{Data: obj, Raw: append([]byte(nil), statusData...)}
		proposalPath := filepath.Join(changeDir, "proposal.md")
		proposalData, err := readProjectFile(changeDir, proposalPath)
		if err != nil {
			if os.IsNotExist(err) {
				return &ValidationError{Path: proposalPath, Message: "missing required file"}
			}
			return err
		}
		if err := v.ValidateProposalMarkdown(proposalPath, proposalData); err != nil {
			return err
		}
		if err := indexMarkdownArtifact(index, contentRoot, proposalPath, proposalData, false); err != nil {
			return err
		}
		standardsPath := filepath.Join(changeDir, "standards.md")
		standardsData, err := readProjectFile(changeDir, standardsPath)
		if err != nil {
			if os.IsNotExist(err) {
				return &ValidationError{Path: standardsPath, Message: "missing required file"}
			}
			return err
		}
		if err := v.ValidateStandardsMarkdown(standardsPath, standardsData); err != nil {
			return err
		}
		standardsDoc, err := parseStandardsMarkdown(standardsPath, standardsData)
		if err != nil {
			return err
		}
		record.StandardRefs = append([]string(nil), standardsDoc.Refs...)
		record.ApplicableStandards = append([]string(nil), standardsDoc.RefsBySection["Applicable Standards"]...)
		record.AddedStandards = append([]string(nil), standardsDoc.RefsBySection["Standards Added Since Last Refresh"]...)
		record.ExcludedStandards = append([]string(nil), standardsDoc.RefsBySection["Standards Considered But Excluded"]...)
		if err := indexMarkdownArtifact(index, contentRoot, standardsPath, standardsData, false); err != nil {
			return err
		}
		entries, err := os.ReadDir(changeDir)
		if err != nil {
			return &ValidationError{Path: changeDir, Message: err.Error()}
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
				continue
			}
			if entry.Name() == "proposal.md" || entry.Name() == "standards.md" {
				continue
			}
			path := filepath.Join(changeDir, entry.Name())
			data, err := readProjectFile(changeDir, path)
			if err != nil {
				return err
			}
			if err := indexMarkdownArtifact(index, contentRoot, path, data, false); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := walkProjectFiles(filepath.Join(contentRoot, "specs"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "specs"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseSpec(path, data)
		if err != nil {
			return err
		}
		record, err := buildSpecRecord(path, doc)
		if err != nil {
			return err
		}
		for _, key := range []string{"originating_changes", "revised_by_changes"} {
			if err := validateChangeIDRefs(path, key, doc.Frontmatter[key], index.ChangeIDs); err != nil {
				return err
			}
		}
		rel, err := filepath.Rel(contentRoot, path)
		if err != nil {
			return err
		}
		record.Path = filepath.ToSlash(rel)
		index.SpecPaths[record.Path] = struct{}{}
		index.Specs[record.Path] = record
		if err := indexMarkdownArtifact(index, contentRoot, path, data, true); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := walkProjectFiles(filepath.Join(contentRoot, "decisions"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "decisions"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseDecision(path, data)
		if err != nil {
			return err
		}
		record, err := buildDecisionRecord(path, doc)
		if err != nil {
			return err
		}
		for _, key := range []string{"originating_changes", "related_changes"} {
			if err := validateChangeIDRefs(path, key, doc.Frontmatter[key], index.ChangeIDs); err != nil {
				return err
			}
		}
		rel, err := filepath.Rel(contentRoot, path)
		if err != nil {
			return err
		}
		record.Path = filepath.ToSlash(rel)
		index.DecisionPaths[record.Path] = struct{}{}
		index.Decisions[record.Path] = record
		if err := indexMarkdownArtifact(index, contentRoot, path, data, true); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := walkProjectFiles(filepath.Join(contentRoot, "standards"), func(path string) error {
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := readProjectFile(filepath.Join(contentRoot, "standards"), path)
		if err != nil {
			return err
		}
		doc, err := v.ParseStandard(path, data)
		if err != nil {
			return err
		}
		record, err := buildStandardRecord(path, doc)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(contentRoot, path)
		if err != nil {
			return err
		}
		record.Path = filepath.ToSlash(rel)
		index.StandardPaths[record.Path] = struct{}{}
		index.Standards[record.Path] = record
		return indexMarkdownArtifact(index, contentRoot, path, data, true)
	}); err != nil {
		return nil, err
	}
	return index, nil
}

func walkChangeDirectories(root string, visit func(changeDir string) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := visit(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func walkProjectFiles(root string, visit func(path string) error) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return &ValidationError{Path: root, Message: "expected a directory root"}
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return &ValidationError{Path: root, Message: err.Error()}
	}
	return walkContainedFiles(resolvedRoot, root, map[string]struct{}{}, visit)
}

func walkContainedFiles(boundaryResolved, currentPath string, active map[string]struct{}, visit func(path string) error) error {
	resolvedPath, err := filepath.EvalSymlinks(currentPath)
	if err != nil {
		return &ValidationError{Path: currentPath, Message: err.Error()}
	}
	if !isWithinRoot(boundaryResolved, resolvedPath) {
		return &ValidationError{Path: currentPath, Message: fmt.Sprintf("resolved path %q escapes the selected project subtree", resolvedPath)}
	}
	info, err := os.Stat(currentPath)
	if err != nil {
		return &ValidationError{Path: currentPath, Message: err.Error()}
	}
	if info.IsDir() {
		resolvedKey := filepath.Clean(resolvedPath)
		if _, ok := active[resolvedKey]; ok {
			return &ValidationError{Path: currentPath, Message: fmt.Sprintf("symlink cycle detected at %q", currentPath)}
		}
		active[resolvedKey] = struct{}{}
		defer delete(active, resolvedKey)
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			return &ValidationError{Path: currentPath, Message: err.Error()}
		}
		for _, entry := range entries {
			if err := walkContainedFiles(boundaryResolved, filepath.Join(currentPath, entry.Name()), active, visit); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return &ValidationError{Path: currentPath, Message: fmt.Sprintf("resolved path %q is not a regular file", resolvedPath)}
	}
	return visit(currentPath)
}

func readProjectFile(boundaryPath, path string) ([]byte, error) {
	resolvedBoundary, err := filepath.EvalSymlinks(boundaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, &ValidationError{Path: boundaryPath, Message: err.Error()}
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	if !isWithinRoot(resolvedBoundary, resolvedPath) {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("resolved path %q escapes the selected project subtree", resolvedPath)}
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	if info.IsDir() {
		return nil, &ValidationError{Path: path, Message: "expected a file, found a directory"}
	}
	if !info.Mode().IsRegular() {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("resolved path %q is not a regular file", resolvedPath)}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	return data, nil
}

func expectObject(path string, value any, context string) (map[string]any, error) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("%s must decode to an object", context)}
	}
	return obj, nil
}

func validateChangeIDRefs(path, field string, raw any, known map[string]struct{}) error {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	for _, item := range items {
		id := fmt.Sprint(item)
		if !changeIDPattern.MatchString(id) {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s contains invalid change ID %q", field, id)}
		}
		if _, ok := known[id]; !ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s references missing change %q", field, id)}
		}
	}
	return nil
}

func validatePathRefs(path, field string, raw any, known map[string]struct{}) error {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	for _, item := range items {
		ref := filepath.ToSlash(fmt.Sprint(item))
		if _, ok := known[ref]; !ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s references missing artifact %q", field, ref)}
		}
	}
	return nil
}

func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
