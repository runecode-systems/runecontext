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
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.10"

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

	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	validator := contracts.NewValidator(filepath.Join(root, "schemas"))
	if err := validator.ValidateValue(metadataSchemaName, "metadata-output.json", payload); err != nil {
		t.Fatalf("metadata output should validate against schema: %v", err)
	}
}

func TestRunMetadataDescriptorRuntimeProfilesAndResolutionTokens(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.10"

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
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.10"

	descriptor := buildCapabilityDescriptor()
	if len(descriptor.Compatibility.SupportedProjectVersions) == 0 {
		t.Fatal("expected supported project versions to be populated")
	}
	if len(descriptor.Compatibility.ExplicitUpgradeEdges) == 0 {
		t.Fatal("expected explicit upgrade edges to be populated")
	}

	hasAlpha5 := false
	hasAlpha8To9 := false
	for _, version := range descriptor.Compatibility.SupportedProjectVersions {
		if version == "0.1.0-alpha.5" {
			hasAlpha5 = true
			break
		}
	}
	for _, edge := range descriptor.Compatibility.ExplicitUpgradeEdges {
		if edge.From == "0.1.0-alpha.8" && edge.To == "0.1.0-alpha.9" {
			hasAlpha8To9 = true
			break
		}
	}
	if !hasAlpha5 {
		t.Fatalf("expected supported project versions to include alpha.5 compatibility range: %#v", descriptor.Compatibility.SupportedProjectVersions)
	}
	if !hasAlpha8To9 {
		t.Fatalf("expected explicit upgrade edges to include alpha.8->alpha.9 edge: %#v", descriptor.Compatibility.ExplicitUpgradeEdges)
	}
}
