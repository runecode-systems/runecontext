package contracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ContextPackOutputMode string

const (
	ContextPackOutputModeHuman   ContextPackOutputMode = "human"
	ContextPackOutputModeMachine ContextPackOutputMode = "machine"
)

func (r *ContextPackReport) Render(mode ContextPackOutputMode) ([]byte, error) {
	if r == nil || r.Pack == nil {
		return nil, fmt.Errorf("context pack report is unavailable")
	}
	switch mode {
	case ContextPackOutputModeHuman:
		return r.renderHuman(), nil
	case ContextPackOutputModeMachine:
		return r.renderMachine()
	default:
		return nil, fmt.Errorf("unknown context-pack output mode %q", mode)
	}
}

func (r *ContextPackReport) renderMachine() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r *ContextPackReport) renderHuman() []byte {
	var buf bytes.Buffer
	writeContextPackHumanHeader(&buf, r)
	writeContextPackHumanWarnings(&buf, r.Warnings)
	if r.Explain != nil {
		writeContextPackHumanExplain(&buf, r.Explain)
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func writeContextPackHumanHeader(buf *bytes.Buffer, report *ContextPackReport) {
	buf.WriteString(fmt.Sprintf("context pack %s\n", quoteContextPackHumanText(report.Pack.ID)))
	buf.WriteString(fmt.Sprintf("  pack hash: %s\n", report.Pack.PackHash))
	buf.WriteString(fmt.Sprintf("  requested bundles: %s\n", joinContextPackHumanText(report.Pack.RequestedBundleIDs)))
	buf.WriteString(fmt.Sprintf("  resolved bundles: %s\n", joinContextPackHumanText(report.Pack.ResolvedFrom.ContextBundleIDs)))
	buf.WriteString(fmt.Sprintf("  selected files: %d\n", report.Summary.SelectedFiles))
	buf.WriteString(fmt.Sprintf("  excluded files: %d\n", report.Summary.ExcludedFiles))
	buf.WriteString(fmt.Sprintf("  referenced content bytes: %d\n", report.Summary.ReferencedContentBytes))
	buf.WriteString(fmt.Sprintf("  provenance bytes: %d\n", report.Summary.ProvenanceBytes))
	buf.WriteString(fmt.Sprintf("  generated_at: %s\n", report.Pack.GeneratedAt))
	buf.WriteByte('\n')
}

func writeContextPackHumanWarnings(buf *bytes.Buffer, warnings []ContextPackAdvisory) {
	buf.WriteString("warnings:\n")
	if len(warnings) == 0 {
		buf.WriteString("  none\n\n")
		return
	}
	for _, warning := range warnings {
		buf.WriteString(fmt.Sprintf("  - %s: %s\n", quoteContextPackHumanText(warning.Code), quoteContextPackHumanText(warning.Message)))
	}
	buf.WriteByte('\n')
}

func writeContextPackHumanExplain(buf *bytes.Buffer, explain *ContextPackExplainReport) {
	buf.WriteString("selected:\n")
	writeContextPackHumanSelectedAspect(buf, BundleAspectProject, explain.Selected.Project)
	writeContextPackHumanSelectedAspect(buf, BundleAspectStandards, explain.Selected.Standards)
	writeContextPackHumanSelectedAspect(buf, BundleAspectSpecs, explain.Selected.Specs)
	writeContextPackHumanSelectedAspect(buf, BundleAspectDecisions, explain.Selected.Decisions)
	buf.WriteByte('\n')
	buf.WriteString("excluded:\n")
	writeContextPackHumanExcludedAspect(buf, BundleAspectProject, explain.Excluded.Project)
	writeContextPackHumanExcludedAspect(buf, BundleAspectStandards, explain.Excluded.Standards)
	writeContextPackHumanExcludedAspect(buf, BundleAspectSpecs, explain.Excluded.Specs)
	writeContextPackHumanExcludedAspect(buf, BundleAspectDecisions, explain.Excluded.Decisions)
}

func writeContextPackHumanSelectedAspect(buf *bytes.Buffer, aspect BundleAspect, items []ContextPackExplainSelectedFile) {
	buf.WriteString(fmt.Sprintf("  [%s]\n", aspect))
	if len(items) == 0 {
		buf.WriteString("    none\n")
		return
	}
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("    %s\n", quoteContextPackHumanText(item.Path)))
		for _, rule := range item.SelectedBy {
			buf.WriteString(fmt.Sprintf("      - %s bundle=%s pattern=%s kind=%s\n", quoteContextPackHumanText(string(rule.Rule)), quoteContextPackHumanText(rule.Bundle), quoteContextPackHumanText(rule.Pattern), quoteContextPackHumanText(string(rule.Kind))))
		}
	}
}

func writeContextPackHumanExcludedAspect(buf *bytes.Buffer, aspect BundleAspect, items []ContextPackExplainExcludedFile) {
	buf.WriteString(fmt.Sprintf("  [%s]\n", aspect))
	if len(items) == 0 {
		buf.WriteString("    none\n")
		return
	}
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("    %s\n", quoteContextPackHumanText(item.Path)))
		buf.WriteString(fmt.Sprintf("      - %s bundle=%s pattern=%s kind=%s\n", quoteContextPackHumanText(string(item.LastRule.Rule)), quoteContextPackHumanText(item.LastRule.Bundle), quoteContextPackHumanText(item.LastRule.Pattern), quoteContextPackHumanText(string(item.LastRule.Kind))))
	}
}

func joinContextPackHumanText(items []string) string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = quoteContextPackHumanText(item)
	}
	return strings.Join(result, ", ")
}

func quoteContextPackHumanText(value string) string {
	for _, r := range value {
		if unicode.IsControl(r) {
			return strconv.Quote(value)
		}
	}
	return value
}
