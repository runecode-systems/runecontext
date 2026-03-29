package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
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

func TestRunValidateReportsUnsupportedProjectVersionDiagnostics(t *testing.T) {
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v0.1.0-alpha.9"

	root := t.TempDir()
	config := "schema_version: 1\nrunecontext_version: 9.9.9\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir content root: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected validate success with diagnostics, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "unsupported_project_version") {
		t.Fatalf("expected unsupported version diagnostic, got %q", stdout.String())
	}
}

func TestRunValidateReportsStaleTreeWithoutMutating(t *testing.T) {
	root := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "reference-projects", "embedded"), root)

	if code := Run([]string{"adapter", "sync", "--path", root, "opencode"}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("expected adapter sync success")
	}
	staleRel := ".opencode/skills/runecontext-stale-merge-validate.md"
	staleAbs := filepath.Join(root, filepath.FromSlash(staleRel))
	managed := "<!-- runecontext-managed-artifact: host-native-v1 -->\n<!-- runecontext-tool: opencode -->\n<!-- runecontext-kind: flow_asset -->\n<!-- runecontext-id: runecontext:stale-merge-validate -->\n"
	if err := os.WriteFile(staleAbs, []byte(managed), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected validate success with warnings, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mixed_or_stale_tree") {
		t.Fatalf("expected stale-tree diagnostic, got %q", stdout.String())
	}
	if _, err := os.Stat(staleAbs); err != nil {
		t.Fatalf("expected validate to avoid hidden upgrade mutation, err=%v", err)
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

func TestRunValidateSignedTagEmptyExpectCommitFailsClearly(t *testing.T) {
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)

	repoDir, details := createSignedGitSourceRepoForCLI(t)
	projectRoot := t.TempDir()
	config := fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: \"\"\n  subdir: runecontext\n", repoDir, details.signedTagName)
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "git expect_commit must not be empty") {
		t.Fatalf("expected explicit empty expect_commit error, got %q", stderr.String())
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
	if !strings.Contains(stderr.String(), "usage=runectx validate [--json] [--non-interactive] [--explain] [--ssh-allowed-signers PATH] [--path PATH] [path]") {
		t.Fatalf("expected usage output, got %q", stderr.String())
	}
}

func TestRunValidateAcceptsPathFlag(t *testing.T) {
	root := fixtureRoot(t, "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--path", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "result=ok") {
		t.Fatalf("expected success output, got %q", stdout.String())
	}
}

func TestRunValidateRejectsPathConflict(t *testing.T) {
	root := fixtureRoot(t, "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"validate", "--path", root, root}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot use both --path and a positional path argument") {
		t.Fatalf("expected --path conflict output, got %q", stderr.String())
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
	foundKeyValue := false
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "=") {
			t.Fatalf("expected key=value output, got %q", line)
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("malformed CLI output line: %q", line)
		}
		foundKeyValue = true
		fields[parts[0]] = unsanitizeCLIValue(parts[1])
	}
	if !foundKeyValue {
		t.Fatalf("expected at least one key=value line in output: %q", output)
	}
	return fields
}

func parseCLIJSONEnvelopeData(t *testing.T, payload []byte) map[string]string {
	t.Helper()
	var envelope struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("expected JSON output, got err=%v payload=%q", err, string(payload))
	}
	if envelope.Data == nil {
		t.Fatalf("expected JSON envelope data, got payload=%q", string(payload))
	}
	return envelope.Data
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

func TestSanitizeValueRoundTripsEscapedSequences(t *testing.T) {
	cases := []string{
		"plain",
		"has=equals",
		"has\\backslash",
		"line1\nline2",
		"carriage\rreturn",
		"tab\tvalue",
		"null\x00byte",
		"combo\\=\n\t\r\x00",
	}
	for _, value := range cases {
		if got := unsanitizeCLIValue(sanitizeValue(value)); got != value {
			t.Fatalf("expected sanitize/unsanitize round trip for %q, got %q", value, got)
		}
	}
}
