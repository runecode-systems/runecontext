package contracts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type unsafeGitInputScenario struct {
	name   string
	config string
	want   string
}

func unsafeGitInputScenarios(repoDir, commit string) []unsafeGitInputScenario {
	return []unsafeGitInputScenario{
		{name: "url starts with dash", config: unsafeGitConfig("-bad-url", "commit: "+commit, "subdir: runecontext"), want: "git source url"},
		{name: "url uses remote helper", config: unsafeGitConfig("ext::helper", "commit: "+commit, "subdir: runecontext"), want: "remote-helper"},
		{name: "ref starts with dash", config: unsafeGitConfig(repoDir, "ref: -main", "allow_mutable_ref: true", "subdir: runecontext"), want: "git ref"},
		{name: "ref contains dot dot", config: unsafeGitConfig(repoDir, "ref: feature..branch", "allow_mutable_ref: true", "subdir: runecontext"), want: "must not contain '..'"},
		{name: "ref ends with slash", config: unsafeGitConfig(repoDir, "ref: feature/", "allow_mutable_ref: true", "subdir: runecontext"), want: "start or end with '/'"},
		{name: "subdir escapes repo", config: unsafeGitConfig(repoDir, "commit: "+commit, "subdir: ../outside"), want: "git subdir"},
	}
}

func unsafeGitConfig(url string, lines ...string) string {
	parts := []string{
		"schema_version: 1",
		"runecontext_version: 0.1.0-alpha.3",
		"assurance_tier: plain",
		"source:",
		"  type: git",
		"  url: " + url,
	}
	for _, line := range lines {
		parts = append(parts, "  "+line)
	}
	return strings.Join(parts, "\n") + "\n"
}

func TestSourceResolutionRejectsPathSymlinkEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	outside := filepath.Join(projectRoot, "outside.txt")
	if err := os.MkdirAll(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution"), 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := tryCreateSymlink("../outside.txt", filepath.Join(localRoot, "escape-link")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "escapes declared local source tree") {
		t.Fatalf("expected path symlink escape to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsPathSymlinkCycle(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	if err := os.MkdirAll(localRoot, 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	if err := tryCreateSymlink(".", filepath.Join(localRoot, "loop")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "symlink cycle detected") {
		t.Fatalf("expected path symlink cycle to fail, got %v", err)
	}
}

func TestSourceResolutionRejectsEmbeddedPathSymlinkEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	outsideRoot := filepath.Join(filepath.Dir(projectRoot), "outside-runecontext")
	if err := os.MkdirAll(outsideRoot, 0o755); err != nil {
		t.Fatalf("mkdir outside root: %v", err)
	}
	if err := tryCreateSymlink(filepath.Join("..", filepath.Base(outsideRoot)), filepath.Join(projectRoot, "linked-runecontext")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: linked-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "escapes the selected project root") {
		t.Fatalf("expected embedded symlink escape to fail, got %v", err)
	}
}

func TestSanitizedGitEnvSetsProtocolAndConfigGuards(t *testing.T) {
	env := sanitizedGitEnv()
	joined := strings.Join(env, "\n")
	for _, expected := range []string{
		"GIT_ALLOW_PROTOCOL=file:git:http:https:ssh",
		"GIT_CONFIG_NOSYSTEM=1",
		"GNUPGHOME=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + os.TempDir(),
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected sanitized git env to contain %q, got %s", expected, joined)
		}
	}
}

func TestSanitizeGitMessageRedactsURLsAndCredentials(t *testing.T) {
	message := "fatal: could not fetch https://user:token@example.com/private/repo and contact admin@example.org about SHA256:abcdef"
	sanitized := sanitizeGitMessage(message)
	if strings.Contains(sanitized, "token") || strings.Contains(sanitized, "example.com/private/repo") || strings.Contains(sanitized, "admin@example.org") || strings.Contains(sanitized, "SHA256:abcdef") {
		t.Fatalf("expected sanitized git message to redact secrets, got %q", sanitized)
	}
	if !strings.Contains(sanitized, "<redacted-url>") {
		t.Fatalf("expected sanitized git message to contain redacted marker, got %q", sanitized)
	}
	if !strings.Contains(sanitized, "<redacted-fingerprint>") {
		t.Fatalf("expected sanitized git message to redact fingerprint, got %q", sanitized)
	}
	if !strings.Contains(sanitized, "<redacted-identity>") {
		t.Fatalf("expected sanitized git message to redact identity, got %q", sanitized)
	}
}

func TestSanitizeGitMessagePreservesReflogSyntax(t *testing.T) {
	for _, message := range []string{
		"fatal: ambiguous argument HEAD@{upstream}: unknown revision",
		"fatal: ambiguous argument HEAD@{0}: unknown revision",
		"fatal: ambiguous argument refs/heads/main@{yesterday}: unknown revision",
	} {
		sanitized := sanitizeGitMessage(message)
		if !strings.Contains(sanitized, "@{") {
			t.Fatalf("expected reflog syntax to be preserved, got %q", sanitized)
		}
		if strings.Contains(sanitized, "<redacted-identity>") {
			t.Fatalf("expected reflog syntax to avoid identity redaction, got %q", sanitized)
		}
	}
}

func TestSSHAllowedSignersVerifierSurfacesGitExecutionFailures(t *testing.T) {
	verifier, err := NewSSHAllowedSignersVerifierWithGitExecutable([]byte("alice@example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE5XQmFkRHVtbXlLZXlNYXRlcmlhbEZvclRlc3Rz\n"), filepath.Join(t.TempDir(), "missing-git"))
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	_, err = verifier.VerifySignedTag(t.TempDir(), "v1.0.0")
	var verificationErr *SignedTagVerificationError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected signed-tag verification error, got %v", err)
	}
	if verificationErr.Reason != SignedTagFailureVerificationFailed {
		t.Fatalf("expected verification_failed reason, got %q", verificationErr.Reason)
	}
	if verificationErr.Message == "" {
		t.Fatal("expected non-empty execution failure detail")
	}
	if len(verificationErr.Diagnostics) == 0 {
		t.Fatal("expected execution failure diagnostic")
	}
	if verificationErr.Diagnostics[0].Code != string(SignedTagFailureVerificationFailed) {
		t.Fatalf("expected verification_failed diagnostic code, got %q", verificationErr.Diagnostics[0].Code)
	}
}

func TestSSHAllowedSignersVerifierSurfacesGitTimeoutsAsStructuredFailures(t *testing.T) {
	verifier, err := NewSSHAllowedSignersVerifier([]byte("alice@example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE5XQmFkRHVtbXlLZXlNYXRlcmlhbEZvclRlc3Rz\n"))
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	originalTimeout := gitCommandTimeout
	originalRunner := gitCommandRunner
	gitCommandTimeout = 10 * time.Millisecond
	t.Cleanup(func() {
		gitCommandTimeout = originalTimeout
		gitCommandRunner = originalRunner
	})
	gitCommandRunner = func(ctx context.Context, executable string, args []string, env []string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	_, err = verifier.VerifySignedTag(t.TempDir(), "v1.0.0")
	var verificationErr *SignedTagVerificationError
	if !errors.As(err, &verificationErr) {
		t.Fatalf("expected signed-tag verification error, got %v", err)
	}
	if verificationErr.Reason != SignedTagFailureVerificationFailed {
		t.Fatalf("expected verification_failed reason, got %q", verificationErr.Reason)
	}
	if !strings.Contains(verificationErr.Message, "timed out") {
		t.Fatalf("expected timeout detail, got %q", verificationErr.Message)
	}
}

func TestParseTrustedSSHVerifyTagOutputAcceptsIdentityContainingWith(t *testing.T) {
	identity, fingerprint, err := parseTrustedSSHVerifyTagOutput(`Good "git" signature for Team with Ops <ops@example.com> with ED25519 key SHA256:abc123`)
	if err != nil {
		t.Fatalf("expected parser success, got %v", err)
	}
	if identity != "Team with Ops <ops@example.com>" {
		t.Fatalf("expected identity capture, got %q", identity)
	}
	if fingerprint != "SHA256:abc123" {
		t.Fatalf("expected fingerprint capture, got %q", fingerprint)
	}
}

func TestParseTrustedSSHVerifyTagOutputRejectsUnexpectedFingerprintPrefix(t *testing.T) {
	_, _, err := parseTrustedSSHVerifyTagOutput(`Good "git" signature for Team Ops <ops@example.com> with ED25519 key MD5:abc123`)
	if err == nil {
		t.Fatal("expected parser to reject non-SHA256 fingerprint")
	}
}

func TestSanitizeGitMessageRedactsMultipleIdentityTokensAndSCPLikeHost(t *testing.T) {
	message := "fetch failed for git@github.com:runecode-systems/runecontext and admin@example.org and trailing user@"
	sanitized := sanitizeGitMessage(message)
	if strings.Contains(sanitized, "git@github.com") || strings.Contains(sanitized, "admin@example.org") || strings.Contains(sanitized, "user@") {
		t.Fatalf("expected identities to be redacted, got %q", sanitized)
	}
	if strings.Count(sanitized, "<redacted-identity>") < 2 {
		t.Fatalf("expected multiple identity redactions, got %q", sanitized)
	}
}

func TestSourceResolutionSkipsDotGitDirectoryInSnapshots(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	gitDir := filepath.Join(localRoot, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("mkdir .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write fake git head: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution"), 0o755); err != nil {
		t.Fatalf("mkdir changes dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localRoot, "changes", "CHG-2026-001-a3f2-source-resolution", "status.yaml"), []byte("schema_version: 1\nid: CHG-2026-001-a3f2-source-resolution\ntitle: Test snapshot exclusions\nstatus: proposed\ntype: feature\nsize: small\nverification_status: pending\ncontext_bundles: []\nrelated_specs: []\nrelated_decisions: []\nrelated_changes: []\ndepends_on: []\ninformed_by: []\nsupersedes: []\nsuperseded_by: []\ncreated_at: \"2026-03-17\"\nclosed_at: null\npromotion_assessment:\n  status: pending\n  suggested_targets: []\n"), 0o644); err != nil {
		t.Fatalf("write status file: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected path source with .git directory to resolve: %v", err)
	}
	defer loaded.Close()
	if _, err := os.Stat(filepath.Join(loaded.Resolution.MaterializedRoot(), ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot to exclude .git directory, got err=%v", err)
	}
}

func TestSourceResolutionRejectsOversizedPathSnapshot(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	localRoot := filepath.Join(projectRoot, "local-runecontext")
	if err := os.MkdirAll(localRoot, 0o755); err != nil {
		t.Fatalf("mkdir local root: %v", err)
	}
	data := strings.Repeat("a", int(localSnapshotLimits.MaxBytes)+1)
	if err := os.WriteFile(filepath.Join(localRoot, "large.txt"), []byte(data), 0o644); err != nil {
		t.Fatalf("write oversized file: %v", err)
	}
	rootConfig := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: path\n  path: local-runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(rootConfig), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	_, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err == nil || !strings.Contains(err.Error(), "maximum snapshot size") {
		t.Fatalf("expected oversized snapshot to fail, got %v", err)
	}
}
