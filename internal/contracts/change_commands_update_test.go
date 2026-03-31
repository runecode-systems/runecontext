package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateChangeWritesNonTerminalLifecycleStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "planned"})
	if err != nil {
		t.Fatalf("update change: %v", err)
	}
	if got, want := result.Status, "planned"; got != want {
		t.Fatalf("expected status %q, got %q", want, got)
	}
	if len(result.ChangedFiles) != 1 || result.ChangedFiles[0].Action != "updated" {
		t.Fatalf("expected one status mutation, got %#v", result.ChangedFiles)
	}

	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	if !strings.Contains(text, "status: planned") {
		t.Fatalf("expected planned lifecycle in status.yaml, got:\n%s", text)
	}
}

func TestUpdateChangeLeavesPromotionAssessmentUntouched(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	rewriteUpdatePromotionAssessmentFixture(t, statusPath)

	before := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	if _, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "implemented"}); err != nil {
		t.Fatalf("update change: %v", err)
	}
	after := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	assertPromotionAssessmentUntouched(t, before, after)
}

func TestUpdateChangeAllowsVerifiedWhenVerificationStatusAlreadyCompleted(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	statusPath := filepath.Join(root, "runecontext", "changes", created.ID, "status.yaml")
	rewriteStatusVerificationStatus(t, statusPath, "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "verified"})
	if err != nil {
		t.Fatalf("update change to verified: %v", err)
	}
	if result.Status != "verified" {
		t.Fatalf("expected verified status, got %q", result.Status)
	}
}

func TestUpdateChangeAllowsVerifiedWithExplicitVerificationStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "verified", VerificationStatus: "failed"})
	if err != nil {
		t.Fatalf("update change to verified with verification status: %v", err)
	}
	if result.Status != "verified" {
		t.Fatalf("expected verified status, got %q", result.Status)
	}
}

func TestUpdateChangeRejectsVerifiedWhenVerificationStatusPending(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	if _, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "verified"}); err == nil || !strings.Contains(err.Error(), "verified changes must record a completed verification_status") {
		t.Fatalf("expected pending verification_status rejection, got %v", err)
	}
}

func TestUpdateChangeRejectsSettingPendingVerificationStatus(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	if _, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "verified", VerificationStatus: "pending"}); err == nil || !strings.Contains(err.Error(), "must not set verification_status to pending") {
		t.Fatalf("expected pending verification_status set rejection, got %v", err)
	}
}

func TestUpdateChangeRejectsVerificationStatusOnNonVerifiedTransition(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	if _, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "implemented", VerificationStatus: "passed"}); err == nil || !strings.Contains(err.Error(), "--verification-status is only supported when --status verified") {
		t.Fatalf("expected verification_status flag rejection for non-verified status, got %v", err)
	}
}

func TestUpdateChangeAddsReciprocalRelatedChanges(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, first := mustCreateDefaultFeatureChange(t, root)
	_, second := mustCreateChange(t, root, defaultFeatureChangeOptions("Second feature", []byte{0x51, 0x52}))

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, first.ID, ChangeUpdateOptions{Status: "planned", AddRelatedChanges: []string{second.ID}})
	if err != nil {
		t.Fatalf("update change with related addition: %v", err)
	}
	if len(result.RelatedChanges) != 1 || result.RelatedChanges[0] != second.ID {
		t.Fatalf("expected updated related_changes [%s], got %#v", second.ID, result.RelatedChanges)
	}

	firstStatusPath := filepath.Join(root, "runecontext", "changes", first.ID, "status.yaml")
	firstText := strings.ReplaceAll(string(mustReadBytes(t, firstStatusPath)), "\r\n", "\n")
	if !strings.Contains(firstText, "related_changes:\n  - "+second.ID) {
		t.Fatalf("expected first status related_changes to include %s, got:\n%s", second.ID, firstText)
	}
	secondStatusPath := filepath.Join(root, "runecontext", "changes", second.ID, "status.yaml")
	secondText := strings.ReplaceAll(string(mustReadBytes(t, secondStatusPath)), "\r\n", "\n")
	if !strings.Contains(secondText, "related_changes:\n  - "+first.ID) {
		t.Fatalf("expected reciprocal related_changes to include %s, got:\n%s", first.ID, secondText)
	}
}

func TestUpdateChangeRemovesReciprocalRelatedChanges(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, first := mustCreateDefaultFeatureChange(t, root)
	_, second := mustCreateChange(t, root, defaultFeatureChangeOptions("Second feature", []byte{0x61, 0x62}))
	wireBidirectionalRelatedChangeLink(t, root, first.ID, second.ID)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, first.ID, ChangeUpdateOptions{Status: "planned", RemoveRelatedChanges: []string{second.ID}})
	if err != nil {
		t.Fatalf("update change with related removal: %v", err)
	}
	if len(result.RelatedChanges) != 0 {
		t.Fatalf("expected updated related_changes to be empty, got %#v", result.RelatedChanges)
	}

	firstStatusPath := filepath.Join(root, "runecontext", "changes", first.ID, "status.yaml")
	firstText := strings.ReplaceAll(string(mustReadBytes(t, firstStatusPath)), "\r\n", "\n")
	if !strings.Contains(firstText, "related_changes: []") {
		t.Fatalf("expected first status related_changes to be empty, got:\n%s", firstText)
	}
	secondStatusPath := filepath.Join(root, "runecontext", "changes", second.ID, "status.yaml")
	secondText := strings.ReplaceAll(string(mustReadBytes(t, secondStatusPath)), "\r\n", "\n")
	if !strings.Contains(secondText, "related_changes: []") {
		t.Fatalf("expected reciprocal related_changes to be empty, got:\n%s", secondText)
	}
}

func TestUpdateChangeRejectsConflictingRelationshipEdits(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, first := mustCreateDefaultFeatureChange(t, root)
	_, second := mustCreateChange(t, root, defaultFeatureChangeOptions("Second feature", []byte{0x71, 0x72}))

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := UpdateChange(v, loaded, first.ID, ChangeUpdateOptions{Status: "planned", AddRelatedChanges: []string{second.ID}, RemoveRelatedChanges: []string{second.ID}})
	if err == nil || !strings.Contains(err.Error(), "relationship edit lists conflict") {
		t.Fatalf("expected relationship conflict rejection, got %v", err)
	}
}

func TestUpdateChangeRejectsSelfRelationshipEdit(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "planned", AddRelatedChanges: []string{created.ID}})
	if err == nil || !strings.Contains(err.Error(), "must not relate a change to itself") {
		t.Fatalf("expected self-related-change rejection, got %v", err)
	}
}

func rewriteUpdatePromotionAssessmentFixture(t *testing.T, statusPath string) {
	t.Helper()
	rewriteFile(t, statusPath, func(text string) string {
		oldBlock := "promotion_assessment:\n  status: pending\n  suggested_targets: []"
		newBlock := strings.Join([]string{
			"promotion_assessment:",
			"  status: suggested",
			"  suggested_targets:",
			"    - target_type: spec",
			"      target_path: specs/example.md",
			"      summary: Keep custom promotion suggestions untouched.",
		}, "\n")
		replaced := strings.Replace(text, oldBlock, newBlock, 1)
		if replaced == text {
			t.Fatalf("expected promotion_assessment block in %s", statusPath)
		}
		return replaced
	})
}

func rewriteStatusVerificationStatus(t *testing.T, statusPath, value string) {
	t.Helper()
	rewriteFile(t, statusPath, func(text string) string {
		lines := strings.Split(text, "\n")
		const prefix = "verification_status: "
		for i, line := range lines {
			if strings.HasPrefix(line, prefix) {
				if value == "" {
					lines[i] = fmt.Sprintf("%s\"\"", prefix)
				} else {
					lines[i] = prefix + value
				}
				return strings.Join(lines, "\n")
			}
		}
		t.Fatalf("status file %s missing verification_status field", statusPath)
		return text
	})
}

func assertPromotionAssessmentUntouched(t *testing.T, before, after string) {
	t.Helper()
	if strings.Contains(after, "promotion_assessment:\n  status: pending") {
		t.Fatalf("expected promotion assessment to remain untouched, got:\n%s", after)
	}
	if !strings.Contains(after, "promotion_assessment:\n  status: suggested") {
		t.Fatalf("expected suggested promotion assessment to remain, got:\n%s", after)
	}
	if !strings.Contains(after, "target_path: specs/example.md") {
		t.Fatalf("expected custom promotion target to remain, got:\n%s", after)
	}
	if strings.Count(before, "promotion_assessment:") != strings.Count(after, "promotion_assessment:") {
		t.Fatalf("expected promotion_assessment structure to remain stable")
	}
}

func TestUpdateChangeRejectsBackwardAndTerminalTransitions(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, created := mustCreateDefaultFeatureChange(t, root)
	mustCloseChange(t, v, root, created.ID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := UpdateChange(v, loaded, created.ID, ChangeUpdateOptions{Status: "verified"})
	if err == nil || !strings.Contains(err.Error(), "already in terminal status") {
		t.Fatalf("expected terminal-status rejection, got %v", err)
	}

	root2 := copyChangeWorkflowTemplate(t)
	v2, created2 := mustCreateDefaultFeatureChange(t, root2)
	loaded2 := mustReloadWorkflowProject(t, v2, root2)
	if _, err := UpdateChange(v2, loaded2, created2.ID, ChangeUpdateOptions{Status: "implemented"}); err != nil {
		loaded2.Close()
		t.Fatalf("advance to implemented: %v", err)
	}
	loaded2.Close()

	loaded2 = mustReloadWorkflowProject(t, v2, root2)
	defer loaded2.Close()
	_, err = UpdateChange(v2, loaded2, created2.ID, ChangeUpdateOptions{Status: "planned"})
	if err == nil || !strings.Contains(err.Error(), "cannot transition backward") {
		t.Fatalf("expected backward-transition rejection, got %v", err)
	}
}
