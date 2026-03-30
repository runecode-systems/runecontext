package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocumentationReferenceArtifactIncludesCanonicalDescriptorSurfaces(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		artifact := buildDocumentationReferenceArtifact()
		descriptor := buildCapabilityDescriptor()

		assertDocumentationReferenceVersions(t, artifact, descriptor)
		assertDocumentationReferenceParity(t, artifact, descriptor)
		assertDocumentationReferenceCapabilities(t, artifact, descriptor)
	})
}

func assertDocumentationReferenceVersions(t *testing.T, artifact documentationReferenceArtifact, descriptor capabilityDescriptor) {
	t.Helper()
	if got, want := artifact.ReferenceSchemaVersion, documentationReferenceSchemaVersion; got != want {
		t.Fatalf("expected reference_schema_version %d, got %d", want, got)
	}
	if got, want := artifact.DescriptorSchema, descriptor.SchemaVersion; got != want {
		t.Fatalf("expected descriptor_schema_version %d, got %d", want, got)
	}
}

func assertDocumentationReferenceParity(t *testing.T, artifact documentationReferenceArtifact, descriptor capabilityDescriptor) {
	t.Helper()
	assertDocumentationReferenceFieldParity(t, "compatibility", artifact.Compatibility, descriptor.Compatibility)
	assertDocumentationReferenceFieldParity(t, "distribution layouts", artifact.DistributionLayouts, descriptor.DistributionLayouts)
	assertDocumentationReferenceFieldParity(t, "project profiles", artifact.ProjectProfiles, descriptor.ProjectProfiles)
	assertDocumentationReferenceFieldParity(t, "features", artifact.Features, descriptor.Features)
	assertDocumentationReferenceFieldParity(t, "assurance", artifact.Assurance, descriptor.Assurance)
	assertDocumentationReferenceFieldParity(t, "canonicalization", artifact.Canonicalization, descriptor.Canonicalization)
	assertDocumentationReferenceFieldParity(t, "resolution", artifact.Resolution, descriptor.Resolution)
}

func assertDocumentationReferenceFieldParity(t *testing.T, name string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %s parity with descriptor, got %#v want %#v", name, got, want)
	}
}

func assertDocumentationReferenceCapabilities(t *testing.T, artifact documentationReferenceArtifact, descriptor capabilityDescriptor) {
	t.Helper()
	if len(artifact.Commands.Items) != len(descriptor.Capabilities.Commands) {
		t.Fatalf("expected commands parity with descriptor, got %d commands and %d descriptor commands", len(artifact.Commands.Items), len(descriptor.Capabilities.Commands))
	}
	if len(artifact.Capabilities.CommandTokens) == 0 {
		t.Fatal("expected non-empty capability command_tokens")
	}
	if len(artifact.Capabilities.MachineFlags) == 0 {
		t.Fatal("expected non-empty capability machine_flags")
	}
}

func TestDocumentationReferenceGeneratedArtifactsParity(t *testing.T) {
	withReleaseMetadataVersionForTests(t, func() {
		jsonPayload, err := documentationReferenceArtifacts()
		if err != nil {
			t.Fatalf("build docs reference artifacts: %v", err)
		}
		root, err := repoRootForTests()
		if err != nil {
			t.Fatalf("locate repo root: %v", err)
		}

		jsonPath := filepath.Join(root, docsReferenceJSONRelativePath)

		expectedJSON, err := os.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("read generated docs reference json: %v", err)
		}

		if !jsonEqual(expectedJSON, jsonPayload) {
			t.Fatalf("generated docs reference json is out of date; run `go run ./tools/syncmetadataartifacts --root .`\nexpected=%s\nactual=%s", string(expectedJSON), string(jsonPayload))
		}
	})
}

func jsonEqual(left, right []byte) bool {
	var l any
	if err := json.Unmarshal(left, &l); err != nil {
		return false
	}
	var r any
	if err := json.Unmarshal(right, &r); err != nil {
		return false
	}

	ln, _ := json.Marshal(l)
	rn, _ := json.Marshal(r)
	return string(ln) == string(rn)
}
