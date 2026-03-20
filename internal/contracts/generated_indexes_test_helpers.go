package contracts

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func assertGeneratedArtifactValidAgainstSchema(t *testing.T, v *Validator, schema, path string, doc any) {
	t.Helper()
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal generated artifact: %v", err)
	}
	if err := v.ValidateYAMLFile(schema, path, data); err != nil {
		t.Fatalf("expected generated artifact to satisfy schema %s: %v\n%s", schema, err, string(data))
	}
}

func assertGeneratedArtifactMatchesGolden(t *testing.T, doc any, goldenPath string) {
	t.Helper()
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal generated artifact: %v", err)
	}
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("missing golden fixture %s\n%s", goldenPath, string(data))
		}
		t.Fatalf("read golden fixture %s: %v", goldenPath, err)
	}
	expected := normalizeGeneratedValue(t, mustParseGeneratedYAML(t, string(goldenData)))
	actual := normalizeGeneratedValue(t, mustParseGeneratedYAML(t, string(data)))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("generated artifact mismatch\nexpected: %#v\nactual:   %#v\nactual_yaml:\n%s", expected, actual, string(data))
	}
}

func normalizeGeneratedValue(t *testing.T, value any) any {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeGeneratedValue(t, item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeGeneratedValue(t, item)
		}
		return result
	case []string:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = normalizeGeneratedValue(t, item)
		}
		return result
	case string:
		return filepath.ToSlash(typed)
	default:
		return typed
	}
}

func mustParseGeneratedYAML(t *testing.T, text string) any {
	t.Helper()
	value, err := parseYAML([]byte(text))
	if err != nil {
		t.Fatalf("parse YAML: %v", err)
	}
	return value
}
