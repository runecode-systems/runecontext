package cli

import (
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func appendStatusNode(builder *strings.Builder, node statusTreeNode, section statusSectionTree, options statusRenderOptions, prefix string, isLast bool) {
	entry, ok := section.entryByID[node.entryID]
	if !ok {
		return
	}
	appendStatusNodeRow(builder, entry, options, prefix, node.hasParent, isLast)
	appendStatusNodeHints(builder, entry, section.fallback, options, prefix, node.hasParent, isLast)
	nextPrefix, ok := statusChildPrefix(node, prefix, isLast)
	if !ok {
		return
	}
	for i, child := range node.children {
		appendStatusNode(builder, child, section, options, nextPrefix, i == len(node.children)-1)
	}
}

func appendStatusNodeRow(builder *strings.Builder, entry contracts.ChangeStatusEntry, options statusRenderOptions, prefix string, hasParent, isLast bool) {
	linePrefix := "- "
	if hasParent {
		linePrefix = prefix + "|- "
		if isLast {
			linePrefix = prefix + "\\- "
		}
	}
	appendStatusEntryRow(builder, entry, options, linePrefix, statusDetailPrefix(prefix, hasParent, isLast))
}

func appendStatusNodeHints(builder *strings.Builder, entry contracts.ChangeStatusEntry, fallback bool, options statusRenderOptions, prefix string, hasParent, isLast bool) {
	hintLines := statusHintLines(entry, fallback, options)
	if options.verbose {
		hintLines = statusRelationshipLines(entry, options)
	}
	appendStatusHintLines(builder, hintLines, prefix, hasParent, isLast)
}

func statusChildPrefix(node statusTreeNode, prefix string, isLast bool) (string, bool) {
	if len(node.children) == 0 {
		return "", false
	}
	if !node.hasParent {
		return "  ", true
	}
	if isLast {
		return prefix + "   ", true
	}
	return prefix + "|  ", true
}

func appendStatusHintLines(builder *strings.Builder, hintLines []string, prefix string, hasParent, isLast bool) {
	if len(hintLines) == 0 {
		return
	}
	hintPrefix := statusDetailPrefix(prefix, hasParent, isLast)
	available := statusTargetRowWidth - displayTextWidth(hintPrefix)
	if available < statusMinTitleWrapWidth {
		available = statusMinTitleWrapWidth
	}
	for _, hint := range hintLines {
		for _, line := range wrapStatusHintText(hint, available) {
			builder.WriteString(hintPrefix + line + "\n")
		}
	}
}

func statusDetailPrefix(prefix string, hasParent, isLast bool) string {
	if !hasParent {
		return "  | "
	}
	if isLast {
		return prefix + "   "
	}
	return prefix + "|  "
}

func buildStatusSectionTree(entries []contracts.ChangeStatusEntry) statusSectionTree {
	entryByID, rank := indexStatusEntries(entries)
	parents, childrenByParent, fallback := buildStatusRelationships(entries, entryByID)
	if fallback {
		return fallbackStatusSectionTree(entries, rank, entryByID)
	}
	roots := buildStatusRoots(entries, childrenByParent, parents, entryByID, rank)
	if statusTreeNeedsFallback(roots, entries, entryByID) {
		return fallbackStatusSectionTree(entries, rank, entryByID)
	}
	return statusSectionTree{roots: roots, rank: rank, entryByID: entryByID}
}

func indexStatusEntries(entries []contracts.ChangeStatusEntry) (map[string]contracts.ChangeStatusEntry, map[string]int) {
	entryByID := make(map[string]contracts.ChangeStatusEntry, len(entries))
	rank := make(map[string]int, len(entries))
	for i, entry := range entries {
		entryByID[entry.ID] = entry
		rank[entry.ID] = i
	}
	return entryByID, rank
}

func buildStatusRelationships(entries []contracts.ChangeStatusEntry, entryByID map[string]contracts.ChangeStatusEntry) (map[string]string, map[string][]string, bool) {
	parents := make(map[string]string, len(entries))
	childrenByParent := make(map[string][]string, len(entries))
	for _, entry := range entries {
		candidates := projectParentCandidates(entry, entryByID)
		if len(candidates) > 1 {
			return nil, nil, true
		}
		if len(candidates) == 1 {
			parentID := candidates[0]
			parents[entry.ID] = parentID
			childrenByParent[parentID] = append(childrenByParent[parentID], entry.ID)
		}
	}
	return parents, childrenByParent, false
}

func buildStatusRoots(entries []contracts.ChangeStatusEntry, childrenByParent map[string][]string, parents map[string]string, entryByID map[string]contracts.ChangeStatusEntry, rank map[string]int) []statusTreeNode {
	roots := make([]statusTreeNode, 0, len(entries))
	for _, entry := range entries {
		if _, hasParent := parents[entry.ID]; hasParent {
			continue
		}
		roots = append(roots, buildStatusNodeTree(entry.ID, childrenByParent, parents, entryByID, rank, map[string]struct{}{}))
	}
	return roots
}

func statusTreeNeedsFallback(roots []statusTreeNode, entries []contracts.ChangeStatusEntry, entryByID map[string]contracts.ChangeStatusEntry) bool {
	if len(entries) > 0 && len(roots) == 0 {
		return true
	}
	seen := make(map[string]struct{}, len(entries))
	for _, root := range roots {
		collectStatusTreeNodeIDs(root, seen)
	}
	return len(seen) != len(entryByID)
}

func fallbackStatusSectionTree(entries []contracts.ChangeStatusEntry, rank map[string]int, entryByID map[string]contracts.ChangeStatusEntry) statusSectionTree {
	roots := make([]statusTreeNode, 0, len(entries))
	for _, entry := range entries {
		roots = append(roots, statusTreeNode{entryID: entry.ID})
	}
	return statusSectionTree{roots: roots, fallback: true, rank: rank, entryByID: entryByID}
}

func collectStatusTreeNodeIDs(node statusTreeNode, seen map[string]struct{}) {
	if seen == nil {
		return
	}
	if _, ok := seen[node.entryID]; ok {
		return
	}
	seen[node.entryID] = struct{}{}
	for _, child := range node.children {
		collectStatusTreeNodeIDs(child, seen)
	}
}

func buildStatusNodeTree(id string, childrenByParent map[string][]string, parents map[string]string, entryByID map[string]contracts.ChangeStatusEntry, rank map[string]int, visited map[string]struct{}) statusTreeNode {
	if _, seen := visited[id]; seen {
		return statusTreeNode{entryID: id, hasParent: hasStatusParent(id, parents)}
	}
	visited[id] = struct{}{}
	childIDs := append([]string(nil), childrenByParent[id]...)
	sortStatusChildren(childIDs, entryByID, rank)
	node := statusTreeNode{entryID: id, hasParent: hasStatusParent(id, parents)}
	for _, childID := range childIDs {
		node.children = append(node.children, buildStatusNodeTree(childID, childrenByParent, parents, entryByID, rank, visited))
	}
	delete(visited, id)
	return node
}

func hasStatusParent(id string, parents map[string]string) bool {
	_, ok := parents[id]
	return ok
}

func projectParentCandidates(entry contracts.ChangeStatusEntry, entryByID map[string]contracts.ChangeStatusEntry) []string {
	candidates := make([]string, 0, 2)
	for _, relatedID := range entry.RelatedChanges {
		parent, ok := entryByID[relatedID]
		if !ok || parent.ID == entry.ID {
			continue
		}
		if parent.Type != "project" {
			continue
		}
		if !containsStatusID(parent.RelatedChanges, entry.ID) {
			continue
		}
		candidates = append(candidates, parent.ID)
	}
	return uniqueStatusIDs(candidates)
}

func sortStatusChildren(childIDs []string, entryByID map[string]contracts.ChangeStatusEntry, rank map[string]int) {
	sort.SliceStable(childIDs, func(i, j int) bool {
		left := entryByID[childIDs[i]]
		right := entryByID[childIDs[j]]
		leftDependsRight := containsStatusID(left.DependsOn, right.ID)
		rightDependsLeft := containsStatusID(right.DependsOn, left.ID)
		if leftDependsRight != rightDependsLeft {
			return rightDependsLeft
		}
		leftRank, leftOK := rank[left.ID]
		rightRank, rightOK := rank[right.ID]
		if leftOK && rightOK && leftRank != rightRank {
			return leftRank < rightRank
		}
		return left.ID < right.ID
	})
}

func containsStatusID(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func uniqueStatusIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}
