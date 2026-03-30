package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const (
	documentationReferenceSchemaVersion = 1
	docsReferenceJSONRelativePath       = "docs/reference/generated/runecontext-reference.json"
)

type documentationReferenceArtifact struct {
	ReferenceSchemaVersion int                              `json:"reference_schema_version" yaml:"reference_schema_version"`
	DescriptorSchema       int                              `json:"descriptor_schema_version" yaml:"descriptor_schema_version"`
	Binary                 string                           `json:"binary" yaml:"binary"`
	Release                descriptorRelease                `json:"release" yaml:"release"`
	Commands               documentationCommandReference    `json:"commands" yaml:"commands"`
	Capabilities           documentationCapabilityReference `json:"capabilities" yaml:"capabilities"`
	Compatibility          descriptorCompatibility          `json:"compatibility" yaml:"compatibility"`
	DistributionLayouts    []descriptorLayout               `json:"distribution_layouts" yaml:"distribution_layouts"`
	ProjectProfiles        []descriptorProject              `json:"project_profiles" yaml:"project_profiles"`
	Features               []string                         `json:"features" yaml:"features"`
	Assurance              descriptorAssurance              `json:"assurance" yaml:"assurance"`
	Canonicalization       descriptorCanonicalization       `json:"canonicalization" yaml:"canonicalization"`
	Resolution             descriptorResolution             `json:"resolution" yaml:"resolution"`
}

type documentationCommandReference struct {
	Items []descriptorCommand `json:"items" yaml:"items"`
}

type documentationCapabilityReference struct {
	CommandTokens []string    `json:"command_tokens" yaml:"command_tokens"`
	MachineFlags  []string    `json:"machine_flags" yaml:"machine_flags"`
	ValueKinds    []ValueKind `json:"value_kinds" yaml:"value_kinds"`
}

// WriteDocumentationReferenceArtifacts refreshes generated docs reference files.
func WriteDocumentationReferenceArtifacts(repoRoot string) error {
	jsonPayload, err := documentationReferenceArtifacts()
	if err != nil {
		return err
	}

	jsonPath := filepath.Join(repoRoot, docsReferenceJSONRelativePath)
	if err := writeDocsReferenceFile(jsonPath, jsonPayload); err != nil {
		return err
	}
	return nil
}

func documentationReferenceArtifacts() ([]byte, error) {
	artifact := buildDocumentationReferenceArtifact()
	jsonPayload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal docs reference json: %w", err)
	}
	jsonPayload = append(jsonPayload, '\n')
	return jsonPayload, nil
}

func buildDocumentationReferenceArtifact() documentationReferenceArtifact {
	descriptor := buildCapabilityDescriptor()
	tokens := make([]string, 0, len(descriptor.Capabilities.Commands))
	for _, command := range descriptor.Capabilities.Commands {
		tokens = append(tokens, command.Token)
	}
	slices.Sort(tokens)
	tokens = slices.Compact(tokens)

	return documentationReferenceArtifact{
		ReferenceSchemaVersion: documentationReferenceSchemaVersion,
		DescriptorSchema:       descriptor.SchemaVersion,
		Binary:                 descriptor.Binary,
		Release:                descriptor.Release,
		Commands: documentationCommandReference{
			Items: descriptor.Capabilities.Commands,
		},
		Capabilities: documentationCapabilityReference{
			CommandTokens: tokens,
			MachineFlags:  descriptor.Capabilities.MachineFlags,
			ValueKinds:    descriptor.Capabilities.ValueKinds,
		},
		Compatibility:       descriptor.Compatibility,
		DistributionLayouts: descriptor.DistributionLayouts,
		ProjectProfiles:     descriptor.ProjectProfiles,
		Features:            descriptor.Features,
		Assurance:           descriptor.Assurance,
		Canonicalization:    descriptor.Canonicalization,
		Resolution:          descriptor.Resolution,
	}
}

func writeDocsReferenceFile(path string, payload []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create docs reference directory %q: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write docs reference file %q: %w", path, err)
	}
	return nil
}
