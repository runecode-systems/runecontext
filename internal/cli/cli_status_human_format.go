package cli

import (
	"fmt"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func renderStatusEntrySummary(entry contracts.ChangeStatusEntry, options statusRenderOptions) string {
	return fmt.Sprintf("%s %s [%s %s]", renderLifecycleBadge(entry.Status, options.color), displayStatusID(entry.ID, options.verbose), emptyAsDash(entry.Type), emptyAsDash(entry.Size))
}

func appendStatusEntryRow(builder *strings.Builder, entry contracts.ChangeStatusEntry, options statusRenderOptions, linePrefix, detailPrefix string) {
	head := renderStatusEntrySummary(entry, options)
	builder.WriteString(linePrefix + head + "\n")
	title := strings.TrimSpace(sanitizeStatusText(entry.Title))
	if title == "" {
		return
	}
	available := statusTargetRowWidth - displayTextWidth(detailPrefix)
	if available < statusMinTitleWrapWidth {
		available = statusMinTitleWrapWidth
	}
	for _, line := range wrapStatusText(title, available) {
		builder.WriteString(detailPrefix + line + "\n")
	}
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
	id = sanitizeStatusText(id)
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
	value = sanitizeStatusText(value)
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
