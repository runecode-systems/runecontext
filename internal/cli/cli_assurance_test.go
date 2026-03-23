package cli

import (
	"bytes"
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
	content := "# top comment\nschema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n# tail comment\n"
	configPath := filepath.Join(root, "runecontext.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
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
