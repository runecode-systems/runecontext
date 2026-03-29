package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRootForTests() (string, error) {
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

func releaseMetadataVersionForTests(t *testing.T) string {
	t.Helper()
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	version, err := ReadReleaseMetadataVersion(root)
	if err != nil {
		t.Fatalf("read release metadata version: %v", err)
	}
	return strings.TrimPrefix(version, "v")
}

func withReleaseMetadataVersionForTests(t *testing.T, fn func()) {
	t.Helper()
	original := runecontextVersion
	t.Cleanup(func() { runecontextVersion = original })
	runecontextVersion = "v" + releaseMetadataVersionForTests(t)
	fn()
}
