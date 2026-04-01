package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunChangeDecompositionPlanAllowsTerminalSubChange(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	writeCLIStatusVerificationStatus(t, projectRoot, featureID, "passed")
	runCLIChangeClose(t, projectRoot, featureID, []string{"--closed-at", "2026-03-21", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"change", "decomposition-plan", umbrellaID,
		"--sub-change", featureID,
		"--path", projectRoot,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected plan to allow terminal sub-change, got %d (%s)", code, stderr.String())
	}
	fields := parseCLIKeyValueOutput(t, stdout.String())
	if got, want := fields["command"], "change_decomposition_plan"; got != want {
		t.Fatalf("expected command %q, got %q", want, got)
	}
}

func TestRunChangeDecompositionApplyRejectsTerminalSubChange(t *testing.T) {
	projectRoot := prepareCLIWorkflowProject(t)
	umbrellaID := runCLIProjectChangeNewForTest(t, projectRoot, "Umbrella project")
	featureID := runCLIChangeNewForTest(t, projectRoot, "Feature sub-change")
	writeCLIStatusVerificationStatus(t, projectRoot, featureID, "passed")
	runCLIChangeClose(t, projectRoot, featureID, []string{"--closed-at", "2026-03-21", "--path", projectRoot})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{
		"change", "decomposition-apply", umbrellaID,
		"--sub-change", featureID,
		"--path", projectRoot,
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected invalid exit for terminal sub-change apply, got %d (%s)", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "cannot accept decomposition apply edits") {
		t.Fatalf("expected terminal apply rejection output, got %q", stderr.String())
	}
}
