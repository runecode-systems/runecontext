package contracts

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildGeneratedBundlesIndexMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	bundleIndex, err := index.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build generated bundle index: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "bundles-index.schema.json", "generated-bundles-index.yaml", bundleIndex)
	assertGeneratedArtifactMatchesGolden(t, bundleIndex, fixturePath(t, "generated-indexes", "golden", "bundle-resolution-bundles.yaml"))
}

func TestBuildGeneratedBundlesDeterministic(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate bundle fixture project: %v", err)
	}
	defer index.Close()

	first, err := index.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build first bundles index: %v", err)
	}
	second, err := index.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build second bundles index: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic bundles output\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestBundlesIndexSchemaRejectsInvalidBundlePaths(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	invalidPaths := []string{
		"bundles/../x.yaml",
		"bundles/.hidden.yaml",
		"bundles//x.yaml",
		"bundles/./x.yaml",
	}
	for _, path := range invalidPaths {
		index := newBundleSchemaFixture(path)
		if err := v.ValidateValue("bundles-index.schema.json", "bundles-index.yaml", index); err == nil {
			t.Fatalf("expected bundles index schema to reject invalid path %q", path)
		}
	}
}

func TestBundlesIndexSchemaAcceptsValidBundlePaths(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	validPaths := []string{
		"bundles/base.yaml",
		"bundles/auth-review.yaml",
		"bundles/a/b.yaml",
	}
	for _, path := range validPaths {
		index := newBundleSchemaFixture(path)
		if err := v.ValidateValue("bundles-index.schema.json", "bundles-index.yaml", index); err != nil {
			t.Fatalf("expected bundles index schema to accept valid path %q: %v", path, err)
		}
	}
}

func TestBundlesIndexSchemaRejectsUnknownFields(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	bundlesMap := buildBundlesIndexSchemaMap(t, v)
	rejectBundlesUnknownTopLevelFields(t, v, bundlesMap)
	rejectBundlesUnknownEntryFields(t, v, bundlesMap)
	rejectBundlesUnknownNestedPatternFields(t, v, bundlesMap)
}

func buildBundlesIndexSchemaMap(t *testing.T, v *Validator) map[string]any {
	t.Helper()
	bundleIndexProject, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate bundle fixture project: %v", err)
	}
	defer bundleIndexProject.Close()

	bundlesIndex, err := bundleIndexProject.BuildGeneratedBundlesIndex()
	if err != nil {
		t.Fatalf("build bundles index: %v", err)
	}
	bundlesData, err := yaml.Marshal(bundlesIndex)
	if err != nil {
		t.Fatalf("marshal bundles index: %v", err)
	}
	bundlesValue, err := parseYAML(bundlesData)
	if err != nil {
		t.Fatalf("parse bundles index yaml: %v", err)
	}
	return bundlesValue.(map[string]any)
}

func rejectBundlesUnknownTopLevelFields(t *testing.T, v *Validator, bundlesMap map[string]any) {
	t.Helper()
	bundlesMap["unexpected"] = true
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown top-level fields")
	}
	delete(bundlesMap, "unexpected")
}

func rejectBundlesUnknownEntryFields(t *testing.T, v *Validator, bundlesMap map[string]any) {
	t.Helper()
	bundles := bundlesMap["bundles"].([]any)
	if len(bundles) == 0 {
		t.Fatal("expected bundle entries")
	}
	bundleEntry := bundles[0].(map[string]any)
	bundleEntry["unexpected"] = "value"
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown bundle entry fields")
	}
	delete(bundleEntry, "unexpected")
}

func rejectBundlesUnknownNestedPatternFields(t *testing.T, v *Validator, bundlesMap map[string]any) {
	t.Helper()
	bundles := bundlesMap["bundles"].([]any)
	bundleEntry := bundles[0].(map[string]any)
	referencedPatterns := bundleEntry["referenced_patterns"].(map[string]any)
	projectPatterns := referencedPatterns["project"].(map[string]any)
	includes := projectPatterns["includes"].([]any)
	if len(includes) == 0 {
		t.Fatal("expected project includes entries")
	}
	firstInclude := includes[0].(map[string]any)
	firstInclude["unexpected"] = "value"
	if err := v.ValidateValue("bundles-index.schema.json", "generated-bundles-index.yaml", bundlesMap); err == nil {
		t.Fatal("expected bundles schema to reject unknown nested pattern fields")
	}
}

func newBundleSchemaFixture(path string) map[string]any {
	return map[string]any{
		"schema_version": 1,
		"bundles": []any{
			map[string]any{
				"id":               "test",
				"path":             path,
				"extends":          []any{},
				"resolved_parents": []any{},
				"referenced_patterns": map[string]any{
					"project":   map[string]any{"includes": []any{}, "excludes": []any{}},
					"standards": map[string]any{"includes": []any{}, "excludes": []any{}},
					"specs":     map[string]any{"includes": []any{}, "excludes": []any{}},
					"decisions": map[string]any{"includes": []any{}, "excludes": []any{}},
				},
			},
		},
	}
}
