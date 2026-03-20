package contracts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

type ResolutionDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
}

type ValidationDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
	Path     string             `json:"path,omitempty" yaml:"path,omitempty"`
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

func runGitCommandContext(ctx context.Context, executable string, args []string, env []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}
