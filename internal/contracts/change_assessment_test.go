package contracts

import (
	"strings"
	"testing"
)

func TestAssessChangeIntakeReturnsAdvisorySignals(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	result, err := AssessChangeIntake(v, loaded, ChangeAssessIntakeOptions{
		Title:       "Launch payments platform",
		Type:        "project",
		Description: "Multi-step migration across services",
	})
	if err != nil {
		t.Fatalf("assess intake: %v", err)
	}
	if got, want := result.RecommendedMode, ChangeModeFull; got != want {
		t.Fatalf("expected recommended mode %q, got %q", want, got)
	}
	if got, want := result.IntakeReadiness, "needs_clarification"; got != want {
		t.Fatalf("expected intake_readiness %q, got %q", want, got)
	}
	if !result.ClarificationNeeded || len(result.ClarificationPrompts) == 0 {
		t.Fatalf("expected clarification prompts, got %#v", result.ClarificationPrompts)
	}
	if got, want := result.DecompositionSignal, "consider_decomposition"; got != want {
		t.Fatalf("expected decomposition signal %q, got %q", want, got)
	}
	if len(result.ApplicableStandards) == 0 {
		t.Fatalf("expected standards from bundle resolution")
	}
	if len(result.Assumptions) == 0 {
		t.Fatalf("expected assumptions to be populated")
	}
}

func TestAssessChangeIntakeRejectsMissingTitle(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	_, err := AssessChangeIntake(v, loaded, ChangeAssessIntakeOptions{Type: "feature"})
	if err == nil || !strings.Contains(err.Error(), "requires title") {
		t.Fatalf("expected missing title error, got %v", err)
	}
}

func TestAssessChangeDecompositionDetectsUmbrellaGraph(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, _ := createUmbrellaAndFeatureSubChange(t, root)

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()

	result, err := AssessChangeDecomposition(v, loaded, umbrellaID)
	if err != nil {
		t.Fatalf("assess decomposition: %v", err)
	}
	if got, want := result.DecompositionSignal, "umbrella_graph_detected"; got != want {
		t.Fatalf("expected decomposition signal %q, got %q", want, got)
	}
	if got := len(result.EligibleSubChangeIDs); got != 1 {
		t.Fatalf("expected one eligible sub-change, got %d (%#v)", got, result.EligibleSubChangeIDs)
	}
	if got, want := result.RecommendedMode, ChangeModeFull; got != want {
		t.Fatalf("expected recommended mode %q, got %q", want, got)
	}
}

func TestAssessChangeDecompositionRejectsUnknownChange(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	_, err := AssessChangeDecomposition(v, loaded, "CHG-2099-001-missing")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected unknown change error, got %v", err)
	}
}
