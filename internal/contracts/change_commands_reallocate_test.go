package contracts

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReallocateChangeUpdatesLocalMarkdownReferences(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	appendReallocationReferences(t, root, result.ID)
	reallocated := mustReallocateChange(t, v, root, result.ID, []byte{0xcc, 0xdd})
	assertReallocationMetadata(t, result.ID, reallocated)
	assertReallocatedProposalReferences(t, root, result.ID, reallocated.ID)
	assertValidatedWorkflowProject(t, v, root)
	if len(reallocated.ChangedFiles) == 0 {
		t.Fatalf("expected changed files to be reported")
	}
}

func appendReallocationReferences(t *testing.T, root, changeID string) {
	t.Helper()
	proposalPath := filepath.Join(root, "runecontext", "changes", changeID, "proposal.md")
	rewriteFile(t, proposalPath, func(text string) string {
		return text + "\nSee changes/" + changeID + " and changes/" + changeID + "/proposal.md#summary and changes/" + changeID + "/standards.md#applicable-standards for the local review flow.\n"
	})
}

func assertReallocationMetadata(t *testing.T, oldID string, result *ChangeReallocationResult) {
	t.Helper()
	if result.OldID != oldID || result.ID != "CHG-2026-002-ccdd-add-cache-invalidation" {
		t.Fatalf("unexpected reallocation ids: %#v", result)
	}
	if result.RewrittenReferenceCount != 3 || len(result.Warnings) != 0 {
		t.Fatalf("unexpected reallocation summary: %#v", result)
	}
}

func assertReallocatedProposalReferences(t *testing.T, root, oldID, newID string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, "runecontext", "changes", oldID)); !os.IsNotExist(err) {
		t.Fatalf("expected old change path to disappear, got err=%v", err)
	}
	newChangeDir := filepath.Join(root, "runecontext", "changes", newID)
	statusText := string(mustReadBytes(t, filepath.Join(newChangeDir, "status.yaml")))
	if !strings.Contains(statusText, "id: "+newID) {
		t.Fatalf("expected status ID rewrite, got:\n%s", statusText)
	}
	proposalText := strings.ReplaceAll(string(mustReadBytes(t, filepath.Join(newChangeDir, "proposal.md"))), "\r\n", "\n")
	if strings.Contains(proposalText, oldID) {
		t.Fatalf("expected old change ID refs to be rewritten, got:\n%s", proposalText)
	}
	for _, want := range []string{"changes/" + newID + "/proposal.md#summary", "changes/" + newID + " and", "changes/" + newID + "/standards.md#applicable-standards"} {
		if !strings.Contains(proposalText, want) {
			t.Fatalf("expected rewritten proposal to contain %q, got:\n%s", want, proposalText)
		}
	}
}

func TestReallocateChangeRejectsExternalReferences(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xaa, 0xbb})})
	if err == nil || !strings.Contains(err.Error(), "alpha.3 reallocation only rewrites local references inside the change") {
		t.Fatalf("expected external-reference rejection, got %v", err)
	}
	statusData := mustReadBytes(t, filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml"))
	if !strings.Contains(string(statusData), "id: CHG-2026-001-a3f2-auth-gateway") {
		t.Fatalf("expected failed reallocation to leave original status intact, got:\n%s", string(statusData))
	}
}

func TestReallocateChangeRejectsTerminalChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	mustCloseChange(t, v, root, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "terminal status") {
		t.Fatalf("expected terminal-status rejection, got %v", err)
	}
}

func TestReallocateChangeRejectsExistingBackupPath(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		t.Fatalf("mkdir backup path: %v", err)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "backup path") {
		t.Fatalf("expected backup-path rejection, got %v", err)
	}
}

func TestReallocateChangeRejectsSymlinksInChangeDirectory(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	changeDir := filepath.Join(root, "runecontext", "changes", result.ID)
	if err := tryCreateSymlink("proposal.md", filepath.Join(changeDir, "proposal-link.md")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err)
		}
		t.Fatalf("create symlink: %v", err)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "does not support symlinks") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestReallocateChangeRejectsSymlinkedDirectoryInChangeTree(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	createSymlinkedNotesDir(t, root, result.ID)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "does not support symlinks") {
		t.Fatalf("expected symlinked-directory rejection, got %v", err)
	}
}

func createSymlinkedNotesDir(t *testing.T, root, changeID string) {
	t.Helper()
	changeDir := filepath.Join(root, "runecontext", "changes", changeID)
	realNotes := filepath.Join(changeDir, "real-notes")
	if err := os.MkdirAll(realNotes, 0o755); err != nil {
		t.Fatalf("mkdir real-notes: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realNotes, "review.md"), []byte("## Review\n\nchanges/"+changeID+"/proposal.md#summary\n"), 0o644); err != nil {
		t.Fatalf("write nested review: %v", err)
	}
	if err := tryCreateSymlink("real-notes", filepath.Join(changeDir, "linked-notes")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err)
		}
		t.Fatalf("create symlinked dir: %v", err)
	}
}

func TestReallocateChangeRejectsSymlinkedRenameTargets(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	changesRoot := filepath.Clean(filepath.Join(root, "runecontext", "changes"))
	originalLstat := lstatPath
	t.Cleanup(func() { lstatPath = originalLstat })
	lstatPath = func(path string) (os.FileInfo, error) {
		if filepath.Clean(path) == changesRoot {
			return fakeFileInfo{name: filepath.Base(path), mode: os.ModeSymlink}, nil
		}
		return os.Lstat(path)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "symlinked targets") {
		t.Fatalf("expected rename-target symlink rejection, got %v", err)
	}
}

func TestReallocateChangeSurfacesRollbackFailures(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	originalRename, originalValidate := renamePath, validateProjectAfterChangeMutation
	t.Cleanup(func() { renamePath = originalRename; validateProjectAfterChangeMutation = originalValidate })
	validateProjectAfterChangeMutation = func(*Validator, string) (*ProjectIndex, error) { return nil, fmt.Errorf("forced validation failure") }
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	originalChangePath := filepath.Join(root, "runecontext", "changes", result.ID)
	renamePath = func(oldPath, newPath string) error {
		if oldPath == backupPath && newPath == originalChangePath {
			return fmt.Errorf("forced rollback rename failure")
		}
		return os.Rename(oldPath, newPath)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ReallocateChange(v, loaded, result.ID, ChangeReallocateOptions{Entropy: bytes.NewReader([]byte{0xcc, 0xdd})})
	if err == nil || !strings.Contains(err.Error(), "manual recovery may be required") || !strings.Contains(err.Error(), "forced validation failure") || !strings.Contains(err.Error(), "forced rollback rename failure") {
		t.Fatalf("expected rollback failure details, got %v", err)
	}
}

func TestReallocateChangeReturnsCleanupWarning(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	originalRemoveAll := removeAllPath
	t.Cleanup(func() { removeAllPath = originalRemoveAll })
	backupPath := filepath.Join(root, "runecontext", ".reallocate-"+result.ID+"-backup")
	removeAllPath = func(path string) error {
		if path == backupPath {
			return fmt.Errorf("forced cleanup failure")
		}
		return os.RemoveAll(path)
	}
	reallocated := mustReallocateChange(t, v, root, result.ID, []byte{0xcc, 0xdd})
	if len(reallocated.Warnings) != 1 || !strings.Contains(reallocated.Warnings[0], "forced cleanup failure") {
		t.Fatalf("expected cleanup warning, got %#v", reallocated.Warnings)
	}
}

func TestReallocateChangeRewritesNestedMarkdownFiles(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	nestedPath := writeNestedReviewFile(t, root, result.ID)
	reallocated := mustReallocateChange(t, v, root, result.ID, []byte{0xcc, 0xdd})
	rewritten := string(mustReadBytes(t, filepath.Join(root, "runecontext", "changes", reallocated.ID, "notes", "review.md")))
	if !strings.Contains(rewritten, "changes/"+reallocated.ID+"/proposal.md#summary") {
		t.Fatalf("expected nested markdown rewrite, got:\n%s", rewritten)
	}
	if !containsMutation(reallocated.ChangedFiles, filepath.ToSlash(filepath.Join("changes", reallocated.ID, "notes", "review.md")), "updated") {
		t.Fatalf("expected nested markdown file mutation, got %#v", reallocated.ChangedFiles)
	}
	_ = nestedPath
}

func writeNestedReviewFile(t *testing.T, root, changeID string) string {
	t.Helper()
	nestedDir := filepath.Join(root, "runecontext", "changes", changeID, "notes")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested notes: %v", err)
	}
	nestedPath := filepath.Join(nestedDir, "review.md")
	if err := os.WriteFile(nestedPath, []byte("## Review\n\nSee changes/"+changeID+"/proposal.md#summary for context.\n"), 0o644); err != nil {
		t.Fatalf("write nested markdown: %v", err)
	}
	return nestedPath
}

func TestRewriteMarkdownChangePathMentionsPreservesCRLFWhenUnchanged(t *testing.T) {
	input := []byte("## Summary\r\nNo change-path references here.\r\n")
	rewritten, count, err := rewriteMarkdownChangePathMentions(input, "changes/CHG-2026-001-a3f2-auth-gateway", "changes/CHG-2026-002-b4c3-auth-gateway")
	if err != nil {
		t.Fatalf("rewrite markdown change paths: %v", err)
	}
	if count != 0 || !bytes.Equal(rewritten, input) {
		t.Fatalf("expected unchanged bytes to preserve original line endings\nwant: %q\ngot:  %q", string(input), string(rewritten))
	}
}

func TestRewriteMarkdownChangePathMentionsPreservesCRLFWhenChanged(t *testing.T) {
	input := []byte("## Summary\r\nSee changes/CHG-2026-001-a3f2-auth-gateway/proposal.md#summary\r\n")
	rewritten, count, err := rewriteMarkdownChangePathMentions(input, "changes/CHG-2026-001-a3f2-auth-gateway", "changes/CHG-2026-002-b4c3-auth-gateway")
	if err != nil {
		t.Fatalf("rewrite markdown change paths: %v", err)
	}
	want := []byte("## Summary\r\nSee changes/CHG-2026-002-b4c3-auth-gateway/proposal.md#summary\r\n")
	if count != 1 || !bytes.Equal(rewritten, want) {
		t.Fatalf("expected rewritten bytes to preserve CRLF\nwant: %q\ngot:  %q", string(want), string(rewritten))
	}
}

func TestRewriteLiteralPathRootInTextUsesUTF8Boundaries(t *testing.T) {
	text := "preéchanges/CHG-2026-001-a3f2-auth-gateway and changes/CHG-2026-001-a3f2-auth-gateway and changes/CHG-2026-001-a3f2-auth-gatewayé"
	rewritten, count := rewriteLiteralPathRootInText(text, "changes/CHG-2026-001-a3f2-auth-gateway", "changes/CHG-2026-002-b4c3-auth-gateway")
	if count != 1 {
		t.Fatalf("expected exactly one UTF-8-safe rewrite, got %d", count)
	}
	for _, want := range []string{"preéchanges/CHG-2026-001-a3f2-auth-gateway", "changes/CHG-2026-002-b4c3-auth-gateway and", "changes/CHG-2026-001-a3f2-auth-gatewayé"} {
		if !strings.Contains(rewritten, want) {
			t.Fatalf("expected rewritten text to contain %q, got %q", want, rewritten)
		}
	}
}

func TestStatusDocumentFromMapRejectsInvalidPromotionAssessmentStatus(t *testing.T) {
	_, err := statusDocumentFromMap(map[string]any{"schema_version": 1, "id": "CHG-2026-001-a3f2-auth-gateway", "title": "Add auth gateway", "status": "proposed", "type": "feature", "verification_status": "pending", "context_bundles": []any{"base"}, "related_specs": []any{}, "related_decisions": []any{}, "related_changes": []any{}, "depends_on": []any{}, "informed_by": []any{}, "supersedes": []any{}, "superseded_by": []any{}, "closed_at": nil, "promotion_assessment": map[string]any{"status": "not-valid"}})
	if err == nil || !strings.Contains(err.Error(), "promotion_assessment.status") {
		t.Fatalf("expected invalid promotion assessment status error, got %v", err)
	}
}
