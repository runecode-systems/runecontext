package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type upgradeState string

const (
	upgradeStateCurrent                   upgradeState = "current"
	upgradeStateUpgradeable               upgradeState = "upgradeable"
	upgradeStateUnsupportedProjectVersion upgradeState = "unsupported_project_version"
	upgradeStateMixedOrStaleTree          upgradeState = "mixed_or_stale_tree"
	upgradeStateConflicted                upgradeState = "conflicted"
)

type upgradePlan struct {
	State          upgradeState
	CurrentVersion string
	TargetVersion  string
	NetworkAccess  bool
	PlanActions    []string
	NextActions    []string
	Conflicts      []string
	Warnings       []string
	ApplyMutations []string

	ConfigPath   string
	ProjectRoot  string
	SourceType   string
	AdapterPlans map[string]adapterSyncState
}

type upgradeEdgeKey struct {
	From string
	To   string
}

type upgradePlannerRegistry struct {
	edges map[upgradeEdgeKey]struct{}
}

func defaultUpgradePlannerRegistry() upgradePlannerRegistry {
	registry := upgradePlannerRegistry{edges: map[upgradeEdgeKey]struct{}{}}
	registry.registerEdge("0.1.0-alpha.8", "0.1.0-alpha.9")
	installed := normalizedRunecontextVersion()
	if installed != "" && installed != "0.0.0-dev" {
		registry.registerEdge("0.1.0-alpha.8", installed)
	}
	return registry
}

func (r *upgradePlannerRegistry) registerEdge(from, to string) {
	if r == nil {
		return
	}
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" {
		return
	}
	r.edges[upgradeEdgeKey{From: from, To: to}] = struct{}{}
}

func (r *upgradePlannerRegistry) hasEdge(from, to string) bool {
	if r == nil {
		return false
	}
	_, ok := r.edges[upgradeEdgeKey{From: strings.TrimSpace(from), To: strings.TrimSpace(to)}]
	return ok
}

func resolveUpgradeTargetVersion(current, requested string) (string, bool) {
	target := strings.TrimSpace(requested)
	if target == "" {
		return current, false
	}
	switch strings.ToLower(target) {
	case "current":
		return current, false
	case "installed", "latest":
		installed := normalizedRunecontextVersion()
		if installed == "" || installed == "0.0.0-dev" {
			return current, true
		}
		return installed, strings.EqualFold(target, "latest")
	default:
		return target, false
	}
}

func isSupportedProjectVersion(version string, registry upgradePlannerRegistry) bool {
	version = strings.TrimSpace(version)
	if version == "" {
		return false
	}
	installed := normalizedRunecontextVersion()
	if installed == "0.0.0-dev" {
		return true
	}
	if version == installed {
		return true
	}
	for edge := range registry.edges {
		if edge.From == version || edge.To == version {
			return true
		}
	}
	return false
}

func upgradePlanDiagnostics(plan upgradePlan) []emittedDiagnostic {
	switch plan.State {
	case upgradeStateUnsupportedProjectVersion:
		return []emittedDiagnostic{{
			Severity: contracts.DiagnosticSeverityError,
			Code:     "unsupported_project_version",
			Message:  fmt.Sprintf("project runecontext_version %s is not supported for upgrade to %s", plan.CurrentVersion, plan.TargetVersion),
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
		diagnostics := make([]emittedDiagnostic, 0, len(plan.Warnings))
		for _, warning := range plan.Warnings {
			diagnostics = append(diagnostics, emittedDiagnostic{
				Severity: contracts.DiagnosticSeverityWarning,
				Code:     "optional_adapter_pack_unavailable",
				Message:  warning,
			})
		}
		return diagnostics
	}
}
