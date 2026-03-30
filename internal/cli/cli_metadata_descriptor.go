package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	runecontextPackageName     = "runecontext"
	assuranceTierPlain         = "plain"
	metadataSchemaName         = "capability-descriptor.schema.json"
	metadataOutputInstancePath = "metadata-output.json"
)

type capabilityDescriptor struct {
	SchemaVersion       int                        `json:"schema_version" yaml:"schema_version"`
	Binary              string                     `json:"binary" yaml:"binary"`
	Release             descriptorRelease          `json:"release" yaml:"release"`
	Compatibility       descriptorCompatibility    `json:"compatibility" yaml:"compatibility"`
	DistributionLayouts []descriptorLayout         `json:"distribution_layouts" yaml:"distribution_layouts"`
	ProjectProfiles     []descriptorProject        `json:"project_profiles" yaml:"project_profiles"`
	Capabilities        descriptorCapabilities     `json:"capabilities" yaml:"capabilities"`
	Features            []string                   `json:"features" yaml:"features"`
	Assurance           descriptorAssurance        `json:"assurance" yaml:"assurance"`
	Canonicalization    descriptorCanonicalization `json:"canonicalization" yaml:"canonicalization"`
	Resolution          descriptorResolution       `json:"resolution" yaml:"resolution"`
}

type descriptorRelease struct {
	PackageName string `json:"package_name" yaml:"package_name"`
	Version     string `json:"version" yaml:"version"`
	Tag         string `json:"tag" yaml:"tag"`
}

type descriptorCompatibility struct {
	DirectlySupportedProjectVersions []string                `json:"directly_supported_project_versions" yaml:"directly_supported_project_versions"`
	UpgradeableFromProjectVersions   []string                `json:"upgradeable_from_project_versions" yaml:"upgradeable_from_project_versions"`
	ExplicitUpgradeEdges             []descriptorUpgradeEdge `json:"explicit_upgrade_edges" yaml:"explicit_upgrade_edges"`
}

type descriptorUpgradeEdge struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

type descriptorCapabilities struct {
	Commands     []descriptorCommand `json:"commands" yaml:"commands"`
	MachineFlags []string            `json:"machine_flags" yaml:"machine_flags"`
	ValueKinds   []ValueKind         `json:"value_kinds" yaml:"value_kinds"`
}

type descriptorCommand struct {
	Path    string   `json:"path" yaml:"path"`
	Token   string   `json:"token" yaml:"token"`
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

type descriptorResolution struct {
	SourceModes         []contracts.SourceMode          `json:"source_modes" yaml:"source_modes"`
	VerificationPosture []contracts.VerificationPosture `json:"verification_postures" yaml:"verification_postures"`
}

func buildCapabilityDescriptor() capabilityDescriptor {
	registry := CommandMetadataRegistry()
	version := normalizedRunecontextVersion()
	planner := defaultUpgradePlannerRegistry()

	return capabilityDescriptor{
		SchemaVersion: contracts.CapabilityDescriptorSchemaVersionForCLI(),
		Binary:        registry.Binary,
		Release: descriptorRelease{
			PackageName: runecontextPackageName,
			Version:     version,
			Tag:         "v" + version,
		},
		Compatibility: descriptorCompatibility{
			DirectlySupportedProjectVersions: deriveDirectlySupportedProjectVersions(version),
			UpgradeableFromProjectVersions:   deriveUpgradeableFromProjectVersions(planner),
			ExplicitUpgradeEdges:             deriveExplicitUpgradeEdges(planner),
		},
		DistributionLayouts: descriptorDistributionLayouts(),
		ProjectProfiles:     descriptorProjectProfiles(),
		Capabilities: descriptorCapabilities{
			Commands:     deriveDescriptorCommands(registry.Commands),
			MachineFlags: deriveMachineFlagNames(),
			ValueKinds:   []ValueKind{ValueKindNone, ValueKindText, ValueKindEnum},
		},
		Features:         descriptorFeatures(),
		Assurance:        buildDescriptorAssurance(),
		Canonicalization: buildDescriptorCanonicalization(),
		Resolution: descriptorResolution{
			SourceModes:         []contracts.SourceMode{contracts.SourceModeEmbedded, contracts.SourceModeGit, contracts.SourceModePath},
			VerificationPosture: descriptorVerificationPostures(),
		},
	}
}

func descriptorVerificationPostures() []contracts.VerificationPosture {
	return []contracts.VerificationPosture{
		contracts.VerificationPostureEmbedded,
		contracts.VerificationPosturePinnedCommit,
		contracts.VerificationPostureVerifiedSignedTag,
		contracts.VerificationPostureUnverifiedMutableRef,
		contracts.VerificationPostureUnverifiedLocal,
	}
}

func validateCapabilityDescriptorSchema(descriptor capabilityDescriptor) error {
	schemaRoot, err := locateSchemaRoot()
	if err != nil {
		return err
	}
	return validateCapabilityDescriptorSchemaAtRoot(schemaRoot, descriptor)
}

func validateCapabilityDescriptorSchemaAtRoot(schemaRoot string, descriptor capabilityDescriptor) error {
	validator := contracts.NewValidator(schemaRoot)
	payload, err := descriptorMap(descriptor)
	if err != nil {
		return fmt.Errorf("build capability descriptor validation payload: %w", err)
	}
	if err := validator.ValidateValue(metadataSchemaName, metadataOutputInstancePath, payload); err != nil {
		return err
	}
	return nil
}

func deriveDirectlySupportedProjectVersions(installedVersion string) []string {
	seen := map[string]struct{}{}
	candidates := make([]string, 0, 16)

	add := func(version string) {
		version = strings.TrimSpace(version)
		if version == "" {
			return
		}
		if _, ok := seen[version]; ok {
			return
		}
		seen[version] = struct{}{}
		candidates = append(candidates, version)
	}

	if installedVersion != "" {
		add(installedVersion)
	}
	for ordinal := 1; ordinal <= 40; ordinal++ {
		candidate := fmt.Sprintf("0.1.0-alpha.%d", ordinal)
		if isCompatibleProjectVersionForInstalled(candidate, installedVersion) {
			add(candidate)
		}
	}

	sortVersions(candidates)
	return candidates
}

func deriveUpgradeableFromProjectVersions(planner upgradePlannerRegistry) []string {
	seen := map[string]struct{}{}
	versions := make([]string, 0, len(planner.edges))
	for edge := range planner.edges {
		if _, ok := seen[edge.From]; ok {
			continue
		}
		seen[edge.From] = struct{}{}
		versions = append(versions, edge.From)
	}
	sortVersions(versions)
	return versions
}

func deriveExplicitUpgradeEdges(planner upgradePlannerRegistry) []descriptorUpgradeEdge {
	edges := make([]descriptorUpgradeEdge, 0, len(planner.edges))
	for edge := range planner.edges {
		edges = append(edges, descriptorUpgradeEdge{From: edge.From, To: edge.To})
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
	return edges
}

func sortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		li, lok := alphaOrdinal(versions[i])
		lj, jok := alphaOrdinal(versions[j])
		if lok && jok {
			return li < lj
		}
		if lok != jok {
			return lok
		}
		return versions[i] < versions[j]
	})
}

func deriveDescriptorCommands(commands []CommandMetadata) []descriptorCommand {
	flattened := flattenRegistryCommands(commands)
	out := make([]descriptorCommand, 0, len(flattened))
	for _, command := range flattened {
		if strings.HasPrefix(command.Path, "-") {
			continue
		}
		item := descriptorCommand{
			Path:  command.Path,
			Token: commandToken(command.Path),
		}
		if command.Path == "version" {
			item.Aliases = []string{"--version", "-v"}
		}
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func flattenRegistryCommands(commands []CommandMetadata) []CommandMetadata {
	items := make([]CommandMetadata, 0, len(commands))
	for _, command := range commands {
		items = append(items, command)
		items = append(items, flattenRegistryCommands(command.Subcommands)...)
	}
	return items
}

func commandToken(path string) string {
	parts := strings.Fields(path)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		parts[i] = strings.TrimPrefix(parts[i], "--")
		parts[i] = strings.TrimPrefix(parts[i], "-")
		parts[i] = strings.ReplaceAll(parts[i], "-", "_")
	}
	return strings.Join(parts, "_")
}

func deriveMachineFlagNames() []string {
	flags := make([]string, 0, len(machineFlagHandlers))
	for name := range machineFlagHandlers {
		flags = append(flags, name)
	}
	slices.Sort(flags)
	return flags
}

func releaseManifestDescriptorFromJSON(raw []byte) (map[string]any, error) {
	var manifest map[string]any
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}
	value, ok := manifest["metadata_descriptor"]
	if !ok {
		return nil, fmt.Errorf("release manifest missing metadata_descriptor")
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("release manifest metadata_descriptor must be an object")
	}
	return obj, nil
}
