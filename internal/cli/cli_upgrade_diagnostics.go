package cli

import (
	"fmt"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func buildUpgradePlanDiagnostics(plan upgradePlan) []emittedDiagnostic {
	switch plan.State {
	case upgradeStateUpgradeable:
		return upgradeableDiagnostics(plan)
	case upgradeStateUnsupportedProjectVersion:
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityError,
			Code:     "unsupported_project_version",
			Message:  fmt.Sprintf("project runecontext_version %s is not supported for upgrade to %s", plan.CurrentVersion, plan.TargetVersion),
		}}
	case upgradeStateProjectNewerThanCLI:
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityError,
			Code:     "project_newer_than_cli",
			Message:  fmt.Sprintf("project runecontext_version %s is newer than installed runectx %s; run runectx upgrade cli apply to upgrade the runectx binary", plan.CurrentVersion, normalizedRunecontextVersion()),
		}}
	case upgradeStateMixedOrStaleTree:
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityWarning,
			Code:     "mixed_or_stale_tree",
			Message:  "stale RuneContext-managed artifacts detected; rerun runectx upgrade apply",
		}}
	case upgradeStateConflicted:
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityError,
			Code:     "conflicted",
			Message:  "managed artifact conflicts detected; resolve ownership conflicts before upgrade apply",
		}}
	default:
		return warningDiagnostics(plan.Warnings)
	}
}

func upgradeableDiagnostics(plan upgradePlan) []emittedDiagnostic {
	if len(plan.UpgradeHops) == 0 {
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityWarning,
			Code:     "upgrade_available",
			Message:  fmt.Sprintf("project runecontext_version %s is compatible but older than target %s; run runectx upgrade apply to bump the pinned version", plan.CurrentVersion, plan.TargetVersion),
		}}
	}
	return []emittedDiagnostic{{
		Severity: contracts.DiagnosticSeverityWarning,
		Code:     "upgrade_available",
		Message:  fmt.Sprintf("project runecontext_version %s can upgrade to %s with %d required migration hop(s); review runectx upgrade and then run runectx upgrade apply", plan.CurrentVersion, plan.TargetVersion, len(plan.UpgradeHops)),
	}}
}

func warningDiagnostics(warnings []string) []emittedDiagnostic {
	diagnostics := make([]emittedDiagnostic, 0, len(warnings))
	for _, warning := range warnings {
		diagnostics = append(diagnostics, emittedDiagnostic{
			Severity: contracts.DiagnosticSeverityWarning,
			Code:     "optional_adapter_pack_unavailable",
			Message:  warning,
		})
	}
	return diagnostics
}
