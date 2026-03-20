package contracts

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBuildGeneratedManifestMatchesGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	manifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build generated manifest: %v", err)
	}
	assertGeneratedArtifactValidAgainstSchema(t, v, "manifest.schema.json", "generated-manifest.yaml", manifest)
	assertGeneratedArtifactMatchesGolden(t, manifest, fixturePath(t, "generated-indexes", "golden", "traceability-manifest.yaml"))
}

func TestBuildGeneratedManifestDeterministic(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()

	first, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build first manifest: %v", err)
	}
	second, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build second manifest: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic manifest output\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestManifestSchemaRejectsInvalidArtifactPaths(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	invalidPaths := []struct {
		field string
		value string
	}{
		{"standards", "standards/../x.md"},
		{"standards", "standards/.hidden.md"},
		{"standards", "standards//x.md"},
		{"specs", "specs/../x.md"},
		{"specs", "specs/.hidden.md"},
		{"specs", "specs//x.md"},
		{"decisions", "decisions/../x.md"},
		{"decisions", "decisions/.hidden.md"},
		{"decisions", "decisions//x.md"},
	}
	for _, tc := range invalidPaths {
		manifest := newManifestSchemaFixture()
		manifest[tc.field] = []any{tc.value}
		if err := v.ValidateValue("manifest.schema.json", "manifest.yaml", manifest); err == nil {
			t.Fatalf("expected manifest schema to reject invalid %s path %q", tc.field, tc.value)
		}
	}
}

func TestManifestSchemaAcceptsValidArtifactPaths(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	validPaths := []struct {
		field string
		value string
	}{
		{"standards", "standards/global/base.md"},
		{"standards", "standards/security/heavy-audit.md"},
		{"standards", "standards/a/b/c.md"},
		{"specs", "specs/auth-gateway.md"},
		{"specs", "specs/a/b.md"},
		{"decisions", "decisions/DEC-0001-trust-boundary-model.md"},
		{"decisions", "decisions/a/b.md"},
	}
	for _, tc := range validPaths {
		manifest := newManifestSchemaFixture()
		manifest[tc.field] = []any{tc.value}
		if err := v.ValidateValue("manifest.schema.json", "manifest.yaml", manifest); err != nil {
			t.Fatalf("expected manifest schema to accept valid %s path %q: %v", tc.field, tc.value, err)
		}
	}
}

func TestManifestSchemaRejectsUnknownFields(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "traceability", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	manifest, err := index.BuildGeneratedManifest()
	if err != nil {
		t.Fatalf("build manifest: %v", err)
	}
	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	manifestValue, err := parseYAML(manifestData)
	if err != nil {
		t.Fatalf("parse manifest yaml: %v", err)
	}
	manifestMap := manifestValue.(map[string]any)
	manifestMap["unexpected"] = "value"
	if err := v.ValidateValue("manifest.schema.json", "generated-manifest.yaml", manifestMap); err == nil {
		t.Fatal("expected manifest schema to reject unknown fields")
	}
}

func newManifestSchemaFixture() map[string]any {
	return map[string]any{
		"schema_version": 1,
		"indexes": map[string]any{
			"changes_by_status": "indexes/changes-by-status.yaml",
			"bundles":           "indexes/bundles.yaml",
		},
		"counts": map[string]any{
			"standards": 1,
			"bundles":   0,
			"changes":   0,
			"specs":     0,
			"decisions": 0,
		},
		"standards": []any{},
		"bundles":   []any{},
		"changes":   []any{},
		"specs":     []any{},
		"decisions": []any{},
	}
}
