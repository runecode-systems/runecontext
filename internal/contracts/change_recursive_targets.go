package contracts

import (
	"fmt"
	"sort"
)

func resolveRecursiveFeatureSubChangeTargets(index *ProjectIndex, umbrella *ChangeRecord) ([]*ChangeRecord, error) {
	if umbrella == nil {
		return nil, fmt.Errorf("recursive lifecycle propagation requires a selected change")
	}
	if umbrella.Type != "project" {
		return nil, fmt.Errorf("change %q is type %q; --recursive is only supported for type project umbrella changes", umbrella.ID, umbrella.Type)
	}
	targets := make([]*ChangeRecord, 0, len(umbrella.RelatedChanges))
	for _, relatedID := range uniqueSortedStrings(umbrella.RelatedChanges) {
		related := index.Changes[relatedID]
		if related == nil {
			return nil, fmt.Errorf("umbrella change %q references missing related change %q", umbrella.ID, relatedID)
		}
		if !containsString(related.RelatedChanges, umbrella.ID) {
			continue
		}
		if related.Type != "feature" {
			return nil, fmt.Errorf("change %q is not an eligible feature sub-change for umbrella %q", related.ID, umbrella.ID)
		}
		targets = append(targets, related)
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].ID < targets[j].ID
	})
	return targets, nil
}
