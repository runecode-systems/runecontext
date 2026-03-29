package cli

import (
	"fmt"
	"strings"

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
	builder.WriteString(fmt.Sprintf("Root: %s\n", sanitizeStatusText(absRoot)))
	builder.WriteString(fmt.Sprintf("Config: %s\n", sanitizeStatusText(summary.SelectedConfigPath)))
	builder.WriteString(fmt.Sprintf("Version: %s  Assurance: %s\n", sanitizeStatusText(summary.RuneContextVersion), sanitizeStatusText(summary.AssuranceTier)))
	builder.WriteString(renderBundleSummary(summary.BundleIDs))
	builder.WriteString("\n\n")
}

func renderBundleSummary(bundleIDs []string) string {
	if len(bundleIDs) == 0 {
		return "Bundles (0): none"
	}
	display := make([]string, 0, len(bundleIDs))
	for _, id := range bundleIDs {
		display = append(display, sanitizeStatusText(id))
	}
	return fmt.Sprintf("Bundles (%d): %s", len(bundleIDs), strings.Join(display, ", "))
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
