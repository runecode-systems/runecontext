package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunChangeDecompositionPlanOutputsAdvisoryGraph(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureAID := runCLIChangeNewForTest(t, projectRoot, "Feature A")
	featureBID := runCLIChangeNewForTest(t, projectRoot, "Feature B")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"change", "decomposition-plan", umbrellaID,
		"--sub-change", featureAID,
		"--sub-change", featureBID,
		"--depends-on", featureBID + ":" + featureAID,
		"--path", projectRoot,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "change_decomposition_plan"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got, want := fields["mutation_performed"], "false"; got != want {
		t.Fatalf("expected mutation_performed %q, got %q", want, got)
	}
	if got, want := fields["umbrella_change_id"], umbrellaID; got != want {
		t.Fatalf("expected umbrella_change_id %q, got %q", want, got)
	}
	if got, want := fields["graph_"+umbrellaID+"_related_change_count"], "2"; got != want {
		t.Fatalf("expected umbrella related count %q, got %q", want, got)
	}
	if got, want := fields["graph_"+featureBID+"_depends_on_1"], featureAID; got != want {
		t.Fatalf("expected %q dependency %q, got %q", featureBID, want, got)
	}
}

func TestRunChangeDecompositionApplyUpdatesGraph(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureAID := runCLIChangeNewForTest(t, projectRoot, "Feature A")
	featureBID := runCLIChangeNewForTest(t, projectRoot, "Feature B")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"change", "decomposition-apply", umbrellaID,
		"--sub-change", featureAID,
		"--sub-change", featureBID,
		"--depends-on", featureBID + ":" + featureAID,
		"--path", projectRoot,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected success exit code, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "change_decomposition_apply"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
	if got := fields["changed_file_count"]; got != "3" {
		t.Fatalf("expected changed_file_count 3, got %q", got)
	}
	assertDecompositionApplyStatusLinks(t, projectRoot, umbrellaID, featureAID, featureBID)
}

func assertDecompositionApplyStatusLinks(t *testing.T, projectRoot, umbrellaID, featureAID, featureBID string) {
	t.Helper()
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", umbrellaID, "status.yaml")
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read umbrella status: %v", err)
	}
	text := strings.ReplaceAll(string(statusData), "\r\n", "\n")
	if !strings.Contains(text, "related_changes:\n  - "+featureAID+"\n  - "+featureBID) {
		t.Fatalf("expected umbrella related_changes to include sub-changes, got:\n%s", text)
	}
	featureBStatusPath := filepath.Join(projectRoot, "runecontext", "changes", featureBID, "status.yaml")
	featureBData, err := os.ReadFile(featureBStatusPath)
	if err != nil {
		t.Fatalf("read feature status: %v", err)
	}
	featureBText := strings.ReplaceAll(string(featureBData), "\r\n", "\n")
	if !strings.Contains(featureBText, "depends_on:\n  - "+featureAID) {
		t.Fatalf("expected feature B depends_on %s, got:\n%s", featureAID, featureBText)
	}
}

func TestRunChangeDecompositionApplyDryRunDoesNotPersistMutation(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	statusPath := filepath.Join(projectRoot, "runecontext", "changes", umbrellaID, "status.yaml")
	before, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status before dry-run decomposition apply: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"change", "decomposition-apply", umbrellaID,
		"--sub-change", featureID,
		"--dry-run",
		"--path", projectRoot,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected dry-run success, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["dry_run"], "true"; got != want {
		t.Fatalf("expected dry_run %q, got %q", want, got)
	}
	after, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read status after dry-run decomposition apply: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("expected dry-run decomposition apply to avoid status mutation")
	}
}

func TestRunChangeDecompositionUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		message string
	}{
		{name: "missing umbrella", args: []string{"change", "decomposition-plan"}, message: "requires exactly one umbrella change ID"},
		{name: "plan dry-run", args: []string{"change", "decomposition-plan", "CHG-2026-001-a3f2-auth-gateway", "--dry-run"}, message: "--dry-run is not supported"},
		{name: "malformed dependency", args: []string{"change", "decomposition-plan", "CHG-2026-001-a3f2-auth-gateway", "--sub-change", "CHG-2026-002-b4c3-api", "--depends-on", "CHG-2026-002-b4c3-api"}, message: "--depends-on must use SUB_CHANGE_ID:CHANGE_ID"},
		{name: "depends-on before sub-change", args: []string{"change", "decomposition-plan", "CHG-2026-001-a3f2-auth-gateway", "--depends-on", "CHG-2026-002-b4c3-api:CHG-2026-003-c5d4-ui", "--sub-change", "CHG-2026-002-b4c3-api"}, message: "before it is declared with --sub-change"},
		{name: "missing sub-change", args: []string{"change", "decomposition-apply", "CHG-2026-001-a3f2-auth-gateway"}, message: "requires at least one --sub-change ID"},
		{name: "unknown apply flag", args: []string{"change", "decomposition-apply", "CHG-2026-001-a3f2-auth-gateway", "--sub-change", "CHG-2026-002-b4c3-api", "--recursive"}, message: "unknown change decomposition-apply flag"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertChangeUsageError(t, tc.args, tc.message)
		})
	}
}

func assertChangeUsageError(t *testing.T, args []string, wantSubstring string) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), wantSubstring) {
		t.Fatalf("expected usage output containing %q, got %q", wantSubstring, stderr.String())
	}
}
