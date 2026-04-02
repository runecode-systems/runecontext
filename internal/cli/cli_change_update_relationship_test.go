package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunChangeUpdateRelationshipEditsPersistAndRemainReciprocal(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "First change")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Second change")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "update", firstID, "--status", "planned", "--add-related-change", secondID, "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["change_status"], "planned"; got != want {
		t.Fatalf("expected change_status %q, got %q", want, got)
	}
	if got, want := fields["related_change_count"], "1"; got != want {
		t.Fatalf("expected related_change_count %q, got %q", want, got)
	}
	if got, want := fields["related_change_1"], secondID; got != want {
		t.Fatalf("expected related_change_1 %q, got %q", want, got)
	}
	assertCLIReciprocalRelatedLink(t, projectRoot, firstID, secondID)
}

func TestRunChangeUpdateRelationshipEditUsageErrors(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	firstID := runCLIChangeNewForTest(t, projectRoot, "First change")
	secondID := runCLIChangeNewForTest(t, projectRoot, "Second change")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "update", firstID, "--status", "planned", "--add-related-change", secondID, "--remove-related-change", secondID, "--path", projectRoot}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit code for conflicting related-change edits, got %d", code)
	}
	if !strings.Contains(stderr.String(), "relationship edit lists conflict") {
		t.Fatalf("expected conflict validation output, got %q", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"change", "update", firstID, "--status", "planned", "--add-related-change", " ", "--path", projectRoot}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code for blank related-change ID, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--add-related-change requires a value") {
		t.Fatalf("expected blank-related-change usage output, got %q", stderr.String())
	}
}

func assertCLIReciprocalRelatedLink(t *testing.T, projectRoot, firstID, secondID string) {
	t.Helper()
	assertCLIStatusContainsRelated(t, filepath.Join(projectRoot, "runecontext", "changes", firstID, "status.yaml"), secondID)
	assertCLIStatusContainsRelated(t, filepath.Join(projectRoot, "runecontext", "changes", secondID, "status.yaml"), firstID)
}

func assertCLIStatusContainsRelated(t *testing.T, statusPath, relatedID string) {
	t.Helper()
	statusBytes, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status %s: %v", statusPath, err)
	}
	statusText := strings.ReplaceAll(string(statusBytes), "\r\n", "\n")
	if !strings.Contains(statusText, "  - "+relatedID) {
		t.Fatalf("expected %s to include related change %s, got:\n%s", statusPath, relatedID, statusText)
	}
}
