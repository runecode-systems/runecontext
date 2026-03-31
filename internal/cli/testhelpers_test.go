package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

var runecontextVersionTestMu sync.Mutex

func repoRootForTests() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			return "", os.ErrNotExist
		}
		wd = next
	}
}

func releaseMetadataVersionForTests(t *testing.T) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	version, err := ReadReleaseMetadataVersion(root)
	if err != nil {
		t.Fatalf("read release metadata version: %v", err)
	}
	return strings.TrimPrefix(version, "v")
}

func withReleaseMetadataVersionForTests(t *testing.T, fn func()) {
	t.Helper()
	setRunecontextVersionForTests(t, "v"+releaseMetadataVersionForTests(t))
	fn()
}

func setRunecontextVersionForTests(t *testing.T, value string) {
	t.Helper()
	runecontextVersionTestMu.Lock()
	original := runecontextVersion
	runecontextVersion = value
	t.Cleanup(func() {
		runecontextVersion = original
		runecontextVersionTestMu.Unlock()
	})
}

type signedGitSourceDetails struct {
	commit            string
	signedTagName     string
	signerIdentity    string
	signerFingerprint string
	allowedSigners    []byte
}

func fixtureRoot(t *testing.T, elems ...string) string {
	t.Helper()
	if len(elems) == 1 {
		switch elems[0] {
		case "valid-project", "reject-bundle-invalid", "reject-change-missing-related-spec", "reject-proposal-invalid":
			return repoFixtureRoot(t, "traceability", elems[0])
		}
	}
	return repoFixtureRoot(t, elems...)
}

func repoFixtureRoot(t *testing.T, elems ...string) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	parts := append([]string{root, "fixtures"}, elems...)
	return filepath.Join(parts...)
}

func copyDirForCLI(t *testing.T, srcRoot, dstRoot string) {
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

func requireToolForCLITests(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

func runGitForCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	runCommandForCLI(t, dir, sanitizedGitEnv(), "git", args...)
}

func gitOutputForCLI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return runCommandOutputForCLI(t, dir, sanitizedGitEnv(), nil, "git", args...)
}

func runCommandForCLI(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	_ = runCommandOutputForCLI(t, dir, env, nil, name, args...)
}

func runCommandOutputForCLI(t *testing.T, dir string, env []string, stdin io.Reader, name string, args ...string) string {
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

func sanitizedGitEnv() []string {
	env := []string{
		"HOME=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + os.TempDir(),
		"GNUPGHOME=" + os.TempDir(),
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_ALLOW_PROTOCOL=file:git:http:https:ssh",
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

func createSignedGitSourceRepoForCLI(t *testing.T) (string, signedGitSourceDetails) {
	t.Helper()
	requireToolForCLITests(t, "git")
	requireToolForCLITests(t, "ssh-keygen")
	repoDir := t.TempDir()
	runGitForCLI(t, repoDir, "init", "--initial-branch=main")
	copyDirForCLI(t, repoFixtureRoot(t, "source-resolution", "templates", "minimal-runecontext"), filepath.Join(repoDir, "runecontext"))
	runGitForCLI(t, repoDir, "add", ".")
	runGitForCLI(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	commit := strings.TrimSpace(gitOutputForCLI(t, repoDir, "rev-parse", "HEAD"))

	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "signer")
	runCommandForCLI(t, repoDir, sanitizedGitEnv(), "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", keyPath)
	publicKeyBytes, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		t.Fatalf("read signer public key: %v", err)
	}
	publicKey := strings.TrimSpace(string(publicKeyBytes))
	allowedSigners := []byte(fmt.Sprintf("alice@example.com %s\n", publicKey))
	signedTagName := "v1.0.0-signed"
	runGitForCLI(t, repoDir, "-c", "gpg.format=ssh", "-c", "user.signingkey="+keyPath, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "tag", "-s", "-m", "signed tag", signedTagName)

	verifier, err := contracts.NewSSHAllowedSignersVerifier(allowedSigners)
	if err != nil {
		t.Fatalf("create allowed-signers verifier: %v", err)
	}
	verification, err := verifier.VerifySignedTag(repoDir, signedTagName)
	if err != nil {
		t.Fatalf("verify signed test tag: %v", err)
	}

	return repoDir, signedGitSourceDetails{
		commit:            commit,
		signedTagName:     signedTagName,
		signerIdentity:    verification.SignerIdentity,
		signerFingerprint: verification.SignerFingerprint,
		allowedSigners:    allowedSigners,
	}
}
