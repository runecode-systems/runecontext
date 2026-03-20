package contracts

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCanonicalContextPackHashInputIgnoresGeneratedAtAndPackHash(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	first, err := canonicalContextPackHashInput(pack)
	if err != nil {
		t.Fatalf("canonical hash input: %v", err)
	}
	copyPack := *pack
	copyPack.GeneratedAt = "2026-03-21T12:00:00Z"
	copyPack.PackHash = strings.Repeat("f", 64)
	second, err := canonicalContextPackHashInput(&copyPack)
	if err != nil {
		t.Fatalf("canonical hash input copy: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("expected canonical input to ignore generated_at and pack_hash\nfirst:  %s\nsecond: %s", string(first), string(second))
	}
}

func TestBuildContextPackPrimaryIDMatchesFirstRequestedBundleID(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"left", "right"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	if pack.ID != pack.RequestedBundleIDs[0] {
		t.Fatalf("expected pack id %q to equal first requested bundle ID %q", pack.ID, pack.RequestedBundleIDs[0])
	}
}

func TestMarshalCanonicalJSONSortsKeysAndDisablesHTMLEscaping(t *testing.T) {
	encoded, err := marshalCanonicalJSON(map[string]any{"z": "a<b>&c", "a": "snowman ☃", "n": "line1\nline2"})
	if err != nil {
		t.Fatalf("marshal canonical JSON: %v", err)
	}
	want := `{"a":"snowman ☃","n":"line1\nline2","z":"a<b>&c"}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, string(encoded))
	}
	if strings.Contains(string(encoded), `\u003c`) || strings.Contains(string(encoded), `\u003e`) || strings.Contains(string(encoded), `\u0026`) {
		t.Fatalf("expected HTML characters to remain unescaped, got %s", string(encoded))
	}
}

func TestMarshalCanonicalJSONSupportsUnsignedIntegers(t *testing.T) {
	encoded, err := marshalCanonicalJSON(map[string]any{"count": uint32(7)})
	if err != nil {
		t.Fatalf("marshal canonical JSON: %v", err)
	}
	if string(encoded) != `{"count":7}` {
		t.Fatalf("expected unsigned integer support, got %s", string(encoded))
	}
}

func TestMarshalCanonicalJSONRejectsFloatValues(t *testing.T) {
	_, err := marshalCanonicalJSON(map[string]any{"ratio": 1.5})
	if err == nil || !strings.Contains(err.Error(), "unsupported canonical JSON value float64") {
		t.Fatalf("expected float rejection, got %v", err)
	}
}

func TestMarshalCanonicalJSONRejectsInvalidUTF8Strings(t *testing.T) {
	_, err := marshalCanonicalJSON(map[string]any{"bad": string([]byte{0xff})})
	if err == nil || !strings.Contains(err.Error(), "valid UTF-8") {
		t.Fatalf("expected invalid UTF-8 rejection, got %v", err)
	}
}

func TestContextPackSchemaConstantsMatchMachineContracts(t *testing.T) {
	data := readFixture(t, filepath.Join(schemaRoot(t), "context-pack.schema.json"))
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse context-pack schema: %v", err)
	}
	properties := schemaProperties(t, schema)
	assertSchemaConstValue(t, properties, "canonicalization", contextPackCanonicalization)
	assertSchemaConstValue(t, properties, "pack_hash_alg", contextPackHashAlgorithm)
	assertRequiredContextPackFields(t, schema)
	assertRequestedBundleIDsSchema(t, properties)
	assertSchemaPatternValue(t, properties, "id", `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties object in context-pack schema")
	}
	return properties
}

func assertRequiredContextPackFields(t *testing.T, schema map[string]any) {
	t.Helper()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("expected required array in context-pack schema")
	}
	for _, field := range []string{"requested_bundle_ids", "excluded", "generated_at"} {
		if !schemaRequiredField(required, field) {
			t.Fatalf("expected context-pack schema required fields to include %q", field)
		}
	}
}

func assertRequestedBundleIDsSchema(t *testing.T, properties map[string]any) {
	t.Helper()
	requestedBundles, ok := properties["requested_bundle_ids"].(map[string]any)
	if !ok {
		t.Fatal("expected requested_bundle_ids schema property")
	}
	if unique, ok := requestedBundles["uniqueItems"].(bool); !ok || !unique {
		t.Fatalf("expected requested_bundle_ids to require unique items, got %#v", requestedBundles["uniqueItems"])
	}
	items, ok := requestedBundles["items"].(map[string]any)
	if !ok {
		t.Fatal("expected requested_bundle_ids items schema")
	}
	if pattern, ok := items["pattern"].(string); !ok || pattern != `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` {
		t.Fatalf("expected requested_bundle_ids item pattern to match bundle ID grammar, got %#v", items["pattern"])
	}
}

func assertSchemaConstValue(t *testing.T, properties map[string]any, field, want string) {
	t.Helper()
	property, ok := properties[field].(map[string]any)
	if !ok {
		t.Fatalf("expected schema property %q", field)
	}
	got, ok := property["const"].(string)
	if !ok {
		t.Fatalf("expected schema property %q to define string const", field)
	}
	if got != want {
		t.Fatalf("expected schema property %q const %q, got %q", field, want, got)
	}
}

func assertSchemaPatternValue(t *testing.T, properties map[string]any, field, want string) {
	t.Helper()
	property, ok := properties[field].(map[string]any)
	if !ok {
		t.Fatalf("expected schema property %q", field)
	}
	got, ok := property["pattern"].(string)
	if !ok {
		t.Fatalf("expected schema property %q to define string pattern", field)
	}
	if got != want {
		t.Fatalf("expected schema property %q pattern %q, got %q", field, want, got)
	}
}

func schemaRequiredField(required []any, want string) bool {
	for _, field := range required {
		if field == want {
			return true
		}
	}
	return false
}
