package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	releaseMetadataRelativePath                    = "nix/release/metadata.nix"
	metadataJSONGoldenRelativePath                 = "fixtures/cli-json-golden/metadata-success.json"
	releaseManifestFixtureRelativePath             = "fixtures/release/release-manifest-with-metadata.json"
	releaseMetadataVersionFieldPattern             = `(?m)^[ \t]*version[ \t]*=[ \t]*"([^"]+)"`
	releaseManifestFixtureFilePerm     os.FileMode = 0o644
)

var releaseMetadataVersionRegexp = regexp.MustCompile(releaseMetadataVersionFieldPattern)

// ReadReleaseMetadataVersion returns the release version declared in nix/release/metadata.nix.
func ReadReleaseMetadataVersion(repoRoot string) (string, error) {
	path := filepath.Join(repoRoot, releaseMetadataRelativePath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read release metadata %q: %w", path, err)
	}
	match := releaseMetadataVersionRegexp.FindSubmatch(raw)
	if len(match) != 2 {
		return "", fmt.Errorf("extract version from %q", path)
	}
	version := strings.TrimSpace(string(match[1]))
	if version == "" {
		return "", fmt.Errorf("empty version in %q", path)
	}
	return strings.TrimPrefix(version, "v"), nil
}

// WriteMetadataSyncArtifacts regenerates docs/reference and fixture artifacts from canonical metadata.
func WriteMetadataSyncArtifacts(repoRoot, releaseVersion string) error {
	normalized := normalizeReleaseVersion(releaseVersion)
	if normalized == "" {
		return fmt.Errorf("release version must be non-empty")
	}

	return withRunecontextVersion(normalized, func() error {
		return writeMetadataSyncArtifactsAtVersion(repoRoot)
	})
}

func writeMetadataSyncArtifactsAtVersion(repoRoot string) error {
	descriptor := buildCapabilityDescriptor()
	if err := writeDocumentationReferenceJSON(repoRoot); err != nil {
		return err
	}
	if err := writeMetadataGoldenFixture(repoRoot, descriptor); err != nil {
		return err
	}
	return writeReleaseManifestFixture(repoRoot, descriptor)
}

func writeDocumentationReferenceJSON(repoRoot string) error {
	docsReferenceJSON, err := documentationReferenceArtifacts()
	if err != nil {
		return err
	}
	return writeGeneratedJSONFile(filepath.Join(repoRoot, docsReferenceJSONRelativePath), docsReferenceJSON)
}

func writeMetadataGoldenFixture(repoRoot string, descriptor capabilityDescriptor) error {
	metadataGoldenJSON, err := json.Marshal(descriptorMap(descriptor))
	if err != nil {
		return fmt.Errorf("marshal metadata golden fixture: %w", err)
	}
	return writeGeneratedJSONFile(filepath.Join(repoRoot, metadataJSONGoldenRelativePath), metadataGoldenJSON)
}

func writeReleaseManifestFixture(repoRoot string, descriptor capabilityDescriptor) error {
	releaseManifestJSON, err := buildReleaseManifestFixtureJSON(descriptor)
	if err != nil {
		return err
	}
	return writeGeneratedJSONFile(filepath.Join(repoRoot, releaseManifestFixtureRelativePath), releaseManifestJSON)
}

func normalizeReleaseVersion(version string) string {
	trimmed := strings.TrimSpace(version)
	trimmed = strings.TrimPrefix(trimmed, "v")
	return strings.TrimSpace(trimmed)
}

func withRunecontextVersion(version string, fn func() error) error {
	original := runecontextVersion
	runecontextVersion = version
	defer func() {
		runecontextVersion = original
	}()
	return fn()
}

func buildReleaseManifestFixtureJSON(descriptor capabilityDescriptor) ([]byte, error) {
	manifest := map[string]any{
		"package_name":        descriptor.Release.PackageName,
		"version":             descriptor.Release.Version,
		"tag":                 descriptor.Release.Tag,
		"metadata_descriptor": descriptorMap(descriptor),
		"archives":            []any{},
	}
	formatted, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal release manifest fixture: %w", err)
	}
	return append(formatted, '\n'), nil
}

func writeGeneratedJSONFile(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create generated artifact directory %q: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, payload, releaseManifestFixtureFilePerm); err != nil {
		return fmt.Errorf("write generated artifact %q: %w", path, err)
	}
	return nil
}
