package contracts

import (
	"path/filepath"
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

func TestPlanChangeDecompositionBuildsGraph(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, _ := createUmbrellaAndFeatureSubChange(t, root)
	_, secondFeature := mustCreateChange(t, root, defaultFeatureChangeOptions("Second sub-change", []byte{0x91, 0x92}))

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := PlanChangeDecomposition(v, loaded, ChangeDecompositionPlanOptions{
		UmbrellaID: umbrellaID,
		SubChanges: []SplitSubChange{
			{ID: secondFeature.ID},
			{ID: "CHG-2026-002-3344-feature-sub-change", DependsOn: []string{secondFeature.ID}},
		},
	})
	if err != nil {
		t.Fatalf("plan decomposition: %v", err)
	}
	if got, want := result.UmbrellaID, umbrellaID; got != want {
		t.Fatalf("expected umbrella %q, got %q", want, got)
	}
	if got := len(result.Graph); got != 3 {
		t.Fatalf("expected graph with 3 nodes, got %d", got)
	}
	if got := result.Graph[umbrellaID].RelatedChanges; len(got) != 2 {
		t.Fatalf("expected umbrella related links to both sub-changes, got %#v", got)
	}
}

func TestApplyChangeDecompositionRewritesRelationships(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, umbrellaID, _ := createUmbrellaAndFeatureSubChange(t, root)
	_, secondFeature := mustCreateChange(t, root, defaultFeatureChangeOptions("Second sub-change", []byte{0xa1, 0xa2}))

	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	result, err := ApplyChangeDecomposition(v, loaded, ChangeDecompositionApplyOptions{
		UmbrellaID: umbrellaID,
		SubChanges: []SplitSubChange{
			{ID: secondFeature.ID},
			{ID: "CHG-2026-002-3344-feature-sub-change", DependsOn: []string{secondFeature.ID}},
		},
	})
	if err != nil {
		t.Fatalf("apply decomposition: %v", err)
	}
	if got := len(result.ChangedFiles); got != 3 {
		t.Fatalf("expected 3 changed status files, got %d (%#v)", got, result.ChangedFiles)
	}
	umbrellaStatusPath := filepath.Join(root, "runecontext", "changes", umbrellaID, "status.yaml")
	umbrellaText := strings.ReplaceAll(string(mustReadBytes(t, umbrellaStatusPath)), "\r\n", "\n")
	if !strings.Contains(umbrellaText, "related_changes:\n  - CHG-2026-002-3344-feature-sub-change\n  - "+secondFeature.ID) {
		t.Fatalf("expected umbrella related_changes rewrite, got:\n%s", umbrellaText)
	}
	featureStatusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-002-3344-feature-sub-change", "status.yaml")
	featureText := strings.ReplaceAll(string(mustReadBytes(t, featureStatusPath)), "\r\n", "\n")
	if !strings.Contains(featureText, "depends_on:\n  - "+secondFeature.ID) {
		t.Fatalf("expected feature depends_on rewrite, got:\n%s", featureText)
	}
	assertValidatedWorkflowProject(t, v, root)
}

func TestApplyChangeDecompositionRejectsNonProjectUmbrella(t *testing.T) {
	root := copyChangeWorkflowTemplate(t)
	v, feature := mustCreateDefaultFeatureChange(t, root)
	_, secondFeature := mustCreateChange(t, root, defaultFeatureChangeOptions("Second feature", []byte{0xb1, 0xb2}))
	loaded := mustReloadWorkflowProject(t, v, root)
	defer loaded.Close()
	_, err := ApplyChangeDecomposition(v, loaded, ChangeDecompositionApplyOptions{
		UmbrellaID: feature.ID,
		SubChanges: []SplitSubChange{{ID: secondFeature.ID}},
	})
	if err == nil || !strings.Contains(err.Error(), "decomposition umbrella must be type project") {
		t.Fatalf("expected non-project umbrella rejection, got %v", err)
	}
}
