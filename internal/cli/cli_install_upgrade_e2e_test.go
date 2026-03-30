package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManualRepoInstallFlowOverReferenceFixture(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	bundleRoot := t.TempDir()
	copyDirForCLI(t, filepath.Join(repoRoot, "schemas"), filepath.Join(bundleRoot, "schemas"))
	projectRoot := filepath.Join(bundleRoot, "projects", "embedded")
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), projectRoot)

	originalWD := t.TempDir()
	if cwd, err := filepath.Abs("."); err == nil {
		originalWD = cwd
	}
	t.Cleanup(func() { t.Chdir(originalWD) })
	t.Chdir(bundleRoot)

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	if code := Run([]string{"validate", "--path", projectRoot}, &validateOut, &validateErr); code != exitOK {
		t.Fatalf("expected validate success from bundle-style layout, got %d (%s)", code, validateErr.String())
	}
	if !strings.Contains(validateOut.String(), "result=ok") {
		t.Fatalf("expected validate success output, got %q", validateOut.String())
	}

	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected doctor success from bundle-style layout, got %d (%s)", code, doctorErr.String())
	}
	if !strings.Contains(doctorOut.String(), "command=doctor") {
		t.Fatalf("expected doctor output, got %q", doctorOut.String())
	}
}

func TestQuickInstallConfirmationViaVersionAndDoctor(t *testing.T) {
	setRunecontextVersionForTests(t, "v0.1.0-alpha.8")

	var versionOut bytes.Buffer
	var versionErr bytes.Buffer
	if code := Run([]string{"--version"}, &versionOut, &versionErr); code != exitOK {
		t.Fatalf("expected --version success, got %d (%s)", code, versionErr.String())
	}
	fields := parseCLIKeyValueOutput(t, versionOut.String())
	if got, want := fields["version"], "0.1.0-alpha.8"; got != want {
		t.Fatalf("expected normalized version %q, got %q", want, got)
	}

	projectRoot := repoFixtureRoot(t, "reference-projects", "embedded")
	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected doctor success, got %d (%s)", code, doctorErr.String())
	}
	doctorFields := parseCLIKeyValueOutput(t, doctorOut.String())
	if got, want := doctorFields["result"], "ok"; got != want {
		t.Fatalf("expected doctor result %q, got %q", want, got)
	}
}

func TestLocalCLIManagedInitFlow(t *testing.T) {
	projectRoot := filepath.Join(t.TempDir(), "local-init-project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("create init project root: %v", err)
	}

	var initOut bytes.Buffer
	var initErr bytes.Buffer
	if code := Run([]string{"init", "--path", projectRoot, "--seed-bundle", "base"}, &initOut, &initErr); code != exitOK {
		t.Fatalf("expected init success, got %d (%s)", code, initErr.String())
	}
	initFields := parseCLIKeyValueOutput(t, initOut.String())
	if got, want := initFields["result"], "ok"; got != want {
		t.Fatalf("expected init result %q, got %q", want, got)
	}

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	if code := Run([]string{"validate", "--path", projectRoot}, &validateOut, &validateErr); code != exitOK {
		t.Fatalf("expected validate success after init, got %d (%s)", code, validateErr.String())
	}

	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected doctor success after init, got %d (%s)", code, doctorErr.String())
	}
	doctorFields := parseCLIKeyValueOutput(t, doctorOut.String())
	if got, want := doctorFields["result"], "ok"; got != want {
		t.Fatalf("expected doctor result %q, got %q", want, got)
	}
}

func TestPreviewFirstUpgradeFlowOverReferenceFixture(t *testing.T) {
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), projectRoot)
	configPath := filepath.Join(projectRoot, "runecontext.yaml")
	originalConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read original config: %v", err)
	}
	runUpgradePreviewAndAssert(t, projectRoot)

	postPreviewConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after preview: %v", err)
	}
	if string(postPreviewConfig) != string(originalConfig) {
		t.Fatalf("expected preview to avoid mutation")
	}
	runUpgradeApplyAndAssert(t, projectRoot)

	updatedConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after apply: %v", err)
	}
	if !strings.Contains(string(updatedConfig), "runecontext_version: 0.1.0-alpha.9") {
		t.Fatalf("expected updated runecontext_version, got %q", string(updatedConfig))
	}
}

func runUpgradePreviewAndAssert(t *testing.T, projectRoot string) {
	t.Helper()
	var previewOut bytes.Buffer
	var previewErr bytes.Buffer
	if code := Run([]string{"upgrade", "--path", projectRoot, "--target-version", "0.1.0-alpha.9"}, &previewOut, &previewErr); code != exitOK {
		t.Fatalf("expected preview upgrade success, got %d (%s)", code, previewErr.String())
	}
	previewFields := parseCLIKeyValueOutput(t, previewOut.String())
	if got, want := previewFields["phase"], "preview"; got != want {
		t.Fatalf("expected preview phase %q, got %q", want, got)
	}
	if got, want := previewFields["state"], "upgradeable"; got != want {
		t.Fatalf("expected preview state %q, got %q", want, got)
	}
}

func runUpgradeApplyAndAssert(t *testing.T, projectRoot string) {
	t.Helper()
	var applyOut bytes.Buffer
	var applyErr bytes.Buffer
	if code := Run([]string{"upgrade", "apply", "--path", projectRoot, "--target-version", "0.1.0-alpha.9"}, &applyOut, &applyErr); code != exitOK {
		t.Fatalf("expected apply success, got %d (%s)", code, applyErr.String())
	}
	applyFields := parseCLIKeyValueOutput(t, applyOut.String())
	if got, want := applyFields["phase"], "apply"; got != want {
		t.Fatalf("expected apply phase %q, got %q", want, got)
	}
	if got, want := applyFields["current_version"], "0.1.0-alpha.9"; got != want {
		t.Fatalf("expected current_version %q, got %q", want, got)
	}
}

func TestReferenceFixturesCLIValidationCoverage(t *testing.T) {
	t.Run("embedded", testReferenceFixtureEmbedded)
	t.Run("verified", testReferenceFixtureVerified)
	t.Run("monorepo nested root", testReferenceFixtureMonorepoNestedRoot)
	t.Run("linked by commit", testReferenceFixtureLinkedByCommit)
	t.Run("linked by signed tag", testReferenceFixtureLinkedBySignedTag)
}

func testReferenceFixtureEmbedded(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "reference-projects", "embedded")
	assertCLIValidateOK(t, projectRoot)
	assertCLIDoctorOK(t, projectRoot)
}

func testReferenceFixtureVerified(t *testing.T) {
	projectRoot := repoFixtureRoot(t, "reference-projects", "verified")
	assertCLIValidateOK(t, projectRoot)
	assertCLIDoctorOK(t, projectRoot)
}

func testReferenceFixtureMonorepoNestedRoot(t *testing.T) {
	monorepoRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "monorepo"), monorepoRoot)
	copyDirForCLI(t, repoFixtureRoot(t, "..", "schemas"), filepath.Join(monorepoRoot, "schemas"))
	projectRoot := filepath.Join(monorepoRoot, "packages", "service", "app")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatalf("create nested service path: %v", err)
	}

	originalWD := t.TempDir()
	if cwd, err := filepath.Abs("."); err == nil {
		originalWD = cwd
	}
	t.Cleanup(func() { t.Chdir(originalWD) })
	t.Chdir(projectRoot)

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	if code := Run([]string{"validate"}, &validateOut, &validateErr); code != exitOK {
		t.Fatalf("expected nested monorepo validate success, got %d (%s)", code, validateErr.String())
	}
	validateFields := parseCLIKeyValueOutput(t, validateOut.String())
	if got, want := validateFields["result"], "ok"; got != want {
		t.Fatalf("expected validate result %q, got %q", want, got)
	}
	if got := filepath.ToSlash(validateFields["selected_config_path"]); !strings.Contains(got, "/packages/service/runecontext.yaml") {
		t.Fatalf("expected service runecontext.yaml selection, got %q", got)
	}

	assertDoctorFromWorkingDir(t)
}

func assertDoctorFromWorkingDir(t *testing.T) {
	t.Helper()
	var doctorOut bytes.Buffer
	var doctorErr bytes.Buffer
	if code := Run([]string{"doctor"}, &doctorOut, &doctorErr); code != exitOK {
		t.Fatalf("expected nested monorepo doctor success, got %d (%s)", code, doctorErr.String())
	}
	doctorFields := parseCLIKeyValueOutput(t, doctorOut.String())
	if got, want := doctorFields["result"], "ok"; got != want {
		t.Fatalf("expected doctor result %q, got %q", want, got)
	}
}

func testReferenceFixtureLinkedByCommit(t *testing.T) {
	repoDir := createLinkedReferenceSourceRepoForCLI(t)
	commit := strings.TrimSpace(gitOutputForCLI(t, repoDir, "rev-parse", "HEAD"))
	projectRoot := materializeReferenceLinkedFixtureForCLI(t, "linked-by-commit", map[string]string{
		"__GIT_URL__": repoDir,
		"__COMMIT__":  commit,
	})
	assertCLIValidateOK(t, projectRoot)
	assertCLIDoctorOK(t, projectRoot)
}

func testReferenceFixtureLinkedBySignedTag(t *testing.T) {
	repoDir, details := createSignedGitSourceRepoForCLI(t)
	projectRoot := materializeReferenceLinkedFixtureForCLI(t, "linked-by-signed-tag", map[string]string{
		"__GIT_URL__":    repoDir,
		"__SIGNED_TAG__": details.signedTagName,
		"__COMMIT__":     details.commit,
	})
	allowedSignersPath := filepath.Join(t.TempDir(), "allowed_signers")
	if err := os.WriteFile(allowedSignersPath, details.allowedSigners, 0o644); err != nil {
		t.Fatalf("write allowed signers: %v", err)
	}

	var validateOut bytes.Buffer
	var validateErr bytes.Buffer
	if code := Run([]string{"validate", "--path", projectRoot, "--ssh-allowed-signers", allowedSignersPath}, &validateOut, &validateErr); code != exitOK {
		t.Fatalf("expected linked signed-tag validate success, got %d (%s)", code, validateErr.String())
	}
	fields := parseCLIKeyValueOutput(t, validateOut.String())
	if got, want := fields["verification_posture"], "verified_signed_tag"; got != want {
		t.Fatalf("expected verification_posture %q, got %q", want, got)
	}
}

func assertCLIValidateOK(t *testing.T, projectRoot string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"validate", "--path", projectRoot}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected validate success for %q, got %d (%s)", projectRoot, code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["result"], "ok"; got != want {
		t.Fatalf("expected validate result %q, got %q", want, got)
	}
}

func assertCLIDoctorOK(t *testing.T, projectRoot string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if code := Run([]string{"doctor", "--path", projectRoot}, &stdout, &stderr); code != exitOK {
		t.Fatalf("expected doctor success for %q, got %d (%s)", projectRoot, code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["result"], "ok"; got != want {
		t.Fatalf("expected doctor result %q, got %q", want, got)
	}
}

func materializeReferenceLinkedFixtureForCLI(t *testing.T, fixtureName string, replacements map[string]string) string {
	t.Helper()
	targetRoot := t.TempDir()
	templatePath := repoFixtureRoot(t, "reference-projects", fixtureName, "runecontext.yaml.tmpl")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read linked fixture template: %v", err)
	}
	content := string(data)
	for from, to := range replacements {
		content = strings.ReplaceAll(content, from, to)
	}
	if err := os.WriteFile(filepath.Join(targetRoot, "runecontext.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write linked fixture config: %v", err)
	}
	return targetRoot
}

func createLinkedReferenceSourceRepoForCLI(t *testing.T) string {
	t.Helper()
	repoDir := t.TempDir()
	runGitForCLI(t, repoDir, "init", "--initial-branch=main")
	templateRoot := repoFixtureRoot(t, "source-resolution", "templates", "minimal-runecontext")
	copyDirForCLI(t, templateRoot, filepath.Join(repoDir, "runecontext"))
	runGitForCLI(t, repoDir, "add", ".")
	runGitForCLI(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	return repoDir
}
