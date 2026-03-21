package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func skipIfSymlinkUnsupported(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" || os.IsPermission(err) {
			t.Skipf("symlink not supported: %v", err)
		}
		t.Fatalf("create symlink: %v", err)
	}
}

func TestCloneFileEntryCountsSymlinkTowardsLimits(t *testing.T) {
	skipIfSymlinkUnsupported(t)
	srcRoot := t.TempDir()
	targetFile := filepath.Join(srcRoot, "origin.txt")
	if err := os.WriteFile(targetFile, []byte("payload"), 0o644); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	symlinkPath := filepath.Join(srcRoot, "link")
	if err := os.Symlink("origin.txt", symlinkPath); err != nil {
		t.Fatalf("create relative symlink: %v", err)
	}
	targetRoot := t.TempDir()
	state := &snapshotState{}
	limits := snapshotLimits{MaxFiles: 1, MaxBytes: 1 << 20}
	if err := cloneFileEntry(state, limits, srcRoot, symlinkPath, filepath.Join(targetRoot, "link")); err != nil {
		t.Fatalf("clone symlink: %v", err)
	}
	if state.files != 1 {
		t.Fatalf("expected symlink to count as 1 file, got %d", state.files)
	}
	if state.bytes != 0 {
		t.Fatalf("expected symlink size 0, got %d", state.bytes)
	}
	if err := cloneFileEntry(state, limits, srcRoot, targetFile, filepath.Join(targetRoot, "origin")); err == nil {
		t.Fatalf("expected file-count limit error after symlink counted, got nil")
	} else if !strings.Contains(err.Error(), "dry-run clone exceeds maximum file count") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneSymlinkRejectsOutsideRoot(t *testing.T) {
	skipIfSymlinkUnsupported(t)
	root := t.TempDir()
	parent := filepath.Dir(root)
	outside := filepath.Join(parent, fmt.Sprintf("outside-%s", t.Name()))
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("create outside dir: %v", err)
	}
	targetFile := filepath.Join(outside, "target.txt")
	if err := os.WriteFile(targetFile, []byte("far"), 0o644); err != nil {
		t.Fatalf("write outside target: %v", err)
	}
	link := filepath.Join(root, "foreign")
	linkTarget := filepath.Join("..", filepath.Base(outside), "target.txt")
	if err := os.Symlink(linkTarget, link); err != nil {
		t.Fatalf("create escaping symlink: %v", err)
	}
	if err := cloneSymlink(root, link, filepath.Join(t.TempDir(), "dest")); err == nil {
		t.Fatalf("expected escaping symlink to fail")
	} else if !strings.Contains(err.Error(), "resolving outside project root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloneFileEntryRejectsUnsupportedFileType(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("named pipes unsupported on Windows")
	}
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
