//go:build !windows
// +build !windows

package cli

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestCloneFileEntryRejectsUnsupportedFileType(t *testing.T) {
	srcRoot := t.TempDir()
	special := filepath.Join(srcRoot, "pipe")
	if err := syscall.Mkfifo(special, 0o644); err != nil {
		if err == syscall.ENOSYS || err == syscall.EPERM {
			t.Skipf("mkfifo not available: %v", err)
		}
		t.Fatalf("mkfifo: %v", err)
	}
	if err := cloneFileEntry(&snapshotState{}, snapshotLimits{}, srcRoot, special, filepath.Join(t.TempDir(), "pipe")); err == nil {
		t.Fatalf("expected unsupported file type error")
	} else if !strings.Contains(err.Error(), "rejects unsupported file type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
