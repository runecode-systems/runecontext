package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunUpgradeApplySkipsDirenvSymlinkedPaths(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	if err := os.MkdirAll(filepath.Join(root, ".direnv"), 0o755); err != nil {
		t.Fatalf("mkdir .direnv: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".direnv", "target"), []byte("target\n"), 0o644); err != nil {
		t.Fatalf("write .direnv target: %v", err)
	}
	if err := os.Symlink("target", filepath.Join(root, ".direnv", "flake-profile")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create .direnv symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success with .direnv symlink present, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["current_version"], "0.1.0-alpha.12"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
}

func TestRunUpgradeApplySkipsGitignoredNonProtectedSymlink(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("local-symlink\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	runGitForCLI(t, root, "init", "--initial-branch=main")
	if err := os.WriteFile(filepath.Join(root, "target.txt"), []byte("target\n"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(root, "local-symlink")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create ignored symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected apply success with ignored symlink present, got %d (%s)", code, stderr.String())
	}
}

func TestRunUpgradeApplyRejectsProtectedManagedSymlink(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	if err := os.MkdirAll(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode parent: %v", err)
	}
	outside := filepath.Join(root, "outside-opencode")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("mkdir outside target: %v", err)
	}
	if err := os.RemoveAll(filepath.Join(root, ".opencode")); err != nil {
		t.Fatalf("remove .opencode dir: %v", err)
	}
	if err := os.Symlink("outside-opencode", filepath.Join(root, ".opencode")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create protected managed symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected protected symlink rejection, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "rejects symlinked path") || !strings.Contains(stderr.String(), ".opencode") {
		t.Fatalf("expected protected symlink rejection message, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyRejectsSymlinkedConfigPath(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	configTarget := filepath.Join(root, "real-runecontext.yaml")
	configContent, err := os.ReadFile(filepath.Join(root, "runecontext.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if err := os.WriteFile(configTarget, configContent, 0o644); err != nil {
		t.Fatalf("write config target: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "runecontext.yaml")); err != nil {
		t.Fatalf("remove original config: %v", err)
	}
	if err := os.Symlink("real-runecontext.yaml", filepath.Join(root, "runecontext.yaml")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create config symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected config symlink rejection, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "upgrade staging rejects symlinked path runecontext.yaml") {
		t.Fatalf("expected config symlink rejection message, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyRejectsNonIgnoredSymlinkWhenGitDirIsOverridden(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	runGitForCLI(t, root, "init", "--initial-branch=main")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("different-path\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "target.txt"), []byte("target\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(root, "local-symlink")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlink: %v", err)
	}
	t.Setenv("GIT_DIR", filepath.Join(root, ".git"))
	t.Setenv("GIT_WORK_TREE", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected non-ignored symlink rejection, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "upgrade staging rejects symlinked path local-symlink") {
		t.Fatalf("expected local symlink rejection message, got %q", stderr.String())
	}
}

func TestRunUpgradeApplyRejectsNonIgnoredSymlinkWhenGlobalIgnoreWouldMatch(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)
	setRunecontextVersionForTests(t, "v0.1.0-alpha.12")
	runGitForCLI(t, root, "init", "--initial-branch=main")
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "git")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir git config dir: %v", err)
	}
	ignorePath := filepath.Join(configDir, "ignore")
	if err := os.WriteFile(ignorePath, []byte("local-symlink\n"), 0o644); err != nil {
		t.Fatalf("write global ignore file: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"))
	if err := os.WriteFile(filepath.Join(root, "target.txt"), []byte("target\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink("target.txt", filepath.Join(root, "local-symlink")); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlink: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "apply", "--path", root, "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected global-ignore-neutralized symlink rejection, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "upgrade staging rejects symlinked path local-symlink") {
		t.Fatalf("expected local symlink rejection message, got %q", stderr.String())
	}
}

func TestCollectUpgradeProtectedRelPathsIgnoresSourcePathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "runecontext.yaml")
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.12\nassurance_tier: plain\nsource:\n  type: embedded\n  path: ../../etc/secret\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	paths := collectUpgradeProtectedRelPaths(root, upgradePlan{ConfigPath: configPath, SourceType: "embedded"})
	for _, path := range paths {
		if strings.HasPrefix(path, "../") {
			t.Fatalf("expected protected paths to exclude outside-root source path, got %v", paths)
		}
	}
}
