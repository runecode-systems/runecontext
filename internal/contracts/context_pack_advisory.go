package contracts

import "fmt"

type ContextPackSummary struct {
	SelectedFiles          int   `json:"selected_files"`
	ExcludedFiles          int   `json:"excluded_files"`
	ReferencedContentBytes int64 `json:"referenced_content_bytes"`
	ProvenanceBytes        int64 `json:"provenance_bytes"`
}

type ContextPackAdvisory struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Value     int64  `json:"value"`
	Threshold int64  `json:"threshold"`
}

func buildContextPackReport(pack *ContextPack, digests []contextPackFileDigest, includeExplain bool, thresholds ContextPackAdvisoryThresholds) (*ContextPackReport, error) {
	explain := contextPackExplainReportFromPack(pack)
	provenanceBytes, err := contextPackProvenanceBytes(explain)
	if err != nil {
		return nil, err
	}
	if len(digests) != countContextPackSelectedFiles(pack.Selected) {
		return nil, fmt.Errorf("context pack selected-file digest count %d does not match selected inventory count %d", len(digests), countContextPackSelectedFiles(pack.Selected))
	}
	summary := ContextPackSummary{
		SelectedFiles:          len(digests),
		ExcludedFiles:          countContextPackExcludedFiles(pack.Excluded),
		ReferencedContentBytes: sumContextPackReferencedBytes(digests),
		ProvenanceBytes:        provenanceBytes,
	}
	report := &ContextPackReport{ReportSchemaVersion: contextPackReportSchemaVersion, Pack: pack, Summary: summary, Warnings: buildContextPackAdvisories(summary, thresholds)}
	if includeExplain {
		report.Explain = explain
	}
	return report, nil
}

func countContextPackExcludedFiles(excluded ContextPackExcludedAspectSet) int {
	return len(excluded.Project) + len(excluded.Standards) + len(excluded.Specs) + len(excluded.Decisions)
}

func countContextPackSelectedFiles(selected ContextPackAspectSet) int {
	return len(selected.Project) + len(selected.Standards) + len(selected.Specs) + len(selected.Decisions)
}

func sumContextPackReferencedBytes(digests []contextPackFileDigest) int64 {
	var total int64
	for _, digest := range digests {
		total += digest.ReferencedBytes
	}
	return total
}

func contextPackProvenanceBytes(explain *ContextPackExplainReport) (int64, error) {
	// Advisory provenance sizing uses the canonical explain representation so the
	// byte count is stable across runs, even though it remains a reporting metric
	// rather than part of the pack hash contract itself.
	data, err := marshalCanonicalJSON(contextPackExplainCanonicalValue(explain))
	if err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

func buildContextPackAdvisories(summary ContextPackSummary, thresholds ContextPackAdvisoryThresholds) []ContextPackAdvisory {
	warnings := make([]ContextPackAdvisory, 0, 3)
	if summary.SelectedFiles > thresholds.SelectedFiles {
		warnings = append(warnings, ContextPackAdvisory{Code: "selected_files_threshold_exceeded", Message: fmt.Sprintf("selected file count %d exceeds advisory threshold %d", summary.SelectedFiles, thresholds.SelectedFiles), Value: int64(summary.SelectedFiles), Threshold: int64(thresholds.SelectedFiles)})
	}
	if summary.ReferencedContentBytes > thresholds.ReferencedContentBytes {
		warnings = append(warnings, ContextPackAdvisory{Code: "referenced_content_bytes_threshold_exceeded", Message: fmt.Sprintf("referenced content bytes %d exceed advisory threshold %d", summary.ReferencedContentBytes, thresholds.ReferencedContentBytes), Value: summary.ReferencedContentBytes, Threshold: thresholds.ReferencedContentBytes})
	}
	if summary.ProvenanceBytes > thresholds.ProvenanceBytes {
		warnings = append(warnings, ContextPackAdvisory{Code: "provenance_bytes_threshold_exceeded", Message: fmt.Sprintf("provenance bytes %d exceed advisory threshold %d", summary.ProvenanceBytes, thresholds.ProvenanceBytes), Value: summary.ProvenanceBytes, Threshold: thresholds.ProvenanceBytes})
	}
	return warnings
}
