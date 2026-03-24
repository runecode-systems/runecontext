package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRunAssuranceBackfillRequiresVerifiedTier(t *testing.T) {
	repoRoot, commits := createAssuranceBackfillRepo(t)
	config := "schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: plain\nsource:\n  type: embedded\n  path: runecontext\n"
	if err := os.WriteFile(filepath.Join(repoRoot, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write runecontext config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir runecontext source: %v", err)
	}
	writeAssuranceBackfillBaselineFixture(t, repoRoot, commits[0])

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--path", repoRoot}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "assurance_tier must be verified") {
		t.Fatalf("expected verified-tier error, got %q", stderr.String())
	}
}

func TestRunAssuranceBackfillHelpTokens(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--help"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected help exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage=runectx assurance backfill") {
		t.Fatalf("expected assurance backfill usage output, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunAssuranceBackfillHelpRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--help", "extra"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "help does not accept additional arguments") {
		t.Fatalf("expected help extra-arg error, got %q", stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
}

func TestRunAssuranceBackfillDryRun(t *testing.T) {
	repoRoot, commits := createAssuranceBackfillRepo(t)
	writeBackfillConfigFixture(t, repoRoot, "verified")
	writeAssuranceBackfillBaselineFixture(t, repoRoot, commits[1])

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--dry-run", "--path", repoRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected dry-run success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got := fields["command"]; got != "assurance backfill" {
		t.Fatalf("unexpected command %q", got)
	}
	if got := fields["mode"]; got != "imported-git-history" {
		t.Fatalf("unexpected mode %q", got)
	}
	if got := fields["plan_action_1"]; got == "" {
		t.Fatalf("expected first plan action, got %#v", fields)
	}
	if got := fields["dry_run"]; got != "true" {
		t.Fatalf("expected dry_run=true, got %q", got)
	}
	if !strings.Contains(stderr.String(), "Dry run: would run assurance backfill validation") {
		t.Fatalf("expected dry-run stderr message, got %q", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "assurance", "backfill")); !os.IsNotExist(err) {
		t.Fatalf("expected no backfill artifacts created during dry-run")
	}
}

func TestRunAssuranceBackfillDryRunFailsInNonGitRepo(t *testing.T) {
	root := t.TempDir()
	writeBackfillConfigFixture(t, root, "verified")
	writeAssuranceBackfillBaselineFixture(t, root, "1234567890abcdef1234567890abcdef12345678")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--dry-run", "--path", root}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "assurance backfill requires a git repository") {
		t.Fatalf("expected non-git repository error, got %q", stderr.String())
	}
}

func TestRunAssuranceBackfillNonGitRepoError(t *testing.T) {
	root := t.TempDir()
	writeBackfillConfigFixture(t, root, "verified")
	writeAssuranceBackfillBaselineFixture(t, root, "1234567890abcdef1234567890abcdef12345678")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--path", root}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "assurance backfill requires a git repository") {
		t.Fatalf("expected non-git repository error, got %q", stderr.String())
	}
}

func TestRunAssuranceBackfillRejectsNonCanonicalAdoptionCommit(t *testing.T) {
	repoRoot, _ := createAssuranceBackfillRepo(t)
	writeBackfillConfigFixture(t, repoRoot, "verified")
	writeAssuranceBackfillBaselineFixture(t, repoRoot, "HEAD")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--path", repoRoot}, &stdout, &stderr)
	if code != exitInvalid {
		t.Fatalf("expected invalid exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "canonical lowercase 40-char hex SHA") {
		t.Fatalf("expected canonical adoption_commit error, got %q", stderr.String())
	}
}

func TestRunAssuranceBackfillCreatesHistoryAndUpdatesBaselineAdditively(t *testing.T) {
	repoRoot, commits := createAssuranceBackfillRepo(t)
	writeBackfillConfigFixture(t, repoRoot, "verified")
	writeAssuranceBackfillBaselineFixture(t, repoRoot, commits[1])
	receiptPath, receiptContent := writeBackfillReceiptFixture(t, repoRoot)

	fields := runAssuranceBackfillForTest(t, repoRoot)
	assertBackfillFirstRunState(t, repoRoot, commits, fields)

	secondFields := runAssuranceBackfillForTest(t, repoRoot)
	assertBackfillSecondRunState(t, repoRoot, secondFields)
	assertBackfillReceiptUntouched(t, receiptPath, receiptContent)
}

func createAssuranceBackfillRepo(t *testing.T) (string, []string) {
	t.Helper()
	requireToolForCLITests(t, "git")
	repoRoot := t.TempDir()
	runGitForCLI(t, repoRoot, "init", "--initial-branch=main")
	runGitForCLI(t, repoRoot, "config", "user.name", "RuneContext Backfill Tests")
	runGitForCLI(t, repoRoot, "config", "user.email", "backfill-tests@example.com")

	trackedPath := filepath.Join(repoRoot, "history.txt")
	commits := make([]string, 0, 3)
	for i := 1; i <= 3; i++ {
		content := fmt.Sprintf("commit-%d\n", i)
		if err := os.WriteFile(trackedPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write history commit file: %v", err)
		}
		runGitForCLI(t, repoRoot, "add", "history.txt")
		runGitForCLI(t, repoRoot, "commit", "-m", fmt.Sprintf("commit %d", i))
		commit := strings.TrimSpace(gitOutputForCLI(t, repoRoot, "rev-parse", "HEAD"))
		commits = append(commits, commit)
	}
	return repoRoot, commits
}

func writeAssuranceBackfillBaselineFixture(t *testing.T, root, adoptionCommit string) {
	t.Helper()
	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
		t.Fatalf("mkdir baseline directory: %v", err)
	}
	baseline := fmt.Sprintf("schema_version: 1\nkind: baseline\nsubject_id: project-root\ncreated_at: 1710000000\ncanonicalization: runecontext-canonical-json-v1\nvalue:\n  adoption_commit: %s\n  source_posture: embedded\n", adoptionCommit)
	if err := os.WriteFile(baselinePath, []byte(baseline), 0o644); err != nil {
		t.Fatalf("write baseline fixture: %v", err)
	}
}

func writeBackfillConfigFixture(t *testing.T, root, tier string) {
	t.Helper()
	config := fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.3\nassurance_tier: %s\nsource:\n  type: embedded\n  path: runecontext\n", tier)
	if err := os.WriteFile(filepath.Join(root, "runecontext.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write runecontext config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "runecontext"), 0o755); err != nil {
		t.Fatalf("mkdir runecontext source: %v", err)
	}
}

func writeBackfillReceiptFixture(t *testing.T, root string) (string, string) {
	t.Helper()
	receiptPath := filepath.Join(root, "assurance", "receipts", "changes", "receipt.json")
	if err := os.MkdirAll(filepath.Dir(receiptPath), 0o755); err != nil {
		t.Fatalf("mkdir receipts dir: %v", err)
	}
	receiptContent := "{\"provenance\":\"captured_verified\",\"receipt_id\":\"r-1\"}\n"
	if err := os.WriteFile(receiptPath, []byte(receiptContent), 0o644); err != nil {
		t.Fatalf("write receipt fixture: %v", err)
	}
	return receiptPath, receiptContent
}

func runAssuranceBackfillForTest(t *testing.T, root string) map[string]string {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"assurance", "backfill", "--path", root}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success, got %d (%s)", code, stderr.String())
	}
	return parseCLIKeyValueOutput(t, stdout.String())
}

func assertBackfillFirstRunState(t *testing.T, repoRoot string, commits []string, fields map[string]string) {
	t.Helper()
	assertBackfillFirstRunCLIFields(t, fields)
	historyPath := assertBackfillHistoryPathPresent(t, fields)
	assertBackfillFirstRunHistoryRecord(t, commits, historyPath)
	assertBackfillImportedEvidenceAdded(t, repoRoot)
}

func assertBackfillFirstRunCLIFields(t *testing.T, fields map[string]string) {
	t.Helper()
	if got := fields["command"]; got != "assurance backfill" {
		t.Fatalf("unexpected command %q", got)
	}
	if got := fields["result"]; got != "ok" {
		t.Fatalf("unexpected result %q", got)
	}
	if got := fields["history_commit_count"]; got != "2" {
		t.Fatalf("unexpected history commit count %q", got)
	}
	if got := fields["imported_evidence_added"]; got != "true" {
		t.Fatalf("expected first run to append imported evidence, got %q", got)
	}
}

func assertBackfillHistoryPathPresent(t *testing.T, fields map[string]string) string {
	t.Helper()
	historyPath := fields["history_path"]
	if historyPath == "" {
		t.Fatalf("expected history path in output")
	}
	return historyPath
}

func assertBackfillFirstRunHistoryRecord(t *testing.T, commits []string, historyPath string) {
	t.Helper()
	history := readImportedHistoryRecord(t, historyPath)
	if history.Provenance != "imported_git_history" {
		t.Fatalf("unexpected history provenance %q", history.Provenance)
	}
	if history.AdoptionCommit != commits[1] {
		t.Fatalf("unexpected adoption commit %q", history.AdoptionCommit)
	}
	if len(history.Commits) != 2 {
		t.Fatalf("expected commits bounded to adoption point, got %d", len(history.Commits))
	}
	if history.Commits[1].Commit != commits[1] {
		t.Fatalf("expected adoption commit as final imported commit, got %q", history.Commits[1].Commit)
	}
}

func assertBackfillImportedEvidenceAdded(t *testing.T, repoRoot string) {
	t.Helper()
	baselinePath := filepath.Join(repoRoot, "assurance", "baseline.yaml")
	importedEvidence := readImportedEvidence(t, readBaselineMapForBackfill(t, baselinePath))
	if len(importedEvidence) != 1 {
		t.Fatalf("expected one imported evidence entry, got %d", len(importedEvidence))
	}
	if readOptionalString(importedEvidence[0], "provenance") != "imported_git_history" {
		t.Fatalf("unexpected imported provenance %q", readOptionalString(importedEvidence[0], "provenance"))
	}
}

func assertBackfillSecondRunState(t *testing.T, repoRoot string, fields map[string]string) {
	t.Helper()
	if got := fields["imported_evidence_added"]; got != "false" {
		t.Fatalf("expected second run to avoid duplicate imported evidence, got %q", got)
	}
	baselinePath := filepath.Join(repoRoot, "assurance", "baseline.yaml")
	if len(readImportedEvidence(t, readBaselineMapForBackfill(t, baselinePath))) != 1 {
		t.Fatalf("expected imported evidence to remain de-duplicated")
	}
}

func TestEmitAssuranceBackfillSuccessNoChangesMessage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	machine := machineOptions{}
	result := assuranceBackfillResult{
		baselinePath:   "/tmp/project/assurance/baseline.yaml",
		historyPath:    "/tmp/project/assurance/backfill/imported-git-history-abc.json",
		adoptionCommit: "1234567890abcdef1234567890abcdef12345678",
		commitCount:    0,
		importedAdded:  false,
	}

	code := emitAssuranceBackfillSuccess(&stdout, &stderr, machine, "/tmp/project", result)
	if code != exitOK {
		t.Fatalf("expected exitOK, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Backfill is up to date") {
		t.Fatalf("expected idempotent backfill message, got %q", stderr.String())
	}
}

func assertBackfillReceiptUntouched(t *testing.T, receiptPath, expectedContent string) {
	t.Helper()
	receiptAfter, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatalf("read receipt after backfill: %v", err)
	}
	if string(receiptAfter) != expectedContent {
		t.Fatalf("expected receipts to remain untouched; got %q", string(receiptAfter))
	}
}

func TestBackfillRelativePathWithinRootRejectsEscape(t *testing.T) {
	repoRoot := t.TempDir()
	escapePath := filepath.Join(filepath.Dir(repoRoot), "outside.json")
	_, err := backfillRelativePathWithinRoot(repoRoot, escapePath)
	if err == nil {
		t.Fatalf("expected escape path error")
	}
	if !strings.Contains(err.Error(), "escapes repository root") {
		t.Fatalf("unexpected error %q", err)
	}
}

func readImportedHistoryRecord(t *testing.T, path string) importedGitHistoryRecord {
	t.Helper()
	historyData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read history record: %v", err)
	}
	var history importedGitHistoryRecord
	if err := json.Unmarshal(historyData, &history); err != nil {
		t.Fatalf("parse history json: %v", err)
	}
	return history
}

func readBaselineMapForBackfill(t *testing.T, baselinePath string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var baseline map[string]any
	if err := yaml.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("parse baseline yaml: %v", err)
	}
	return baseline
}

func readImportedEvidence(t *testing.T, baseline map[string]any) []map[string]any {
	t.Helper()
	value, ok := baseline["value"].(map[string]any)
	if !ok {
		t.Fatalf("baseline value should be object, got %#v", baseline["value"])
	}
	rawEntries, ok := value["imported_evidence"].([]any)
	if !ok {
		t.Fatalf("baseline imported_evidence should be list, got %#v", value["imported_evidence"])
	}
	entries := make([]map[string]any, 0, len(rawEntries))
	for _, raw := range rawEntries {
		entry, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("unexpected imported evidence entry %#v", raw)
		}
		entries = append(entries, entry)
	}
	return entries
}
