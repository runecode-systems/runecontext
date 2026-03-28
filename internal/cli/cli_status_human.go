package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	defaultStatusHistoryLimit = 5
	statusHistoryModeRecent   = "recent"
	statusHistoryModeAll      = "all"
	statusHistoryModeNone     = "none"
	statusTargetRowWidth      = 96
	statusMinTitleWrapWidth   = 24

	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiRed    = "\x1b[31m"
)

type statusRenderOptions struct {
	color        bool
	explain      bool
	historyMode  string
	historyLimit int
	verbose      bool
}

type statusSectionTree struct {
	roots     []statusTreeNode
	fallback  bool
	rank      map[string]int
	entryByID map[string]contracts.ChangeStatusEntry
}

type statusTreeNode struct {
	entryID   string
	children  []statusTreeNode
	hasParent bool
}

func renderHumanStatus(absRoot string, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary, options statusRenderOptions) string {
	if summary == nil {
		return ""
	}
	var builder strings.Builder
	options = normalizeStatusRenderOptions(options)
	closedEntries, closedHidden := selectHistoryEntries(summary.Closed, options)
	supersededEntries, supersededHidden := selectHistoryEntries(summary.Superseded, options)
	appendStatusHeader(&builder, absRoot, summary, options)
	appendStatusSection(&builder, "In Flight", summary.Active, options)
	builder.WriteString("\n")
	appendStatusSection(&builder, "Recently Completed", closedEntries, options)
	appendStatusHistoryHint(&builder, "closed", len(summary.Closed), len(closedEntries), closedHidden, options)
	appendStatusSection(&builder, "Replaced", supersededEntries, options)
	appendStatusHistoryHint(&builder, "superseded", len(summary.Superseded), len(supersededEntries), supersededHidden, options)
	if options.explain {
		appendStatusExplainHuman(&builder, loaded, summary, options)
	}
	return builder.String()
}

func appendStatusHeader(builder *strings.Builder, absRoot string, summary *contracts.ProjectStatusSummary, options statusRenderOptions) {
	builder.WriteString(styleStatusText("RuneContext Status", ansiBold+ansiBlue, options.color))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Root: %s\n", absRoot))
	builder.WriteString(fmt.Sprintf("Config: %s\n", summary.SelectedConfigPath))
	builder.WriteString(fmt.Sprintf("Version: %s  Assurance: %s\n", summary.RuneContextVersion, summary.AssuranceTier))
	builder.WriteString(renderBundleSummary(summary.BundleIDs))
	builder.WriteString("\n\n")
}

func renderBundleSummary(bundleIDs []string) string {
	if len(bundleIDs) == 0 {
		return "Bundles (0): none"
	}
	return fmt.Sprintf("Bundles (%d): %s", len(bundleIDs), strings.Join(bundleIDs, ", "))
}

func appendStatusSection(builder *strings.Builder, title string, entries []contracts.ChangeStatusEntry, options statusRenderOptions) {
	builder.WriteString(styleStatusText(title, ansiBold, options.color))
	builder.WriteString(fmt.Sprintf(" (%d)\n", len(entries)))
	if len(entries) == 0 {
		builder.WriteString("  (none)\n\n")
		return
	}
	section := buildStatusSectionTree(entries)
	for i, root := range section.roots {
		isLastRoot := i == len(section.roots)-1
		appendStatusNode(builder, root, section, options, "", isLastRoot)
	}
	builder.WriteString("\n")
}

func appendStatusHistoryHint(builder *strings.Builder, label string, total, shown int, hidden bool, options statusRenderOptions) {
	if !hidden {
		return
	}
	builder.WriteString(fmt.Sprintf("  showing %d of %d %s changes; use --history all to show more\n\n", shown, total, label))
}

func appendStatusNode(builder *strings.Builder, node statusTreeNode, section statusSectionTree, options statusRenderOptions, prefix string, isLast bool) {
	entry, ok := section.entryByID[node.entryID]
	if !ok {
		return
	}
	if !node.hasParent {
		appendStatusEntryRow(builder, entry, options, "- ", statusDetailPrefix(prefix, false, isLast))
	} else {
		connector := "|- "
		if isLast {
			connector = "\\- "
		}
		appendStatusEntryRow(builder, entry, options, prefix+connector, statusDetailPrefix(prefix, true, isLast))
	}
	hintLines := statusHintLines(entry, section.fallback, options)
	if options.verbose {
		hintLines = statusRelationshipLines(entry, options)
	}
	if len(hintLines) > 0 {
		appendStatusHintLines(builder, hintLines, prefix, node.hasParent, isLast)
	}
	if len(node.children) == 0 {
		return
	}
	nextPrefix := prefix
	if node.hasParent {
		if isLast {
			nextPrefix += "   "
		} else {
			nextPrefix += "|  "
		}
	} else {
		nextPrefix = "  "
	}
	for i, child := range node.children {
		appendStatusNode(builder, child, section, options, nextPrefix, i == len(node.children)-1)
	}
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

func renderStatusEntrySummary(entry contracts.ChangeStatusEntry, options statusRenderOptions) string {
	return fmt.Sprintf("%s %s [%s %s]", renderLifecycleBadge(entry.Status, options.color), displayStatusID(entry.ID, options.verbose), emptyAsDash(entry.Type), emptyAsDash(entry.Size))
}

func appendStatusEntryRow(builder *strings.Builder, entry contracts.ChangeStatusEntry, options statusRenderOptions, linePrefix, detailPrefix string) {
	head := renderStatusEntrySummary(entry, options)
	builder.WriteString(linePrefix + head + "\n")
	title := strings.TrimSpace(entry.Title)
	if title == "" {
		return
	}
	available := statusTargetRowWidth - displayTextWidth(detailPrefix)
	if available < statusMinTitleWrapWidth {
		available = statusMinTitleWrapWidth
	}
	wrappedTitle := wrapStatusText(title, available)
	for _, line := range wrappedTitle {
		builder.WriteString(detailPrefix + line + "\n")
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

func statusHintLines(entry contracts.ChangeStatusEntry, fallback bool, options statusRenderOptions) []string {
	lines := make([]string, 0, 3)
	appendStatusRelationLine(&lines, "depends on", entry.DependsOn, options)
	appendStatusRelationLine(&lines, "superseded by", entry.SupersededBy, options)
	if fallback {
		appendStatusRelationLine(&lines, "related", entry.RelatedChanges, options)
	}
	return lines
}

func normalizeStatusRenderOptions(options statusRenderOptions) statusRenderOptions {
	if strings.TrimSpace(options.historyMode) == "" {
		options.historyMode = statusHistoryModeRecent
	}
	if options.historyLimit < 1 {
		options.historyLimit = defaultStatusHistoryLimit
	}
	return options
}

func selectHistoryEntries(entries []contracts.ChangeStatusEntry, options statusRenderOptions) ([]contracts.ChangeStatusEntry, bool) {
	if len(entries) == 0 {
		return entries, false
	}
	ordered := sortedStatusEntriesByRecency(entries)
	switch options.historyMode {
	case statusHistoryModeAll:
		return ordered, false
	case statusHistoryModeNone:
		return nil, len(ordered) > 0
	default:
		if len(ordered) <= options.historyLimit {
			return ordered, false
		}
		return ordered[:options.historyLimit], true
	}
}

func sortedStatusEntriesByRecency(entries []contracts.ChangeStatusEntry) []contracts.ChangeStatusEntry {
	ordered := append([]contracts.ChangeStatusEntry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := statusEntryRecencyKey(ordered[i])
		right := statusEntryRecencyKey(ordered[j])
		if left != right {
			return left > right
		}
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

func buildStatusSectionTree(entries []contracts.ChangeStatusEntry) statusSectionTree {
	entryByID := make(map[string]contracts.ChangeStatusEntry, len(entries))
	rank := make(map[string]int, len(entries))
	for i, entry := range entries {
		entryByID[entry.ID] = entry
		rank[entry.ID] = i
	}
	parents := make(map[string]string, len(entries))
	childrenByParent := make(map[string][]string, len(entries))
	fallback := false
	for _, entry := range entries {
		candidates := projectParentCandidates(entry, entryByID)
		if len(candidates) > 1 {
			fallback = true
			break
		}
		if len(candidates) == 1 {
			parentID := candidates[0]
			parents[entry.ID] = parentID
			childrenByParent[parentID] = append(childrenByParent[parentID], entry.ID)
		}
	}
	if fallback {
		roots := make([]statusTreeNode, 0, len(entries))
		for _, entry := range entries {
			roots = append(roots, statusTreeNode{entryID: entry.ID})
		}
		return statusSectionTree{roots: roots, fallback: true, rank: rank, entryByID: entryByID}
	}
	roots := make([]statusTreeNode, 0, len(entries))
	for _, entry := range entries {
		if _, hasParent := parents[entry.ID]; hasParent {
			continue
		}
		roots = append(roots, buildStatusNodeTree(entry.ID, childrenByParent, parents, entryByID, rank, map[string]struct{}{}))
	}
	return statusSectionTree{roots: roots, rank: rank, entryByID: entryByID}
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

func statusEntryRecencyKey(entry contracts.ChangeStatusEntry) string {
	for _, candidate := range []string{entry.ClosedAt, entry.CreatedAt} {
		if parsed, ok := parseStatusDate(candidate); ok {
			return parsed.Format("2006-01-02")
		}
	}
	return ""
}

func parseStatusDate(raw string) (time.Time, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func statusRelationshipLines(entry contracts.ChangeStatusEntry, options statusRenderOptions) []string {
	lines := make([]string, 0, 8)
	lines = append(lines, fmt.Sprintf("verification: %s", emptyAsDash(entry.VerificationStatus)))
	appendStatusRelationLine(&lines, "depends on", entry.DependsOn, options)
	appendStatusRelationLine(&lines, "related", entry.RelatedChanges, options)
	appendStatusRelationLine(&lines, "supersedes", entry.Supersedes, options)
	appendStatusRelationLine(&lines, "superseded by", entry.SupersededBy, options)
	if entry.CreatedAt != "" {
		lines = append(lines, fmt.Sprintf("created: %s", entry.CreatedAt))
	}
	if entry.ClosedAt != "" {
		lines = append(lines, fmt.Sprintf("closed: %s", entry.ClosedAt))
	}
	lines = append(lines, fmt.Sprintf("path: %s", entry.Path))
	return lines
}

func appendStatusRelationLine(lines *[]string, label string, ids []string, options statusRenderOptions) {
	if len(ids) == 0 {
		return
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	display := make([]string, 0, len(sorted))
	for _, id := range sorted {
		display = append(display, displayStatusID(id, options.verbose))
	}
	*lines = append(*lines, fmt.Sprintf("%s: %s", label, strings.Join(display, ", ")))
}

func renderVerificationBadge(status string, useColor bool) string {
	label := fmt.Sprintf("[%s]", emptyAsDash(status))
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "passed":
		return styleStatusText(label, ansiGreen, useColor)
	case "pending":
		return styleStatusText(label, ansiYellow, useColor)
	case "failed":
		return styleStatusText(label, ansiRed, useColor)
	case "skipped":
		return styleStatusText(label, ansiCyan, useColor)
	default:
		return styleStatusText(label, ansiDim, useColor)
	}
}

func renderLifecycleBadge(status string, useColor bool) string {
	label := fmt.Sprintf("[%s]", emptyAsDash(status))
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "proposed":
		return styleStatusText(label, ansiDim, useColor)
	case "planned":
		return styleStatusText(label, ansiBlue, useColor)
	case "implemented":
		return styleStatusText(label, ansiCyan, useColor)
	case "verified", "closed":
		return styleStatusText(label, ansiGreen, useColor)
	case "superseded":
		return styleStatusText(label, ansiYellow, useColor)
	default:
		return styleStatusText(label, ansiDim, useColor)
	}
}

func compactChangeID(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) >= 3 && parts[0] == "CHG" {
		return strings.Join(parts[:3], "-")
	}
	return id
}

func displayStatusID(id string, verbose bool) string {
	if verbose {
		return id
	}
	return compactChangeID(id)
}

func displayTextWidth(value string) int {
	return len([]rune(value))
}

func wrapStatusText(text string, maxWidth int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{""}
	}
	if maxWidth < 1 {
		return []string{trimmed}
	}
	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return []string{trimmed}
	}
	lines := make([]string, 0, 2)
	line := words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if displayTextWidth(candidate) <= maxWidth {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
	}
	lines = append(lines, line)
	return lines
}

func wrapStatusHintText(text string, maxWidth int) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []string{""}
	}
	label, body, ok := strings.Cut(trimmed, ": ")
	if !ok || strings.TrimSpace(body) == "" {
		return wrapStatusText(trimmed, maxWidth)
	}
	labelPrefix := label + ": "
	bodyWidth := maxWidth - displayTextWidth(labelPrefix)
	if bodyWidth < statusMinTitleWrapWidth {
		bodyWidth = statusMinTitleWrapWidth
	}
	bodyLines := wrapStatusText(body, bodyWidth)
	if len(bodyLines) == 0 {
		return []string{labelPrefix}
	}
	lines := make([]string, 0, len(bodyLines))
	lines = append(lines, labelPrefix+bodyLines[0])
	for _, line := range bodyLines[1:] {
		lines = append(lines, strings.Repeat(" ", displayTextWidth(labelPrefix))+line)
	}
	return lines
}

func styleStatusText(value, code string, enabled bool) string {
	if !enabled || value == "" {
		return value
	}
	return code + value + ansiReset
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func appendStatusExplainHuman(builder *strings.Builder, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary, options statusRenderOptions) {
	lines := appendStatusExplainLines(nil, loaded, summary)
	if len(lines) == 0 {
		return
	}
	builder.WriteString(styleStatusText("Explain", ansiBold, options.color))
	builder.WriteString("\n")
	for _, item := range lines {
		builder.WriteString(fmt.Sprintf("- %s: %s\n", item.key, item.value))
	}
	builder.WriteString("\n")
}

func shouldUseStatusColor(w io.Writer) bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if term := strings.TrimSpace(strings.ToLower(os.Getenv("TERM"))); term == "" || term == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
