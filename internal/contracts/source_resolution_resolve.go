package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
)

func resolveSourceFromConfig(configPath, projectRoot string, rootData []byte, options ResolveOptions) (*SourceResolution, error) {
	sourceMap, err := decodeSourceMap(configPath, rootData)
	if err != nil {
		return nil, err
	}
	base := &SourceResolution{SelectedConfigPath: filepath.Clean(configPath), ProjectRoot: filepath.Clean(projectRoot)}
	return resolveConfiguredSource(base, configPath, projectRoot, sourceMap, options)
}

func decodeSourceMap(configPath string, rootData []byte) (map[string]any, error) {
	parsed, err := parseYAML(rootData)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	rootMap, err := expectObject(configPath, parsed, "root config")
	if err != nil {
		return nil, err
	}
	sourceRaw, ok := rootMap["source"]
	if !ok {
		return nil, &ValidationError{Path: configPath, Message: "root config is missing source"}
	}
	sourceMap, ok := sourceRaw.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: configPath, Message: "source must decode to an object"}
	}
	return sourceMap, nil
}

func resolveConfiguredSource(base *SourceResolution, configPath, projectRoot string, sourceMap map[string]any, options ResolveOptions) (*SourceResolution, error) {
	sourceType := fmt.Sprint(sourceMap["type"])
	switch sourceType {
	case string(SourceModeEmbedded):
		return resolveEmbeddedSource(base, configPath, projectRoot, sourceMap)
	case string(SourceModeGit):
		return resolveGitSource(base, configPath, sourceMap, options.GitTrust)
	case string(SourceModePath):
		return resolvePathSource(base, configPath, projectRoot, sourceMap, options.ExecutionMode)
	default:
		return nil, &ValidationError{Path: configPath, Message: fmt.Sprintf("unsupported source type %q", sourceType)}
	}
}

func resolveEmbeddedSource(base *SourceResolution, configPath, projectRoot string, sourceMap map[string]any) (*SourceResolution, error) {
	declaredRoot, absRoot, err := resolveEmbeddedSourceRoot(configPath, projectRoot, sourceMap)
	if err != nil {
		return nil, err
	}
	base.SourceRoot = declaredRoot
	base.SourceMode = SourceModeEmbedded
	base.SourceRef = "embedded"
	base.VerificationPosture = VerificationPostureEmbedded
	base.Tree = &LocalSourceTree{Root: absRoot, SnapshotKind: "live"}
	return base, nil
}

func resolvePathSource(base *SourceResolution, configPath, projectRoot string, sourceMap map[string]any, executionMode ExecutionMode) (*SourceResolution, error) {
	if executionMode == ExecutionModeRemoteCI {
		return nil, &ValidationError{Path: configPath, Message: "source.type=path is invalid in execution mode remote_ci"}
	}
	declaredRoot, absRoot, err := resolveDeclaredLocalSourceRoot(configPath, projectRoot, sourceMap)
	if err != nil {
		return nil, err
	}
	tree, err := snapshotLocalTree(absRoot)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	base.SourceRoot = declaredRoot
	base.SourceMode = SourceModePath
	base.SourceRef = declaredRoot
	base.VerificationPosture = VerificationPostureUnverifiedLocal
	base.Diagnostics = []ResolutionDiagnostic{{Severity: DiagnosticSeverityWarning, Code: "unverified_local_source", Message: "local path sources are unverified and non-auditable"}}
	base.Tree = tree
	return base, nil
}

func resolveGitSource(base *SourceResolution, configPath string, sourceMap map[string]any, gitTrust GitTrustInputs) (*SourceResolution, error) {
	resolver, subdir, err := newGitSourceResolver(configPath, sourceMap)
	if err != nil {
		return nil, err
	}
	result, err := resolveGitSourceMaterialization(resolver, subdir, sourceMap, gitTrust)
	if err != nil {
		return nil, err
	}
	applyGitResolution(base, subdir, result)
	return base, nil
}

type gitSourceResult struct {
	tree                  *LocalSourceTree
	commit                string
	ref                   string
	posture               VerificationPosture
	signedTagVerification *SignedTagVerification
	diagnostics           []ResolutionDiagnostic
}

func newGitSourceResolver(configPath string, sourceMap map[string]any) (gitResolver, string, error) {
	url := strings.TrimSpace(fmt.Sprint(sourceMap["url"]))
	if url == "" {
		return gitResolver{}, "", &ValidationError{Path: configPath, Message: "git source url must not be empty"}
	}
	if err := validateGitURL(url); err != nil {
		return gitResolver{}, "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	subdir, err := gitSourceSubdir(configPath, sourceMap)
	if err != nil {
		return gitResolver{}, "", err
	}
	return gitResolver{configPath: configPath, url: url}, subdir, nil
}

func gitSourceSubdir(configPath string, sourceMap map[string]any) (string, error) {
	subdir := "runecontext"
	if rawSubdir, ok := sourceMap["subdir"]; ok && strings.TrimSpace(fmt.Sprint(rawSubdir)) != "" {
		normalizedSubdir, err := normalizeContainedRelativePath(fmt.Sprint(rawSubdir))
		if err != nil {
			return "", &ValidationError{Path: configPath, Message: fmt.Sprintf("git subdir %v", err)}
		}
		return normalizedSubdir, nil
	}
	return subdir, nil
}
