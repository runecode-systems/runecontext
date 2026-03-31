package cli

import (
	"fmt"
	"sort"
	"strings"
)

func normalizeMigrationIntervalBounds(current, target string) (string, string, bool) {
	current = strings.TrimSpace(current)
	target = strings.TrimSpace(target)
	if current == "" || target == "" {
		return "", "", false
	}
	return current, target, true
}

func isForwardMigrationInterval(current, target string) bool {
	comparison, comparable := compareKnownRunecontextVersions(current, target)
	if !comparable || comparison >= 0 {
		return false
	}
	return true
}

func selectMigrationEdgesWithinInterval(edges map[upgradeEdgeKey]struct{}, current, target string) ([]upgradeHop, error) {
	selected := make([]upgradeHop, 0, len(edges))
	for edge := range edges {
		if !isValidForwardMigrationEdge(edge) {
			return nil, fmt.Errorf("invalid migration edge %s -> %s", edge.From, edge.To)
		}
		if !isMigrationEdgeWithinClosedInterval(edge, current, target) {
			continue
		}
		selected = append(selected, upgradeHop{From: edge.From, To: edge.To})
	}
	sortUpgradeHops(selected)
	return selected, nil
}

func isMigrationEdgeWithinClosedInterval(edge upgradeEdgeKey, current, target string) bool {
	return versionInClosedInterval(edge.From, current, target) && versionInClosedInterval(edge.To, current, target)
}

func isValidForwardMigrationEdge(edge upgradeEdgeKey) bool {
	comparison, comparable := compareKnownRunecontextVersions(edge.From, edge.To)
	return comparable && comparison < 0
}

func sortUpgradeHops(hops []upgradeHop) {
	sort.Slice(hops, func(i, j int) bool {
		if hops[i].From == hops[j].From {
			return lessUpgradeVersion(hops[i].To, hops[j].To)
		}
		return lessUpgradeVersion(hops[i].From, hops[j].From)
	})
}

func ensureMigrationEdgesNonOverlapping(hops []upgradeHop, current, target string) error {
	cursor := current
	for _, hop := range hops {
		if lessUpgradeVersion(hop.From, cursor) {
			return fmt.Errorf("ambiguous migration edges overlap within upgrade interval %s -> %s", current, target)
		}
		cursor = hop.To
	}
	return nil
}

func versionInClosedInterval(version, lower, upper string) bool {
	if lessUpgradeVersion(version, lower) {
		return false
	}
	if lessUpgradeVersion(upper, version) {
		return false
	}
	return true
}

func lessUpgradeVersion(left, right string) bool {
	comparison, comparable := compareKnownRunecontextVersions(left, right)
	if comparable {
		return comparison < 0
	}
	return left < right
}
