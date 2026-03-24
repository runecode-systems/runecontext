package cli

import (
	"fmt"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

func buildBundleResolveOutput(absRoot string, loaded *contracts.LoadedProject, request bundleResolveRequest, resolution *contracts.BundleResolution, report *contracts.ContextPackReport, diagnostics []emittedDiagnostic) []line {
	output := []line{
		{"result", "ok"},
		{"command", bundleResolveCommand},
		{"root", absRoot},
		{"selected_config_path", selectedConfigPath(loaded)},
	}
	if loaded != nil && loaded.Resolution != nil {
		output = append(output,
			line{"project_root", loaded.Resolution.ProjectRoot},
			line{"source_root", loaded.Resolution.SourceRoot},
			line{"source_mode", string(loaded.Resolution.SourceMode)},
		)
	}
	output = append(output, line{"requested_bundle_count", fmt.Sprintf("%d", len(request.bundleIDs))})
	output = appendStringItems(output, "requested_bundle", request.bundleIDs)
	if resolution != nil {
		output = append(output,
			line{"bundle_resolution_id", resolution.ID},
			line{"resolved_bundle_count", fmt.Sprintf("%d", len(resolution.Linearization))},
		)
		output = appendStringItems(output, "resolved_bundle", resolution.Linearization)
	}
	output = appendContextPackReportLines(output, report)
	output = append(output, line{"diagnostic_count", fmt.Sprintf("%d", len(diagnostics))})
	return appendValidateDiagnosticLines(output, diagnostics)
}

func appendBundleResolveExplainLines(lines []line, loaded *contracts.LoadedProject, resolution *contracts.BundleResolution, report *contracts.ContextPackReport, diagnostics []emittedDiagnostic) []line {
	lines = append(lines,
		line{"explain_scope", "resolution,bundle-linearization,context-pack-report"},
		line{"explain_diagnostic_count", fmt.Sprintf("%d", len(diagnostics))},
	)
	if loaded != nil && loaded.Resolution != nil {
		lines = append(lines,
			line{"explain_resolution_source_mode", string(loaded.Resolution.SourceMode)},
			line{"explain_resolution_selected_config_path", loaded.Resolution.SelectedConfigPath},
		)
	}
	if resolution != nil {
		lines = append(lines, line{"explain_resolved_bundle_count", fmt.Sprintf("%d", len(resolution.Linearization))})
	}
	if report != nil {
		lines = append(lines,
			line{"explain_context_pack_warning_count", fmt.Sprintf("%d", len(report.Warnings))},
			line{"explain_context_pack_selected_files", fmt.Sprintf("%d", report.Summary.SelectedFiles)},
			line{"explain_context_pack_excluded_files", fmt.Sprintf("%d", report.Summary.ExcludedFiles)},
		)
	}
	return lines
}

func appendContextPackReportLines(lines []line, report *contracts.ContextPackReport) []line {
	if report == nil || report.Pack == nil {
		return lines
	}
	lines = append(lines,
		line{"context_pack_report_schema_version", fmt.Sprintf("%d", report.ReportSchemaVersion)},
		line{"context_pack_schema_version", fmt.Sprintf("%d", report.Pack.SchemaVersion)},
		line{"context_pack_id", report.Pack.ID},
		line{"context_pack_hash", report.Pack.PackHash},
		line{"context_pack_hash_alg", report.Pack.PackHashAlg},
		line{"context_pack_canonicalization", report.Pack.Canonicalization},
		line{"context_pack_selected_file_count", fmt.Sprintf("%d", report.Summary.SelectedFiles)},
		line{"context_pack_excluded_file_count", fmt.Sprintf("%d", report.Summary.ExcludedFiles)},
		line{"context_pack_referenced_content_bytes", fmt.Sprintf("%d", report.Summary.ReferencedContentBytes)},
		line{"context_pack_provenance_bytes", fmt.Sprintf("%d", report.Summary.ProvenanceBytes)},
		line{"context_pack_warning_count", fmt.Sprintf("%d", len(report.Warnings))},
	)
	for i, warning := range report.Warnings {
		prefix := fmt.Sprintf("context_pack_warning_%d", i+1)
		lines = append(lines,
			line{prefix + "_code", warning.Code},
			line{prefix + "_message", warning.Message},
			line{prefix + "_value", fmt.Sprintf("%d", warning.Value)},
			line{prefix + "_threshold", fmt.Sprintf("%d", warning.Threshold)},
		)
	}
	return lines
}
