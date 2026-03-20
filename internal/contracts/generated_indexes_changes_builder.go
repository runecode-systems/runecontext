package contracts

import "fmt"

func (p *ProjectIndex) BuildGeneratedChangesByStatusIndex() (*GeneratedChangesByStatusIndex, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	index := &GeneratedChangesByStatusIndex{
		SchemaVersion: generatedChangeIndexSchemaVersion,
		Statuses:      newGeneratedChangeStatusGroups(),
	}
	for _, changeID := range SortedKeys(p.Changes) {
		record := p.Changes[changeID]
		if record == nil {
			continue
		}
		entry, err := buildGeneratedChangeStatusEntry(p.ContentRoot, record)
		if err != nil {
			return nil, err
		}
		if err := appendChangeStatusEntry(index, record.Status, entry); err != nil {
			return nil, err
		}
	}
	return index, nil
}

func newGeneratedChangeStatusGroups() GeneratedChangeStatusGroups {
	return GeneratedChangeStatusGroups{
		Proposed:    []GeneratedChangeStatusEntry{},
		Planned:     []GeneratedChangeStatusEntry{},
		Implemented: []GeneratedChangeStatusEntry{},
		Verified:    []GeneratedChangeStatusEntry{},
		Closed:      []GeneratedChangeStatusEntry{},
		Superseded:  []GeneratedChangeStatusEntry{},
	}
}

func buildGeneratedChangeStatusEntry(contentRoot string, record *ChangeRecord) (GeneratedChangeStatusEntry, error) {
	statusPath, err := generatedRelativeArtifactPath(contentRoot, record.StatusPath)
	if err != nil {
		return GeneratedChangeStatusEntry{}, fmt.Errorf("build generated changes index: %w", err)
	}
	return GeneratedChangeStatusEntry{
		ID:    record.ID,
		Title: record.Title,
		Type:  record.Type,
		Size:  record.Size,
		Path:  statusPath,
	}, nil
}

func appendChangeStatusEntry(index *GeneratedChangesByStatusIndex, status LifecycleStatus, entry GeneratedChangeStatusEntry) error {
	switch status {
	case StatusProposed:
		index.Statuses.Proposed = append(index.Statuses.Proposed, entry)
	case StatusPlanned:
		index.Statuses.Planned = append(index.Statuses.Planned, entry)
	case StatusImplemented:
		index.Statuses.Implemented = append(index.Statuses.Implemented, entry)
	case StatusVerified:
		index.Statuses.Verified = append(index.Statuses.Verified, entry)
	case StatusClosed:
		index.Statuses.Closed = append(index.Statuses.Closed, entry)
	case StatusSuperseded:
		index.Statuses.Superseded = append(index.Statuses.Superseded, entry)
	default:
		return fmt.Errorf("build generated changes index: change %q has unsupported lifecycle status %q", entry.ID, status)
	}
	return nil
}
