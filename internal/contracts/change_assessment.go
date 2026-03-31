package contracts

import (
	"fmt"
	"strings"
)

func AssessChangeIntake(v *Validator, loaded *LoadedProject, options ChangeAssessIntakeOptions) (*ChangeAssessIntakeResult, error) {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return nil, err
	}
	changeType := strings.TrimSpace(options.Type)
	if err := validateChangeTypeValue(changeType); err != nil {
		return nil, err
	}
	if strings.TrimSpace(options.Title) == "" {
		return nil, fmt.Errorf("change intake assessment requires title")
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()

	assessment := assessChangeIntake(options.Title, changeType, options.Size, options.Description)
	contextBundles, contextAssumptions, err := resolveContextBundlesForChange(index, options.ContextBundles)
	if err != nil {
		return nil, err
	}
	standards, standardAssumptions, err := resolveApplicableStandards(index, contextBundles)
	if err != nil {
		return nil, err
	}
	assumptions := uniqueStringsInOrder(append([]string{}, assessment.Assumptions...))
	assumptions = uniqueStringsInOrder(append(assumptions, contextAssumptions...))
	assumptions = uniqueStringsInOrder(append(assumptions, standardAssumptions...))
	if note := verificationAssumption(loaded.Resolution.ProjectRoot); note != "" {
		assumptions = append(assumptions, note)
	}

	return &ChangeAssessIntakeResult{
		Type:                 changeType,
		Size:                 assessment.Size,
		RecommendedMode:      assessment.RecommendedMode,
		IntakeReadiness:      intakeReadinessFromAssessment(assessment),
		ClarificationNeeded:  len(assessment.FollowUpPrompts) > 0,
		ClarificationPrompts: append([]string(nil), assessment.FollowUpPrompts...),
		DecompositionSignal:  decompositionSignalForIntake(changeType, assessment.Size, options.Description),
		ContextBundles:       append([]string(nil), contextBundles...),
		ApplicableStandards:  append([]string(nil), standards...),
		Reasons:              append([]string(nil), assessment.Reasons...),
		Assumptions:          assumptions,
	}, nil
}

func AssessChangeDecomposition(v *Validator, loaded *LoadedProject, changeID string) (*ChangeAssessDecompositionResult, error) {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return nil, err
	}
	trimmedID := strings.TrimSpace(changeID)
	if trimmedID == "" {
		return nil, fmt.Errorf("change decomposition assessment requires a change ID")
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()

	record := index.Changes[trimmedID]
	if record == nil {
		return nil, fmt.Errorf("change %q does not exist", trimmedID)
	}
	related := uniqueSortedStrings(record.RelatedChanges)
	eligible, prerequisites := decompositionGraphSignals(index, record)
	reasons := decompositionReasons(record, eligible, prerequisites)
	prompts := decompositionClarificationPrompts(record)

	return &ChangeAssessDecompositionResult{
		ID:                    record.ID,
		Status:                string(record.Status),
		Type:                  record.Type,
		Size:                  record.Size,
		RecommendedMode:       recommendedModeForExisting(record),
		DecompositionSignal:   decompositionSignalForRecord(record, eligible, prerequisites),
		ClarificationNeeded:   len(prompts) > 0,
		ClarificationPrompts:  prompts,
		RelatedChanges:        related,
		EligibleSubChangeIDs:  eligible,
		PrerequisiteChangeIDs: prerequisites,
		Reasons:               reasons,
	}, nil
}

func intakeReadinessFromAssessment(assessment changeIntakeAssessment) string {
	if assessment.RecommendedMode == ChangeModeFull && len(assessment.FollowUpPrompts) > 0 {
		return "needs_clarification"
	}
	if assessment.RecommendedMode == ChangeModeFull {
		return "ready_with_full_mode"
	}
	return "ready"
}

func decompositionSignalForIntake(changeType, size, description string) string {
	if shouldFlagIntakeDecomposition(changeType, size, description) {
		return "consider_decomposition"
	}
	return "none"
}

func shouldFlagIntakeDecomposition(changeType, size, description string) bool {
	if changeType == "project" {
		return true
	}
	if size == "large" {
		return true
	}
	return containsHeuristicKeyword(description, []string{"umbrella", "decompose", "split", "phased", "multi-step", "multiple teams"})
}

func decompositionGraphSignals(index *ProjectIndex, record *ChangeRecord) ([]string, []string) {
	if index == nil || record == nil {
		return nil, nil
	}
	eligible := make([]string, 0)
	for _, relatedID := range uniqueSortedStrings(record.RelatedChanges) {
		related := index.Changes[relatedID]
		if related == nil {
			continue
		}
		if related.Type == "feature" && containsString(related.RelatedChanges, record.ID) {
			eligible = append(eligible, related.ID)
		}
	}
	return uniqueSortedStrings(eligible), uniqueSortedStrings(record.DependsOn)
}

func recommendedModeForExisting(record *ChangeRecord) ChangeMode {
	if record == nil {
		return ChangeModeMinimum
	}
	if strings.TrimSpace(record.Type) == "project" || strings.TrimSpace(record.Size) == "large" {
		return ChangeModeFull
	}
	return ChangeModeMinimum
}

func decompositionSignalForRecord(record *ChangeRecord, eligible, prerequisites []string) string {
	if record == nil {
		return "none"
	}
	if len(eligible) > 0 {
		return "umbrella_graph_detected"
	}
	if strings.TrimSpace(record.Type) == "project" || strings.TrimSpace(record.Size) == "large" {
		return "consider_decomposition"
	}
	if len(prerequisites) > 0 {
		return "prerequisites_present"
	}
	return "none"
}

func decompositionReasons(record *ChangeRecord, eligible, prerequisites []string) []string {
	reasons := make([]string, 0, 3)
	if record == nil {
		return reasons
	}
	if strings.TrimSpace(record.Type) == "project" {
		reasons = append(reasons, "Project changes usually benefit from explicit decomposition planning.")
	}
	if strings.TrimSpace(record.Size) == "large" {
		reasons = append(reasons, "Large changes are better tracked as linked sub-change plans.")
	}
	if len(eligible) > 0 {
		reasons = append(reasons, "Reciprocal related feature changes indicate an umbrella/sub-change graph.")
	}
	if len(prerequisites) > 0 {
		reasons = append(reasons, "depends_on links indicate prerequisite ordering that decomposition planning should preserve.")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "No decomposition pressure detected from current change metadata.")
	}
	return reasons
}

func decompositionClarificationPrompts(record *ChangeRecord) []string {
	if record == nil {
		return nil
	}
	prompts := make([]string, 0, 3)
	if strings.TrimSpace(record.Type) == "project" {
		prompts = append(prompts, "Define umbrella scope boundaries and expected feature sub-changes.")
	}
	if strings.TrimSpace(record.Size) == "large" {
		prompts = append(prompts, "Identify milestones that should become separate linked changes.")
	}
	if len(record.DependsOn) > 0 {
		prompts = append(prompts, "Confirm dependency ordering for related change execution.")
	}
	return prompts
}
