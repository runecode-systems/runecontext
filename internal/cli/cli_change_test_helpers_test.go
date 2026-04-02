package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func prepareCLIWorkflowProject(t *testing.T) string {
	t.Helper()
	repoRoot, err := repoRootForTests()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(repoRoot)
	projectRoot := t.TempDir()
	copyDirForCLI(t, repoFixtureRoot(t, "change-workflow", "template-project"), projectRoot)
	return projectRoot
}

func appendCLIProposalSelfReference(t *testing.T, projectRoot, changeID string) {
	t.Helper()
	proposalPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "proposal.md")
	data, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("read proposal: %v", err)
	}
	updated := strings.ReplaceAll(string(data), "\r\n", "\n") + "\nSee changes/" + changeID + "/proposal.md#summary for the current change summary.\n"
	if err := os.WriteFile(proposalPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
}

func runCLIChangeReallocate(t *testing.T, projectRoot, changeID string) map[string]string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "reallocate", changeID, "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())
}

func assertCLIReallocateFields(t *testing.T, fields map[string]string, changeID string) string {
	t.Helper()
	if got, want := fields["command"], "change_reallocate"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["old_change_id"], changeID; got != want {
		t.Fatalf("expected old_change_id %q, got %q", want, got)
	}
	newID := fields["change_id"]
	if newID == "" || newID == changeID {
		t.Fatalf("expected a new change ID, got %#v", fields)
	}
	if got := fields["rewritten_reference_count"]; got != "1" {
		t.Fatalf("expected one rewritten reference, got %q", got)
	}
	if got := fields["warning_count"]; got != "0" {
		t.Fatalf("expected no warnings, got %q", got)
	}
	return newID
}

func assertCLIReallocatedProposal(t *testing.T, projectRoot, oldID, newID string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(projectRoot, "runecontext", "changes", oldID)); !os.IsNotExist(err) {
		t.Fatalf("expected old change directory to be removed, got err=%v", err)
	}
	proposalData, err := os.ReadFile(filepath.Join(projectRoot, "runecontext", "changes", newID, "proposal.md"))
	if err != nil {
		t.Fatalf("read reallocated proposal: %v", err)
	}
	if !strings.Contains(string(proposalData), "changes/"+newID+"/proposal.md#summary") {
		t.Fatalf("expected CLI reallocation to rewrite local reference, got:\n%s", string(proposalData))
	}
}

func runCLIChangeClose(t *testing.T, projectRoot, changeID string, args []string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	fullArgs := append([]string{"change", "close", changeID}, args...)
	if code := Run(fullArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("change close failed: %d (%s)", code, stderr.String())
	}
}

func runCLIStandardChangeNewForTest(t *testing.T, projectRoot, title string) string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--title", title, "--type", "standard", "--size", "small", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("standard change new failed: %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())["change_id"]
}
