package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type upgradeState string

const (
	upgradeStateCurrent                   upgradeState = "current"
	upgradeStateUpgradeable               upgradeState = "upgradeable"
	upgradeStateUnsupportedProjectVersion upgradeState = "unsupported_project_version"
	upgradeStateProjectNewerThanCLI       upgradeState = "project_newer_than_cli"
	upgradeStateMixedOrStaleTree          upgradeState = "mixed_or_stale_tree"
	upgradeStateConflicted                upgradeState = "conflicted"
)

type upgradePlan struct {
	State          upgradeState
	CurrentVersion string
	TargetVersion  string
	UpgradeHops    []upgradeHop
	HopActions     []string
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

type upgradeHop struct {
	From string
	To   string
}

type upgradePlannerRegistry struct {
	edges map[upgradeEdgeKey]struct{}
	next  map[string][]string
}

func defaultUpgradePlannerRegistry() upgradePlannerRegistry {
	registry := upgradePlannerRegistry{edges: map[upgradeEdgeKey]struct{}{}, next: map[string][]string{}}
	registry.registerEdge("0.1.0-alpha.8", "0.1.0-alpha.9")
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
	key := upgradeEdgeKey{From: from, To: to}
	if _, exists := r.edges[key]; exists {
		return
	}
	r.edges[key] = struct{}{}
	r.next[from] = append(r.next[from], to)
	sort.Strings(r.next[from])
}

func (r *upgradePlannerRegistry) planPath(from, to string) ([]upgradeHop, bool) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if r == nil || from == "" || to == "" {
		return nil, false
	}
	if from == to {
		return nil, true
	}
	parents, found := r.searchPathParents(from, to)
	if !found {
		return nil, false
	}
	return buildUpgradeHops(parents, from, to), true
}

func (r *upgradePlannerRegistry) searchPathParents(from, to string) (map[string]string, bool) {

	queue := []string{from}
	seen := map[string]struct{}{from: {}}
	parent := map[string]string{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range r.next[current] {
			if _, visited := seen[next]; visited {
				continue
			}
			seen[next] = struct{}{}
			parent[next] = current
			if next == to {
				return parent, true
			}
			queue = append(queue, next)
		}
	}

	return nil, false
}

func buildUpgradeHops(parent map[string]string, from, to string) []upgradeHop {
	versions := []string{to}
	current := to
	for current != from {
		current = parent[current]
		versions = append(versions, current)
	}
	reverseStrings(versions)
	hops := make([]upgradeHop, 0, len(versions)-1)
	for i := 0; i+1 < len(versions); i++ {
		hops = append(hops, upgradeHop{From: versions[i], To: versions[i+1]})
	}
	return hops
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func buildUpgradeHopActions(hops []upgradeHop) []string {
	actions := make([]string, 0, len(hops))
	for _, hop := range hops {
		actions = append(actions, fmt.Sprintf("migrate runecontext_version %s -> %s", hop.From, hop.To))
	}
	return actions
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
	if isCompatibleProjectVersionForInstalled(version, installed) {
		return true
	}
	for edge := range registry.edges {
		if edge.From == version || edge.To == version {
			return true
		}
	}
	return false
}

func isCompatibleProjectVersionForInstalled(projectVersion, installedVersion string) bool {
	if !strings.HasPrefix(installedVersion, "0.1.0-alpha.") {
		return false
	}
	projectOrdinal, ok := alphaOrdinal(projectVersion)
	if !ok {
		return false
	}
	installedOrdinal, ok := alphaOrdinal(installedVersion)
	if !ok {
		return false
	}
	return projectOrdinal >= 5 && projectOrdinal <= 8 && installedOrdinal >= projectOrdinal
}

func alphaOrdinal(version string) (int, bool) {
	const prefix = "0.1.0-alpha."
	if !strings.HasPrefix(version, prefix) {
		return 0, false
	}
	value := strings.TrimPrefix(version, prefix)
	if value == "" {
		return 0, false
	}
	ordinal, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return ordinal, true
}

func upgradePlanDiagnostics(plan upgradePlan) []emittedDiagnostic {
	return buildUpgradePlanDiagnostics(plan)
}
