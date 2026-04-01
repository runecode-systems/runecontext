package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func enforceNonInteractiveChangeNew(machine machineOptions, request changeNewRequest) error {
	if !machine.nonInteractive {
		return nil
	}
	missing := missingNonInteractiveChangeNewFields(request)
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("--non-interactive requires explicit %s to avoid inferred defaults", strings.Join(missing, ", "))
}

func missingNonInteractiveChangeNewFields(request changeNewRequest) []string {
	missing := make([]string, 0, 3)
	if !request.sizeProvided || strings.TrimSpace(request.size) == "" {
		missing = append(missing, "--size")
	}
	if !request.modeProvided || strings.TrimSpace(request.mode) == "" {
		missing = append(missing, "--shape")
	}
	if !request.bundleProvided || !hasNonInteractiveBundles(request.contextBundles) {
		missing = append(missing, "--bundle")
	}
	return missing
}

func hasNonInteractiveBundles(bundles []string) bool {
	for _, bundle := range bundles {
		if strings.TrimSpace(bundle) != "" {
			return true
		}
	}
	return false
}

func appendStatusExplainLines(lines []line, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary) []line {
	lines = append(lines,
		line{"explain_scope", "resolution,status-classification"},
		line{"explain_status_rule_active", "changes not in closed or superseded lifecycle states"},
		line{"explain_status_rule_closed", "changes in closed lifecycle state"},
		line{"explain_status_rule_superseded", "changes in superseded lifecycle state"},
	)
	if loaded == nil || loaded.Resolution == nil || summary == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_resolution_selected_config_path", loaded.Resolution.SelectedConfigPath},
		line{"explain_resolution_source_mode", string(loaded.Resolution.SourceMode)},
		line{"explain_resolution_source_ref", loaded.Resolution.SourceRef},
		line{"explain_resolution_verification_posture", string(loaded.Resolution.VerificationPosture)},
		line{"explain_status_active_count_reason", fmt.Sprintf("%d non-terminal changes discovered", len(summary.Active))},
		line{"explain_status_closed_count_reason", fmt.Sprintf("%d closed changes discovered", len(summary.Closed))},
		line{"explain_status_superseded_count_reason", fmt.Sprintf("%d superseded changes discovered", len(summary.Superseded))},
	)
	return lines
}

func appendValidateExplainLines(lines []line, request validateRequest, index *contracts.ProjectIndex, diagnostics []emittedDiagnostic) []line {
	strategy := "nearest_ancestor"
	if request.explicitRoot {
		strategy = "explicit_root"
	}
	lines = append(lines,
		line{"explain_scope", "resolution,diagnostics"},
		line{"explain_resolution_strategy", strategy},
		line{"explain_diagnostic_count_reason", fmt.Sprintf("%d diagnostics collected from resolution, project, and bundle validation", len(diagnostics))},
	)
	if index == nil || index.Resolution == nil {
		return lines
	}
	if index.Resolution.ResolvedCommit != "" {
		lines = append(lines, line{"explain_resolution_resolved_commit_reason", "source resolves to a specific commit for deterministic validation"})
	}
	return lines
}

func appendChangeNewExplainLines(lines []line, result *contracts.ChangeOperationResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "standards-selection,change-intake"},
		line{"explain_context_bundle_count_reason", fmt.Sprintf("%d context bundles selected for this change", len(result.ContextBundles))},
		line{"explain_standards_count_reason", fmt.Sprintf("%d standards selected from bundle resolution", len(result.ApplicableStandards))},
		line{"explain_recommended_mode", string(result.RecommendedMode)},
	)
	return lines
}

func appendChangeShapeExplainLines(lines []line, result *contracts.ChangeOperationResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "standards-selection"},
		line{"explain_standards_refresh_reason", result.StandardsRefreshAction},
		line{"explain_added_standard_count_reason", fmt.Sprintf("%d standards added since last refresh", len(result.AddedStandards))},
	)
	return lines
}

func appendChangeCloseExplainLines(lines []line, result *contracts.ChangeOperationResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "promotion-suggestions"},
		line{"explain_promotion_status", result.PromotionAssessmentStatus},
		line{"explain_promotion_target_count", fmt.Sprintf("%d", len(result.SuggestedPromotionTargets))},
	)
	if result.Recursive {
		lines = append(lines,
			line{"explain_scope_2", "recursive-lifecycle-cascade"},
			line{"explain_recursive_target_count", fmt.Sprintf("%d", result.RecursiveTargetCount)},
		)
	}
	for i, target := range result.SuggestedPromotionTargets {
		lines = append(lines, line{fmt.Sprintf("explain_promotion_target_%d", i+1), target})
	}
	if result.Recursive {
		for i, targetID := range result.RecursiveTargetIDs {
			lines = append(lines, line{fmt.Sprintf("explain_recursive_target_%d", i+1), targetID})
		}
	}
	return lines
}

func appendPromoteExplainLines(lines []line, result *contracts.ChangeOperationResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "promotion"},
		line{"explain_promotion_status", result.PromotionAssessmentStatus},
		line{"explain_promotion_target_count", fmt.Sprintf("%d", len(result.SuggestedPromotionTargets))},
	)
	for i, target := range result.SuggestedPromotionTargets {
		lines = append(lines, line{fmt.Sprintf("explain_promotion_target_%d", i+1), target})
	}
	return lines
}

func appendChangeReallocateExplainLines(lines []line, result *contracts.ChangeReallocationResult) []line {
	if result == nil {
		return lines
	}
	return append(lines,
		line{"explain_scope", "reference-rewrite"},
		line{"explain_rewrite_count_reason", fmt.Sprintf("%d local markdown references rewritten from old change path to new change path", result.RewrittenReferenceCount)},
	)
}

func appendChangeUpdateExplainLines(lines []line, result *contracts.ChangeOperationResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "lifecycle-transition"},
		line{"explain_lifecycle_status", result.Status},
		line{"explain_related_change_count", fmt.Sprintf("%d", len(result.RelatedChanges))},
	)
	for i, relatedID := range result.RelatedChanges {
		lines = append(lines, line{fmt.Sprintf("explain_related_change_%d", i+1), relatedID})
	}
	if result.Recursive {
		lines = append(lines,
			line{"explain_scope_2", "recursive-lifecycle-cascade"},
			line{"explain_recursive_target_count", fmt.Sprintf("%d", result.RecursiveTargetCount)},
		)
		for i, targetID := range result.RecursiveTargetIDs {
			lines = append(lines, line{fmt.Sprintf("explain_recursive_target_%d", i+1), targetID})
		}
	}
	return lines
}

func appendChangeAssessIntakeExplainLines(lines []line, result *contracts.ChangeAssessIntakeResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "change-intake-advisory"},
		line{"explain_advisory_only", "true"},
		line{"explain_recommended_mode", string(result.RecommendedMode)},
		line{"explain_intake_readiness", result.IntakeReadiness},
		line{"explain_decomposition_signal", result.DecompositionSignal},
		line{"explain_clarification_needed", boolString(result.ClarificationNeeded)},
	)
	for i, prompt := range result.ClarificationPrompts {
		lines = append(lines, line{fmt.Sprintf("explain_clarification_prompt_%d", i+1), prompt})
	}
	return lines
}

func appendChangeAssessDecompositionExplainLines(lines []line, result *contracts.ChangeAssessDecompositionResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "change-decomposition-advisory"},
		line{"explain_advisory_only", "true"},
		line{"explain_decomposition_signal", result.DecompositionSignal},
		line{"explain_recommended_mode", string(result.RecommendedMode)},
		line{"explain_clarification_needed", boolString(result.ClarificationNeeded)},
		line{"explain_eligible_sub_change_count", fmt.Sprintf("%d", len(result.EligibleSubChangeIDs))},
		line{"explain_prerequisite_change_count", fmt.Sprintf("%d", len(result.PrerequisiteChangeIDs))},
	)
	for i, prompt := range result.ClarificationPrompts {
		lines = append(lines, line{fmt.Sprintf("explain_clarification_prompt_%d", i+1), prompt})
	}
	return lines
}

func appendChangeDecompositionPlanExplainLines(lines []line, result *contracts.ChangeDecompositionPlanResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "change-decomposition-plan"},
		line{"explain_advisory_only", "true"},
		line{"explain_umbrella_change_id", result.UmbrellaID},
		line{"explain_graph_node_count", fmt.Sprintf("%d", len(result.NodeIDs))},
	)
	return lines
}

func appendChangeDecompositionApplyExplainLines(lines []line, result *contracts.ChangeDecompositionApplyResult) []line {
	if result == nil {
		return lines
	}
	lines = append(lines,
		line{"explain_scope", "change-decomposition-apply"},
		line{"explain_umbrella_change_id", result.UmbrellaID},
		line{"explain_graph_node_count", fmt.Sprintf("%d", len(result.NodeIDs))},
		line{"explain_changed_file_count", fmt.Sprintf("%d", len(result.ChangedFiles))},
	)
	return lines
}

func appendAssuranceEnableExplainLines(lines []line, root string, plans []string) []line {
	lines = append(lines,
		line{"explain_scope", "assurance-enable"},
		line{"explain_assurance_tier_target", "verified"},
		line{"explain_assurance_root", root},
		line{"explain_baseline_subject_id", "project-root"},
		line{"explain_baseline_canonicalization", "runecontext-canonical-json-v1"},
	)
	for i, action := range plans {
		lines = append(lines, line{fmt.Sprintf("explain_plan_action_%d", i+1), action})
	}
	return lines
}
