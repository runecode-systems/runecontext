package contracts

import (
	"fmt"
	"testing"
)

func TestSourceResolutionGitSignedTagAcceptsFullyQualifiedTagRef(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedGitSourceRepo(t)
	verifier := newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: refs/tags/%s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))

	loaded, err := v.LoadProject(projectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
		GitTrust:        GitTrustInputs{SignedTagVerifier: verifier},
	})
	if err != nil {
		t.Fatalf("expected fully-qualified signed_tag to verify successfully: %v", err)
	}
	defer loaded.Close()
}
