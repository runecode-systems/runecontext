package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunMetadataOutputsDescriptorJSON(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		payload := runMetadataPayload(t)
		assertMetadataSchemaAndShape(t, payload)
		assertMetadataOutputValidAgainstSchema(t, payload)
	})
}

func TestRunMetadataDescriptorRuntimeProfilesAndResolutionTokens(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		descriptor := buildCapabilityDescriptor()
		profiles := map[string]bool{}
		for _, layout := range descriptor.Runtime.Layouts {
			profiles[layout.Profile] = true
		}
		if !profiles["repo_bundle"] || !profiles["installed_share_layout"] {
			t.Fatalf("expected runtime layouts to include repo_bundle and installed_share_layout, got %#v", descriptor.Runtime.Layouts)
		}

		if len(descriptor.Resolution.SourceModes) != 3 {
			t.Fatalf("expected three source modes, got %#v", descriptor.Resolution.SourceModes)
		}
		if len(descriptor.Resolution.VerificationPosture) != 5 {
			t.Fatalf("expected five verification postures, got %#v", descriptor.Resolution.VerificationPosture)
		}
	})
}

func TestRunMetadataUsageAndRejections(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := Run([]string{"metadata", "--help"}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected help success, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage="+metadataUsage) {
		t.Fatalf("expected metadata usage output, got %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"metadata", "--json"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "usage="+metadataUsage) {
		t.Fatalf("expected metadata usage output for flag rejection, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"metadata", "extra"}, &stdout, &stderr); code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "metadata does not accept positional arguments") {
		t.Fatalf("expected positional rejection, got %q", stderr.String())
	}
}

func TestReleaseManifestDescriptorParityRoundTrip(t *testing.T) {
	descriptor := buildCapabilityDescriptor()
	rawDescriptor, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatalf("marshal descriptor: %v", err)
	}
	rawManifest := []byte(`{"metadata_descriptor":` + string(rawDescriptor) + `,"archives":[]}`)
	manifestDescriptor, err := releaseManifestDescriptorFromJSON(rawManifest)
	if err != nil {
		t.Fatalf("parse release manifest descriptor: %v", err)
	}

	want, err := json.Marshal(descriptorMap(descriptor))
	if err != nil {
		t.Fatalf("marshal expected descriptor map: %v", err)
	}
	got, err := json.Marshal(manifestDescriptor)
	if err != nil {
		t.Fatalf("marshal parsed descriptor map: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected release manifest descriptor parity\nwant=%s\ngot=%s", string(want), string(got))
	}
}

func TestCapabilityDescriptorCompatibilitySplitsSupportedVersionsAndUpgradeEdges(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		descriptor := buildCapabilityDescriptor()
		assertCompatibilityPopulation(t, descriptor)
		assertCompatibilityIncludesExpectedVersions(t, descriptor)
	})
}

func runMetadataPayload(t *testing.T) map[string]any {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"metadata"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal metadata json: %v", err)
	}
	return payload
}

func assertMetadataSchemaAndShape(t *testing.T, payload map[string]any) {
	t.Helper()
	if got, want := int(payload["schema_version"].(float64)), 1; got != want {
		t.Fatalf("expected schema_version %d, got %d", want, got)
	}
	if got, want := payload["descriptor_schema_version"], "1"; got != want {
		t.Fatalf("expected descriptor_schema_version %q, got %#v", want, got)
	}
	compatibility, ok := payload["compatibility"].(map[string]any)
	if !ok {
		t.Fatalf("expected compatibility object, got %#v", payload["compatibility"])
	}
	if _, ok := compatibility["supported_project_versions"]; !ok {
		t.Fatalf("expected supported_project_versions in compatibility: %#v", compatibility)
	}
	if _, ok := compatibility["explicit_upgrade_edges"]; !ok {
		t.Fatalf("expected explicit_upgrade_edges in compatibility: %#v", compatibility)
	}
	runtime, ok := payload["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("expected runtime object, got %#v", payload["runtime"])
	}
	layouts, ok := runtime["layouts"].([]any)
	if !ok || len(layouts) < 2 {
		t.Fatalf("expected runtime.layouts with repo and installed layouts, got %#v", runtime["layouts"])
	}
}

func assertMetadataOutputValidAgainstSchema(t *testing.T, payload map[string]any) {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	validator := contracts.NewValidator(filepath.Join(root, "schemas"))
	if err := validator.ValidateValue(metadataSchemaName, "metadata-output.json", payload); err != nil {
		t.Fatalf("metadata output should validate against schema: %v", err)
	}
}

func assertCompatibilityPopulation(t *testing.T, descriptor capabilityDescriptor) {
	t.Helper()
	if len(descriptor.Compatibility.SupportedProjectVersions) == 0 {
		t.Fatal("expected supported project versions to be populated")
	}
	if len(descriptor.Compatibility.ExplicitUpgradeEdges) == 0 {
		t.Fatal("expected explicit upgrade edges to be populated")
	}
}

func assertCompatibilityIncludesExpectedVersions(t *testing.T, descriptor capabilityDescriptor) {
	t.Helper()
	if !containsString(descriptor.Compatibility.SupportedProjectVersions, "0.1.0-alpha.5") {
		t.Fatalf("expected supported project versions to include alpha.5 compatibility range: %#v", descriptor.Compatibility.SupportedProjectVersions)
	}
	if !containsUpgradeEdge(descriptor.Compatibility.ExplicitUpgradeEdges, "0.1.0-alpha.8", "0.1.0-alpha.9") {
		t.Fatalf("expected explicit upgrade edges to include alpha.8->alpha.9 edge: %#v", descriptor.Compatibility.ExplicitUpgradeEdges)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsUpgradeEdge(edges []descriptorUpgradeEdge, from, to string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to {
			return true
		}
	}
	return false
}
