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

func TestRunValidateSuccess(t *testing.T) {
	root := fixtureRoot(t, "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if stdout.String() == "" {
		t.Fatalf("expected success output, got empty stdout")
	}
	if !strings.Contains(stdout.String(), "result=ok") {
		t.Fatalf("expected success result line, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "command=validate") {
		t.Fatalf("expected command line, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "root=") {
		t.Fatalf("expected success output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "selected_config_path=") {
		t.Fatalf("expected selected config metadata, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "source_mode=embedded") {
		t.Fatalf("expected source metadata, got %q", stdout.String())
	}
}

func TestRunValidateSurfacesDeprecatedStandardDiagnostics(t *testing.T) {
	root := filepath.Join(repoFixtureRoot(t, "bundle-resolution"), "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "deprecated_standard_selected") {
		t.Fatalf("expected deprecated standard diagnostic, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "_bundle=child-reinclude") {
		t.Fatalf("expected bundle metadata in diagnostics, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "_aspect=standards") {
		t.Fatalf("expected aspect metadata in diagnostics, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "_matches=standards/global/legacy.md") {
		t.Fatalf("expected match metadata in diagnostics, got %q", stdout.String())
	}
}

func TestRunValidateSurfacesProjectValidationDiagnostics(t *testing.T) {
	root := filepath.Join(repoFixtureRoot(t, "traceability"), "valid-project")
	projectRoot := t.TempDir()
	copyDirForCLI(t, root, projectRoot)
	standardPath := filepath.Join(projectRoot, "runecontext", "standards", "global", "deterministic-check-write.md")
	data, err := os.ReadFile(standardPath)
	if err != nil {
		t.Fatalf("read standard fixture: %v", err)
	}
	updated := strings.Replace(string(data), "status: active", "status: deprecated\nreplaced_by: standards/global/deterministic-check-write-v2.md", 1)
	if err := os.WriteFile(standardPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write standard fixture: %v", err)
	}
	replacementPath := filepath.Join(projectRoot, "runecontext", "standards", "global", "deterministic-check-write-v2.md")
	if err := os.WriteFile(replacementPath, []byte("---\nschema_version: 1\nid: global/deterministic-check-write-v2\ntitle: Deterministic Check Write v2\nstatus: active\n---\n\n# Deterministic Check Write v2\n\nUse the newer wording.\n"), 0o644); err != nil {
		t.Fatalf("write replacement standard: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "deprecated_standard_referenced") {
		t.Fatalf("expected project validation diagnostic, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "diagnostic_1_path=") && !strings.Contains(stdout.String(), "diagnostic_2_path=") {
		t.Fatalf("expected diagnostic path metadata, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "changes/CHG-2026-001-a3f2-auth-gateway/standards.md") {
		t.Fatalf("expected relative diagnostic path, got %q", stdout.String())
	}
}

func TestCollectDiagnosticsDeduplicatesBundleWarnings(t *testing.T) {
	index := &contracts.ProjectIndex{}
	first := emittedDiagnostic{Severity: contracts.DiagnosticSeverityWarning, Code: "deprecated_standard_selected", Message: "same", Bundle: "bundle-a", Aspect: "standards", Rule: "include", Pattern: "standards/global/legacy.md", Matches: []string{"standards/global/legacy.md"}}
	second := emittedDiagnostic{Severity: contracts.DiagnosticSeverityWarning, Code: "deprecated_standard_selected", Message: "same", Bundle: "bundle-a", Aspect: "standards", Rule: "include", Pattern: "standards/global/legacy.md", Matches: []string{"standards/global/legacy.md"}}
	_ = index
	items := dedupeDiagnostics([]emittedDiagnostic{first, second})
	if len(items) != 1 {
		t.Fatalf("expected diagnostics to dedupe, got %#v", items)
	}
}

func TestRunValidateNearestAncestorDiscoveryReportsSelectedConfig(t *testing.T) {
	nested := filepath.Join(repoFixtureRoot(t, "source-resolution", "monorepo"), "packages", "service", "internal")
	t.Chdir(nested)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	normalizedStdout := filepath.ToSlash(strings.ReplaceAll(stdout.String(), "\\\\", "\\"))
	if !strings.Contains(normalizedStdout, "selected_config_path=") || !strings.Contains(normalizedStdout, "packages/service/runecontext.yaml") {
		t.Fatalf("expected nested selected config path, got %q", stdout.String())
	}
	if !strings.Contains(normalizedStdout, "project_root=") || !strings.Contains(normalizedStdout, "packages/service") {
		t.Fatalf("expected nested project root, got %q", stdout.String())
	}
}

func TestRunValidateExternalProjectUsesRepoSchemas(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)

	projectRoot := t.TempDir()
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir source root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "selected_config_path=") {
		t.Fatalf("expected selected config output, got %q", stdout.String())
	}
	if strings.Contains(stderr.String(), "schemas/runecontext.schema.json") {
		t.Fatalf("expected CLI to use repo schemas, got %q", stderr.String())
	}
}

func TestRunValidateOutputsSignedTagSignerMetadata(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)

	repoDir, details := createSignedGitSourceRepoForCLI(t)
	projectRoot := t.TempDir()
	config := fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.signedTagName, details.commit)
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	allowedSignersPath := filepath.Join(projectRoot, "trusted_signers")
	if err := os.WriteFile(allowedSignersPath, details.allowedSigners, 0o600); err != nil {
		t.Fatalf("write allowed signers file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--ssh-allowed-signers", allowedSignersPath, projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "verification_posture=verified_signed_tag") {
		t.Fatalf("expected signed-tag verification posture, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "verified_signer_identity="+details.signerIdentity) {
		t.Fatalf("expected signer identity output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "verified_signer_fingerprint="+details.signerFingerprint) {
		t.Fatalf("expected signer fingerprint output, got %q", stdout.String())
	}
}

func TestRunValidateSignedTagFailureOutputsStructuredReason(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)

	repoDir, details := createSignedGitSourceRepoForCLI(t)
	projectRoot := t.TempDir()
	config := fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.signedTagName, details.commit)
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	wrongAllowedSignersPath := filepath.Join(projectRoot, "wrong_trusted_signers")
	if err := os.WriteFile(wrongAllowedSignersPath, []byte("bob@example.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE5XQmFkRHVtbXlLZXlNYXRlcmlhbEZvclRlc3Rz\n"), 0o600); err != nil {
		t.Fatalf("write wrong allowed signers file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--ssh-allowed-signers", wrongAllowedSignersPath, projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_reason=untrusted_signer") {
		t.Fatalf("expected structured error reason, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_tag="+details.signedTagName) {
		t.Fatalf("expected structured error tag, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "diagnostic_count=1") {
		t.Fatalf("expected structured diagnostic count, got %q", stderr.String())
	}
}

func TestRunValidateFailure(t *testing.T) {
	root := fixtureRoot(t, "reject-change-missing-related-spec")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=invalid") {
		t.Fatalf("expected invalid result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_path=") {
		t.Fatalf("expected error path output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=") {
		t.Fatalf("expected validation failure output, got %q", stderr.String())
	}
}

func TestRunChangeNewCreatesChange(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Add cache invalidation", "--type", "feature", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "change_new"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["change_mode"], "minimum"; got != want {
		t.Fatalf("expected change_mode %q, got %q", want, got)
	}
	if got := fields["change_id"]; !strings.HasPrefix(got, "CHG-20") {
		t.Fatalf("expected change_id output, got %q", stdout.String())
	}
	if got, want := fields["review_diff_required"], "true"; got != want {
		t.Fatalf("expected review_diff_required %q, got %q", want, got)
	}
	changeDir := filepath.Join(projectRoot, "runecontext", "changes", fields["change_id"])
	if _, err := os.Stat(filepath.Join(changeDir, "proposal.md")); err != nil {
		t.Fatalf("expected proposal.md to exist: %v", err)
	}
}

func TestRunChangeNewExplicitDotPathUsesExplicitRoot(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot, err := os.MkdirTemp(repoRoot, "cli-explicit-root-")
	if err != nil {
		t.Fatalf("mktemp under repo root: %v", err)
	}
	defer os.RemoveAll(projectRoot)
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	nestedRoot := filepath.Join(projectRoot, "nested")
	if err := os.MkdirAll(nestedRoot, 0o755); err != nil {
		t.Fatalf("mkdir nested root: %v", err)
	}
	t.Chdir(nestedRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", "Add cache invalidation", "--type", "feature", "--size", "small", "--bundle", "base", "--path", "."}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code for explicit current-dir root, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "runecontext.yaml") {
		t.Fatalf("expected explicit-root lookup failure mentioning runecontext.yaml, got %q", stderr.String())
	}
}

func TestRunChangeShapeRefreshesStandards(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	changeID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	updated := strings.Replace(string(data), "- base", "- security", 1)
	if err := os.WriteFile(statusPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "shape", changeID, "--task", "Implement cache invalidation flow.", "--reference", "docs/cache.md", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["standards_refresh"], "updated"; got != want {
		t.Fatalf("expected standards_refresh %q, got %q", want, got)
	}
	if got, want := fields["added_standard_1"], "standards/security/review.md"; got != want {
		t.Fatalf("expected added standard %q, got %q", want, got)
	}
	if got, want := fields["review_diff_required"], "true"; got != want {
		t.Fatalf("expected review_diff_required %q, got %q", want, got)
	}
	if got, want := fields["changed_file_1_action"], "created"; got != want {
		t.Fatalf("expected first file action %q, got %q", want, got)
	}
}

func TestRunChangeCloseOutputsClosedChange(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	changeID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", changeID, "--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["change_status"], "closed"; got != want {
		t.Fatalf("expected change_status %q, got %q", want, got)
	}
	if got, want := fields["closed_at"], "2026-03-20"; got != want {
		t.Fatalf("expected closed_at %q, got %q", want, got)
	}
}

func TestRunStatusOutputsCounts(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	firstID := runCLIChangeNewForTest(t, projectRoot, "Add cache invalidation")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Revise cache invalidation")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", firstID, "--verification-status", "skipped", "--superseded-by", secondID, "--closed-at", "2026-03-20", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("close superseded change: %d (%s)", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"change", "close", secondID, "--verification-status", "passed", "--closed-at", "2026-03-21", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("close remaining change: %d (%s)", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"status", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("status command failed: %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["active_count"], "0"; got != want {
		t.Fatalf("expected active_count %q, got %q", want, got)
	}
	if got, want := fields["closed_count"], "1"; got != want {
		t.Fatalf("expected closed_count %q, got %q", want, got)
	}
	if got, want := fields["superseded_count"], "1"; got != want {
		t.Fatalf("expected superseded_count %q, got %q", want, got)
	}
	if got, want := fields["superseded_1_id"], firstID; got != want {
		t.Fatalf("expected superseded change %q, got %q", want, got)
	}
	if got, want := fields["closed_1_id"], secondID; got != want {
		t.Fatalf("expected closed change %q, got %q", want, got)
	}
}

func TestRunValidateRejectsInvalidProposal(t *testing.T) {
	root := fixtureRoot(t, "reject-proposal-invalid")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_path=") || !strings.Contains(stderr.String(), "proposal.md") {
		t.Fatalf("expected proposal path in output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsInvalidBundle(t *testing.T) {
	root := fixtureRoot(t, "reject-bundle-invalid")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", root}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected validation failure exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "error_path=") || !strings.Contains(stderr.String(), "bundles") {
		t.Fatalf("expected bundle path in output, got %q", stderr.String())
	}
}

func TestRunValidateUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "a", "b"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage=runectx validate [--ssh-allowed-signers PATH] [path]") {
		t.Fatalf("expected usage output, got %q", stderr.String())
	}
}

func runCLIChangeNewForTest(t *testing.T, projectRoot, title string) string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", title, "--type", "feature", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("change new failed: %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())["change_id"]
}

func parseCLIKeyValueOutput(t *testing.T, output string) map[string]string {
	t.Helper()
	fields := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("unexpected CLI output line %q", line)
		}
		fields[parts[0]] = unsanitizeCLIValue(parts[1])
	}
	return fields
}

func unsanitizeCLIValue(value string) string {
	var builder strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '\\' || i+1 >= len(value) {
			builder.WriteByte(value[i])
			continue
		}
		i++
		switch value[i] {
		case '\\':
			builder.WriteByte('\\')
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case '0':
			builder.WriteByte('\x00')
		case '=':
			builder.WriteByte('=')
		default:
			builder.WriteByte('\\')
			builder.WriteByte(value[i])
		}
	}
	return builder.String()
}

func TestRunValidateRejectsUnknownFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown validate flag") {
		t.Fatalf("expected unknown-flag output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsMissingAllowedSignersPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected missing-path output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsEmptyAllowedSignersEqualsValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers="}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected empty-value usage output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsEmptyAllowedSignersSeparateValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers", ""}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected empty separate-value usage output, got %q", stderr.String())
	}
}

func TestRunValidateRejectsBlankAllowedSignersEqualsValue(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--ssh-allowed-signers=   "}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires a path") {
		t.Fatalf("expected blank equals-value usage output, got %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "result=usage_error") {
		t.Fatalf("expected usage result output, got %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error_message=unknown command") {
		t.Fatalf("expected unknown command output, got %q", stderr.String())
	}
}

func TestRunNoCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected help exit code, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "runectx help") {
		t.Fatalf("expected help subcommand in usage output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "help       Show CLI usage") {
		t.Fatalf("expected help command description, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

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
