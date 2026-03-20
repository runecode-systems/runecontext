package contracts

import "fmt"

type ContextPackExplainReport struct {
	RequestedBundleIDs []string                            `json:"requested_bundle_ids"`
	ResolvedBundleIDs  []string                            `json:"resolved_bundle_ids"`
	Selected           ContextPackExplainAspectSet         `json:"selected"`
	Excluded           ContextPackExplainExcludedAspectSet `json:"excluded"`
}

type ContextPackExplainAspectSet struct {
	Project   []ContextPackExplainSelectedFile `json:"project"`
	Standards []ContextPackExplainSelectedFile `json:"standards"`
	Specs     []ContextPackExplainSelectedFile `json:"specs"`
	Decisions []ContextPackExplainSelectedFile `json:"decisions"`
}

type ContextPackExplainExcludedAspectSet struct {
	Project   []ContextPackExplainExcludedFile `json:"project"`
	Standards []ContextPackExplainExcludedFile `json:"standards"`
	Specs     []ContextPackExplainExcludedFile `json:"specs"`
	Decisions []ContextPackExplainExcludedFile `json:"decisions"`
}

type ContextPackExplainSelectedFile struct {
	Path       string                     `json:"path"`
	SelectedBy []ContextPackRuleReference `json:"selected_by"`
}

type ContextPackExplainExcludedFile struct {
	Path     string                   `json:"path"`
	LastRule ContextPackRuleReference `json:"last_rule"`
}

func contextPackExplainReportFromPack(pack *ContextPack) *ContextPackExplainReport {
	if pack == nil {
		return nil
	}
	return &ContextPackExplainReport{
		RequestedBundleIDs: append([]string(nil), pack.RequestedBundleIDs...),
		ResolvedBundleIDs:  append([]string(nil), pack.ResolvedFrom.ContextBundleIDs...),
		Selected:           contextPackExplainSelectedFromPack(pack.Selected),
		Excluded:           contextPackExplainExcludedFromPack(pack.Excluded),
	}
}

func contextPackExplainReportFromResolution(requested []string, resolution *BundleResolution) *ContextPackExplainReport {
	if resolution == nil {
		return nil
	}
	return &ContextPackExplainReport{
		RequestedBundleIDs: append([]string(nil), requested...),
		ResolvedBundleIDs:  append([]string(nil), resolution.Linearization...),
		Selected:           contextPackExplainSelectedFromResolution(resolution),
		Excluded:           contextPackExplainExcludedFromResolution(resolution),
	}
}

func contextPackExplainSelectedFromPack(selected ContextPackAspectSet) ContextPackExplainAspectSet {
	return ContextPackExplainAspectSet{
		Project:   contextPackExplainSelectedFilesFromPack(selected.Project),
		Standards: contextPackExplainSelectedFilesFromPack(selected.Standards),
		Specs:     contextPackExplainSelectedFilesFromPack(selected.Specs),
		Decisions: contextPackExplainSelectedFilesFromPack(selected.Decisions),
	}
}

func contextPackExplainExcludedFromPack(excluded ContextPackExcludedAspectSet) ContextPackExplainExcludedAspectSet {
	return ContextPackExplainExcludedAspectSet{
		Project:   contextPackExplainExcludedFilesFromPack(excluded.Project),
		Standards: contextPackExplainExcludedFilesFromPack(excluded.Standards),
		Specs:     contextPackExplainExcludedFilesFromPack(excluded.Specs),
		Decisions: contextPackExplainExcludedFilesFromPack(excluded.Decisions),
	}
}

func contextPackExplainSelectedFromResolution(resolution *BundleResolution) ContextPackExplainAspectSet {
	return ContextPackExplainAspectSet{
		Project:   contextPackExplainSelectedFilesFromInventory(resolution.Aspects[BundleAspectProject].Selected),
		Standards: contextPackExplainSelectedFilesFromInventory(resolution.Aspects[BundleAspectStandards].Selected),
		Specs:     contextPackExplainSelectedFilesFromInventory(resolution.Aspects[BundleAspectSpecs].Selected),
		Decisions: contextPackExplainSelectedFilesFromInventory(resolution.Aspects[BundleAspectDecisions].Selected),
	}
}

func contextPackExplainExcludedFromResolution(resolution *BundleResolution) ContextPackExplainExcludedAspectSet {
	return ContextPackExplainExcludedAspectSet{
		Project:   contextPackExplainExcludedFilesFromInventory(resolution.Aspects[BundleAspectProject].Excluded),
		Standards: contextPackExplainExcludedFilesFromInventory(resolution.Aspects[BundleAspectStandards].Excluded),
		Specs:     contextPackExplainExcludedFilesFromInventory(resolution.Aspects[BundleAspectSpecs].Excluded),
		Decisions: contextPackExplainExcludedFilesFromInventory(resolution.Aspects[BundleAspectDecisions].Excluded),
	}
}

func contextPackExplainSelectedFilesFromPack(items []ContextPackSelectedFile) []ContextPackExplainSelectedFile {
	result := make([]ContextPackExplainSelectedFile, len(items))
	for i, item := range items {
		result[i] = ContextPackExplainSelectedFile{Path: item.Path, SelectedBy: append([]ContextPackRuleReference(nil), item.SelectedBy...)}
	}
	return result
}

func contextPackExplainExcludedFilesFromPack(items []ContextPackExcludedFile) []ContextPackExplainExcludedFile {
	result := make([]ContextPackExplainExcludedFile, len(items))
	for i, item := range items {
		result[i] = ContextPackExplainExcludedFile{Path: item.Path, LastRule: item.LastRule}
	}
	return result
}

func contextPackExplainSelectedFilesFromInventory(items []BundleInventoryEntry) []ContextPackExplainSelectedFile {
	result := make([]ContextPackExplainSelectedFile, len(items))
	for i, item := range items {
		result[i] = ContextPackExplainSelectedFile{Path: item.Path, SelectedBy: contextPackRuleReferences(item.MatchedBy)}
	}
	return result
}

func contextPackExplainExcludedFilesFromInventory(items []BundleInventoryEntry) []ContextPackExplainExcludedFile {
	result := make([]ContextPackExplainExcludedFile, len(items))
	for i, item := range items {
		result[i] = ContextPackExplainExcludedFile{Path: item.Path, LastRule: contextPackRuleReference(item.FinalRule)}
	}
	return result
}

func contextPackExplainReportsEqual(left, right *ContextPackExplainReport) (bool, error) {
	leftBytes, err := marshalCanonicalJSON(contextPackExplainCanonicalValue(left))
	if err != nil {
		return false, fmt.Errorf("canonicalize explain report: %w", err)
	}
	rightBytes, err := marshalCanonicalJSON(contextPackExplainCanonicalValue(right))
	if err != nil {
		return false, fmt.Errorf("canonicalize explain report: %w", err)
	}
	return string(leftBytes) == string(rightBytes), nil
}

func contextPackExplainCanonicalValue(explain *ContextPackExplainReport) map[string]any {
	if explain == nil {
		return map[string]any{}
	}
	return map[string]any{
		"requested_bundle_ids": append([]string(nil), explain.RequestedBundleIDs...),
		"resolved_bundle_ids":  append([]string(nil), explain.ResolvedBundleIDs...),
		"selected": map[string]any{
			"project":   contextPackExplainSelectedValue(explain.Selected.Project),
			"standards": contextPackExplainSelectedValue(explain.Selected.Standards),
			"specs":     contextPackExplainSelectedValue(explain.Selected.Specs),
			"decisions": contextPackExplainSelectedValue(explain.Selected.Decisions),
		},
		"excluded": map[string]any{
			"project":   contextPackExplainExcludedValue(explain.Excluded.Project),
			"standards": contextPackExplainExcludedValue(explain.Excluded.Standards),
			"specs":     contextPackExplainExcludedValue(explain.Excluded.Specs),
			"decisions": contextPackExplainExcludedValue(explain.Excluded.Decisions),
		},
	}
}

func contextPackExplainSelectedValue(items []ContextPackExplainSelectedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":        item.Path,
			"selected_by": contextPackRuleReferencesValue(item.SelectedBy),
		}
	}
	return result
}

func contextPackExplainExcludedValue(items []ContextPackExplainExcludedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":      item.Path,
			"last_rule": contextPackRuleReferenceValue(item.LastRule),
		}
	}
	return result
}
