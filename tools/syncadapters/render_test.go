package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOutputRejectsUnsafeTargets(t *testing.T) {
	root := t.TempDir()
	for _, output := range []string{string(filepath.Separator), root, filepath.Join(root, "..", "outside")} {
		output := output
		t.Run(output, func(t *testing.T) {
			if _, err := resolveOutput(root, output); err == nil {
				t.Fatalf("expected resolveOutput to reject %q", output)
			}
		})
	}
}

func TestResolveOutputAllowsRepositoryDescendants(t *testing.T) {
	root := t.TempDir()
	assertRepositoryDescendantOutput(t, root, "build/generated/adapters")
	assertRepositoryDescendantOutput(t, root, filepath.Join(root, "build", "generated", "adapters"))
}

func TestResolveOutputRejectsSymlinkedOutputAncestor(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "build")); err != nil {
		if os.IsPermission(err) {
			t.Skipf("symlink creation not permitted: %v", err)
		}
		t.Fatalf("create symlinked output ancestor: %v", err)
	}
	if _, err := resolveOutput(root, "build/generated/adapters"); err == nil {
		t.Fatal("expected resolveOutput to reject symlinked output ancestor")
	}
}

func assertRepositoryDescendantOutput(t *testing.T, root, output string) {
	t.Helper()
	t.Run(output, func(t *testing.T) {
		got, err := resolveOutput(root, output)
		if err != nil {
			t.Fatalf("resolveOutput returned error: %v", err)
		}
		rel, err := filepath.Rel(root, got)
		if err != nil {
			t.Fatalf("compute relative path: %v", err)
		}
		if rel == "." || rel == ".." || rel == "" {
			t.Fatalf("expected repository descendant output, got %q", got)
		}
	})
}
