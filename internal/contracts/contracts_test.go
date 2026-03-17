package contracts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	root := fixturePath(t, "schema-contracts")
	validRootConfig := readFixture(t, filepath.Join(root, "valid-runecontext-with-extensions-optin.yaml"))

	validCases := map[string]string{
		"valid-runecontext-no-extensions.yaml":         "runecontext.schema.json",
		"valid-runecontext-with-extensions-optin.yaml": "runecontext.schema.json",
		"valid-git-source-signed-tag.yaml":             "runecontext.schema.json",
		"valid-bundle-closed-schema.yaml":              "bundle.schema.json",
		"valid-bundle-with-extensions.yaml":            "bundle.schema.json",
		"valid-change-status.yaml":                     "change-status.schema.json",
		"valid-custom-type.yaml":                       "change-status.schema.json",
		"valid-superseded-change.yaml":                 "change-status.schema.json",
		"valid-context-pack.yaml":                      "context-pack.schema.json",
	}
	for name, schema := range validCases {
		t.Run(name, func(t *testing.T) {
			data := readFixture(t, filepath.Join(root, name))
			if err := v.ValidateYAMLFile(schema, name, data); err != nil {
				t.Fatalf("expected fixture to validate: %v", err)
			}
			if name == "valid-bundle-with-extensions.yaml" {
				if err := v.ValidateExtensionOptIn("runecontext.yaml", validRootConfig, name, data); err != nil {
					t.Fatalf("expected extension opt-in fixture to validate: %v", err)
				}
			}
		})
	}

	rejectCases := map[string]string{
		"reject-unknown-field-runecontext.yaml":  "runecontext.schema.json",
		"reject-unknown-schema-version.yaml":     "runecontext.schema.json",
		"reject-bad-extension-key.yaml":          "change-status.schema.json",
		"reject-context-pack-unknown-field.yaml": "context-pack.schema.json",
		"reject-yaml-anchors-aliases.yaml":       "change-status.schema.json",
		"reject-yaml-custom-tag.yaml":            "change-status.schema.json",
		"reject-yaml-flow-style.yaml":            "bundle.schema.json",
		"reject-yaml-multiline-string.yaml":      "bundle.schema.json",
	}
	for name, schema := range rejectCases {
		t.Run(name, func(t *testing.T) {
			data := readFixture(t, filepath.Join(root, name))
			if err := v.ValidateYAMLFile(schema, name, data); err == nil {
				t.Fatalf("expected fixture to fail validation")
			}
		})
	}

	t.Run("reject-extensions-without-optin.yaml", func(t *testing.T) {
		rootData := readFixture(t, filepath.Join(root, "valid-runecontext-no-extensions.yaml"))
		artifactData := readFixture(t, filepath.Join(root, "reject-extensions-without-optin.yaml"))
		if err := v.ValidateYAMLFile("change-status.schema.json", "reject-extensions-without-optin.yaml", artifactData); err != nil {
			t.Fatalf("expected standalone schema validation to pass before project-level extension enforcement: %v", err)
		}
		if err := v.ValidateExtensionOptIn("runecontext.yaml", rootData, "reject-extensions-without-optin.yaml", artifactData); err == nil {
			t.Fatalf("expected project-level extension rejection")
		}
	})

	t.Run("reject-related-specs-wrong-type.yaml", func(t *testing.T) {
		data := readFixture(t, filepath.Join(root, "reject-related-specs-wrong-type.yaml"))
		if err := v.ValidateYAMLFile("change-status.schema.json", "reject-related-specs-wrong-type.yaml", data); err == nil {
			t.Fatalf("expected wrong-type reference fixture to fail schema validation")
		}
	})
}

func TestProposalMarkdownFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	valid := readFixture(t, fixturePath(t, "markdown-contracts", "valid-proposal.md"))
	if err := v.ValidateProposalMarkdown("valid-proposal.md", valid); err != nil {
		t.Fatalf("expected valid proposal fixture: %v", err)
	}

	for _, name := range []string{"reject-proposal-out-of-order.md", "reject-proposal-empty-section.md"} {
		t.Run(name, func(t *testing.T) {
			data := readFixture(t, fixturePath(t, "markdown-contracts", name))
			if err := v.ValidateProposalMarkdown(name, data); err == nil {
				t.Fatalf("expected %s to fail", name)
			}
		})
	}
}

func TestStandardsMarkdownFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	valid := readFixture(t, fixturePath(t, "markdown-contracts", "valid-standards.md"))
	if err := v.ValidateStandardsMarkdown("valid-standards.md", valid); err != nil {
		t.Fatalf("expected valid standards fixture: %v", err)
	}

	for _, name := range []string{"reject-standards-missing-applicable.md", "reject-standards-out-of-order.md"} {
		t.Run(name, func(t *testing.T) {
			data := readFixture(t, fixturePath(t, "markdown-contracts", name))
			if err := v.ValidateStandardsMarkdown(name, data); err == nil {
				t.Fatalf("expected %s to fail", name)
			}
		})
	}
}

func TestSpecAndDecisionFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	validSpec := readFixture(t, fixturePath(t, "traceability", "valid-project", "runecontext", "specs", "auth-gateway.md"))
	if _, err := v.ParseSpec("fixtures/traceability/valid-project/runecontext/specs/auth-gateway.md", validSpec); err != nil {
		t.Fatalf("expected valid spec fixture: %v", err)
	}

	validDecision := readFixture(t, fixturePath(t, "traceability", "valid-project", "runecontext", "decisions", "DEC-0001-trust-boundary-model.md"))
	if _, err := v.ParseDecision("fixtures/traceability/valid-project/runecontext/decisions/DEC-0001-trust-boundary-model.md", validDecision); err != nil {
		t.Fatalf("expected valid decision fixture: %v", err)
	}

	badSpec := readFixture(t, fixturePath(t, "traceability", "reject-spec-id-mismatch", "runecontext", "specs", "auth-gateway.md"))
	if _, err := v.ParseSpec("fixtures/traceability/reject-spec-id-mismatch/runecontext/specs/auth-gateway.md", badSpec); err == nil {
		t.Fatalf("expected bad spec fixture to fail")
	}
}

func TestTraceabilityProjectFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	validRoot := fixturePath(t, "traceability", "valid-project")
	if _, err := v.ValidateProject(validRoot); err != nil {
		t.Fatalf("expected valid traceability project: %v", err)
	}
	validCustomRoot := fixturePath(t, "traceability", "valid-project-custom-root")
	if _, err := v.ValidateProject(validCustomRoot); err != nil {
		t.Fatalf("expected valid custom-root traceability project: %v", err)
	}

	rejectCases := []struct {
		name       string
		fixtureDir string
		contains   string
	}{
		{name: "reject decision missing change", fixtureDir: "reject-decision-missing-change", contains: "missing change"},
		{name: "reject change missing related spec", fixtureDir: "reject-change-missing-related-spec", contains: "missing artifact"},
		{name: "reject extensions without opt-in", fixtureDir: "reject-extensions-without-optin", contains: "extensions require `allow_extensions: true`"},
		{name: "reject bundle invalid", fixtureDir: "reject-bundle-invalid", contains: "missing property 'includes'"},
		{name: "reject proposal invalid", fixtureDir: "reject-proposal-invalid", contains: "appears where \"Problem\" is required"},
		{name: "reject spec ancestor path collision", fixtureDir: "reject-spec-ancestor-path-collision", contains: "must match path-relative stem"},
	}
	for _, tc := range rejectCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := v.ValidateProject(fixturePath(t, "traceability", tc.fixtureDir))
			if err == nil {
				t.Fatalf("expected validation failure")
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("expected error to contain %q, got %v", tc.contains, err)
			}
		})
	}
}

func TestValidateProjectRejectsSpecSymlinkEscape(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := t.TempDir()
	contentRoot := filepath.Join(projectRoot, "runecontext")
	if err := os.MkdirAll(filepath.Join(contentRoot, "changes", "CHG-2026-001-a3f2-auth-gateway"), 0o755); err != nil {
		t.Fatalf("mkdir change dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(contentRoot, "specs"), 0o755); err != nil {
		t.Fatalf("mkdir specs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "runecontext.yaml"), []byte("schema_version: 1\nrunecontext_version: 0.1.0-alpha.2\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"), 0o644); err != nil {
		t.Fatalf("write root config: %v", err)
	}
	changeDir := filepath.Join(contentRoot, "changes", "CHG-2026-001-a3f2-auth-gateway")
	if err := os.WriteFile(filepath.Join(changeDir, "status.yaml"), []byte("schema_version: 1\nid: CHG-2026-001-a3f2-auth-gateway\ntitle: Test\nstatus: proposed\ntype: feature\nsize: small\ncontext_bundles: []\nrelated_specs: []\nrelated_decisions: []\nrelated_changes: []\ndepends_on: []\ninformed_by: []\nsupersedes: []\nsuperseded_by: []\ncreated_at: \"2026-03-17\"\nclosed_at: null\nverification_status: pending\npromotion_assessment:\n  status: pending\n  suggested_targets: []\n"), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "proposal.md"), []byte("## Summary\n\nN/A\n\n## Problem\n\nN/A\n\n## Proposed Change\n\nAdd a test.\n\n## Why Now\n\nN/A\n\n## Assumptions\n\nN/A\n\n## Out of Scope\n\nN/A\n\n## Impact\n\nN/A\n"), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(changeDir, "standards.md"), []byte("## Applicable Standards\n\n- `standards/global/base.md`\n"), 0o644); err != nil {
		t.Fatalf("write standards: %v", err)
	}
	outside := filepath.Join(projectRoot, "outside-spec.md")
	if err := os.WriteFile(outside, []byte("---\nschema_version: 1\nid: auth-gateway\ntitle: Bad\noriginating_changes: []\nrevised_by_changes: []\n---\n\n# Bad\n"), 0o644); err != nil {
		t.Fatalf("write outside spec: %v", err)
	}
	if err := tryCreateSymlink(filepath.Join("..", "..", "outside-spec.md"), filepath.Join(contentRoot, "specs", "auth-gateway.md")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	_, err := v.ValidateProject(projectRoot)
	if err == nil || !strings.Contains(err.Error(), "escapes the selected project subtree") {
		t.Fatalf("expected spec symlink escape to fail, got %v", err)
	}
}

func TestWalkProjectFilesAllowsSymlinkedRootDirectory(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "real-specs")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir real root: %v", err)
	}
	file := filepath.Join(target, "example.md")
	if err := os.WriteFile(file, []byte("# Example\n"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	linked := filepath.Join(root, "specs")
	if err := tryCreateSymlink("real-specs", linked); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	paths := make([]string, 0)
	if err := walkProjectFiles(linked, func(path string) error {
		paths = append(paths, filepath.Base(path))
		return nil
	}); err != nil {
		t.Fatalf("expected symlinked root directory to be walkable: %v", err)
	}
	if len(paths) != 1 || paths[0] != "example.md" {
		t.Fatalf("expected to visit example.md through symlinked root, got %v", paths)
	}
}

func TestParseSpecAllowsClosingFrontmatterDelimiterAtEOF(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	data := []byte("---\nschema_version: 1\nid: auth-gateway\ntitle: Auth Gateway\noriginating_changes:\n  - CHG-2026-001-a3f2-auth-gateway\nrevised_by_changes: []\n---\n# Auth Gateway")
	if _, err := v.ParseSpec("fixtures/specs/auth-gateway.md", data); err != nil {
		t.Fatalf("expected EOF frontmatter delimiter form to validate: %v", err)
	}
	dataNoTrailingNewline := []byte("---\nschema_version: 1\nid: auth-gateway\ntitle: Auth Gateway\noriginating_changes:\n  - CHG-2026-001-a3f2-auth-gateway\nrevised_by_changes: []\n---")
	if _, err := v.ParseSpec("fixtures/specs/auth-gateway.md", dataNoTrailingNewline); err != nil {
		t.Fatalf("expected closing delimiter at EOF without body to parse: %v", err)
	}
}

func TestParseSpecRejectsNonDelimiterFrontmatterLine(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	data := []byte("---\nschema_version: 1\nid: auth-gateway\ntitle: Auth Gateway\noriginating_changes:\n  - CHG-2026-001-a3f2-auth-gateway\nrevised_by_changes: []\n---oops\n# Auth Gateway")
	if _, err := v.ParseSpec("fixtures/specs/auth-gateway.md", data); err == nil {
		t.Fatal("expected malformed closing delimiter to fail")
	}
}

func TestResolveContentRoot(t *testing.T) {
	projectRoot := fixturePath(t, "traceability", "valid-project-custom-root")
	rootData := readFixture(t, filepath.Join(projectRoot, "runecontext.yaml"))
	resolution, err := resolveContentRoot(projectRoot, rootData)
	if err != nil {
		t.Fatalf("expected content root to resolve: %v", err)
	}
	defer resolution.Close()
	contentRoot := resolution.MaterializedRoot()
	expected := filepath.Join(projectRoot, "docs-context")
	if contentRoot != expected {
		t.Fatalf("expected content root %q, got %q", expected, contentRoot)
	}
}

func schemaRoot(t *testing.T) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "schemas")
}

func fixturePath(t *testing.T, elems ...string) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	parts := append([]string{root, "fixtures"}, elems...)
	return filepath.Join(parts...)
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return data
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			return "", os.ErrNotExist
		}
		wd = next
	}
}
