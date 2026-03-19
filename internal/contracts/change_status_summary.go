package contracts

import (
	"fmt"
	"path/filepath"
)

func BuildProjectStatusSummary(v *Validator, loaded *LoadedProject) (*ProjectStatusSummary, error) {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	summary := newProjectStatusSummary(loaded)
	appendProjectChangeStatus(summary, index)
	appendProjectBundleStatus(summary, index)
	return summary, nil
}

func newProjectStatusSummary(loaded *LoadedProject) *ProjectStatusSummary {
	return &ProjectStatusSummary{
		Root:               loaded.Resolution.ProjectRoot,
		SelectedConfigPath: loaded.Resolution.SelectedConfigPath,
		RuneContextVersion: fmt.Sprint(loaded.RootConfig["runecontext_version"]),
		AssuranceTier:      fmt.Sprint(loaded.RootConfig["assurance_tier"]),
		Active:             make([]ChangeStatusEntry, 0),
		Closed:             make([]ChangeStatusEntry, 0),
		Superseded:         make([]ChangeStatusEntry, 0),
	}
}

func appendProjectChangeStatus(summary *ProjectStatusSummary, index *ProjectIndex) {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		entry := changeStatusEntryFromRecord(index, record)
		appendSummaryEntry(summary, entry, record.Status)
	}
}

func changeStatusEntryFromRecord(index *ProjectIndex, record *ChangeRecord) ChangeStatusEntry {
	return ChangeStatusEntry{
		ID:     record.ID,
		Title:  record.Title,
		Status: string(record.Status),
		Type:   record.Type,
		Size:   record.Size,
		Path:   runeContextRelativePath(index.ContentRoot, filepath.Join(record.DirPath, "status.yaml")),
	}
}

func appendSummaryEntry(summary *ProjectStatusSummary, entry ChangeStatusEntry, status LifecycleStatus) {
	switch status {
	case StatusClosed:
		summary.Closed = append(summary.Closed, entry)
	case StatusSuperseded:
		summary.Superseded = append(summary.Superseded, entry)
	default:
		summary.Active = append(summary.Active, entry)
	}
}

func appendProjectBundleStatus(summary *ProjectStatusSummary, index *ProjectIndex) {
	if index.Bundles != nil {
		summary.BundleIDs = SortedKeys(index.Bundles.bundles)
	}
}
