package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMVPReadinessReferenceFixtureMatrix(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	runReferenceFixtureMatrixEmbedded(t, v)
	runReferenceFixtureMatrixVerified(t, v)
	runReferenceFixtureMatrixMonorepo(t, v)
	runReferenceFixtureMatrixLinkedByCommit(t, v)
	runReferenceFixtureMatrixLinkedBySignedTag(t, v)
}

func runReferenceFixtureMatrixEmbedded(t *testing.T, v *Validator) {
	t.Helper()
	root := fixturePath(t, "reference-projects", "embedded")
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate embedded fixture: %v", err)
	}
	defer index.Close()
	if got, want := string(index.Resolution.SourceMode), string(SourceModeEmbedded); got != want {
		t.Fatalf("expected embedded source mode %q, got %q", want, got)
	}
}

func runReferenceFixtureMatrixVerified(t *testing.T, v *Validator) {
	t.Helper()
	root := fixturePath(t, "reference-projects", "verified")
	loaded, err := v.LoadProject(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("load verified fixture: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("validate verified fixture: %v", err)
	}
	defer index.Close()
	if got, want := fmt.Sprint(loaded.RootConfig["assurance_tier"]), AssuranceTierVerified; got != want {
		t.Fatalf("expected assurance tier %q, got %q", want, got)
	}
	if _, err := os.Stat(filepath.Join(root, "assurance", "baseline.yaml")); err != nil {
		t.Fatalf("expected verified baseline fixture: %v", err)
	}
}

func runReferenceFixtureMatrixMonorepo(t *testing.T, v *Validator) {
	t.Helper()
	root := t.TempDir()
	copyDirForTest(t, fixturePath(t, "reference-projects", "monorepo"), root)
	nestedStart := filepath.Join(root, "packages", "service", "app", "src")
	if err := os.MkdirAll(nestedStart, 0o755); err != nil {
		t.Fatalf("create nested start path: %v", err)
	}
	index, err := v.ValidateProjectWithOptions(nestedStart, ResolveOptions{ConfigDiscovery: ConfigDiscoveryNearestAncestor, ExecutionMode: ExecutionModeLocal})
	if err != nil {
		t.Fatalf("validate nested monorepo fixture: %v", err)
	}
	defer index.Close()
	if got := filepath.ToSlash(index.RootConfigPath); !strings.Contains(got, "/packages/service/runecontext.yaml") {
		t.Fatalf("expected nested selected config path, got %q", got)
	}
}

func runReferenceFixtureMatrixLinkedByCommit(t *testing.T, v *Validator) {
	t.Helper()
	repoDir, commit := createGitSourceRepo(t)
	projectRoot := materializeReferenceLinkedFixture(t, "linked-by-commit", map[string]string{
		"__GIT_URL__": repoDir,
		"__COMMIT__":  commit,
	})
	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("validate linked-by-commit fixture: %v", err)
	}
	defer index.Close()
	if got, want := index.Resolution.ResolvedCommit, commit; got != want {
		t.Fatalf("expected resolved commit %q, got %q", want, got)
	}
}

func runReferenceFixtureMatrixLinkedBySignedTag(t *testing.T, v *Validator) {
	t.Helper()
	repoDir, details := createSignedGitSourceRepo(t)
	projectRoot := materializeReferenceLinkedFixture(t, "linked-by-signed-tag", map[string]string{
		"__GIT_URL__":    repoDir,
		"__SIGNED_TAG__": details.SignedTagName,
		"__COMMIT__":     details.Commit,
	})
	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal, GitTrust: GitTrustInputs{SignedTagVerifier: newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)}})
	if err != nil {
		t.Fatalf("load linked-by-signed-tag fixture: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("validate linked-by-signed-tag fixture: %v", err)
	}
	defer index.Close()
	if got, want := index.Resolution.VerificationPosture, VerificationPostureVerifiedSignedTag; got != want {
		t.Fatalf("expected verification posture %q, got %q", want, got)
	}
}
