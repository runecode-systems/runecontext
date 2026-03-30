package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateChangeNonRecursiveLeavesFeatureSubChangeUntouched(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, umbrellaID, ChangeUpdateOptions{Status: "planned"})
	if err != nil {
		t.Fatalf("update umbrella non-recursive: %v", err)
	}
	if result.Recursive {
		t.Fatalf("expected non-recursive update result, got recursive=true")
	}
	if got, want := result.RecursiveTargetCount, 0; got != want {
		t.Fatalf("expected recursive_target_count %d, got %d", want, got)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "planned")
	assertStatusLifecycleEquals(t, root, featureID, "proposed")
}

func TestUpdateChangeRecursiveUmbrellaUpdatesFeatureSubChanges(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := UpdateChange(v, loaded, umbrellaID, ChangeUpdateOptions{Status: "planned", Recursive: true})
	if err != nil {
		t.Fatalf("update umbrella recursive: %v", err)
	}
	if !result.Recursive {
		t.Fatalf("expected recursive update result")
	}
	if got, want := result.RecursiveTargetCount, 1; got != want {
		t.Fatalf("expected recursive_target_count %d, got %d", want, got)
	}
	if len(result.RecursiveTargetIDs) != 1 || result.RecursiveTargetIDs[0] != featureID {
		t.Fatalf("expected recursive target %q, got %#v", featureID, result.RecursiveTargetIDs)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "planned")
	assertStatusLifecycleEquals(t, root, featureID, "planned")
}

func TestUpdateChangeRecursiveRejectsNonFeatureRelatedChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, _ := createUmbrellaAndFeatureSubChange(t, root)
	otherID := createProjectChangeLinkedToUmbrella(t, root, umbrellaID)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := UpdateChange(v, loaded, umbrellaID, ChangeUpdateOptions{Status: "planned", Recursive: true})
	if err == nil || !strings.Contains(err.Error(), "not an eligible feature sub-change") {
		t.Fatalf("expected non-feature recursive target rejection, got %v", err)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "proposed")
	assertStatusLifecycleEquals(t, root, otherID, "proposed")
}

func TestUpdateChangeRecursiveRollsBackWhenOneTargetCannotTransition(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	mustCloseChange(t, v, root, featureID, ChangeCloseOptions{VerificationStatus: "passed", ClosedAt: time.Date(2026, time.March, 20, 0, 0, 0, 0, time.UTC)})

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := UpdateChange(v, loaded, umbrellaID, ChangeUpdateOptions{Status: "planned", Recursive: true})
	if err == nil || !strings.Contains(err.Error(), featureID) || !strings.Contains(err.Error(), "already in terminal status") {
		t.Fatalf("expected recursive transition block on terminal feature, got %v", err)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "proposed")
	assertStatusLifecycleEquals(t, root, featureID, "closed")
}

func TestCloseChangeRecursiveUmbrellaClosesFeatureSubChanges(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", featureID, "status.yaml"), "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC), Recursive: true})
	if err != nil {
		t.Fatalf("close umbrella recursive: %v", err)
	}
	if !result.Recursive {
		t.Fatalf("expected recursive close result")
	}
	if got, want := result.RecursiveTargetCount, 1; got != want {
		t.Fatalf("expected recursive_target_count %d, got %d", want, got)
	}
	if len(result.RecursiveTargetIDs) != 1 || result.RecursiveTargetIDs[0] != featureID {
		t.Fatalf("expected recursive target %q, got %#v", featureID, result.RecursiveTargetIDs)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "closed")
	assertStatusLifecycleEquals(t, root, featureID, "closed")
}

func TestCloseChangeNonRecursiveLeavesFeatureSubChangeUntouched(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("close umbrella non-recursive: %v", err)
	}
	if result.Recursive {
		t.Fatalf("expected non-recursive close result, got recursive=true")
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "closed")
	assertStatusLifecycleEquals(t, root, featureID, "proposed")
}

func TestCloseChangeRecursiveRejectsNonFeatureRelatedChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, _ := createUmbrellaAndFeatureSubChange(t, root)
	otherID := createProjectChangeLinkedToUmbrella(t, root, umbrellaID)
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", otherID, "status.yaml"), "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC), Recursive: true})
	if err == nil || !strings.Contains(err.Error(), "not an eligible feature sub-change") {
		t.Fatalf("expected non-feature recursive target rejection, got %v", err)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "proposed")
	assertStatusLifecycleEquals(t, root, otherID, "proposed")
}

func TestCloseChangeRecursiveRollsBackWhenOneTargetBlocksCascade(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{ClosedAt: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC), Recursive: true})
	if err == nil || !strings.Contains(err.Error(), featureID) || !strings.Contains(err.Error(), "requires --verification-status") {
		t.Fatalf("expected recursive close rejection with target context, got %v", err)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "proposed")
	assertStatusLifecycleEquals(t, root, featureID, "proposed")
}

func TestCloseChangeRecursiveSupersededByAddsReciprocalLinksForAllTargets(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	_, successor := mustCreateChange(t, root, defaultFeatureChangeOptions("Successor change", []byte{0x77, 0x88}))
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", featureID, "status.yaml"), "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{
		ClosedAt:     time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC),
		Recursive:    true,
		SupersededBy: []string{successor.ID},
	})
	if err != nil {
		t.Fatalf("close umbrella recursive superseded_by: %v", err)
	}
	if !result.Recursive {
		t.Fatalf("expected recursive close result")
	}
	if got, want := result.RecursiveTargetCount, 1; got != want {
		t.Fatalf("expected recursive_target_count %d, got %d", want, got)
	}

	assertStatusLifecycleEquals(t, root, umbrellaID, "superseded")
	assertStatusLifecycleEquals(t, root, featureID, "superseded")
	assertSupersedesLinksContain(t, root, successor.ID, umbrellaID, featureID)
	assertValidatedWorkflowProject(t, v, root)
}

func createUmbrellaAndFeatureSubChange(t *testing.T, root string) (*Validator, string, string) {
	t.Helper()
	v, umbrella := mustCreateChange(t, root, defaultProjectChangeOptions("Umbrella project", []byte{0x11, 0x22}))
	_, feature := mustCreateChange(t, root, defaultFeatureChangeOptions("Feature sub-change", []byte{0x33, 0x44}))
	wireBidirectionalRelatedChangeLink(t, root, umbrella.ID, feature.ID)
	return v, umbrella.ID, feature.ID
}

func createProjectChangeLinkedToUmbrella(t *testing.T, root, umbrellaID string) string {
	t.Helper()
	_, extra := mustCreateChange(t, root, defaultProjectChangeOptions("Extra project node", []byte{0x55, 0x66}))
	wireBidirectionalRelatedChangeLink(t, root, umbrellaID, extra.ID)
	return extra.ID
}

func wireBidirectionalRelatedChangeLink(t *testing.T, root, leftID, rightID string) {
	t.Helper()
	appendRelatedChangeLink(t, filepath.Join(root, "runecontext", "changes", leftID, "status.yaml"), rightID)
	appendRelatedChangeLink(t, filepath.Join(root, "runecontext", "changes", rightID, "status.yaml"), leftID)
}

func appendRelatedChangeLink(t *testing.T, statusPath, relatedID string) {
	t.Helper()
	rewriteFile(t, statusPath, func(text string) string {
		updated, ok := replaceEmptyRelatedChanges(text, relatedID)
		if ok {
			return updated
		}
		updated, ok = appendRelatedChangeUnderBlock(text, relatedID)
		if ok {
			return updated
		}
		t.Fatalf("status %s missing related_changes field", statusPath)
		return text
	})
}

func replaceEmptyRelatedChanges(text, relatedID string) (string, bool) {
	if !strings.Contains(text, "related_changes: []") {
		return "", false
	}
	return strings.Replace(text, "related_changes: []", fmt.Sprintf("related_changes:\n  - %s", relatedID), 1), true
}

func appendRelatedChangeUnderBlock(text, relatedID string) (string, bool) {
	lines := strings.Split(text, "\n")
	for i := range lines {
		if lines[i] != "related_changes:" {
			continue
		}
		return appendRelatedChangeItem(lines, i, relatedID), true
	}
	return "", false
}

func appendRelatedChangeItem(lines []string, blockIndex int, relatedID string) string {
	for j := blockIndex + 1; j < len(lines); j++ {
		if !strings.HasPrefix(lines[j], "  - ") {
			lines = append(lines[:j], append([]string{"  - " + relatedID}, lines[j:]...)...)
			return strings.Join(lines, "\n")
		}
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[j]), "-")) == relatedID {
			return strings.Join(lines, "\n")
		}
	}
	lines = append(lines, "  - "+relatedID)
	return strings.Join(lines, "\n")
}

func assertStatusLifecycleEquals(t *testing.T, root, changeID, want string) {
	t.Helper()
	statusPath := filepath.Join(root, "runecontext", "changes", changeID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status for %s: %v", changeID, err)
	}
	needle := "status: " + want
	if !strings.Contains(strings.ReplaceAll(string(data), "\r\n", "\n"), needle) {
		t.Fatalf("expected %s to contain %q, got:\n%s", statusPath, needle, string(data))
	}
}

func assertSupersedesLinksContain(t *testing.T, root, successorID string, expected ...string) {
	t.Helper()
	statusPath := filepath.Join(root, "runecontext", "changes", successorID, "status.yaml")
	text := strings.ReplaceAll(string(mustReadBytes(t, statusPath)), "\r\n", "\n")
	for _, targetID := range expected {
		needle := "  - " + targetID
		if !strings.Contains(text, needle) {
			t.Fatalf("expected %s to contain supersedes entry %q, got:\n%s", statusPath, targetID, text)
		}
	}
}

func TestCloseChangeRecursiveRejectsSupersededByTargetSelfSupersession(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, featureID := createUmbrellaAndFeatureSubChange(t, root)
	// mark both as verified so close would otherwise proceed
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml"), "passed")
	rewriteStatusVerificationStatus(t, filepath.Join(root, "runecontext", "changes", featureID, "status.yaml"), "passed")

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	// Attempt to close umbrella recursively but declare feature as a successor
	_, err := CloseChange(v, loaded, umbrellaID, ChangeCloseOptions{
		ClosedAt:     time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC),
		Recursive:    true,
		SupersededBy: []string{featureID},
	})
	if err == nil || !strings.Contains(err.Error(), featureID) || !strings.Contains(err.Error(), "recursive close target") {
		t.Fatalf("expected rejection when superseded_by references a recursive target, got %v", err)
	}

	// ensure no statuses were mutated
	assertStatusLifecycleEquals(t, root, umbrellaID, "proposed")
	assertStatusLifecycleEquals(t, root, featureID, "proposed")
}
