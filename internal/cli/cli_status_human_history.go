package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

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
		lines = append(lines, fmt.Sprintf("created: %s", sanitizeStatusText(entry.CreatedAt)))
	}
	if entry.ClosedAt != "" {
		lines = append(lines, fmt.Sprintf("closed: %s", sanitizeStatusText(entry.ClosedAt)))
	}
	lines = append(lines, fmt.Sprintf("path: %s", sanitizeStatusText(entry.Path)))
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
