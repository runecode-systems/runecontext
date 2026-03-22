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
	for i, target := range result.SuggestedPromotionTargets {
		lines = append(lines, line{fmt.Sprintf("explain_promotion_target_%d", i+1), target})
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
