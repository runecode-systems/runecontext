package contracts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	gitCommitPattern    = regexp.MustCompile(`^[a-f0-9]{40}$`)
	gitRefPattern       = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)
	gitURLSchemePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://`)
	gitCommandTimeout   = 30 * time.Second
	gitExecutable       = "git"
	gitURLControlChars  = "\x00\r\n\t"
	gitAllowedProtocols = []string{"file", "git", "http", "https", "ssh"}
	gitCommandRunner    = runGitCommandContext
	localSnapshotLimits = snapshotLimits{
		MaxFiles: 10000,
		MaxBytes: 128 << 20,
		MaxDepth: 64,
		Excludes: map[string]struct{}{
			".git": {},
		},
	}
)

type snapshotLimits struct {
	MaxFiles int
	MaxBytes int64
	MaxDepth int
	Excludes map[string]struct{}
}

type snapshotState struct {
	files int
	bytes int64
}

type ExecutionMode string

const (
	ExecutionModeLocal    ExecutionMode = "local"
	ExecutionModeRemoteCI ExecutionMode = "remote_ci"
)

type ConfigDiscoveryMode string

const (
	ConfigDiscoveryExplicitRoot    ConfigDiscoveryMode = "explicit_root"
	ConfigDiscoveryNearestAncestor ConfigDiscoveryMode = "nearest_ancestor"
)

type SourceMode string

const (
	SourceModeEmbedded SourceMode = "embedded"
	SourceModeGit      SourceMode = "git"
	SourceModePath     SourceMode = "path"
)

type VerificationPosture string

const (
	VerificationPosturePinnedCommit         VerificationPosture = "pinned_commit"
	VerificationPostureVerifiedSignedTag    VerificationPosture = "verified_signed_tag"
	VerificationPostureUnverifiedMutableRef VerificationPosture = "unverified_mutable_ref"
	VerificationPostureUnverifiedLocal      VerificationPosture = "unverified_local_source"
	VerificationPostureEmbedded             VerificationPosture = "embedded"
)

type DiagnosticSeverity string

const (
	DiagnosticSeverityInfo    DiagnosticSeverity = "info"
	DiagnosticSeverityWarning DiagnosticSeverity = "warning"
	DiagnosticSeverityError   DiagnosticSeverity = "error"
)

type ResolveOptions struct {
	ConfigDiscovery ConfigDiscoveryMode
	ExecutionMode   ExecutionMode
	GitTrust        GitTrustInputs
}

type GitTrustInputs struct {
	SignedTagVerifier SignedTagVerifier
}

type SignedTagVerifier interface {
	VerifySignedTag(repoRoot, tagName string) (*SignedTagVerification, error)
}

type SignedTagVerification struct {
	SignerIdentity    string                 `json:"signer_identity,omitempty" yaml:"signer_identity,omitempty"`
	SignerFingerprint string                 `json:"signer_fingerprint,omitempty" yaml:"signer_fingerprint,omitempty"`
	Diagnostics       []ResolutionDiagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

type SignedTagFailureReason string

const (
	SignedTagFailureMissingTrust         SignedTagFailureReason = "missing_trust"
	SignedTagFailureUnsignedTag          SignedTagFailureReason = "unsigned_tag"
	SignedTagFailureInvalidSignature     SignedTagFailureReason = "invalid_signature"
	SignedTagFailureUntrustedSigner      SignedTagFailureReason = "untrusted_signer"
	SignedTagFailureExpectCommitMismatch SignedTagFailureReason = "expect_commit_mismatch"
	SignedTagFailureVerificationFailed   SignedTagFailureReason = "verification_failed"
)

type SignedTagVerificationError struct {
	Path              string
	Tag               string
	Reason            SignedTagFailureReason
	Message           string
	ResolvedCommit    string
	SignerIdentity    string
	SignerFingerprint string
	Diagnostics       []ResolutionDiagnostic
}

func (e *SignedTagVerificationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

func (e *SignedTagVerificationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return &ValidationError{Path: e.Path, Message: e.Message}
}

type SSHAllowedSignersVerifier struct {
	allowedSigners []byte
	gitExecutable  string
}

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

type ResolutionDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
}

type LocalSourceTree struct {
	Root         string `json:"-" yaml:"-"`
	SnapshotKind string `json:"snapshot_kind,omitempty" yaml:"snapshot_kind,omitempty"`
	cleanupRoot  string
}

func (t *LocalSourceTree) Close() error {
	if t == nil || t.cleanupRoot == "" {
		return nil
	}
	err := os.RemoveAll(t.cleanupRoot)
	t.cleanupRoot = ""
	return err
}

type SourceResolution struct {
	SelectedConfigPath        string                 `json:"selected_config_path" yaml:"selected_config_path"`
	ProjectRoot               string                 `json:"project_root" yaml:"project_root"`
	SourceRoot                string                 `json:"source_root" yaml:"source_root"`
	SourceMode                SourceMode             `json:"source_mode" yaml:"source_mode"`
	SourceRef                 string                 `json:"source_ref" yaml:"source_ref"`
	ResolvedCommit            string                 `json:"resolved_commit,omitempty" yaml:"resolved_commit,omitempty"`
	VerificationPosture       VerificationPosture    `json:"verification_posture" yaml:"verification_posture"`
	VerifiedSignerIdentity    string                 `json:"verified_signer_identity,omitempty" yaml:"verified_signer_identity,omitempty"`
	VerifiedSignerFingerprint string                 `json:"verified_signer_fingerprint,omitempty" yaml:"verified_signer_fingerprint,omitempty"`
	Diagnostics               []ResolutionDiagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
	Tree                      *LocalSourceTree       `json:"-" yaml:"-"`
}

func (v *SSHAllowedSignersVerifier) VerifySignedTag(repoRoot, tagName string) (*SignedTagVerification, error) {
	if v == nil {
		return nil, fmt.Errorf("signed tag verifier is required")
	}
	tempRoot, err := os.MkdirTemp("", "runectx-allowed-signers-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempRoot)
	allowedSignersPath := filepath.Join(tempRoot, "allowed_signers")
	if err := os.WriteFile(allowedSignersPath, v.allowedSigners, 0o600); err != nil {
		return nil, err
	}
	result := runGitCaptured(
		v.gitCommandArgs(repoRoot, allowedSignersPath, tagName),
		v.gitExecutable,
	)
	output := strings.TrimSpace(result.Output)
	if result.TimedOut {
		message := fmt.Sprintf("signed tag %q verification failed: git %s: command timed out after %s", tagName, sanitizeGitArgs(result.Args), gitCommandTimeout)
		return nil, &SignedTagVerificationError{
			Tag:     tagName,
			Reason:  SignedTagFailureVerificationFailed,
			Message: message,
			Diagnostics: []ResolutionDiagnostic{{
				Severity: DiagnosticSeverityError,
				Code:     string(SignedTagFailureVerificationFailed),
				Message:  message,
			}},
		}
	}
	if result.Err != nil && result.ExitCode == -1 {
		message := sanitizeGitMessage(strings.TrimSpace(result.Output))
		if message == "" {
			message = sanitizeGitMessage(result.Err.Error())
		}
		return nil, &SignedTagVerificationError{
			Tag:     tagName,
			Reason:  SignedTagFailureVerificationFailed,
			Message: fmt.Sprintf("signed tag %q verification failed: %s", tagName, message),
			Diagnostics: []ResolutionDiagnostic{{
				Severity: DiagnosticSeverityError,
				Code:     string(SignedTagFailureVerificationFailed),
				Message:  fmt.Sprintf("signed tag %q verification failed: %s", tagName, message),
			}},
		}
	}
	if result.ExitCode == 0 {
		identity, fingerprint, err := parseTrustedSSHVerifyTagOutput(output)
		if err != nil {
			return nil, fmt.Errorf("parse trusted signed-tag verification output: %w", err)
		}
		return &SignedTagVerification{
			SignerIdentity:    identity,
			SignerFingerprint: fingerprint,
		}, nil
	}
	reason := classifySignedTagFailure(output)
	message := signedTagFailureMessage(tagName, reason, output)
	return nil, &SignedTagVerificationError{
		Tag:     tagName,
		Reason:  reason,
		Message: message,
		Diagnostics: []ResolutionDiagnostic{{
			Severity: DiagnosticSeverityError,
			Code:     string(reason),
			Message:  message,
		}},
	}
}

func (v *SSHAllowedSignersVerifier) gitCommandArgs(repoRoot, allowedSignersPath, tagName string) []string {
	return []string{"-C", repoRoot, "-c", "gpg.format=ssh", "-c", "gpg.ssh.allowedSignersFile=" + allowedSignersPath, "verify-tag", "--raw", tagName}
}

func validateSignedTagVerification(verification *SignedTagVerification, tagName string) error {
	if verification == nil {
		return &SignedTagVerificationError{
			Tag:     tagName,
			Reason:  SignedTagFailureVerificationFailed,
			Message: fmt.Sprintf("signed tag %q verification failed: verifier returned no verification details", tagName),
			Diagnostics: []ResolutionDiagnostic{{
				Severity: DiagnosticSeverityError,
				Code:     string(SignedTagFailureVerificationFailed),
				Message:  fmt.Sprintf("signed tag %q verification failed: verifier returned no verification details", tagName),
			}},
		}
	}
	if strings.TrimSpace(verification.SignerIdentity) == "" || strings.TrimSpace(verification.SignerFingerprint) == "" {
		return &SignedTagVerificationError{
			Tag:     tagName,
			Reason:  SignedTagFailureVerificationFailed,
			Message: fmt.Sprintf("signed tag %q verification failed: verifier returned incomplete signer details", tagName),
			Diagnostics: []ResolutionDiagnostic{{
				Severity: DiagnosticSeverityError,
				Code:     string(SignedTagFailureVerificationFailed),
				Message:  fmt.Sprintf("signed tag %q verification failed: verifier returned incomplete signer details", tagName),
			}},
		}
	}
	return nil
}

func (r *SourceResolution) Close() error {
	if r == nil || r.Tree == nil {
		return nil
	}
	return r.Tree.Close()
}

func (r *SourceResolution) MaterializedRoot() string {
	if r == nil || r.Tree == nil {
		return ""
	}
	return r.Tree.Root
}

type LoadedProject struct {
	RootConfigData []byte
	RootConfig     map[string]any
	Resolution     *SourceResolution
}

func (p *LoadedProject) Close() error {
	if p == nil || p.Resolution == nil {
		return nil
	}
	return p.Resolution.Close()
}

func (v *Validator) LoadProject(path string, options ResolveOptions) (*LoadedProject, error) {
	options = normalizeResolveOptions(options)
	configPath, projectRoot, err := discoverConfig(path, options.ConfigDiscovery)
	if err != nil {
		return nil, err
	}
	rootData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	if err := v.ValidateYAMLFile("runecontext.schema.json", configPath, rootData); err != nil {
		return nil, err
	}
	parsed, err := parseYAML(rootData)
	if err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	rootConfig, err := expectObject(configPath, parsed, "root config")
	if err != nil {
		return nil, err
	}
	resolution, err := resolveSourceFromConfig(configPath, projectRoot, rootData, options)
	if err != nil {
		return nil, err
	}
	return &LoadedProject{
		RootConfigData: append([]byte(nil), rootData...),
		RootConfig:     rootConfig,
		Resolution:     resolution,
	}, nil
}

func (v *Validator) ValidateLoadedProject(loaded *LoadedProject) (*ProjectIndex, error) {
	if loaded == nil || loaded.Resolution == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	contentRoot := loaded.Resolution.MaterializedRoot()
	if contentRoot == "" {
		return nil, fmt.Errorf("resolved source root is unavailable")
	}
	rootConfigPath := loaded.Resolution.SelectedConfigPath
	rootData := loaded.RootConfigData
	index, err := buildProjectIndex(v, contentRoot)
	if err != nil {
		return nil, err
	}
	index.RootConfigPath = rootConfigPath
	index.ContentRoot = contentRoot
	index.Resolution = loaded.Resolution
	bundles, err := loadBundleCatalog(v, rootConfigPath, rootData, contentRoot)
	if err != nil {
		return nil, err
	}
	index.Bundles = bundles
	for path, record := range index.StatusFiles {
		if err := v.ValidateExtensionOptIn(rootConfigPath, rootData, path, record.Raw); err != nil {
			return nil, err
		}
		for _, key := range []string{"depends_on", "informed_by", "related_changes", "supersedes", "superseded_by"} {
			if err := validateChangeIDRefs(path, key, record.Data[key], index.ChangeIDs); err != nil {
				return nil, err
			}
		}
		if err := validatePathRefs(path, "related_specs", record.Data["related_specs"], index.SpecPaths); err != nil {
			return nil, err
		}
		if err := validatePathRefs(path, "related_decisions", record.Data["related_decisions"], index.DecisionPaths); err != nil {
			return nil, err
		}
	}
	if err := validateChangeLifecycleConsistency(index); err != nil {
		return nil, err
	}
	if err := validateRelatedChangeReciprocity(index); err != nil {
		return nil, err
	}
	if err := validateSupersessionConsistency(index); err != nil {
		return nil, err
	}
	if err := validateArtifactTraceabilityConsistency(index); err != nil {
		return nil, err
	}
	if err := validateMarkdownDeepRefs(index); err != nil {
		return nil, err
	}
	return index, nil
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

func discoverConfig(path string, mode ConfigDiscoveryMode) (string, string, error) {
	start, err := filepath.Abs(path)
	if err != nil {
		return "", "", &ValidationError{Path: path, Message: err.Error()}
	}
	info, err := os.Stat(start)
	if err == nil && !info.IsDir() {
		if filepath.Base(start) == "runecontext.yaml" {
			start = filepath.Dir(start)
		} else {
			start = filepath.Dir(start)
		}
	}
	switch mode {
	case ConfigDiscoveryNearestAncestor:
		current := start
		for {
			candidate := filepath.Join(current, "runecontext.yaml")
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Clean(candidate), filepath.Clean(current), nil
			}
			next := filepath.Dir(current)
			if next == current {
				return "", "", &ValidationError{Path: start, Message: "no runecontext.yaml found in current directory or ancestors"}
			}
			current = next
		}
	case ConfigDiscoveryExplicitRoot:
		candidate := filepath.Join(start, "runecontext.yaml")
		if _, err := os.Stat(candidate); err != nil {
			return "", "", &ValidationError{Path: candidate, Message: err.Error()}
		}
		return filepath.Clean(candidate), filepath.Clean(start), nil
	default:
		return "", "", &ValidationError{Path: start, Message: fmt.Sprintf("unsupported config discovery mode %q", mode)}
	}
}

func resolveSourceFromConfig(configPath, projectRoot string, rootData []byte, options ResolveOptions) (*SourceResolution, error) {
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
	sourceType := fmt.Sprint(sourceMap["type"])
	base := &SourceResolution{
		SelectedConfigPath: filepath.Clean(configPath),
		ProjectRoot:        filepath.Clean(projectRoot),
	}
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
	base.Diagnostics = []ResolutionDiagnostic{
		{
			Severity: DiagnosticSeverityWarning,
			Code:     "unverified_local_source",
			Message:  "local path sources are unverified and non-auditable",
		},
	}
	base.Tree = tree
	return base, nil
}

func resolveGitSource(base *SourceResolution, configPath string, sourceMap map[string]any, gitTrust GitTrustInputs) (*SourceResolution, error) {
	url := strings.TrimSpace(fmt.Sprint(sourceMap["url"]))
	if url == "" {
		return nil, &ValidationError{Path: configPath, Message: "git source url must not be empty"}
	}
	if err := validateGitURL(url); err != nil {
		return nil, &ValidationError{Path: configPath, Message: err.Error()}
	}
	subdir := "runecontext"
	if rawSubdir, ok := sourceMap["subdir"]; ok && strings.TrimSpace(fmt.Sprint(rawSubdir)) != "" {
		normalizedSubdir, err := normalizeContainedRelativePath(fmt.Sprint(rawSubdir))
		if err != nil {
			return nil, &ValidationError{Path: configPath, Message: fmt.Sprintf("git subdir %v", err)}
		}
		subdir = normalizedSubdir
	}
	resolver := gitResolver{configPath: configPath, url: url}
	var (
		tree                  *LocalSourceTree
		commit                string
		ref                   string
		posture               VerificationPosture
		signedTagVerification *SignedTagVerification
		diagnostics           []ResolutionDiagnostic
		err                   error
	)
	if rawCommit, ok := sourceMap["commit"]; ok && strings.TrimSpace(fmt.Sprint(rawCommit)) != "" {
		ref = strings.TrimSpace(fmt.Sprint(rawCommit))
		if err := validateGitCommit(ref); err != nil {
			return nil, &ValidationError{Path: configPath, Message: err.Error()}
		}
		commit = ref
		posture = VerificationPosturePinnedCommit
		tree, err = resolver.materialize(ref, subdir)
	} else if rawRef, ok := sourceMap["ref"]; ok && strings.TrimSpace(fmt.Sprint(rawRef)) != "" {
		ref = strings.TrimSpace(fmt.Sprint(rawRef))
		if err := validateGitRef(ref); err != nil {
			return nil, &ValidationError{Path: configPath, Message: err.Error()}
		}
		if allow, _ := sourceMap["allow_mutable_ref"].(bool); !allow {
			return nil, &ValidationError{Path: configPath, Message: "mutable git refs require allow_mutable_ref: true"}
		}
		posture = VerificationPostureUnverifiedMutableRef
		diagnostics = append(diagnostics, ResolutionDiagnostic{
			Severity: DiagnosticSeverityWarning,
			Code:     "mutable_ref",
			Message:  "mutable git refs are unverified and may resolve differently over time",
		})
		tree, commit, err = resolver.materializeRef(ref, subdir)
	} else if _, ok := sourceMap["signed_tag"]; ok {
		ref = strings.TrimSpace(fmt.Sprint(sourceMap["signed_tag"]))
		if ref == "" {
			return nil, &ValidationError{Path: configPath, Message: "git signed_tag must not be empty"}
		}
		if err := validateGitRef(ref); err != nil {
			return nil, &ValidationError{Path: configPath, Message: strings.Replace(err.Error(), "git ref", "git signed_tag", 1)}
		}
		expectCommit, ok := sourceMap["expect_commit"]
		if !ok || strings.TrimSpace(fmt.Sprint(expectCommit)) == "" {
			return nil, &ValidationError{Path: configPath, Message: "git expect_commit must not be empty"}
		}
		expectCommitValue := strings.TrimSpace(fmt.Sprint(expectCommit))
		if err := validateGitCommit(expectCommitValue); err != nil {
			return nil, &ValidationError{Path: configPath, Message: strings.Replace(err.Error(), "git commit", "git expect_commit", 1)}
		}
		if gitTrust.SignedTagVerifier == nil {
			return nil, &SignedTagVerificationError{
				Path:    configPath,
				Tag:     ref,
				Reason:  SignedTagFailureMissingTrust,
				Message: "signed tag resolution requires explicit trusted signer inputs",
				Diagnostics: []ResolutionDiagnostic{{
					Severity: DiagnosticSeverityError,
					Code:     string(SignedTagFailureMissingTrust),
					Message:  "signed tag resolution requires explicit trusted signer inputs",
				}},
			}
		}
		posture = VerificationPostureVerifiedSignedTag
		tree, commit, signedTagVerification, err = resolver.materializeSignedTag(ref, expectCommitValue, subdir, gitTrust.SignedTagVerifier)
	} else {
		return nil, &ValidationError{Path: configPath, Message: "git source must declare commit, signed_tag, or ref"}
	}
	if err != nil {
		return nil, err
	}
	base.SourceRoot = subdir
	base.SourceMode = SourceModeGit
	base.SourceRef = ref
	base.ResolvedCommit = commit
	base.VerificationPosture = posture
	if signedTagVerification != nil {
		base.VerifiedSignerIdentity = signedTagVerification.SignerIdentity
		base.VerifiedSignerFingerprint = signedTagVerification.SignerFingerprint
		diagnostics = append(diagnostics, signedTagVerification.Diagnostics...)
	}
	base.Diagnostics = diagnostics
	base.Tree = tree
	return base, nil
}

func resolveEmbeddedSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared, err := normalizeContainedRelativePath(rawPath)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %v", err)}
	}
	absRoot := filepath.Clean(filepath.Join(projectRoot, filepath.FromSlash(declared)))
	resolvedProjectRoot, resolvedRoot, err := canonicalizePaths(projectRoot, absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !isWithinRoot(resolvedProjectRoot, resolvedRoot) {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("embedded source path %q escapes the selected project root", rawPath)}
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("resolved source root %q is not a directory", absRoot)}
	}
	return declared, absRoot, nil
}

func resolveDeclaredLocalSourceRoot(configPath, projectRoot string, sourceMap map[string]any) (string, string, error) {
	rawPath := strings.TrimSpace(fmt.Sprint(sourceMap["path"]))
	if rawPath == "" {
		return "", "", &ValidationError{Path: configPath, Message: "content root path must not be empty"}
	}
	declared := cleanSourceRootValue(rawPath)
	absRoot := declared
	if !filepath.IsAbs(rawPath) {
		absRoot = filepath.Join(projectRoot, rawPath)
	}
	absRoot = filepath.Clean(absRoot)
	info, err := os.Stat(absRoot)
	if err != nil {
		return "", "", &ValidationError{Path: configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		return "", "", &ValidationError{Path: configPath, Message: fmt.Sprintf("resolved source root %q is not a directory", absRoot)}
	}
	return declared, absRoot, nil
}

func normalizeContainedRelativePath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("must not be empty")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("must not be absolute")
	}
	cleaned := filepath.Clean(value)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("must not escape its containing root")
	}
	return filepath.ToSlash(cleaned), nil
}

func validateGitURL(url string) error {
	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("git source url must not start with '-'")
	}
	if strings.ContainsAny(url, gitURLControlChars+" ") {
		return fmt.Errorf("git source url contains unsupported whitespace or control characters")
	}
	lower := strings.ToLower(url)
	if strings.HasPrefix(lower, "ext::") {
		return fmt.Errorf("git source url must not use remote-helper forms")
	}
	if gitURLSchemePattern.MatchString(url) {
		scheme := strings.ToLower(url[:strings.Index(url, "://")])
		for _, allowed := range gitAllowedProtocols {
			if scheme == allowed {
				return nil
			}
		}
		return fmt.Errorf("git source url scheme %q is not allowed", scheme)
	}
	if strings.Contains(url, "::") {
		return fmt.Errorf("git source url must not use remote-helper forms")
	}
	return nil
}

func validateGitCommit(commit string) error {
	if strings.HasPrefix(commit, "-") {
		return fmt.Errorf("git commit must not start with '-'")
	}
	if !gitCommitPattern.MatchString(commit) {
		return fmt.Errorf("git commit must be a 40-character lowercase hex SHA")
	}
	return nil
}

func validateGitRef(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("git ref must not start with '-'")
	}
	if !gitRefPattern.MatchString(ref) {
		return fmt.Errorf("git ref contains unsupported characters")
	}
	if strings.Contains(ref, "..") {
		return fmt.Errorf("git ref must not contain '..'")
	}
	if strings.Contains(ref, "//") {
		return fmt.Errorf("git ref must not contain consecutive '/'")
	}
	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") {
		return fmt.Errorf("git ref must not start or end with '/'")
	}
	if strings.HasSuffix(ref, ".lock") {
		return fmt.Errorf("git ref must not end with '.lock'")
	}
	for _, segment := range strings.Split(ref, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("git ref contains an invalid path segment")
		}
		if strings.HasPrefix(segment, ".") {
			return fmt.Errorf("git ref segments must not start with '.'")
		}
	}
	return nil
}

func cleanSourceRootValue(value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func snapshotLocalTree(sourceRoot string) (*LocalSourceTree, error) {
	realRoot, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		return nil, err
	}
	tempRoot, err := os.MkdirTemp("", "runectx-local-source-")
	if err != nil {
		return nil, err
	}
	snapshotRoot := filepath.Join(tempRoot, "snapshot")
	if err := copyResolvedTree(realRoot, snapshotRoot, realRoot, map[string]struct{}{}, localSnapshotLimits, &snapshotState{}, 0); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, err
	}
	return &LocalSourceTree{Root: snapshotRoot, SnapshotKind: "snapshot_copy", cleanupRoot: tempRoot}, nil
}

func copyResolvedTree(sourcePath, destPath, root string, active map[string]struct{}, limits snapshotLimits, state *snapshotState, depth int) error {
	if depth > limits.MaxDepth {
		return fmt.Errorf("local source tree exceeds maximum depth of %d", limits.MaxDepth)
	}
	resolved, err := filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return err
	}
	if !isWithinRoot(root, resolved) {
		return fmt.Errorf("resolved path %q escapes declared local source tree", resolved)
	}
	if _, ok := active[resolved]; ok {
		return fmt.Errorf("symlink cycle detected at %q", resolved)
	}
	active[resolved] = struct{}{}
	defer delete(active, resolved)

	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if depth > 0 {
			if _, excluded := limits.Excludes[filepath.Base(resolved)]; excluded {
				return nil
			}
		}
		if err := os.MkdirAll(destPath, 0o755); err != nil {
			return err
		}
		entries, err := os.ReadDir(resolved)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			childSource := filepath.Join(resolved, entry.Name())
			childDest := filepath.Join(destPath, entry.Name())
			if err := copyResolvedTree(childSource, childDest, root, active, limits, state, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported non-regular file %q in local source tree", resolved)
	}
	if err := validateOpenPathWithinRoot(resolved, root); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	state.files++
	if state.files > limits.MaxFiles {
		return fmt.Errorf("local source tree exceeds maximum file count of %d", limits.MaxFiles)
	}
	state.bytes += info.Size()
	if state.bytes > limits.MaxBytes {
		return fmt.Errorf("local source tree exceeds maximum snapshot size of %d bytes", limits.MaxBytes)
	}
	src, err := os.Open(resolved)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	closeErr := dst.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func isWithinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

type gitResolver struct {
	configPath string
	url        string
}

type gitCommandResult struct {
	Args     []string
	Output   string
	ExitCode int
	Err      error
	TimedOut bool
}

func (r gitResolver) materialize(commit, subdir string) (*LocalSourceTree, error) {
	tree, resolvedCommit, err := r.materializeCommitToTree(commit, subdir)
	if err != nil {
		return nil, err
	}
	if resolvedCommit != commit {
		_ = tree.Close()
		return nil, &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git commit %q did not match pinned commit %q", resolvedCommit, commit)}
	}
	return tree, nil
}

func (r gitResolver) materializeRef(ref, subdir string) (*LocalSourceTree, string, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "--depth", "1", "origin", ref); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", "FETCH_HEAD"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	return r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
}

func (r gitResolver) materializeSignedTag(tagName, expectCommit, subdir string, verifier SignedTagVerifier) (*LocalSourceTree, string, *SignedTagVerification, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", nil, err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "origin", "+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	verification, err := verifier.VerifySignedTag(repoRoot, tagName)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		var verificationErr *SignedTagVerificationError
		if errors.As(err, &verificationErr) {
			if verificationErr.Path == "" {
				verificationErr.Path = r.configPath
			}
			if verificationErr.Tag == "" {
				verificationErr.Tag = tagName
			}
			return nil, "", nil, verificationErr
		}
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if err := validateSignedTagVerification(verification, tagName); err != nil {
		_ = os.RemoveAll(tempRoot)
		var verificationErr *SignedTagVerificationError
		if errors.As(err, &verificationErr) {
			verificationErr.Path = r.configPath
			if verificationErr.Tag == "" {
				verificationErr.Tag = tagName
			}
			return nil, "", nil, verificationErr
		}
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	commitOutput, err := gitOutput("-C", repoRoot, "rev-parse", tagName+"^{commit}")
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	resolvedCommit := strings.TrimSpace(commitOutput)
	if resolvedCommit != expectCommit {
		_ = os.RemoveAll(tempRoot)
		message := fmt.Sprintf("signed tag %q resolved commit %q did not match expect_commit %q", tagName, resolvedCommit, expectCommit)
		return nil, "", nil, &SignedTagVerificationError{
			Path:              r.configPath,
			Tag:               tagName,
			Reason:            SignedTagFailureExpectCommitMismatch,
			Message:           message,
			ResolvedCommit:    resolvedCommit,
			SignerIdentity:    verification.SignerIdentity,
			SignerFingerprint: verification.SignerFingerprint,
			Diagnostics: []ResolutionDiagnostic{{
				Severity: DiagnosticSeverityError,
				Code:     string(SignedTagFailureExpectCommitMismatch),
				Message:  message,
			}},
		}
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", resolvedCommit); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	tree, finalizedCommit, err := r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
	if err != nil {
		return nil, "", nil, err
	}
	if finalizedCommit != resolvedCommit {
		_ = tree.Close()
		return nil, "", nil, &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git commit %q did not match verified signed-tag commit %q", finalizedCommit, resolvedCommit)}
	}
	return tree, resolvedCommit, verification, nil
}

func (r gitResolver) materializeCommitToTree(commit, subdir string) (*LocalSourceTree, string, error) {
	tempRoot, repoRoot, err := r.initializeRepository()
	if err != nil {
		return nil, "", err
	}
	if err := runGit("-C", repoRoot, "fetch", "--quiet", "--no-tags", "origin", "+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if err := runGit("-C", repoRoot, "cat-file", "-e", commit+"^{commit}"); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("pinned git commit %q was not found after fetching advertised refs: %v", commit, err)}
	}
	if err := runGit("-C", repoRoot, "checkout", "--quiet", "--detach", commit); err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	return r.finalizeMaterializedTree(tempRoot, repoRoot, subdir)
}

func (r gitResolver) initializeRepository() (string, string, error) {
	tempRoot, err := os.MkdirTemp("", "runectx-git-source-")
	if err != nil {
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	repoRoot := filepath.Join(tempRoot, "repo")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		_ = os.RemoveAll(tempRoot)
		return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	for _, args := range [][]string{{"init", "--quiet", repoRoot}, {"-C", repoRoot, "remote", "add", "origin", r.url}} {
		if err := runGit(args...); err != nil {
			_ = os.RemoveAll(tempRoot)
			return "", "", &ValidationError{Path: r.configPath, Message: err.Error()}
		}
	}
	return tempRoot, repoRoot, nil
}

func (r gitResolver) finalizeMaterializedTree(tempRoot, repoRoot, subdir string) (*LocalSourceTree, string, error) {
	commitOutput, err := gitOutput("-C", repoRoot, "rev-parse", "HEAD")
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	commit := strings.TrimSpace(commitOutput)
	materializedRoot := filepath.Clean(filepath.Join(repoRoot, filepath.FromSlash(subdir)))
	if !isWithinRoot(repoRoot, materializedRoot) {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("git source subdir %q escapes the fetched repository root", subdir)}
	}
	info, err := os.Stat(materializedRoot)
	if err != nil {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: err.Error()}
	}
	if !info.IsDir() {
		_ = os.RemoveAll(tempRoot)
		return nil, "", &ValidationError{Path: r.configPath, Message: fmt.Sprintf("resolved git source root %q is not a directory", materializedRoot)}
	}
	return &LocalSourceTree{Root: materializedRoot, SnapshotKind: "git_checkout", cleanupRoot: tempRoot}, commit, nil
}

func runGit(args ...string) error {
	result := runGitCaptured(args, gitExecutable)
	if result.TimedOut {
		return fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil {
		message := sanitizeGitMessage(strings.TrimSpace(result.Output))
		if message == "" {
			message = sanitizeGitMessage(result.Err.Error())
		}
		return fmt.Errorf("git %s: %s", sanitizeGitArgs(result.Args), message)
	}
	return nil
}

func gitOutput(args ...string) (string, error) {
	result := runGitCaptured(args, gitExecutable)
	if result.TimedOut {
		return "", fmt.Errorf("git %s: command timed out after %s", sanitizeGitArgs(result.Args), gitCommandTimeout)
	}
	if result.Err != nil {
		message := sanitizeGitMessage(strings.TrimSpace(result.Output))
		if message == "" {
			message = sanitizeGitMessage(result.Err.Error())
		}
		return "", fmt.Errorf("git %s: %s", sanitizeGitArgs(result.Args), message)
	}
	return result.Output, nil
}

func runGitCaptured(args []string, executable string) gitCommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	output, err := gitCommandRunner(ctx, executable, args, sanitizedGitEnv())
	result := gitCommandResult{
		Args:   append([]string(nil), args...),
		Output: string(output),
		Err:    err,
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		return result
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		return result
	}
	return result
}

func runGitCommandContext(ctx context.Context, executable string, args []string, env []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func parseTrustedSSHVerifyTagOutput(output string) (string, string, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		const prefix = `Good "git" signature for `
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		remainder := strings.TrimPrefix(line, prefix)
		fingerprintIndex := strings.LastIndex(remainder, " SHA256:")
		if fingerprintIndex <= 0 {
			continue
		}
		identitySection := strings.TrimSpace(remainder[:fingerprintIndex])
		withIndex := strings.LastIndex(identitySection, " with ")
		if withIndex <= 0 {
			continue
		}
		identity := strings.TrimSpace(identitySection[:withIndex])
		fingerprint := strings.TrimSpace(remainder[fingerprintIndex+1:])
		if identity == "" || fingerprint == "" || !strings.HasPrefix(fingerprint, "SHA256:") {
			continue
		}
		return identity, fingerprint, nil
	}
	return "", "", fmt.Errorf("missing trusted signer identity/fingerprint in verification output")
}

func classifySignedTagFailure(output string) SignedTagFailureReason {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "no signature found"):
		return SignedTagFailureUnsignedTag
	case strings.Contains(lower, "could not verify signature") || strings.Contains(lower, "couldn't verify signature") || strings.Contains(lower, "bad signature") || strings.Contains(lower, "invalid format"):
		return SignedTagFailureInvalidSignature
	case strings.Contains(lower, "no principal matched"):
		return SignedTagFailureUntrustedSigner
	default:
		return SignedTagFailureVerificationFailed
	}
}

func signedTagFailureMessage(tagName string, reason SignedTagFailureReason, output string) string {
	sanitizedOutput := sanitizeGitMessage(strings.TrimSpace(output))
	switch reason {
	case SignedTagFailureUnsignedTag:
		return fmt.Sprintf("signed tag %q is unsigned", tagName)
	case SignedTagFailureInvalidSignature:
		if sanitizedOutput != "" {
			return fmt.Sprintf("signed tag %q has an invalid signature: %s", tagName, sanitizedOutput)
		}
		return fmt.Sprintf("signed tag %q has an invalid signature", tagName)
	case SignedTagFailureUntrustedSigner:
		if sanitizedOutput != "" {
			return fmt.Sprintf("signed tag %q was signed by an untrusted signer: %s", tagName, sanitizedOutput)
		}
		return fmt.Sprintf("signed tag %q was signed by an untrusted signer", tagName)
	default:
		if sanitizedOutput != "" {
			return fmt.Sprintf("signed tag %q verification failed: %s", tagName, sanitizedOutput)
		}
		return fmt.Sprintf("signed tag %q verification failed", tagName)
	}
}

func sanitizedGitEnv() []string {
	env := []string{
		"HOME=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + os.TempDir(),
		"GNUPGHOME=" + os.TempDir(),
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_ALLOW_PROTOCOL=" + strings.Join(gitAllowedProtocols, ":"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"SSH_AUTH_SOCK=",
		"GIT_SSH=",
		"GIT_SSH_COMMAND=",
		"GCM_INTERACTIVE=Never",
		"LANG=C",
		"LC_ALL=C",
	}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok && value != "" {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func runeContextRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func canonicalizePaths(root, target string) (string, string, error) {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve root %q: %w", root, err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", "", fmt.Errorf("resolve path %q: %w", target, err)
	}
	return resolvedRoot, resolvedTarget, nil
}

func validateOpenPathWithinRoot(resolvedPath, root string) error {
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if !isWithinRoot(resolvedRoot, resolvedPath) {
		return fmt.Errorf("resolved path %q escapes declared local source tree", resolvedPath)
	}
	return nil
}

func sanitizeGitArgs(args []string) string {
	sanitized := make([]string, len(args))
	for i, arg := range args {
		sanitized[i] = sanitizeGitMessage(arg)
	}
	return strings.Join(sanitized, " ")
}

func sanitizeGitMessage(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return trimmed
	}
	for _, prefix := range []string{"sha256:", "SHA256:"} {
		for {
			idx := strings.Index(trimmed, prefix)
			if idx < 0 {
				break
			}
			end := idx + len(prefix)
			for end < len(trimmed) && !strings.ContainsRune(" \n\r\t'\"", rune(trimmed[end])) {
				end++
			}
			trimmed = trimmed[:idx] + "<redacted-fingerprint>" + trimmed[end:]
		}
	}
	for _, prefix := range []string{"https://", "http://", "ssh://", "file://"} {
		for {
			idx := strings.Index(strings.ToLower(trimmed), prefix)
			if idx < 0 {
				break
			}
			end := idx + len(prefix)
			for end < len(trimmed) && !strings.ContainsRune(" \n\r\t'\"", rune(trimmed[end])) {
				end++
			}
			trimmed = trimmed[:idx] + "<redacted-url>" + trimmed[end:]
		}
	}
	searchFrom := 0
	for {
		rel := strings.Index(trimmed[searchFrom:], "@")
		if rel < 0 {
			break
		}
		at := searchFrom + rel
		if at <= 0 {
			searchFrom = at + 1
			continue
		}
		if at+1 < len(trimmed) && trimmed[at+1] == '{' {
			searchFrom = at + 1
			continue
		}
		start := strings.LastIndexAny(trimmed[:at], " /\n\r\t\"")
		if start < 0 {
			start = 0
		} else {
			start++
		}
		end := at + 1
		for end < len(trimmed) && !strings.ContainsRune(" /\n\r\t\"'", rune(trimmed[end])) {
			end++
		}
		trimmed = trimmed[:start] + "<redacted-identity>" + trimmed[end:]
		searchFrom = start + len("<redacted-identity>")
	}
	return trimmed
}

func walkRuneContextFiles(root string, visit func(path string, d fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return visit(path, d)
	})
}
