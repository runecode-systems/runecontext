package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunChangeUpdateRecursiveOutputsCascadeMetadata(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	writeCLIBidirectionalRelatedLink(t, projectRoot, umbrellaID, featureID)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "update", umbrellaID, "--status", "planned", "--recursive", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["recursive"], "true"; got != want {
		t.Fatalf("expected recursive %q, got %q", want, got)
	}
	if got, want := fields["recursive_target_count"], "1"; got != want {
		t.Fatalf("expected recursive_target_count %q, got %q", want, got)
	}
	if got, want := fields["recursive_target_1"], featureID; got != want {
		t.Fatalf("expected recursive_target_1 %q, got %q", want, got)
	}
}

func TestRunChangeCloseRecursiveOutputsCascadeMetadata(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	writeCLIBidirectionalRelatedLink(t, projectRoot, umbrellaID, featureID)
	writeCLIStatusVerificationStatus(t, projectRoot, umbrellaID, "passed")
	writeCLIStatusVerificationStatus(t, projectRoot, featureID, "passed")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", umbrellaID, "--closed-at", "2026-03-21", "--recursive", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["recursive"], "true"; got != want {
		t.Fatalf("expected recursive %q, got %q", want, got)
	}
	if got, want := fields["recursive_target_count"], "1"; got != want {
		t.Fatalf("expected recursive_target_count %q, got %q", want, got)
	}
	if got, want := fields["recursive_target_1"], featureID; got != want {
		t.Fatalf("expected recursive_target_1 %q, got %q", want, got)
	}
}

func TestRunChangeCloseRecursiveSupersededByAddsReciprocalLinksForAllTargets(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	successorID := runCLIChangeNewForTest(t, projectRoot, "Successor change")
	writeCLIBidirectionalRelatedLink(t, projectRoot, umbrellaID, featureID)
	writeCLIStatusVerificationStatus(t, projectRoot, umbrellaID, "passed")
	writeCLIStatusVerificationStatus(t, projectRoot, featureID, "passed")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", umbrellaID, "--closed-at", "2026-03-21", "--recursive", "--superseded-by", successorID, "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	assertCLISuccessorSupersedesContains(t, projectRoot, successorID, umbrellaID, featureID)
}

func runCLIProjectChangeNewForTest(t *testing.T, projectRoot, title string) string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", title, "--type", "project", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("project change new failed: %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())["change_id"]
}

func writeCLIBidirectionalRelatedLink(t *testing.T, projectRoot, leftID, rightID string) {
	t.Helper()
	appendCLIRelatedLink(t, filepath.Join(projectRoot, "runecontext", "changes", leftID, "status.yaml"), rightID)
	appendCLIRelatedLink(t, filepath.Join(projectRoot, "runecontext", "changes", rightID, "status.yaml"), leftID)
}

func appendCLIRelatedLink(t *testing.T, statusPath, relatedID string) {
	t.Helper()
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status %s: %v", statusPath, err)
	}
	before := strings.ReplaceAll(string(data), "\r\n", "\n")
	after := strings.Replace(before, "related_changes: []", "related_changes:\n  - "+relatedID, 1)
	if before == after {
		t.Fatalf("expected related_changes: [] in %s", statusPath)
	}
	if err := os.WriteFile(statusPath, []byte(after), 0o644); err != nil {
		t.Fatalf("write status %s: %v", statusPath, err)
	}
}

func writeCLIStatusVerificationStatus(t *testing.T, projectRoot, changeID, verificationStatus string) {
	t.Helper()
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status %s: %v", statusPath, err)
	}
	before := strings.ReplaceAll(string(data), "\r\n", "\n")
	after := strings.Replace(before, "verification_status: pending", "verification_status: "+verificationStatus, 1)
	if before == after {
		t.Fatalf("expected verification_status: pending in %s", statusPath)
	}
	if err := os.WriteFile(statusPath, []byte(after), 0o644); err != nil {
		t.Fatalf("write status %s: %v", statusPath, err)
	}
}

func assertCLISuccessorSupersedesContains(t *testing.T, projectRoot, successorID string, targetIDs ...string) {
	t.Helper()
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", successorID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read successor status %s: %v", statusPath, err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, targetID := range targetIDs {
		if !strings.Contains(text, "  - "+targetID) {
			t.Fatalf("expected successor %s to supersede %s, got:\n%s", successorID, targetID, text)
		}
	}
}
