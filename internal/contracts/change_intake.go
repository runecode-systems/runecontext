package contracts

import (
	"fmt"
	"strings"
)

func assessChangeIntake(title, changeType, requestedSize, description string) changeIntakeAssessment {
	assessment := newChangeIntakeAssessment(changeType, requestedSize)
	signal := title + " " + description
	risky := intakeHasHeuristicRisk(title, description)
	ambiguous := intakeHasHeuristicAmbiguity(title, description)
	applyChangeTypeIntakeRules(&assessment, changeType, signal, risky, ambiguous)
	ensureDefaultIntakeReason(&assessment)
	return assessment
}

func newChangeIntakeAssessment(changeType, requestedSize string) changeIntakeAssessment {
	size, assumptions := determineChangeSize(changeType, requestedSize)
	return changeIntakeAssessment{
		Size:             size,
		RecommendedMode:  ChangeModeMinimum,
		Assumptions:      assumptions,
		VerificationNote: "Use the repository's standard verification flow before closing this change.",
	}
}

func determineChangeSize(changeType, requestedSize string) (string, []string) {
	size := strings.TrimSpace(requestedSize)
	if size != "" {
		return size, nil
	}
	inferred := defaultChangeSize(changeType)
	assumption := fmt.Sprintf("Inferred size %q from the change type because no explicit size was provided.", inferred)
	return inferred, []string{assumption}
}

func applyChangeTypeIntakeRules(assessment *changeIntakeAssessment, changeType, signal string, risky, ambiguous bool) {
	if assessment == nil {
		return
	}
	if changeType == "project" {
		applyProjectIntakeRules(assessment)
		return
	}
	if changeType == "bug" {
		applyBugIntakeRules(assessment, risky, ambiguous)
		return
	}
	if changeType == "feature" {
		promoteFeatureIntake(assessment, risky, ambiguous)
		return
	}
	applyOperationalIntakeRules(assessment, changeType, signal, risky, ambiguous)
}

func applyProjectIntakeRules(assessment *changeIntakeAssessment) {
	assessment.RecommendedMode = ChangeModeFull
	assessment.Reasons = append(assessment.Reasons, "Project work uses deeper intake because bad defaults compound.")
	assessment.ChecklistTitle = "Project Intake Checklist"
	assessment.ChecklistItems = []string{
		"Mission and target users.",
		"Stack and runtime constraints.",
		"Deployment and security constraints.",
		"Success criteria.",
		"Non-goals.",
	}
	assessment.FollowUpPrompts = append(assessment.FollowUpPrompts, assessment.ChecklistItems...)
}

func applyBugIntakeRules(assessment *changeIntakeAssessment, risky, ambiguous bool) {
	if !shouldPromoteBugIntake(assessment.Size, risky, ambiguous) {
		return
	}
	assessment.RecommendedMode = ChangeModeFull
	assessment.Reasons = append(assessment.Reasons, "Bugs with unclear root cause, ambiguity, or security/schema/API impact should be shaped in full mode.")
	assessment.ChecklistTitle = "Bug Escalation Checklist"
	assessment.ChecklistItems = []string{
		"Clarify the current behavior and the expected behavior.",
		"Confirm whether security, schema, or API surfaces are affected.",
		"Record the root-cause hypothesis and any open uncertainties.",
	}
	assessment.FollowUpPrompts = append(assessment.FollowUpPrompts,
		"User-facing behavior that materially changes the fix.",
		"API or interface changes introduced by the fix.",
		"Verification and acceptance criteria for the repaired behavior.",
	)
}

func shouldPromoteBugIntake(size string, risky, ambiguous bool) bool {
	return size == "large" || risky || ambiguous
}

func promoteFeatureIntake(assessment *changeIntakeAssessment, risky, ambiguous bool) {
	if assessment.Size != "large" && !risky && !ambiguous {
		return
	}
	assessment.RecommendedMode = ChangeModeFull
	assessment.Reasons = append(assessment.Reasons, "Large, ambiguous, or high-risk feature work should move to full mode early.")
}

func applyOperationalIntakeRules(assessment *changeIntakeAssessment, changeType, signal string, risky, ambiguous bool) {
	if shouldPromoteStandardOrChore(changeType, assessment.Size, signal) {
		assessment.RecommendedMode = ChangeModeFull
		assessment.Reasons = append(assessment.Reasons, "Broad standards or chore changes should be shaped so future impact stays reviewable.")
		return
	}
	if assessment.Size == "large" || risky || ambiguous {
		assessment.RecommendedMode = ChangeModeFull
		assessment.Reasons = append(assessment.Reasons, "The requested change looks large, ambiguous, or risky enough to justify full mode.")
	}
}

func shouldPromoteStandardOrChore(changeType, size, signal string) bool {
	if changeType != "standard" && changeType != "chore" {
		return false
	}
	if size == "large" {
		return true
	}
	return containsHeuristicKeyword(signal, []string{"deprecate", "rename", "migration", "rollout"})
}

func ensureDefaultIntakeReason(assessment *changeIntakeAssessment) {
	if len(assessment.Reasons) > 0 {
		return
	}
	assessment.Reasons = append(assessment.Reasons, "Minimum mode is sufficient for the current size and risk signal.")
}

func defaultChangeSize(changeType string) string {
	switch changeType {
	case "project":
		return "large"
	case "feature":
		return "medium"
	case "bug", "standard", "chore":
		return "small"
	default:
		return "medium"
	}
}

func resolveContextBundlesForChange(index *ProjectIndex, requested []string) ([]string, []string, error) {
	bundles := uniqueSortedStrings(requested)
	if index == nil || index.Bundles == nil || len(index.Bundles.bundles) == 0 {
		return bundles, nil, nil
	}
	if err := ensureKnownContextBundles(index, bundles); err != nil {
		return nil, nil, err
	}
	if len(bundles) > 0 {
		return bundles, nil, nil
	}
	return inferContextBundles(index)
}

func ensureKnownContextBundles(index *ProjectIndex, bundles []string) error {
	for _, id := range bundles {
		if _, ok := index.Bundles.bundles[id]; !ok {
			return fmt.Errorf("context bundle %q does not exist", id)
		}
	}
	return nil
}

func inferContextBundles(index *ProjectIndex) ([]string, []string, error) {
	if len(index.Bundles.bundles) == 1 {
		inferred := SortedKeys(index.Bundles.bundles)
		assumption := fmt.Sprintf("Inferred context bundle %q because it is the only bundle in the project.", inferred[0])
		return inferred, []string{assumption}, nil
	}
	for _, candidate := range []string{"default", "base", "core"} {
		if _, ok := index.Bundles.bundles[candidate]; ok {
			assumption := fmt.Sprintf("Inferred context bundle %q from the repository defaults.", candidate)
			return []string{candidate}, []string{assumption}, nil
		}
	}
	return nil, []string{"No context bundle was selected automatically; standards fall back to all non-draft standards."}, nil
}

func resolveApplicableStandards(index *ProjectIndex, contextBundles []string) ([]string, []string, error) {
	if index == nil {
		return nil, nil, fmt.Errorf("project index is required")
	}
	selected, err := selectedStandardsFromBundles(index, contextBundles)
	if err != nil {
		return nil, nil, err
	}
	if len(selected) > 0 {
		return selected, nil, nil
	}
	return fallbackApplicableStandards(index)
}

func selectedStandardsFromBundles(index *ProjectIndex, contextBundles []string) ([]string, error) {
	selected := make([]string, 0)
	for _, bundleID := range uniqueSortedStrings(contextBundles) {
		resolution, err := index.ResolveBundle(bundleID)
		if err != nil {
			return nil, err
		}
		selected = append(selected, standardsSelectedFromBundleResolution(resolution)...)
	}
	return uniqueSortedStrings(selected), nil
}

func standardsSelectedFromBundleResolution(resolution *BundleResolution) []string {
	if resolution == nil {
		return nil
	}
	aspect, ok := resolution.Aspects[BundleAspectStandards]
	if !ok {
		return nil
	}
	selected := make([]string, 0, len(aspect.Selected))
	for _, entry := range aspect.Selected {
		selected = append(selected, entry.Path)
	}
	return selected
}

func fallbackApplicableStandards(index *ProjectIndex) ([]string, []string, error) {
	fallback := selectableNonDraftStandards(index)
	assumption := "Used all non-draft standards as a conservative fallback because no standards were selected through context bundles."
	if len(fallback) == 0 {
		assumption = "No selectable standards are defined in the project yet; the Applicable Standards section is rendered as N/A."
	}
	return fallback, []string{assumption}, nil
}

func selectableNonDraftStandards(index *ProjectIndex) []string {
	fallback := make([]string, 0)
	for _, path := range SortedKeys(index.Standards) {
		if index.Standards[path].Status == StandardStatusDraft {
			continue
		}
		fallback = append(fallback, path)
	}
	return fallback
}

func intakeHasHeuristicRisk(title, description string) bool {
	return containsHeuristicKeyword(title+" "+description, []string{"security", "schema", "api", "migration", "rollout", "deploy", "auth", "permission", "secret"})
}

func intakeHasHeuristicAmbiguity(title, description string) bool {
	return containsHeuristicKeyword(title+" "+description, []string{"unclear", "unknown", "investigate", "spike", "explore", "ambiguous"})
}

func containsHeuristicKeyword(value string, keywords []string) bool {
	value = strings.ToLower(value)
	for _, keyword := range keywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}
