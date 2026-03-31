package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type failingCLIUpgradeResolver struct{}

func (failingCLIUpgradeResolver) ResolveLatestRelease(currentVersion string) (string, error) {
	return "", errFailLatestRelease
}

var errFailLatestRelease = &cliUpgradeTestError{msg: "release metadata unavailable"}

type cliUpgradeTestError struct{ msg string }

func (e *cliUpgradeTestError) Error() string { return e.msg }

func TestRunUpgradeCLIPreviewDispatchAndContractFields(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "--target-version", "0.1.0-alpha.9", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["command"], "upgrade"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["phase"], "preview"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["scope"], "cli"; got != want {
		t.Fatalf("expected scope %q, got %q", want, got)
	}
	if got, want := fields["availability_state"], "update_available"; got != want {
		t.Fatalf("expected availability_state %q, got %q", want, got)
	}
	if got, want := fields["selected_release"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected selected_release %q, got %q", want, got)
	}
	if got, want := fields["target_release"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected target_release %q, got %q", want, got)
	}
	if got, want := fields["planned_install_action"], "download_and_install"; got != want {
		t.Fatalf("expected planned_install_action %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "false"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
}

func TestRunUpgradeCLIPreviewNonMutatingAndLatestNetworkOptIn(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")
	original := resolveLatestCLIReleaseFn
	t.Cleanup(func() { resolveLatestCLIReleaseFn = original })
	resolveLatestCLIReleaseFn = cliUpgradeResolverFunc(func(currentVersion string) (string, error) {
		return "0.1.0-alpha.9", nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "--target-version", "latest", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["phase"], "preview"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["network_access"], "true"; got != want {
		t.Fatalf("expected network_access %q, got %q", want, got)
	}
}

func TestRunUpgradeCLIPreviewLatestResolverFailureIsInvalid(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")
	original := resolveLatestCLIReleaseFn
	t.Cleanup(func() { resolveLatestCLIReleaseFn = original })
	resolveLatestCLIReleaseFn = failingCLIUpgradeResolver{}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "--target-version", "latest", "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "release metadata unavailable") {
		t.Fatalf("expected latest resolver error in stderr, got %q", stderr.String())
	}
}

func TestRunUpgradeCLIPreviewLatestRejectsInvalidResolvedVersion(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")
	original := resolveLatestCLIReleaseFn
	t.Cleanup(func() { resolveLatestCLIReleaseFn = original })
	resolveLatestCLIReleaseFn = cliUpgradeResolverFunc(func(currentVersion string) (string, error) {
		return "not-a-version", nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "--target-version", "latest", "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "must look like a semantic version") {
		t.Fatalf("expected semantic-version validation failure, got %q", stderr.String())
	}
}

func TestCLIUpgradeRuntimeRootDoesNotSearchCwd(t *testing.T) {
	executableRepo := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(executableRepo, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), []byte(testInstallerScriptContent()), 0o755); err != nil {
		t.Fatalf("write executable installer: %v", err)
	}
	cwdRepo := createCLIUpgradeFixtureRepo(t)
	subdir := filepath.Join(cwdRepo, "nested", "work")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir cwd subdir: %v", err)
	}
	originalExecutable := cliUpgradeExecutablePathFn
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		cliUpgradeExecutablePathFn = originalExecutable
		_ = os.Chdir(originalWD)
	})
	cliUpgradeExecutablePathFn = func() (string, error) {
		return filepath.Join(executableRepo, "bin", "runectx"), nil
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	root, err := cliUpgradeRuntimeRoot()
	if err != nil {
		t.Fatalf("expected runtime root resolution success, got %v", err)
	}
	if mustResolvePath(t, root, "runtime root") != mustResolvePath(t, executableRepo, "executable repo") {
		t.Fatalf("expected runtime root from executable repo, got %q", root)
	}
}

func TestResolveLatestCLIReleaseUsesShippedRuntimeManifest(t *testing.T) {
	repoRoot := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(repoRoot, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), []byte(testInstallerScriptContent()), 0o755); err != nil {
		t.Fatalf("write installer: %v", err)
	}
	manifest := `{"metadata_descriptor":{"release":{"version":"0.1.0-alpha.99"}}}`
	if err := os.WriteFile(filepath.Join(repoRoot, "share", "runecontext", "release-manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write release manifest: %v", err)
	}
	originalExecutable := cliUpgradeExecutablePathFn
	t.Cleanup(func() { cliUpgradeExecutablePathFn = originalExecutable })
	cliUpgradeExecutablePathFn = func() (string, error) {
		return filepath.Join(repoRoot, "bin", "runectx"), nil
	}

	version, err := staticLatestCLIReleaseResolver{}.ResolveLatestRelease("0.1.0-alpha.12")
	if err != nil {
		t.Fatalf("expected latest release resolution success, got %v", err)
	}
	if version != "0.1.0-alpha.99" {
		t.Fatalf("expected latest release from shipped runtime manifest, got %q", version)
	}
}

func TestRunUpgradeCLIApplyUsesInstallerBoundary(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")

	original := applyCLIUpgradePlanFn
	t.Cleanup(func() { applyCLIUpgradePlanFn = original })
	called := false
	applyCLIUpgradePlanFn = cliUpgradeInstallerFunc(func(plan cliUpgradePlan) (cliUpgradePlan, error) {
		called = true
		plan.Changed = true
		plan.UpdatedBinaryPath = "/tmp/runectx"
		return plan, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "apply", "--target-version", "0.1.0-alpha.9", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !called {
		t.Fatalf("expected cli installer boundary to be called")
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["phase"], "apply"; got != want {
		t.Fatalf("expected phase %q, got %q", want, got)
	}
	if got, want := fields["scope"], "cli"; got != want {
		t.Fatalf("expected scope %q, got %q", want, got)
	}
	if got, want := fields["changed"], "true"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}
	if got, want := fields["updated_binary_path"], "/tmp/runectx"; got != want {
		t.Fatalf("expected updated_binary_path %q, got %q", want, got)
	}
}

func TestRunUpgradeCLIApplyReportsInstallerFailure(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")
	originalInstaller := applyCLIUpgradePlanFn
	originalResolver := resolveLatestCLIReleaseFn
	t.Cleanup(func() {
		applyCLIUpgradePlanFn = originalInstaller
		resolveLatestCLIReleaseFn = originalResolver
	})
	applyCLIUpgradePlanFn = cliUpgradeInstallerFunc(func(plan cliUpgradePlan) (cliUpgradePlan, error) {
		return plan, &cliUpgradeTestError{msg: "installer failed"}
	})
	resolveLatestCLIReleaseFn = cliUpgradeResolverFunc(func(currentVersion string) (string, error) {
		return "0.1.0-alpha.9", nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "apply", "--target-version", "latest", "--json"}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "installer failed") {
		t.Fatalf("expected installer failure guidance, got %q", stderr.String())
	}
}

func TestRunUpgradeCLIApplyDefaultsTargetVersionToCurrent(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "apply", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["target_release"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected default target_release %q, got %q", want, got)
	}
}

func TestRunUpgradeCLIApplyShortCircuitsWhenAlreadyCurrent(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")
	originalInstaller := applyCLIUpgradePlanFn
	t.Cleanup(func() { applyCLIUpgradePlanFn = originalInstaller })
	called := false
	applyCLIUpgradePlanFn = cliUpgradeInstallerFunc(func(plan cliUpgradePlan) (cliUpgradePlan, error) {
		called = true
		return plan, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "apply", "--target-version", "current", "--json"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if called {
		t.Fatalf("expected installer boundary not to be called for up-to-date plan")
	}
	fields := parseCLIJSONEnvelopeData(t, stdout.Bytes())
	if got, want := fields["changed"], "false"; got != want {
		t.Fatalf("expected changed %q, got %q", want, got)
	}
}

func TestInstallerCommandForCurrentPlatformAnchorsToExecutableRepoNotCwd(t *testing.T) {
	executableRepo, cwdRepo, subdir := createExecutableAndCwdFixtureRepos(t)
	originalExecutable := cliUpgradeExecutablePathFn
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		cliUpgradeExecutablePathFn = originalExecutable
		_ = os.Chdir(originalWD)
	})
	cliUpgradeExecutablePathFn = func() (string, error) {
		return filepath.Join(executableRepo, "bin", "runectx"), nil
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	script, args, err := installerCommandForCurrentPlatform("0.1.0-alpha.9")
	if err != nil {
		t.Fatalf("expected installer resolution success, got %v", err)
	}
	assertInstallerResolutionAnchorsToExecutableRepo(t, script, args, executableRepo, cwdRepo)
}

func assertInstallerResolutionAnchorsToExecutableRepo(t *testing.T, script string, args []string, executableRepo, cwdRepo string) {
	t.Helper()
	if script != expectedInstallerLauncher() {
		t.Fatalf("expected installer launcher %q, got %q", expectedInstallerLauncher(), script)
	}
	actualPath := mustResolvePath(t, installerScriptArg(t, args), "actual installer path")
	if actualPath != mustResolvePath(t, filepath.Join(executableRepo, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), "expected installer path") {
		t.Fatalf("expected installer path from executable repo, got %#v", args)
	}
	if actualPath == mustResolvePath(t, filepath.Join(cwdRepo, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), "cwd installer path") {
		t.Fatalf("expected installer path not to come from cwd repo, got %#v", args)
	}
}

func installerScriptArg(t *testing.T, args []string) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		if len(args) == 0 {
			t.Fatalf("expected installer args to include script path")
		}
		return args[0]
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "-File" {
			return args[i+1]
		}
	}
	t.Fatalf("expected PowerShell installer args to include -File <script>, got %#v", args)
	return ""
}

func createExecutableAndCwdFixtureRepos(t *testing.T) (string, string, string) {
	t.Helper()
	executableRepo := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(executableRepo, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), []byte(testInstallerScriptContent()), 0o755); err != nil {
		t.Fatalf("write executable repo installer: %v", err)
	}
	cwdRepo := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(cwdRepo, "share", "runecontext", "installers", cliUpgradeInstallerScriptName()), []byte(testInstallerScriptContent()), 0o755); err != nil {
		t.Fatalf("write cwd repo installer: %v", err)
	}
	subdir := filepath.Join(cwdRepo, "nested", "work")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir fake cwd: %v", err)
	}
	return executableRepo, cwdRepo, subdir
}

func TestInstallerCommandForCurrentPlatformRejectsSymlinkedInstallerScript(t *testing.T) {
	repoRoot := createCLIUpgradeFixtureRepo(t)
	if err := os.Symlink(filepath.Join(repoRoot, "real-installer.sh"), filepath.Join(repoRoot, "share", "runecontext", "installers", cliUpgradeInstallerScriptName())); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create installer symlink: %v", err)
	}
	originalExecutable := cliUpgradeExecutablePathFn
	t.Cleanup(func() { cliUpgradeExecutablePathFn = originalExecutable })
	cliUpgradeExecutablePathFn = func() (string, error) {
		return filepath.Join(repoRoot, "bin", "runectx"), nil
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "bin", "runectx"), []byte("binary\n"), 0o755); err != nil {
		t.Fatalf("write fake executable: %v", err)
	}

	_, _, err := installerCommandForCurrentPlatform("0.1.0-alpha.9")
	if err == nil || !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("expected symlinked installer rejection, got %v", err)
	}
}

func createCLIUpgradeFixtureRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	makeCLIUpgradeFixtureDirs(t, repoRoot)
	if err := os.WriteFile(filepath.Join(repoRoot, "nix", "release", "metadata.nix"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "real-installer.sh"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write real installer: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "bin"), 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "bin", "runectx"), []byte("binary\n"), 0o755); err != nil {
		t.Fatalf("write fake executable: %v", err)
	}
	for _, name := range requiredSchemaNames() {
		if err := os.WriteFile(filepath.Join(repoRoot, "share", "runecontext", "schemas", name), []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("write schema %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(repoRoot, "share", "runecontext", "release-manifest.json"), []byte(`{"metadata_descriptor":{"release":{"version":"0.1.0-alpha.12"}}}`), 0o644); err != nil {
		t.Fatalf("write release manifest: %v", err)
	}
	return repoRoot
}

func makeCLIUpgradeFixtureDirs(t *testing.T, repoRoot string) {
	t.Helper()
	for _, dir := range []string{"nix/release", "scripts", "share/runecontext/installers", "share/runecontext/schemas", "bin"} {
		if err := os.MkdirAll(filepath.Join(repoRoot, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
}

func mustResolvePath(t *testing.T, path, label string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve %s: %v", label, err)
	}
	return resolved
}

func testInstallerScriptContent() string {
	if runtime.GOOS == "windows" {
		return "Write-Output 'installer'\n"
	}
	return "#!/usr/bin/env bash\n"
}

func expectedInstallerLauncher() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	return "bash"
}

func TestRunUpgradeCLIAndProjectUpgradeRemainDistinct(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := createEmbeddedProjectForUpgradeTests(t)
	var upgradeStdout bytes.Buffer
	var upgradeStderr bytes.Buffer
	code := Run([]string{"upgrade", "--path", root, "--json"}, &upgradeStdout, &upgradeStderr)
	if code != exitOK {
		t.Fatalf("expected project upgrade preview success, got %d (%s)", code, upgradeStderr.String())
	}
	projectFields := parseCLIJSONEnvelopeData(t, upgradeStdout.Bytes())
	if got := projectFields["scope"]; got != "" {
		t.Fatalf("expected project upgrade output to omit cli scope, got %q", got)
	}
	if got, want := projectFields["state"], "current"; got != want {
		t.Fatalf("expected project upgrade state %q, got %q", want, got)
	}

	var cliStdout bytes.Buffer
	var cliStderr bytes.Buffer
	code = Run([]string{"upgrade", "cli", "--target-version", "current", "--json"}, &cliStdout, &cliStderr)
	if code != exitOK {
		t.Fatalf("expected cli upgrade preview success, got %d (%s)", code, cliStderr.String())
	}
	cliFields := parseCLIJSONEnvelopeData(t, cliStdout.Bytes())
	if got, want := cliFields["scope"], "cli"; got != want {
		t.Fatalf("expected cli scope %q, got %q", want, got)
	}
	if got := cliFields["state"]; got != "" {
		t.Fatalf("expected cli preview to omit project state, got %q", got)
	}
}

type cliUpgradeInstallerFunc func(plan cliUpgradePlan) (cliUpgradePlan, error)

func (fn cliUpgradeInstallerFunc) Apply(plan cliUpgradePlan) (cliUpgradePlan, error) {
	return fn(plan)
}

type cliUpgradeResolverFunc func(currentVersion string) (string, error)

func (fn cliUpgradeResolverFunc) ResolveLatestRelease(currentVersion string) (string, error) {
	return fn(currentVersion)
}
