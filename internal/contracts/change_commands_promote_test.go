package contracts

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPromoteChangeAcceptWritesStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        strings.NewReader("abcd"),
	})
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := PromoteChange(v, loaded, created.ID, PromoteOptions{})
	if err != nil {
		t.Fatalf("promote accept: %v", err)
	}
	if got, want := result.PromotionAssessmentStatus, "accepted"; got != want {
		t.Fatalf("expected promotion status %q, got %q", want, got)
	}
	if len(result.ChangedFiles) != 1 || result.ChangedFiles[0].Action != "updated" {
		t.Fatalf("expected one status mutation, got %#v", result.ChangedFiles)
	}

	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: accepted") {
		t.Fatalf("expected accepted promotion status in status.yaml, got:\n%s", text)
	}
	if !strings.Contains(text, "target_type: standard") || !strings.Contains(text, "target_path: standards/global/base.md") {
		t.Fatalf("expected suggested promotion targets to remain unchanged, got:\n%s", text)
	}
}

func TestPromoteChangeCompleteWritesStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        strings.NewReader("abce"),
	})
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	mustPromoteChange(t, v, root, created.ID, PromoteOptions{})

	result := mustPromoteChange(t, v, root, created.ID, PromoteOptions{Complete: true})
	if got, want := result.PromotionAssessmentStatus, "completed"; got != want {
		t.Fatalf("expected promotion status %q, got %q", want, got)
	}

	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "promotion_assessment:\n  status: completed") {
		t.Fatalf("expected completed promotion status in status.yaml, got:\n%s", text)
	}
	if !strings.Contains(text, "target_type: standard") {
		t.Fatalf("expected suggested targets to remain in status.yaml, got:\n%s", text)
	}
}

func TestPromoteChangeDryRunCloneLeavesOriginalUnchanged(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        strings.NewReader("abcf"),
	})
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	before := mustReadBytes(t, statusPath)

	cloneRoot := t.TempDir()
	copyDirTree(t, root, cloneRoot)
	cloneValidator, cloneLoaded := mustLoadWorkflowProject(t, cloneRoot)
	defer cloneLoaded.Close()
	cloneResult, err := PromoteChange(cloneValidator, cloneLoaded, created.ID, PromoteOptions{})
	if err != nil {
		t.Fatalf("promote accept in dry-run clone: %v", err)
	}
	if got, want := cloneResult.PromotionAssessmentStatus, "accepted"; got != want {
		t.Fatalf("expected clone promotion status %q, got %q", want, got)
	}

	assertFileBytesEqual(t, statusPath, before)
	cloneStatusPath := filepath.Join(cloneRoot, "runecontext", "changes", created.ID, "status.yaml")
	cloneText := strings.ReplaceAll(string(mustReadBytes(t, cloneStatusPath)), "\r\n", "\n")
	if !strings.Contains(cloneText, "promotion_assessment:\n  status: accepted") {
		t.Fatalf("expected accepted status in clone, got:\n%s", cloneText)
	}
}

func TestPromoteChangeRejectsInvalidTransitions(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, defaultFeatureChangeOptions("Add cache invalidation", []byte{0xaa, 0xbb}))
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := PromoteChange(v, loaded, created.ID, PromoteOptions{})
	if err == nil || !strings.Contains(err.Error(), "requires current promotion_assessment.status to be \"suggested\"") {
		t.Fatalf("expected suggested-transition failure, got %v", err)
	}

	standardRoot := copyChangeWorkflowTemplate(t)
	sv, standard := mustCreateChange(t, standardRoot, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        bytes.NewReader([]byte{0xab, 0xcd}),
	})
	mustCloseChange(t, sv, standardRoot, standard.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})
	standardLoaded := mustReloadWorkflowProject(t, sv, standardRoot)
	defer standardLoaded.Close()
	_, err = PromoteChange(sv, standardLoaded, standard.ID, PromoteOptions{Complete: true})
	if err == nil || !strings.Contains(err.Error(), "requires current promotion_assessment.status to be \"accepted\"") {
		t.Fatalf("expected accepted-transition failure, got %v", err)
	}
}

func TestPromoteChangeRejectsUnknownTargetType(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateChange(t, root, ChangeCreateOptions{
		Title:          "Refine base standard wording",
		Type:           "standard",
		ContextBundles: []string{"base"},
		Now:            time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		Entropy:        strings.NewReader("abc0"),
	})
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := PromoteChange(v, loaded, created.ID, PromoteOptions{Targets: []string{"note:notes/agenda.md"}})
	if err == nil || !strings.Contains(err.Error(), "unknown target type") {
		t.Fatalf("expected unknown-target-type rejection, got %v", err)
	}
}

func mustPromoteChange(t *testing.T, v *Validator, root, changeID string, options PromoteOptions) *ChangeOperationResult {
	t.Helper()
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := PromoteChange(v, loaded, changeID, options)
	if err != nil {
		t.Fatalf("promote change: %v", err)
	}
	return result
}
