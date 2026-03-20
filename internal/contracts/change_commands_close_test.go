package contracts

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCloseChangeWritesClosedStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	closeResult := mustCloseChange(t, v, root, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	assertClosedResultMetadata(t, closeResult)
	assertClosedStatusFile(t, root, result.ID)
}

func assertClosedResultMetadata(t *testing.T, result *ChangeOperationResult) {
	t.Helper()
	if len(result.ContextBundles) != 1 || result.ContextBundles[0] != "base" {
		t.Fatalf("expected close result to retain context bundles, got %#v", result.ContextBundles)
	}
	if len(result.ApplicableStandards) != 1 || result.ApplicableStandards[0] != "standards/global/base.md" {
		t.Fatalf("expected close result to retain applicable standards, got %#v", result.ApplicableStandards)
	}
}

func assertClosedStatusFile(t *testing.T, root, changeID string) {
	t.Helper()
	requireFileContent(t, filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"), strings.Join([]string{"schema_version: 1", "id: CHG-2026-001-aabb-add-cache-invalidation", "title: Add cache invalidation", "status: closed", "type: feature", "size: small", "verification_status: passed", "context_bundles:", "  - base", "related_specs: []", "related_decisions: []", "related_changes: []", "depends_on: []", "informed_by: []", "supersedes: []", "superseded_by: []", "created_at: \"2026-03-18\"", "closed_at: \"2026-03-20\"", "promotion_assessment:", "  status: none", "  suggested_targets: []", ""}, "\n"))
}

func TestCloseChangeWritesSupersededStatusAndReciprocalLink(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{VerificationStatus: "skipped", ClosedAt: time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), SupersededBy: []string{"CHG-2026-002-b4c3-auth-revision"}}); err != nil {
		t.Fatalf("supersede change: %v", err)
	}
	assertSupersededStatusContents(t, root)
	assertReciprocalSupersedesLink(t, root)
}

func assertSupersededStatusContents(t *testing.T, root string) {
	t.Helper()
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "status: superseded") || !strings.Contains(text, "closed_at: \"2026-03-18\"") {
		t.Fatalf("unexpected superseded status contents:\n%s", text)
	}
	if !strings.Contains(text, "promotion_assessment:\n  status: suggested") {
		t.Fatalf("expected superseded close to record deterministic promotion suggestions, got:\n%s", text)
	}
}

func assertReciprocalSupersedesLink(t *testing.T, root string) {
	t.Helper()
	successorPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml")
	successorText := string(mustReadBytes(t, successorPath))
	if !strings.Contains(successorText, "supersedes:\n  - CHG-2026-001-a3f2-auth-gateway") {
		t.Fatalf("expected reciprocal supersedes link, got:\n%s", successorText)
	}
}

func TestCloseChangePreservesFilePermissions(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", result.ID, "status.yaml")
	if err := os.Chmod(statusPath, 0o600); err != nil {
		t.Fatalf("chmod status path: %v", err)
	}
	mustCloseChange(t, v, root, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	assertFilePermission(t, statusPath, 0o600)
}

func assertFilePermission(t *testing.T, path string, want fs.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat rewritten status: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != want {
		t.Fatalf("expected close rewrite to preserve perms %o, got %o", want, info.Mode().Perm())
	}
}

func TestCloseChangeRejectsSymlinkedStatusTarget(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", result.ID, "status.yaml")
	original := mustReadBytes(t, statusPath)
	originalLstat := lstatPath
	t.Cleanup(func() { lstatPath = originalLstat })
	lstatPath = func(path string) (os.FileInfo, error) {
		if filepath.Clean(path) == filepath.Clean(statusPath) {
			return fakeFileInfo{name: filepath.Base(path), mode: os.ModeSymlink}, nil
		}
		return os.Lstat(path)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	if err == nil || !strings.Contains(err.Error(), "symlinked targets") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
	assertFileBytesEqual(t, statusPath, original)
}

func TestWriteFileAtomicallyFallsBackWhenDestinationRenameCannotReplace(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "status.yaml")
	if err := os.WriteFile(targetPath, []byte("old\n"), 0o600); err != nil {
		t.Fatalf("write original target: %v", err)
	}
	originalRename, originalFallback := renamePath, atomicReplaceNeedsFallback
	t.Cleanup(func() { renamePath = originalRename; atomicReplaceNeedsFallback = originalFallback })
	atomicReplaceNeedsFallback = true
	renameAttempts := 0
	renamePath = func(oldPath, newPath string) error {
		if filepath.Clean(newPath) == filepath.Clean(targetPath) && renameAttempts == 0 {
			renameAttempts++
			return fmt.Errorf("simulated windows rename collision")
		}
		renameAttempts++
		return os.Rename(oldPath, newPath)
	}
	if err := writeFileAtomically(targetPath, []byte("new\n"), 0o600); err != nil {
		t.Fatalf("write file atomically with fallback: %v", err)
	}
	assertFileBytesEqual(t, targetPath, []byte("new\n"))
	assertFilePermission(t, targetPath, 0o600)
}

func TestCloseChangeRollsBackWhenLaterWriteFails(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v := NewValidator(schemaRoot(t))
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	successorPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml")
	beforeStatus, beforeSuccessor := mustReadBytes(t, statusPath), mustReadBytes(t, successorPath)
	originalWriteFile := writeFilePath
	t.Cleanup(func() { writeFilePath = originalWriteFile })
	writeCount := 0
	writeFilePath = func(path string, data []byte, perm os.FileMode) error {
		writeCount++
		if writeCount == 2 {
			return fmt.Errorf("forced write failure")
		}
		return os.WriteFile(path, data, perm)
	}
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{VerificationStatus: "skipped", ClosedAt: time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), SupersededBy: []string{"CHG-2026-002-b4c3-auth-revision"}})
	if err == nil || !strings.Contains(err.Error(), "forced write failure") {
		t.Fatalf("expected write failure, got %v", err)
	}
	assertFileBytesEqual(t, statusPath, beforeStatus)
	assertFileBytesEqual(t, successorPath, beforeSuccessor)
}

func TestCloseChangeRejectsTerminalSuccessorWithoutReciprocalLink(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, first, second := createTwoWorkflowChanges(t, root)
	mustCloseChange(t, v, root, second.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 19, 0, 0, 0, 0, time.UTC)})
	firstStatusPath := filepath.Join(root, "runecontext", "changes", first.ID, "status.yaml")
	secondStatusPath := filepath.Join(root, "runecontext", "changes", second.ID, "status.yaml")
	firstBefore, secondBefore := mustReadBytes(t, firstStatusPath), mustReadBytes(t, secondStatusPath)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, first.ID, ChangeCloseOptions{VerificationStatus: "skipped", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC), SupersededBy: []string{second.ID}})
	if err == nil || !strings.Contains(err.Error(), "cannot be updated with a reciprocal supersedes link") {
		t.Fatalf("expected terminal successor rejection, got %v", err)
	}
	assertFileBytesEqual(t, firstStatusPath, firstBefore)
	assertFileBytesEqual(t, secondStatusPath, secondBefore)
}

func createTwoWorkflowChanges(t *testing.T, root string) (*Validator, *ChangeOperationResult, *ChangeOperationResult) {
	t.Helper()
	v, first := mustCreateDefaultFeatureChange(t, root)
	_, second := mustCreateChange(t, root, defaultFeatureChangeOptions("Revise cache invalidation", []byte{0xcc, 0xdd}))
	return v, first, second
}

func TestCloseChangeRollsBackStatusOnValidationFailure(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, result := mustCreateDefaultFeatureChange(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", result.ID, "status.yaml")
	before := mustReadBytes(t, statusPath)
	originalValidate := validateProjectAfterChangeMutation
	t.Cleanup(func() { validateProjectAfterChangeMutation = originalValidate })
	validateProjectAfterChangeMutation = func(*Validator, string) (*ProjectIndex, error) { return nil, fmt.Errorf("forced validation failure") }
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, result.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	if err == nil || !strings.Contains(err.Error(), "forced validation failure") {
		t.Fatalf("expected forced validation failure, got %v", err)
	}
	assertFileBytesEqual(t, statusPath, before)
}

func TestBuildProjectStatusSummaryLeavesMissingOptionalSizeEmpty(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	writeExistingChangeWithoutOptionalFields(t, root)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	summary, err := BuildProjectStatusSummary(v, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}
	if len(summary.Active) != 1 || summary.Active[0].Size != "" {
		t.Fatalf("expected one active change with empty size, got %#v", summary.Active)
	}
}

func TestCloseChangeOmitsMissingOptionalFieldsWhenRewritingStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	changeID := writeExistingChangeWithoutOptionalFields(t, root)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, changeID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	assertOptionalFieldsOmitted(t, filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"))
}

func assertOptionalFieldsOmitted(t *testing.T, path string) {
	t.Helper()
	text := strings.ReplaceAll(string(mustReadBytes(t, path)), "\r\n", "\n")
	if strings.Contains(text, "<nil>") || strings.Contains(text, "created_at:") || strings.Contains(text, "size:") {
		t.Fatalf("expected rewritten status to omit missing optional fields, got:\n%s", text)
	}
	if !strings.Contains(text, "closed_at: \"2026-03-20\"") {
		t.Fatalf("expected closed_at to be written, got:\n%s", text)
	}
}

func TestCloseChangeNormalizesEmptyPromotionAssessmentToNoneOnClose(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	changeID := writeExistingChangeWithEmptyPromotionAssessment(t, root)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, changeID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	text := strings.ReplaceAll(string(mustReadBytes(t, filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"))), "\r\n", "\n")
	if strings.Contains(text, "<nil>") || !strings.Contains(text, "promotion_assessment:\n  status: none\n  suggested_targets: []") {
		t.Fatalf("expected close to replace empty promotion assessment with deterministic none state, got:\n%s", text)
	}
}

func TestCloseChangeRecordsPromotionAssessmentNoneWhenNoTargets(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	changeID := writeExistingChangeWithoutOptionalFields(t, root)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, changeID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	text := strings.ReplaceAll(string(mustReadBytes(t, filepath.Join(root, "runecontext", "changes", changeID, "status.yaml"))), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: none\n  suggested_targets: []") {
		t.Fatalf("expected explicit none promotion assessment, got:\n%s", text)
	}
}

func TestCloseChangePromotionTargetsUseStableFormatting(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	needle := strings.Join([]string{
		"promotion_assessment:",
		"  status: suggested",
		"  suggested_targets:",
		"    - target_type: spec",
		"      target_path: specs/auth-gateway.md",
		"      summary: Review and promote durable spec updates from this change.",
		"    - target_type: decision",
		"      target_path: decisions/DEC-0001-trust-boundary-model.md",
		"      summary: Review and promote durable decision updates from this change.",
	}, "\n")
	if !strings.Contains(text, needle) {
		t.Fatalf("expected stable promotion target formatting, got:\n%s", text)
	}
	if strings.Contains(text, "target_type: standard") {
		t.Fatalf("expected non-standard change to avoid standard promotion targets, got:\n%s", text)
	}
}

func TestCloseChangeSuggestsStandardsTargetsForStandardChanges(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        strings.NewReader("abcd"),
	})
	closeResult := mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	if closeResult.Status != "closed" {
		t.Fatalf("expected closed status, got %q", closeResult.Status)
	}
	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: suggested") {
		t.Fatalf("expected suggested promotion status for standard change, got:\n%s", text)
	}
	if !strings.Contains(text, "target_type: standard") || !strings.Contains(text, "target_path: standards/global/base.md") {
		t.Fatalf("expected standard promotion target, got:\n%s", text)
	}
}

func TestCloseChangePreservesAcceptedPromotionAssessment(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	rewritePromotionAssessmentStatus(t, statusPath, "accepted")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: accepted\n  suggested_targets: []") {
		t.Fatalf("expected accepted promotion state to be preserved, got:\n%s", text)
	}
}

func TestCloseChangePreservesCompletedPromotionAssessment(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	rewritePromotionAssessmentStatus(t, statusPath, "completed")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("close change: %v", err)
	}
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: completed\n  suggested_targets: []") {
		t.Fatalf("expected completed promotion state to be preserved, got:\n%s", text)
	}
}

func rewritePromotionAssessmentStatus(t *testing.T, statusPath, status string) {
	t.Helper()
	rewriteFile(t, statusPath, func(text string) string {
		oldBlock := "promotion_assessment:\n  status: pending\n  suggested_targets: []"
		newBlock := "promotion_assessment:\n  status: " + status + "\n  suggested_targets: []"
		replaced := strings.Replace(text, oldBlock, newBlock, 1)
		if replaced == text {
			t.Fatalf("rewritePromotionAssessmentStatus: expected promotion assessment block in %s", statusPath)
		}
		return replaced
	})
}
