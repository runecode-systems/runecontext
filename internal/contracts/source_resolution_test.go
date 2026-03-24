package contracts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSourceResolutionEmbeddedGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "source-resolution", "embedded-project")

	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected embedded fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "embedded.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
	})
	if index.Resolution.MaterializedRoot() != filepath.Join(projectRoot, "runecontext") {
		t.Fatalf("expected embedded source to materialize from live tree")
	}
}

func TestSourceResolutionPathLocalAndRemoteCI(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "source-resolution", "path-project")

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected local path fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "path-local.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
	})
	if index.Resolution.Tree == nil || index.Resolution.Tree.SnapshotKind != "snapshot_copy" {
		t.Fatalf("expected path mode to use a snapshot-friendly local tree")
	}

	_, err = v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeRemoteCI,
	})
	if err == nil || !strings.Contains(err.Error(), "source.type=path is invalid in execution mode remote_ci") {
		t.Fatalf("expected remote/ci path resolution to fail, got %v", err)
	}
}

func TestSourceResolutionMonorepoNearestAncestor(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	monorepoRoot := fixturePath(t, "source-resolution", "monorepo")

	nestedStart := filepath.Join(monorepoRoot, "packages", "service", "internal")
	nestedIndex, err := v.ValidateProjectWithOptions(nestedStart, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryNearestAncestor,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected nested monorepo fixture to validate: %v", err)
	}
	defer nestedIndex.Close()
	assertResolutionMatchesGolden(t, nestedIndex.Resolution, fixturePath(t, "source-resolution", "golden", "monorepo-nested.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(monorepoRoot),
	})

	rootStart := filepath.Join(monorepoRoot, "packages", "worker")
	rootIndex, err := v.ValidateProjectWithOptions(rootStart, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryNearestAncestor,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected root monorepo fixture to validate: %v", err)
	}
	defer rootIndex.Close()
	assertResolutionMatchesGolden(t, rootIndex.Resolution, fixturePath(t, "source-resolution", "golden", "monorepo-root.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(monorepoRoot),
	})
}

func TestSourceResolutionGitPinnedCommitGolden(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  commit: %s\n  subdir: runecontext\n", repoDir, commit))

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected git pinned fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "git-pinned.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
		"${COMMIT}":       commit,
	})
	if index.Resolution.Tree == nil || index.Resolution.Tree.SnapshotKind != "git_checkout" {
		t.Fatalf("expected git source to materialize via checkout")
	}
}

func TestSourceResolutionGitMutableRefRequiresOptInAndWarns(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  ref: main\n  allow_mutable_ref: true\n  subdir: runecontext\n", repoDir))

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		t.Fatalf("expected mutable-ref fixture to validate: %v", err)
	}
	defer index.Close()
	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "git-mutable-ref.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
		"${COMMIT}":       commit,
	})

	rejectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  ref: main\n  subdir: runecontext\n", repoDir))
	_, err = v.LoadProject(rejectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err == nil || !strings.Contains(err.Error(), "allow_mutable_ref") {
		t.Fatalf("expected missing mutable-ref opt-in to fail, got %v", err)
	}
}

func TestSourceResolutionGitSignedTagTrustedSignerSuccess(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust: GitTrustInputs{
			SignedTagVerifier: verifier,
		},
	})
	if err != nil {
		t.Fatalf("expected signed-tag fixture to validate: %v", err)
	}
	defer index.Close()

	assertResolutionMatchesGolden(t, index.Resolution, fixturePath(t, "source-resolution", "golden", "git-signed-tag.yaml"), map[string]string{
		"${PROJECT_ROOT}": filepath.ToSlash(projectRoot),
		"${TAG}":          details.SignedTagName,
		"${COMMIT}":       details.Commit,
		"${SIGNER}":       details.SignerIdentity,
		"${FINGERPRINT}":  details.SignerFingerprint,
	})
	if index.Resolution.Tree == nil || index.Resolution.Tree.SnapshotKind != "git_checkout" {
		t.Fatalf("expected signed git source to materialize via checkout")
	}
}

func TestSourceResolutionGitSignedTagFailsWithoutExplicitTrustInputs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	var verificationErr *SignedTagVerificationError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected signed-tag verification error, got %v", err)
	}
	if verificationErr.Reason != SignedTagFailureMissingTrust {
		t.Fatalf("expected missing trust failure, got %q", verificationErr.Reason)
	}
	if !strings.Contains(verificationErr.Message, "explicit trusted signer inputs") {
		t.Fatalf("expected missing trust message, got %q", verificationErr.Message)
	}
}

func TestSourceResolutionGitSignedTagUntrustedSignerFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.UntrustedAllowedSigners)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: verifier},
	})
	assertSignedTagFailure(t, err, SignedTagFailureUntrustedSigner, "untrusted signer")
}

func TestSourceResolutionGitSignedTagUnsignedFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.UnsignedTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: verifier},
	})
	assertSignedTagFailure(t, err, SignedTagFailureUnsignedTag, "unsigned")
}

func TestSourceResolutionGitSignedTagBadSignatureFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.BadSignatureTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: verifier},
	})
	assertSignedTagFailure(t, err, SignedTagFailureInvalidSignature, "invalid signature")
}

func TestSourceResolutionGitSignedTagEmptyExpectCommitFailsClearly(t *testing.T) {
	repoDir, details := createSignedGitSourceRepo(t)
	_, err := resolveGitSource(&SourceResolution{}, "runecontext.yaml", map[string]any{
		"url":           repoDir,
		"signed_tag":    details.SignedTagName,
		"expect_commit": "",
		"subdir":        "runecontext",
	}, GitTrustInputs{
		SignedTagVerifier: newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners),
	})
	if err == nil || !strings.Contains(err.Error(), "git expect_commit must not be empty") {
		t.Fatalf("expected explicit empty expect_commit failure, got %v", err)
	}
}

func TestSourceResolutionGitSignedTagExpectCommitMismatchFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)
	mismatchedCommit := strings.Repeat("a", 40)
	if mismatchedCommit == details.Commit {
		mismatchedCommit = strings.Repeat("b", 40)
	}
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, mismatchedCommit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: verifier},
	})
	var verificationErr *SignedTagVerificationError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected signed-tag verification error, got %v", err)
	}
	if verificationErr.Reason != SignedTagFailureExpectCommitMismatch {
		t.Fatalf("expected expect_commit mismatch failure, got %q", verificationErr.Reason)
	}
	if verificationErr.ResolvedCommit != details.Commit {
		t.Fatalf("expected resolved commit capture %q, got %q", details.Commit, verificationErr.ResolvedCommit)
	}
	if verificationErr.SignerIdentity != details.SignerIdentity {
		t.Fatalf("expected signer identity capture %q, got %q", details.SignerIdentity, verificationErr.SignerIdentity)
	}
	if verificationErr.SignerFingerprint != details.SignerFingerprint {
		t.Fatalf("expected signer fingerprint capture %q, got %q", details.SignerFingerprint, verificationErr.SignerFingerprint)
	}
	if !strings.Contains(verificationErr.Message, "expect_commit") {
		t.Fatalf("expected expect_commit mismatch message, got %q", verificationErr.Message)
	}
}

func TestSSHAllowedSignersVerifierRejectsEmptyTrustMaterial(t *testing.T) {
	if _, err := NewSSHAllowedSignersVerifier(nil); err == nil {
		t.Fatal("expected empty trust material to fail")
	}
}

func TestSourceResolutionGitSignedTagVerifierReturningNilFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: signedTagVerifierFunc(func(string, string) (*SignedTagVerification, error) { return nil, nil })},
	})
	assertSignedTagFailure(t, err, SignedTagFailureVerificationFailed, "no verification details")
}

func TestSourceResolutionGitSignedTagVerifierReturningIncompleteSignerFailsClosed(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	_, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust: GitTrustInputs{SignedTagVerifier: signedTagVerifierFunc(func(string, string) (*SignedTagVerification, error) {
			return &SignedTagVerification{SignerIdentity: "alice@example.com"}, nil
		})},
	})
	assertSignedTagFailure(t, err, SignedTagFailureVerificationFailed, "incomplete signer details")
}

func TestSourceResolutionRejectsEmbeddedPathEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	outside := filepath.Join(filepath.Dir(projectRoot), "outside-runecontext")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside dir: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: ../outside-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "embedded source path") {
		t.Fatalf("expected embedded escape to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsUnsafeGitInputs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	for _, tc := range unsafeGitInputScenarios(repoDir, commit) {
		t.Run(tc.name, func(t *testing.T) {
			projectRoot := writeRootConfigProject(t, tc.config)
			_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %s to fail, got %v", tc.name, err)
			}
		})
	}
}

func writeRootConfigProject(t *testing.T, config string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	return root
}

type signedGitSourceDetails struct {
	Commit                  string
	SignedTagName           string
	UnsignedTagName         string
	BadSignatureTagName     string
	SignerIdentity          string
	SignerFingerprint       string
	AllowedSigners          []byte
	UntrustedAllowedSigners []byte
}

type signedTagVerifierFunc func(repoRoot, tagName string) (*SignedTagVerification, error)

func (f signedTagVerifierFunc) VerifySignedTag(repoRoot, tagName string) (*SignedTagVerification, error) {
	return f(repoRoot, tagName)
}

func createGitSourceRepo(t *testing.T) (string, string) {
	t.Helper()
	repoDir := t.TempDir()
	runGitTest(t, repoDir, "init", "--initial-branch=main")
	templateRoot := fixturePath(t, "source-resolution", "templates", "minimal-runecontext")
	copyDirForTest(t, templateRoot, filepath.Join(repoDir, "runecontext"))
	runGitTest(t, repoDir, "add", ".")
	runGitTest(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	commit := strings.TrimSpace(gitOutputForTest(t, repoDir, "rev-parse", "HEAD"))
	return repoDir, commit
}

func createSignedGitSourceRepo(t *testing.T) (string, signedGitSourceDetails) {
	t.Helper()
	requireToolForContractsTests(t, "git")
	requireToolForContractsTests(t, "ssh-keygen")
	repoDir, commit := createGitSourceRepo(t)
	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "signer")
	runCommandForTest(t, repoDir, sanitizedGitEnv(), "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", keyPath)
	publicKey := strings.TrimSpace(string(readFixture(t, keyPath+".pub")))
	allowedSigners := []byte(fmt.Sprintf("alice@example.com %s\n", publicKey))
	untrustedKeyPath := filepath.Join(keyDir, "untrusted-signer")
	runCommandForTest(t, repoDir, sanitizedGitEnv(), "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", untrustedKeyPath)
	untrustedPublicKey := strings.TrimSpace(string(readFixture(t, untrustedKeyPath+".pub")))
	untrustedAllowedSigners := []byte(fmt.Sprintf("bob@example.com %s\n", untrustedPublicKey))
	signedTagName := "v1.0.0-signed"
	unsignedTagName := "v1.0.0-unsigned"
	badSignatureTagName := "v1.0.0-bad-signature"
	runGitTest(t, repoDir, "-c", "gpg.format=ssh", "-c", "user.signingkey="+keyPath, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "tag", "-s", "-m", "signed tag", signedTagName)
	runGitTest(t, repoDir, "tag", unsignedTagName)
	corruptSignedTagForTest(t, repoDir, signedTagName, badSignatureTagName)
	verifier := newSSHAllowedSignersVerifierForTest(t, allowedSigners)
	verification, err := verifier.VerifySignedTag(repoDir, signedTagName)
	if err != nil {
		t.Fatalf("verify signed test tag: %v", err)
	}
	return repoDir, signedGitSourceDetails{
		Commit:                  commit,
		SignedTagName:           signedTagName,
		UnsignedTagName:         unsignedTagName,
		BadSignatureTagName:     badSignatureTagName,
		SignerIdentity:          verification.SignerIdentity,
		SignerFingerprint:       verification.SignerFingerprint,
		AllowedSigners:          allowedSigners,
		UntrustedAllowedSigners: untrustedAllowedSigners,
	}
}

func requireToolForContractsTests(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

func newSSHAllowedSignersVerifierForTest(t *testing.T, allowedSigners []byte) *SSHAllowedSignersVerifier {
	t.Helper()
	verifier, err := NewSSHAllowedSignersVerifier(allowedSigners)
	if err != nil {
		t.Fatalf("create allowed-signers verifier: %v", err)
	}
	return verifier
}

func corruptSignedTagForTest(t *testing.T, repoDir, sourceTag, targetTag string) {
	t.Helper()
	tagText := gitOutputForTest(t, repoDir, "cat-file", "tag", sourceTag)
	lines := strings.Split(strings.TrimSuffix(tagText, "\n"), "\n")
	corrupted := false
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i-1], "-----BEGIN SSH SIGNATURE-----") && lines[i] != "" && !strings.HasPrefix(lines[i], "-----") {
			if lines[i][0] == 'A' {
				lines[i] = "B" + lines[i][1:]
			} else {
				lines[i] = "A" + lines[i][1:]
			}
			corrupted = true
			break
		}
	}
	if !corrupted {
		t.Fatal("failed to locate SSH signature payload to corrupt")
	}
	obj := runCommandOutputForTest(t, repoDir, sanitizedGitEnv(), strings.NewReader(strings.Join(lines, "\n")+"\n"), "git", "-C", repoDir, "hash-object", "-t", "tag", "-w", "--stdin")
	runGitTest(t, repoDir, "update-ref", "refs/tags/"+targetTag, strings.TrimSpace(obj))
}

func assertSignedTagFailure(t *testing.T, err error, reason SignedTagFailureReason, contains string) {
	t.Helper()
	var verificationErr *SignedTagVerificationError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected signed-tag verification error, got %v", err)
	}
	if verificationErr.Reason != reason {
		t.Fatalf("expected signed-tag failure reason %q, got %q", reason, verificationErr.Reason)
	}
	if contains != "" && !strings.Contains(strings.ToLower(verificationErr.Message), strings.ToLower(contains)) {
		t.Fatalf("expected signed-tag failure message to contain %q, got %q", contains, verificationErr.Message)
	}
	if len(verificationErr.Diagnostics) == 0 {
		t.Fatal("expected signed-tag failure diagnostics")
	}
}

func TestSourceResolutionGitPinnedCommitWorksFromAdvertisedRefs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  commit: %s\n  subdir: runecontext\n", repoDir, commit))

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected pinned commit to resolve from advertised refs: %v", err)
	}
	defer loaded.Close()
	if loaded.Resolution == nil || loaded.Resolution.ResolvedCommit != commit {
		t.Fatalf("expected resolved commit %q, got %#v", commit, loaded.Resolution)
	}
}

func copyDirForTest(t *testing.T, srcRoot, dstRoot string) {
	t.Helper()
	if err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstRoot, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture directory: %v", err)
	}
}

func runGitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	runCommandForTest(t, dir, sanitizedGitEnv(), "git", args...)
}

func gitOutputForTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return runCommandOutputForTest(t, dir, sanitizedGitEnv(), nil, "git", args...)
}

func runCommandForTest(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	_ = runCommandOutputForTest(t, dir, env, nil, name, args...)
}

func runCommandOutputForTest(t *testing.T, dir string, env []string, stdin io.Reader, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = stdin
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

func tryCreateSymlink(target, path string) error {
	if err := os.Symlink(target, path); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			return fmt.Errorf("symlink tests skipped: %w", err)
		}
		return fmt.Errorf("create symlink: %w", err)
	}
	return nil
}
