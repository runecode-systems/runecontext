package contracts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

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
	_, err := v.ValidateExtensionUsage(rootConfigPath, rootData, artifactPath, artifactData)
	return err
}

func (v *Validator) ValidateExtensionUsage(rootConfigPath string, rootData []byte, artifactPath string, artifactData []byte) (bool, error) {
	rootMap, err := parseRequiredYAMLMap(rootConfigPath, rootData, "root config")
	if err != nil {
		return false, err
	}
	artifactMap, err := parseRequiredYAMLMap(artifactPath, artifactData, "artifact")
	if err != nil {
		return false, err
	}
	if _, hasExtensions := artifactMap["extensions"]; hasExtensions {
		allow, _ := rootMap["allow_extensions"].(bool)
		if !allow {
			return false, &ValidationError{Path: artifactPath, Message: "extensions require `allow_extensions: true` in runecontext.yaml"}
		}
		return true, nil
	}
	return false, nil
}

func parseRequiredYAMLMap(path string, data []byte, context string) (map[string]any, error) {
	parsed, err := parseYAML(data)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	rootMap, ok := parsed.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: context + " must decode to a mapping"}
	}
	return rootMap, nil
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
	return rejectRestrictedYAMLNode(&node)
}

func rejectRestrictedYAMLNode(node *yaml.Node) error {
	if err := validateRestrictedYAMLNode(node); err != nil {
		return err
	}
	for _, child := range node.Content {
		if err := rejectRestrictedYAMLNode(child); err != nil {
			return err
		}
	}
	return nil
}

func validateRestrictedYAMLNode(node *yaml.Node) error {
	if node.Anchor != "" || node.Kind == yaml.AliasNode {
		return fmt.Errorf("YAML anchors and aliases are not allowed")
	}
	if node.Style&yaml.TaggedStyle != 0 {
		return fmt.Errorf("YAML tags are not allowed")
	}
	if isNonEmptyFlowCollection(node) {
		return fmt.Errorf("YAML flow-style collections are not allowed")
	}
	if node.Style&yaml.LiteralStyle != 0 || node.Style&yaml.FoldedStyle != 0 {
		return fmt.Errorf("YAML multiline strings are not allowed")
	}
	return nil
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
