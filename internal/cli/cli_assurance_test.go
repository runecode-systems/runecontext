package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestEnsureAssuranceTierConfigReplacesLine(t *testing.T) {
	original := "schema_version: 1\nassurance_tier: plain\nsource:\n  type: embedded\n"
	updated, replaced := ensureAssuranceTierConfig([]byte(original))
	if !replaced {
		t.Fatalf("expected replacement")
	}
	if !strings.Contains(string(updated), "assurance_tier: verified") {
		t.Fatalf("expected verified tier, got %s", string(updated))
	}
	if strings.Count(string(updated), "assurance_tier: verified") != 1 {
		t.Fatalf("expected exactly one assurance tier line, got %s", string(updated))
	}
}

func TestEnsureAssuranceTierConfigAddsWhenMissing(t *testing.T) {
	original := "schema_version: 1\nsource:\n  type: embedded\n"
	updated, replaced := ensureAssuranceTierConfig([]byte(original))
	if replaced {
		t.Fatalf("did not expect an existing tier to be replaced")
	}
	if !strings.Contains(string(updated), "assurance_tier: verified") {
		t.Fatalf("expected verified tier to be appended, got %s", string(updated))
	}
	if !strings.HasSuffix(string(updated), "assurance_tier: verified\n") {
		t.Fatalf("expected tier line appended at end, got %s", string(updated))
	}
}

func TestEnsureAssuranceTierConfigPreservesInlineComment(t *testing.T) {
	original := "schema_version: 1\nassurance_tier: plain # keep me\n"
	updated, replaced := ensureAssuranceTierConfig([]byte(original))
	if !replaced {
		t.Fatalf("expected replacement")
	}
	text := string(updated)
	if !strings.Contains(text, "assurance_tier: verified # keep me") {
		t.Fatalf("expected inline comment preserved, got %q", text)
	}
}

func TestEnsureAssuranceTierConfigRewritesSpacedColonKey(t *testing.T) {
	original := "schema_version: 1\nassurance_tier : plain\n"
	updated, replaced := ensureAssuranceTierConfig([]byte(original))
	if !replaced {
		t.Fatalf("expected replacement for spaced-colon key")
	}
	text := string(updated)
	if strings.Count(text, "assurance_tier") != 1 {
		t.Fatalf("expected single assurance_tier key, got %q", text)
	}
	if !strings.Contains(text, "assurance_tier : verified") {
		t.Fatalf("expected rewritten spaced-colon key, got %q", text)
	}
}

func TestEnsureAssuranceTierConfigPreservesCRLF(t *testing.T) {
	original := "schema_version: 1\r\nsource:\r\n  type: embedded\r\n"
	updated, replaced := ensureAssuranceTierConfig([]byte(original))
	if replaced {
		t.Fatalf("did not expect replacement")
	}
	text := string(updated)
	if !strings.Contains(text, "\r\nassurance_tier: verified\r\n") {
		t.Fatalf("expected CRLF appended tier line, got %q", text)
	}
}

func TestParseAssuranceEnableArgsSuccess(t *testing.T) {
	cases := []struct {
		name         string
		args         []string
		wantRoot     string
		wantExplicit bool
	}{
		{"verified only", []string{"verified"}, ".", false},
		{"verified path", []string{"verified", "./proj"}, "./proj", true},
		{"path flag", []string{"--path", "./proj", "verified"}, "./proj", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAssuranceEnableArgs(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.root != tc.wantRoot {
				t.Fatalf("unexpected root %q, want %q", got.root, tc.wantRoot)
			}
			if got.explicitRoot != tc.wantExplicit {
				t.Fatalf("unexpected explicit flag %v, want %v", got.explicitRoot, tc.wantExplicit)
			}
		})
	}
}

func TestParseAssuranceEnableArgsErrors(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"missing verified", []string{"--path", "./proj"}},
		{"unknown flag", []string{"--foo", "verified"}},
		{"extra positional", []string{"verified", "./proj", "./extra"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseAssuranceEnableArgs(tc.args); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestRunAssuranceParsesMachineFlagsBeforeSubcommand(t *testing.T) {
	root := t.TempDir()
	_ = writeAssuranceConfigFixture(t, root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "--json", "enable", "verified", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"command\":\"assurance enable\"") {
		t.Fatalf("expected json output for assurance enable, got %q", stdout.String())
	}
}

func TestRunAssuranceEnableExplainAddsExplainLines(t *testing.T) {
	root := t.TempDir()
	_ = writeAssuranceConfigFixture(t, root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "enable", "verified", "--path", root, "--explain"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if fields["explain_scope"] == "" {
		t.Fatalf("expected explain output, got %q", stdout.String())
	}
}

func TestRunAssuranceEnablePreservesConfig(t *testing.T) {
	root := t.TempDir()
	configPath := writeAssuranceConfigFixture(t, root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "enable", "verified", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success, got %d (%s)", code, stderr.String())
	}
	assertVerifiedConfigPreserved(t, configPath)
	assertBaselineEnvelope(t, filepath.Join(root, "assurance", "baseline.yaml"))
}

func writeAssuranceConfigFixture(t *testing.T, root string) string {
	t.Helper()
	content := "# top comment\nschema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n# tail comment\n"
	configPath := filepath.Join(root, "runecontext.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir embedded source path: %v", err)
	}
	return configPath
}

func TestRunAssuranceEnableUsesNearestAncestorDiscovery(t *testing.T) {
	root := writeDiscoveryFixtureProject(t)
	nested := filepath.Join(root, "internal")
	t.Chdir(nested)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "enable", "verified"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success, got %d (%s)", code, stderr.String())
	}
	configPath := filepath.Join(root, "runecontext.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "assurance_tier: verified") {
		t.Fatalf("tier not updated: %s", string(data))
	}
	if _, err := os.Stat(filepath.Join(root, "assurance", "baseline.yaml")); err != nil {
		t.Fatalf("expected baseline at discovered project root: %v", err)
	}
}

func writeDiscoveryFixtureProject(t *testing.T) string {
	t.Helper()
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	projectRoot, err := os.MkdirTemp(repoRoot, "assurance-discovery-*")
	if err != nil {
		t.Fatalf("mkdir temp project root: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(projectRoot) })
	root := filepath.Join(projectRoot, "packages", "service")
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: service-context\n"
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "service-context"), 0o755); err != nil {
		t.Fatalf("mkdir embedded source: %v", err)
	}
	return root
}

func TestRunAssuranceEnableFailsWhenExistingBaselineUnreadable(t *testing.T) {
	root := t.TempDir()
	_ = writeAssuranceConfigFixture(t, root)
	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	if err := os.MkdirAll(baselinePath, 0o755); err != nil {
		t.Fatalf("make unreadable baseline path: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "enable", "verified", "--path", root}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit for unreadable existing baseline, got %d (%s)", code, stderr.String())
	}
}

func TestEmitAssuranceEnableErrorJSONModeSuppressesWarning(t *testing.T) {
	var stderr bytes.Buffer
	machine := machineOptions{jsonOutput: true}
	err := &assuranceEnableError{err: fmt.Errorf("write failed"), rollbackErr: fmt.Errorf("rollback failed")}
	emitAssuranceEnableError(&stderr, machine, "/tmp/project", err)

	text := stderr.String()
	if strings.Contains(text, "Warning:") {
		t.Fatalf("expected no human warning in json mode, got %q", text)
	}
	if !strings.Contains(text, "rollback_error") {
		t.Fatalf("expected rollback error in structured output, got %q", text)
	}
}

func TestSourceSnapshotFieldsUsesExpectCommitFallback(t *testing.T) {
	rootCfg := map[string]any{
		"source": map[string]any{
			"type":          "git",
			"expect_commit": "1234567890abcdef1234567890abcdef12345678",
		},
	}
	commit, posture := sourceSnapshotFields("", rootCfg, nil)
	if commit != "1234567890abcdef1234567890abcdef12345678" {
		t.Fatalf("unexpected adoption commit %q", commit)
	}
	if posture != "git" {
		t.Fatalf("unexpected source posture %q", posture)
	}
}

func TestSourceSnapshotFieldsSynthesizesEmbeddedAdoptionCommit(t *testing.T) {
	rootCfg := map[string]any{
		"source": map[string]any{
			"type": "embedded",
			"path": "runecontext",
		},
	}
	commit, posture := sourceSnapshotFields("", rootCfg, nil)
	if !isCanonicalLowerHex40(commit) {
		t.Fatalf("expected canonical synthetic adoption commit, got %q", commit)
	}
	if posture != "embedded" {
		t.Fatalf("unexpected source posture %q", posture)
	}
}

func assertVerifiedConfigPreserved(t *testing.T, configPath string) {
	t.Helper()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "# tail comment") {
		t.Fatalf("comment lost: %s", string(data))
	}
	if !strings.Contains(string(data), "assurance_tier: verified") {
		t.Fatalf("tier not updated: %s", string(data))
	}
}

func assertBaselineEnvelope(t *testing.T, baselinePath string) {
	t.Helper()
	baselineData, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var env contracts.AssuranceEnvelope
	if err := yaml.Unmarshal(baselineData, &env); err != nil {
		t.Fatalf("parse baseline: %v", err)
	}
	if env.Kind != "baseline" {
		t.Fatalf("unexpected kind %q", env.Kind)
	}
	if env.Canonicalization != "runecontext-canonical-json-v1" {
		t.Fatalf("unexpected canonicalization %q", env.Canonicalization)
	}
	if _, ok := env.Value.(map[string]any); !ok {
		t.Fatalf("expected map value, got %#v", env.Value)
	}
}
