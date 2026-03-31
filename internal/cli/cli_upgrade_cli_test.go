package cli

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestRunUpgradeCLIApplyRequiresTargetVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"upgrade", "cli", "apply"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "upgrade cli apply requires --target-version") {
		t.Fatalf("expected missing target-version error, got %q", stderr.String())
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
	if script != "bash" {
		t.Fatalf("expected bash launcher, got %q", script)
	}
	if len(args) == 0 || args[0] != filepath.Join(executableRepo, "scripts", "install-runectx.sh") {
		t.Fatalf("expected installer path from executable repo, got %#v", args)
	}
	if len(args) > 0 && args[0] == filepath.Join(cwdRepo, "scripts", "install-runectx.sh") {
		t.Fatalf("expected installer path not to come from cwd repo, got %#v", args)
	}
}

func createExecutableAndCwdFixtureRepos(t *testing.T) (string, string, string) {
	t.Helper()
	executableRepo := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(executableRepo, "scripts", "install-runectx.sh"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatalf("write executable repo installer: %v", err)
	}
	cwdRepo := createCLIUpgradeFixtureRepo(t)
	if err := os.WriteFile(filepath.Join(cwdRepo, "scripts", "install-runectx.sh"), []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
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
	if err := os.Symlink(filepath.Join(repoRoot, "real-installer.sh"), filepath.Join(repoRoot, "scripts", "install-runectx.sh")); err != nil {
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
	if err := os.MkdirAll(filepath.Join(repoRoot, "nix", "release"), 0o755); err != nil {
		t.Fatalf("mkdir metadata dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "scripts"), 0o755); err != nil {
		t.Fatalf("mkdir scripts dir: %v", err)
	}
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
	return repoRoot
}

func TestRunUpgradeCLIAndProjectUpgradeRemainDistinct(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.10")

	root := repoFixtureRoot(t, "reference-projects", "embedded")
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
