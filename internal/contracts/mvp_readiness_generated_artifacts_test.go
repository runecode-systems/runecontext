package contracts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMVPReadinessGeneratedArtifactsRemainDerivedWhenAbsent(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate fixture without generated artifacts: %v", err)
	}
	defer index.Close()
	for _, rel := range []string{generatedManifestRelativePath, generatedChangesIndexRelativePath, generatedBundlesIndexRelativePath} {
		if _, err := os.Stat(filepath.Join(index.ContentRoot, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be absent before generation, err=%v", rel, err)
		}
	}
}
