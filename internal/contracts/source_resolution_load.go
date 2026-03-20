package contracts

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

func NewSSHAllowedSignersVerifier(allowedSigners []byte) (*SSHAllowedSignersVerifier, error) {
	if len(bytes.TrimSpace(allowedSigners)) == 0 {
		return nil, fmt.Errorf("ssh allowed signers data must not be empty")
	}
	return &SSHAllowedSignersVerifier{allowedSigners: append([]byte(nil), allowedSigners...), gitExecutable: gitExecutable}, nil
}

func NewSSHAllowedSignersVerifierWithGitExecutable(allowedSigners []byte, executable string) (*SSHAllowedSignersVerifier, error) {
	verifier, err := NewSSHAllowedSignersVerifier(allowedSigners)
	if err != nil {
		return nil, err
	}
	executable = strings.TrimSpace(executable)
	if executable == "" {
		return nil, fmt.Errorf("git executable must not be empty")
	}
	verifier.gitExecutable = executable
	return verifier, nil
}

func (v *Validator) LoadProject(path string, options ResolveOptions) (*LoadedProject, error) {
	options = normalizeResolveOptions(options)
	configPath, projectRoot, err := discoverConfig(path, options.ConfigDiscovery)
	if err != nil {
		return nil, err
	}
	rootData, rootConfig, err := loadValidatedRootConfig(v, configPath)
	if err != nil {
		return nil, err
	}
	resolution, err := resolveSourceFromConfig(configPath, projectRoot, rootData, options)
	if err != nil {
		return nil, err
	}
	return &LoadedProject{RootConfigData: append([]byte(nil), rootData...), RootConfig: rootConfig, Resolution: resolution}, nil
}

func loadValidatedRootConfig(v *Validator, configPath string) ([]byte, map[string]any, error) {
	rootData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	if err := v.ValidateYAMLFile("runecontext.schema.json", configPath, rootData); err != nil {
		return nil, nil, err
	}
	parsed, err := parseYAML(rootData)
	if err != nil {
		return nil, nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	rootConfig, err := expectObject(configPath, parsed, "root config")
	if err != nil {
		return nil, nil, err
	}
	return rootData, rootConfig, nil
}

func (v *Validator) ValidateLoadedProject(loaded *LoadedProject) (*ProjectIndex, error) {
	contentRoot, rootConfigPath, rootData, err := validateLoadedProjectInputs(loaded)
	if err != nil {
		return nil, err
	}
	index, err := buildResolvedProjectIndex(v, contentRoot, rootConfigPath, rootData, loaded.Resolution)
	if err != nil {
		return nil, err
	}
	if err := validateResolvedStatusFiles(v, index, rootConfigPath, rootData); err != nil {
		return nil, err
	}
	if err := validateResolvedProjectReferences(index); err != nil {
		return nil, err
	}
	return index, nil
}

func validateLoadedProjectInputs(loaded *LoadedProject) (string, string, []byte, error) {
	if loaded == nil || loaded.Resolution == nil {
		return "", "", nil, fmt.Errorf("loaded project is required")
	}
	contentRoot := loaded.Resolution.MaterializedRoot()
	if contentRoot == "" {
		return "", "", nil, fmt.Errorf("resolved source root is unavailable")
	}
	return contentRoot, loaded.Resolution.SelectedConfigPath, loaded.RootConfigData, nil
}

func buildResolvedProjectIndex(v *Validator, contentRoot, rootConfigPath string, rootData []byte, resolution *SourceResolution) (*ProjectIndex, error) {
	index, err := buildProjectIndex(v, contentRoot)
	if err != nil {
		return nil, err
	}
	index.RootConfigPath = rootConfigPath
	index.ContentRoot = contentRoot
	index.Resolution = resolution
	bundles, err := loadBundleCatalog(v, rootConfigPath, rootData, contentRoot)
	if err != nil {
		return nil, err
	}
	index.Bundles = bundles
	return index, nil
}

func validateResolvedStatusFiles(v *Validator, index *ProjectIndex, rootConfigPath string, rootData []byte) error {
	for path, record := range index.StatusFiles {
		if err := v.ValidateExtensionOptIn(rootConfigPath, rootData, path, record.Raw); err != nil {
			return err
		}
		if err := validateResolvedStatusFileReferences(path, record, index); err != nil {
			return err
		}
	}
	return nil
}

func validateResolvedStatusFileReferences(path string, record StatusFileRecord, index *ProjectIndex) error {
	for _, key := range []string{"depends_on", "informed_by", "related_changes", "supersedes", "superseded_by"} {
		if err := validateChangeIDRefs(path, key, record.Data[key], index.ChangeIDs); err != nil {
			return err
		}
	}
	if err := validatePathRefs(path, "related_specs", record.Data["related_specs"], index.SpecPaths); err != nil {
		return err
	}
	return validatePathRefs(path, "related_decisions", record.Data["related_decisions"], index.DecisionPaths)
}

func validateResolvedProjectReferences(index *ProjectIndex) error {
	for _, validate := range []func(*ProjectIndex) error{
		validateStandardMetadata,
		validateChangeStandardReferences,
		validateBundleStandardSelections,
		validateStandardReferenceBodies,
		validateChangeLifecycleConsistency,
		validateRelatedChangeReciprocity,
		validateSupersessionConsistency,
		validateArtifactTraceabilityConsistency,
		validateMarkdownDeepRefs,
	} {
		if err := validate(index); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) ValidateProjectWithOptions(path string, options ResolveOptions) (*ProjectIndex, error) {
	loaded, err := v.LoadProject(path, options)
	if err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		_ = loaded.Close()
		return nil, err
	}
	return index, nil
}

func normalizeResolveOptions(options ResolveOptions) ResolveOptions {
	if options.ConfigDiscovery == "" {
		options.ConfigDiscovery = ConfigDiscoveryExplicitRoot
	}
	if options.ExecutionMode == "" {
		options.ExecutionMode = ExecutionModeLocal
	}
	return options
}
