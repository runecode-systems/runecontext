package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoFixtureRoot(t, "traceability"), name)
}

func repoFixtureRoot(t *testing.T, elems ...string) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	parts := append([]string{root, "fixtures"}, elems...)
	return filepath.Join(parts...)
}

type cliSignedTagDetails struct {
	commit            string
	signedTagName     string
	signerIdentity    string
	signerFingerprint string
	allowedSigners    []byte
}

func createSignedGitSourceRepoForCLI(t *testing.T) (string, cliSignedTagDetails) {
	t.Helper()
	requireToolForCLITests(t, "git")
	requireToolForCLITests(t, "ssh-keygen")
	repoDir := t.TempDir()
	runGitForCLI(t, repoDir, "init", "--initial-branch=main")
	templateRoot := repoFixtureRoot(t, "source-resolution", "templates", "minimal-runecontext")
	copyDirForCLI(t, templateRoot, filepath.Join(repoDir, "runecontext"))
	runGitForCLI(t, repoDir, "add", ".")
	runGitForCLI(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	commit := strings.TrimSpace(gitOutputForCLI(t, repoDir, "rev-parse", "HEAD"))
	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "signer")
	runCommandForCLI(t, repoDir, sanitizedGitEnvForCLITests(), "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", keyPath)
	pubKey, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		t.Fatalf("read public key: %v", err)
	}
	allowedSigners := []byte(fmt.Sprintf("alice@example.com %s\n", strings.TrimSpace(string(pubKey))))
	verifier := newCLIAllowedSignersVerifier(t, allowedSigners)
	signedTagName := "v1.0.0-signed"
	runGitForCLI(t, repoDir, "-c", "gpg.format=ssh", "-c", "user.signingkey="+keyPath, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "tag", "-s", "-m", "signed tag", signedTagName)
	verification, err := verifier.VerifySignedTag(repoDir, signedTagName)
	if err != nil {
		t.Fatalf("verify signed tag for CLI fixture: %v", err)
	}
	return repoDir, cliSignedTagDetails{
		commit:            commit,
		signedTagName:     signedTagName,
		signerIdentity:    verification.SignerIdentity,
		signerFingerprint: verification.SignerFingerprint,
		allowedSigners:    allowedSigners,
	}
}

func requireToolForCLITests(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
}

func newCLIAllowedSignersVerifier(t *testing.T, allowedSigners []byte) *contracts.SSHAllowedSignersVerifier {
	t.Helper()
	verifier, err := contracts.NewSSHAllowedSignersVerifier(allowedSigners)
	if err != nil {
		t.Fatalf("create cli allowed-signers verifier: %v", err)
	}
	return verifier
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

func runGitForCLI(t *testing.T, dir string, args ...string) {
	t.Helper()
	runCommandForCLI(t, dir, sanitizedGitEnvForCLITests(), "git", args...)
}

func gitOutputForCLI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	return runCommandOutputForCLI(t, dir, sanitizedGitEnvForCLITests(), nil, "git", args...)
}

func runCommandForCLI(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	_ = runCommandOutputForCLI(t, dir, env, nil, name, args...)
}

func runCommandOutputForCLI(t *testing.T, dir string, env []string, stdin *bytes.Reader, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = stdin
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return string(output)
}

func sanitizedGitEnvForCLITests() []string {
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
