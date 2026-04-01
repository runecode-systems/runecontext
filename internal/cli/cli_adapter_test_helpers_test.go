package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func statMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.Mode()
}

func createUserOwnedConfig(t *testing.T, projectRoot string) string {
	t.Helper()
	userConfigPath := filepath.Join(projectRoot, ".opencode", "config.yml")
	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0o755); err != nil {
		t.Fatalf("mkdir user config dir: %v", err)
	}
	if err := os.WriteFile(userConfigPath, []byte("user_owned: true\n"), 0o644); err != nil {
		t.Fatalf("write user config file: %v", err)
	}
	return userConfigPath
}

func assertUserOwnedConfigPreserved(t *testing.T, userConfigPath string) {
	t.Helper()
	userData, err := os.ReadFile(userConfigPath)
	if err != nil {
		t.Fatalf("read user config after sync: %v", err)
	}
	if string(userData) != "user_owned: true\n" {
		t.Fatalf("expected user config boundary to be preserved, got %q", string(userData))
	}
}

func assertAdapterSyncBoundaries(t *testing.T, userConfigPath, projectRoot string) {
	t.Helper()
	assertUserOwnedConfigPreserved(t, userConfigPath)
	assertNoAdapterTrackingTree(t, projectRoot)
	if _, err := os.Stat(filepath.Join(projectRoot, "adapters")); !os.IsNotExist(err) {
		t.Fatalf("expected sync to avoid user-owned adapter source tree writes, got err=%v", err)
	}
}

func assertNoAdapterTrackingTree(t *testing.T, projectRoot string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectRoot, ".runecontext", "adapters")); !os.IsNotExist(err) {
		t.Fatalf("expected no adapter tracking tree under .runecontext, got err=%v", err)
	}
}

func assertManagedArtifactMarker(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read managed artifact %s: %v", path, err)
	}
	if _, ok := parseHostNativeOwnershipHeader(data); !ok {
		t.Fatalf("expected ownership marker in %s", path)
	}
}

func assertShellInjectionCallPresent(t *testing.T, path, call string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read host-native artifact %s: %v", path, err)
	}
	token := "!`" + call + "`"
	if !strings.Contains(string(data), token) {
		t.Fatalf("expected shell-injection token %q in %s", token, path)
	}
}

func assertNoShellInjectionCall(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read host-native artifact %s: %v", path, err)
	}
	if strings.Contains(string(data), "!`") {
		t.Fatalf("expected no shell injection token in %s", path)
	}
}

func assertFrontmatterContains(t *testing.T, path, token string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read host-native artifact %s: %v", path, err)
	}
	if !strings.HasPrefix(string(data), "---\n") {
		t.Fatalf("expected frontmatter prefix in %s", path)
	}
	if !strings.Contains(string(data), token) {
		t.Fatalf("expected frontmatter token %q in %s", token, path)
	}
}

type validateHookBoundaryTest struct {
	scriptPath  string
	projectRoot string
	fakeBin     string
	calledPath  string
}

func prepareValidateHookBoundaryTest(t *testing.T) validateHookBoundaryTest {
	t.Helper()
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	test := validateHookBoundaryTest{
		scriptPath:  filepath.Join(repoRoot, "adapters", "source", "packs", "opencode", "automation", "validate_after_authoritative_edit.sh"),
		projectRoot: prepareCLIWorkflowProject(t),
		fakeBin:     t.TempDir(),
	}
	test.calledPath = filepath.Join(test.projectRoot, "validate-called")
	writeFakeRunectxExecutable(t, filepath.Join(test.fakeBin, "runectx"))
	return test
}

func writeFakeRunectxExecutable(t *testing.T, path string) {
	t.Helper()
	stub := "#!/usr/bin/env bash\nset -euo pipefail\nprintf '%s\\n' \"$*\" > \"$RUNECTX_ARGS_OUT\"\n"
	if err := os.WriteFile(path, []byte(stub), 0o755); err != nil {
		t.Fatalf("write fake runectx: %v", err)
	}
}

func runValidateHookScript(test validateHookBoundaryTest, changedPath string) (string, error) {
	_ = os.Remove(test.calledPath)
	cmd := exec.Command("bash", test.scriptPath, changedPath)
	cmd.Dir = test.projectRoot
	cmd.Env = append(os.Environ(),
		"PATH="+test.fakeBin+":"+os.Getenv("PATH"),
		"RUNECTX_ARGS_OUT="+test.calledPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("script failed: %w\n%s", err, string(out))
	}
	called, err := os.ReadFile(test.calledPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(called), nil
}
