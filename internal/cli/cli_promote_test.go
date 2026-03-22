package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPromoteUsageMentionsSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"promote", "--accept", "--complete"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "summary auto-filled per target type") {
		t.Fatalf("expected usage note about auto-filled summaries, got %q", stderr.String())
	}
}

func TestRunPromoteTargetSummaryDefaults(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Refresh security baseline")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot})
	target := "standard:standards/cli-promotion-target.md"
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"promote", changeID, "--accept", "--target", target, "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected promote accept success, got %d (%s)", code, stderr.String())
	}
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "status.yaml")
	content, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	if !strings.Contains(text, "target_path: standards/cli-promotion-target.md") {
		t.Fatalf("expected custom target path, got:\n%s", text)
	}
	if !strings.Contains(text, "summary: Review and promote durable standards updates from this change.") {
		t.Fatalf("expected default summary in status.yaml, got:\n%s", text)
	}
}

func TestRunPromoteJSONGolden(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Refresh security baseline")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-20", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"promote", "--json", changeID, "--accept", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected promote --json success, got %d (%s)", code, stderr.String())
	}
	assertJSONGolden(t, "promote-success.json", stdout.Bytes())
}
