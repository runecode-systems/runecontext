package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type line struct {
	key   string
	value string
}

type emittedDiagnostic struct {
	Severity contracts.DiagnosticSeverity
	Code     string
	Message  string
	Path     string
	Bundle   string
	Aspect   string
	Rule     string
	Pattern  string
	Matches  []string
}

func writeLines(w io.Writer, lines ...line) {
	for _, entry := range lines {
		fmt.Fprintf(w, "%s=%s\n", entry.key, sanitizeValue(entry.value))
	}
}

func writeCommandUsageError(w io.Writer, command, usage string, err error) {
	writeLines(w, buildCommandUsageErrorLines(command, usage, err)...)
}

func writeCommandInvalid(w io.Writer, command, root string, err error) {
	writeLines(w, buildCommandInvalidLines(command, root, err)...)
}

func buildCommandUsageErrorLines(command, usage string, err error) []line {
	return []line{
		{"result", "usage_error"},
		{"command", command},
		{"error_message", err.Error()},
		{"usage", usage},
	}
}

func buildCommandInvalidLines(command, root string, err error) []line {
	return []line{
		{"result", "invalid"},
		{"command", command},
		{"root", root},
		{"error_message", err.Error()},
	}
}

func sanitizeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\t", "\\t")
	value = strings.ReplaceAll(value, "\x00", "\\0")
	value = strings.ReplaceAll(value, "=", "\\=")
	return value
}

func appendStatusEntries(lines []line, prefix string, entries []contracts.ChangeStatusEntry) []line {
	for i, entry := range entries {
		index := i + 1
		lines = append(lines,
			line{fmt.Sprintf("%s_%d_id", prefix, index), entry.ID},
			line{fmt.Sprintf("%s_%d_title", prefix, index), entry.Title},
			line{fmt.Sprintf("%s_%d_status", prefix, index), entry.Status},
			line{fmt.Sprintf("%s_%d_type", prefix, index), entry.Type},
			line{fmt.Sprintf("%s_%d_size", prefix, index), entry.Size},
			line{fmt.Sprintf("%s_%d_path", prefix, index), entry.Path},
		)
	}
	return lines
}

func appendStringItems(lines []line, prefix string, items []string) []line {
	for i, item := range items {
		lines = append(lines, line{fmt.Sprintf("%s_%d", prefix, i+1), item})
	}
	return lines
}

func appendReasonsAndAssumptions(lines []line, reasons, assumptions []string) []line {
	lines = append(lines, line{"reason_count", fmt.Sprintf("%d", len(reasons))})
	for i, reason := range reasons {
		lines = append(lines, line{fmt.Sprintf("reason_%d", i+1), reason})
	}
	lines = append(lines, line{"assumption_count", fmt.Sprintf("%d", len(assumptions))})
	for i, assumption := range assumptions {
		lines = append(lines, line{fmt.Sprintf("assumption_%d", i+1), assumption})
	}
	return lines
}

func appendWarnings(lines []line, warnings []string) []line {
	lines = append(lines, line{"warning_count", fmt.Sprintf("%d", len(warnings))})
	for i, warning := range warnings {
		lines = append(lines, line{fmt.Sprintf("warning_%d", i+1), warning})
	}
	return lines
}

func appendChangedFiles(lines []line, changed []contracts.FileMutation) []line {
	lines = append(lines, line{"changed_file_count", fmt.Sprintf("%d", len(changed))})
	for i, file := range changed {
		prefix := fmt.Sprintf("changed_file_%d", i+1)
		lines = append(lines,
			line{prefix + "_path", file.Path},
			line{prefix + "_action", file.Action},
		)
	}
	return lines
}

func collectDiagnostics(index *contracts.ProjectIndex) []emittedDiagnostic {
	if index == nil {
		return nil
	}

	items := make([]emittedDiagnostic, 0)
	if index.Resolution != nil {
		for _, diagnostic := range index.Resolution.Diagnostics {
			items = append(items, emittedDiagnostic{Severity: diagnostic.Severity, Code: diagnostic.Code, Message: diagnostic.Message})
		}
	}
	for _, diagnostic := range index.Diagnostics {
		items = append(items, emittedDiagnostic{Severity: diagnostic.Severity, Code: diagnostic.Code, Message: diagnostic.Message, Path: diagnostic.Path})
	}
	if index.Bundles != nil {
		for _, diagnostic := range index.Bundles.Diagnostics() {
			items = append(items, emittedDiagnostic{
				Severity: diagnostic.Severity,
				Code:     diagnostic.Code,
				Message:  diagnostic.Message,
				Bundle:   diagnostic.Bundle,
				Aspect:   string(diagnostic.Aspect),
				Rule:     string(diagnostic.Rule),
				Pattern:  diagnostic.Pattern,
				Matches:  append([]string(nil), diagnostic.Matches...),
			})
		}
	}
	return dedupeDiagnostics(items)
}

func dedupeDiagnostics(items []emittedDiagnostic) []emittedDiagnostic {
	if len(items) == 0 {
		return nil
	}

	result := make([]emittedDiagnostic, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.Join([]string{
			string(item.Severity),
			item.Code,
			item.Message,
			item.Path,
			item.Bundle,
			item.Aspect,
			item.Rule,
			item.Pattern,
			strings.Join(item.Matches, ","),
		}, "\x1f")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}
