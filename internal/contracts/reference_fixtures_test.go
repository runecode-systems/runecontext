package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferenceFixtureEmbeddedValidates(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	root := fixturePath(t, "reference-projects", "embedded")
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected embedded reference fixture to load: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("expected embedded reference fixture to validate: %v", err)
	}
	defer index.Close()
	if got, want := fmt.Sprint(loaded.RootConfig["assurance_tier"]), "plain"; got != want {
		t.Fatalf("expected assurance tier %q, got %q", want, got)
	}
}

func TestReferenceFixtureVerifiedValidates(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	root := fixturePath(t, "reference-projects", "verified")
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected verified reference fixture to load: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("expected verified reference fixture to validate: %v", err)
	}
	defer index.Close()
	if got, want := fmt.Sprint(loaded.RootConfig["assurance_tier"]), AssuranceTierVerified; got != want {
		t.Fatalf("expected assurance tier %q, got %q", want, got)
	}
}

func TestReferenceFixtureMonorepoNestedDiscovery(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	root := fixturePath(t, "reference-projects", "monorepo")
	start := filepath.Join(root, "packages", "service")
	index, err := v.ValidateProjectWithOptions(start, ResolveOptions{ConfigDiscovery: ConfigDiscoveryNearestAncestor, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("expected nested monorepo reference fixture to validate: %v", err)
	}
	defer index.Close()
	if !strings.Contains(filepath.ToSlash(index.RootConfigPath), "packages/service/runecontext.yaml") {
		t.Fatalf("expected nested root config path, got %q", index.RootConfigPath)
	}
}

func TestReferenceFixtureLinkedByCommitValidates(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := materializeReferenceLinkedFixture(t, "linked-by-commit", map[string]string{
		"__GIT_URL__": repoDir,
		"__COMMIT__":  commit,
	})
	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected linked-by-commit fixture to validate: %v", err)
	}
	defer index.Close()
	if index.Resolution.ResolvedCommit != commit {
		t.Fatalf("expected resolved commit %q, got %q", commit, index.Resolution.ResolvedCommit)
	}
}

func TestReferenceFixtureLinkedBySignedTagValidates(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	projectRoot := materializeReferenceLinkedFixture(t, "linked-by-signed-tag", map[string]string{
		"__GIT_URL__":    repoDir,
		"__SIGNED_TAG__": details.SignedTagName,
		"__COMMIT__":     details.Commit,
	})
	index, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal, GitTrust: GitTrustInputs{SignedTagVerifier: newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)}})
	if err != nil {
		t.Fatalf("expected linked-by-signed-tag fixture to validate: %v", err)
	}
	defer index.Close()
	if got, want := string(index.Resolution.VerificationPosture), string(VerificationPostureVerifiedSignedTag); got != want {
		t.Fatalf("expected verification posture %q, got %q", want, got)
	}
}

func materializeReferenceLinkedFixture(t *testing.T, fixtureName string, replacements map[string]string) string {
	t.Helper()
	targetRoot := t.TempDir()
	templatePath := fixturePath(t, "reference-projects", fixtureName, "runecontext.yaml.tmpl")
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
