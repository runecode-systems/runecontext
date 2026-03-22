package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func TestRunChangeNewNonInteractiveRequiresExplicitInferenceFlags(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--non-interactive", "--title", "Add cache invalidation", "--type", "feature", "--path", projectRoot}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--non-interactive requires explicit --size, --shape, --bundle") {
		t.Fatalf("expected explicit non-interactive inference failure, got %q", stderr.String())
	}
}

func TestRunChangeNewNonInteractiveWithExplicitInputsSucceeds(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "new", "--non-interactive", "--title", "Add cache invalidation", "--type", "feature", "--size", "small", "--shape", "minimum", "--bundle", "base", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["non_interactive"], "true"; got != want {
		t.Fatalf("expected non_interactive %q, got %q", want, got)
	}
}

func TestRunStatusExplainIncludesResolutionDetails(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--explain", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if !strings.Contains(fields["explain_scope"], "resolution") {
		t.Fatalf("expected resolution explain scope, got %#v", fields)
	}
	if fields["explain_resolution_source_mode"] == "" {
		t.Fatalf("expected explain resolution source mode, got %#v", fields)
	}
}

func TestRunValidateExplainIncludesResolutionStrategy(t *testing.T) {
	root := fixtureRoot(t, "valid-project")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--explain", root}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["explain_resolution_strategy"], "explicit_root"; got != want {
		t.Fatalf("expected explain_resolution_strategy %q, got %q", want, got)
	}
}

func TestRunChangeCloseExplainIncludesPromotionSuggestions(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var createOut bytes.Buffer
	var createErr bytes.Buffer
	if code := Run([]string{"change", "new", "--title", "Refresh security baseline", "--type", "standard", "--size", "small", "--bundle", "base", "--path", projectRoot}, &createOut, &createErr); code != 0 {
		t.Fatalf("expected create success, got %d (%s)", code, createErr.String())
	}
	changeID := parseCLIKeyValueOutput(t, createOut.String())["change_id"]
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"change", "close", "--explain", changeID, "--verification-status", "passed", "--closed-at", "2026-03-21", "--path", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["explain_promotion_status"], "suggested"; got != want {
		t.Fatalf("expected explain_promotion_status %q, got %q", want, got)
	}
	if got := fields["explain_promotion_target_count"]; got != "1" {
		t.Fatalf("expected one promotion target, got %q", got)
	}
	if !strings.HasPrefix(fields["explain_promotion_target_1"], "standard:") {
		t.Fatalf("expected standard promotion target, got %q", fields["explain_promotion_target_1"])
	}
}

func TestRunStatusJSONGolden(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"status", "--json", projectRoot}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	assertJSONGolden(t, "status-success.json", stdout.Bytes())
}

func TestRunValidateUsageJSONGolden(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"validate", "--json", "a", "b"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	assertJSONGolden(t, "validate-usage-error.json", stderr.Bytes())
}

func TestCLIParityStatusMatchesLibrarySummary(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	fields := runCLIStatus(t, projectRoot)
	_, validator, loaded, err := loadProjectForCLI(projectRoot, true)
	if err != nil {
		t.Fatalf("load project for library parity: %v", err)
	}
	defer loaded.Close()
	summary, err := contracts.BuildProjectStatusSummary(validator, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}
	if got, want := fields["active_count"], "0"; got != want {
		t.Fatalf("expected active_count %q, got %q", want, got)
	}
	if got, want := fields["bundle_count"], "2"; got != want {
		t.Fatalf("expected bundle_count %q, got %q", want, got)
	}
	if got, want := fields["bundle_1"], summary.BundleIDs[0]; got != want {
		t.Fatalf("expected bundle_1 %q, got %q", want, got)
	}
	if got, want := fields["bundle_2"], summary.BundleIDs[1]; got != want {
		t.Fatalf("expected bundle_2 %q, got %q", want, got)
	}
}

func assertJSONGolden(t *testing.T, fixtureName string, payload []byte) {
	t.Helper()
	normalized := normalizeJSONEnvelopeForGolden(t, payload)
	goldenPath := repoFixtureRoot(t, "cli-json-golden", fixtureName)
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}
	if strings.TrimSpace(string(expected)) != normalized {
		t.Fatalf("unexpected JSON golden for %s\nexpected: %s\nactual:   %s", fixtureName, string(expected), normalized)
	}
}

func normalizeJSONEnvelopeForGolden(t *testing.T, payload []byte) string {
	t.Helper()
	envelope := parseJSONEnvelopeForGolden(t, payload)
	normalizeJSONEnvelopePaths(envelope.Data)
	normalizeJSONEnvelopeChangeFields(envelope.Data)
	normalizeJSONEnvelopeTargets(envelope.Data)
	normalized, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal normalized JSON envelope: %v", err)
	}
	return string(normalized)
}

func parseJSONEnvelopeForGolden(t *testing.T, payload []byte) *struct {
	SchemaVersion int               `json:"schema_version"`
	Result        string            `json:"result"`
	Command       string            `json:"command"`
	ExitCode      int               `json:"exit_code"`
	FailureClass  string            `json:"failure_class"`
	Data          map[string]string `json:"data"`
} {
	var envelope struct {
		SchemaVersion int               `json:"schema_version"`
		Result        string            `json:"result"`
		Command       string            `json:"command"`
		ExitCode      int               `json:"exit_code"`
		FailureClass  string            `json:"failure_class"`
		Data          map[string]string `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("unmarshal JSON envelope: %v (%s)", err, string(payload))
	}
	return &envelope
}

func normalizeJSONEnvelopePaths(data map[string]string) {
	for _, key := range []string{"root", "project_root", "source_root", "selected_config_path", "error_path"} {
		if _, ok := data[key]; ok {
			data[key] = "<path>"
		}
	}
}

func normalizeJSONEnvelopeChangeFields(data map[string]string) {
	if _, ok := data["change_id"]; ok {
		data["change_id"] = "<change-id>"
	}
	if _, ok := data["change_path"]; ok {
		data["change_path"] = "<path>"
	}
}

func normalizeJSONEnvelopeTargets(data map[string]string) {
	for key := range data {
		if strings.HasPrefix(key, "changed_file_") || key == "changed_file_count" {
			delete(data, key)
			continue
		}
		if key == "target_count" {
			data[key] = "<target-count>"
			continue
		}
		if strings.HasPrefix(key, "target_") {
			data[key] = "<target>"
		}
	}
}
