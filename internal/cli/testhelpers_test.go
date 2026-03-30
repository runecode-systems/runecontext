package cli

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var runecontextVersionTestMu sync.Mutex

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
	setRunecontextVersionForTests(t, "v"+releaseMetadataVersionForTests(t))
	fn()
}

func setRunecontextVersionForTests(t *testing.T, value string) {
	t.Helper()
	runecontextVersionTestMu.Lock()
	original := runecontextVersion
	runecontextVersion = value
	t.Cleanup(func() {
		runecontextVersion = original
		runecontextVersionTestMu.Unlock()
	})
}
