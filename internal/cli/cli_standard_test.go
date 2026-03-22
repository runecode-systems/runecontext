package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStandardDiscoverNonInteractiveOutputsCandidates(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--non-interactive", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], standardDiscoverCommand; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["non_interactive"], "true"; got != want {
		t.Fatalf("expected non_interactive %q, got %q", want, got)
	}
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["candidate_standard_count"], "2"; got != want {
		t.Fatalf("expected candidate_standard_count %q, got %q", want, got)
	}
	if got, want := fields["candidate_standard_1"], "standards/global/base.md"; got != want {
		t.Fatalf("expected first candidate standard %q, got %q", want, got)
	}
	if got, want := fields["candidate_promotion_target_1"], "standard:standards/global/base.md"; got != want {
		t.Fatalf("expected first promotion target %q, got %q", want, got)
	}
	if got, want := fields["handoff_requested"], "false"; got != want {
		t.Fatalf("expected handoff_requested %q, got %q", want, got)
	}
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
}

func TestRunStandardDiscoverExplainAddsExplainLines(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--explain", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if !strings.Contains(fields["explain_scope"], "standards-discovery") {
		t.Fatalf("expected standards-discovery explain scope, got %#v", fields)
	}
	if got, want := fields["explain_advisory_only"], "true"; got != want {
		t.Fatalf("expected explain_advisory_only %q, got %q", want, got)
	}
	if fields["explain_candidate_standard_count_reason"] == "" {
		t.Fatalf("expected explain candidate reason, got %#v", fields)
	}
}

func TestRunStandardDiscoverHandoffRequiresExplicitConfirmation(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Refresh standards discovery docs")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--change", changeID, "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["handoff_requested"], "true"; got != want {
		t.Fatalf("expected handoff_requested %q, got %q", want, got)
	}
	if got, want := fields["handoff_confirmed"], "false"; got != want {
		t.Fatalf("expected handoff_confirmed %q, got %q", want, got)
	}
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_blocked_reason"], "missing_explicit_confirmation"; got != want {
		t.Fatalf("expected handoff_blocked_reason %q, got %q", want, got)
	}
	if fields["handoff_command"] != "" {
		t.Fatalf("expected no handoff_command without confirmation, got %#v", fields)
	}
}

func TestRunStandardDiscoverHandoffEmitsExplicitCommandWhenConfirmed(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Promote durable standard update")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--change", changeID, "--confirm-handoff", "--target", "standard:standards/global/base.md", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["handoff_confirmed"], "true"; got != want {
		t.Fatalf("expected handoff_confirmed %q, got %q", want, got)
	}
	if got, want := fields["handoff_eligible"], "true"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_promotion_target"], "standard:standards/global/base.md"; got != want {
		t.Fatalf("expected handoff_promotion_target %q, got %q", want, got)
	}
	if fields["handoff_command"] != "" {
		t.Fatalf("expected no shell handoff command output, got %#v", fields)
	}
	if fields["handoff_blocked_reason"] != "" {
		t.Fatalf("expected no handoff block reason, got %#v", fields)
	}
}

func TestRunStandardDiscoverHandoffBlocksAmbiguousTargetSelection(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Require explicit handoff target")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--change", changeID, "--confirm-handoff", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_target_required"], "true"; got != want {
		t.Fatalf("expected handoff_target_required %q, got %q", want, got)
	}
	if got, want := fields["handoff_blocked_reason"], "ambiguous_candidate_targets"; got != want {
		t.Fatalf("expected handoff_blocked_reason %q, got %q", want, got)
	}
}

func TestRunStandardDiscoverHandoffRejectsUnknownTarget(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Reject unknown handoff target")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--change", changeID, "--confirm-handoff", "--target", "standard:standards/missing.md", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_blocked_reason"], "target_not_in_candidates"; got != want {
		t.Fatalf("expected handoff_blocked_reason %q, got %q", want, got)
	}
}

func TestRunStandardDiscoverHandoffRequiresChangeWhenTargetProvided(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--target", "standard:standards/global/base.md", "--path", projectRoot}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--confirm-handoff and --target require --change") {
		t.Fatalf("expected clear usage error, got %q", stderr.String())
	}
}

func TestRunStandardDiscoverHandoffBlocksWhenPromotionStatusNotSuggested(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Require suggested status before handoff")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	securityPath := filepath.Join(projectRoot, "runecontext", "standards", "security", "review.md")
	setStandardStatusForTest(t, securityPath, "deprecated")
	rewriteChangePromotionAssessmentStatus(t, projectRoot, changeID, "none")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--change", changeID, "--confirm-handoff", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_blocked_reason"], "promotion_status_not_suggested"; got != want {
		t.Fatalf("expected handoff_blocked_reason %q, got %q", want, got)
	}
}

func TestRunStandardDiscoverNonInteractiveConfirmedHandoffStaysAdvisory(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	changeID := runCLIStandardChangeNewForTest(t, projectRoot, "Keep non-interactive advisory")
	runCLIChangeClose(t, projectRoot, changeID, []string{"--verification-status", "passed", "--closed-at", "2026-03-22", "--path", projectRoot})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--non-interactive", "--change", changeID, "--confirm-handoff", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["handoff_eligible"], "false"; got != want {
		t.Fatalf("expected handoff_eligible %q, got %q", want, got)
	}
	if got, want := fields["handoff_confirmed"], "false"; got != want {
		t.Fatalf("expected handoff_confirmed %q, got %q", want, got)
	}
	if got, want := fields["handoff_blocked_reason"], "non_interactive_requires_explicit_confirmation"; got != want {
		t.Fatalf("expected handoff_blocked_reason %q, got %q", want, got)
	}
	if fields["handoff_command"] != "" {
		t.Fatalf("expected no handoff command in non-interactive mode, got %#v", fields)
	}
}

func TestRunStandardDiscoverNoActiveStandardsOutputsReusableEmptyCandidates(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	standardPath := filepath.Join(projectRoot, "runecontext", "standards", "global", "base.md")
	setStandardStatusForTest(t, standardPath, "deprecated")
	securityPath := filepath.Join(projectRoot, "runecontext", "standards", "security", "review.md")
	setStandardStatusForTest(t, securityPath, "deprecated")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"standard", "discover", "--non-interactive", "--path", projectRoot}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["candidate_standard_count"], "0"; got != want {
		t.Fatalf("expected candidate_standard_count %q, got %q", want, got)
	}
	if got, want := fields["candidate_promotion_target_count"], "0"; got != want {
		t.Fatalf("expected candidate_promotion_target_count %q, got %q", want, got)
	}
}

func setStandardStatusForTest(t *testing.T, path, status string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read standard: %v", err)
	}
	updated := strings.Replace(string(data), "status: active", "status: "+status, 1)
	if updated == string(data) {
		t.Fatalf("expected active status in %s", path)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("write standard: %v", err)
	}
}

func rewriteChangePromotionAssessmentStatus(t *testing.T, projectRoot, changeID, status string) {
	t.Helper()
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", changeID, "status.yaml")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	updated := strings.Replace(string(data), "status: suggested", fmt.Sprintf("status: %s", status), 1)
	if updated == string(data) {
		t.Fatalf("expected promotion_assessment block in %s", statusPath)
	}
	if err := os.WriteFile(statusPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
}
