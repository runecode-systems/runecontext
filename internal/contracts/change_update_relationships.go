package contracts

import (
	"fmt"
	"strings"
)

func resolveRelatedChangeTargets(index *ProjectIndex, record *ChangeRecord, adds []string, removes []string) ([]*ChangeRecord, []string, error) {
	addSet, err := normalizedRelationshipTargetSet(adds)
	if err != nil {
		return nil, nil, err
	}
	removeSet, err := normalizedRelationshipTargetSet(removes)
	if err != nil {
		return nil, nil, err
	}
	if err := validateRelationshipEditConflict(addSet, removeSet); err != nil {
		return nil, nil, err
	}
	next := uniqueSortedStrings(append([]string(nil), record.RelatedChanges...))
	next, err = applyRelatedChangeAdds(index, record, next, addSet)
	if err != nil {
		return nil, nil, err
	}
	next = applyRelatedChangeRemovals(record, next, removeSet)
	next = uniqueSortedStrings(next)
	targets, err := collectRelatedChangeRecords(index, record.RelatedChanges, next)
	if err != nil {
		return nil, nil, err
	}
	return targets, next, nil
}

func validateRelationshipEditConflict(addSet, removeSet map[string]struct{}) error {
	for id := range addSet {
		if _, ok := removeSet[id]; ok {
			return fmt.Errorf("change update relationship edit lists conflict on %q", id)
		}
	}
	return nil
}

func applyRelatedChangeAdds(index *ProjectIndex, record *ChangeRecord, next []string, addSet map[string]struct{}) ([]string, error) {
	for _, id := range SortedKeys(addSet) {
		if err := validateRelatedAddTarget(index, record.ID, id); err != nil {
			return nil, err
		}
		if !containsString(next, id) {
			next = append(next, id)
		}
	}
	return next, nil
}

func validateRelatedAddTarget(index *ProjectIndex, changeID, relatedID string) error {
	if relatedID == changeID {
		return fmt.Errorf("change update must not relate a change to itself")
	}
	if index.Changes[relatedID] == nil {
		return fmt.Errorf("change update related change %q does not exist", relatedID)
	}
	return nil
}

func applyRelatedChangeRemovals(record *ChangeRecord, next []string, removeSet map[string]struct{}) []string {
	for _, id := range SortedKeys(removeSet) {
		if !containsString(record.RelatedChanges, id) {
			continue
		}
		next = removeStringValue(next, id)
	}
	return next
}

func collectRelatedChangeRecords(index *ProjectIndex, existing []string, next []string) ([]*ChangeRecord, error) {
	targetSet := map[string]struct{}{}
	for _, id := range existing {
		targetSet[id] = struct{}{}
	}
	for _, id := range next {
		targetSet[id] = struct{}{}
	}
	targets := make([]*ChangeRecord, 0, len(targetSet))
	for _, id := range SortedKeys(targetSet) {
		target := index.Changes[id]
		if target == nil {
			return nil, fmt.Errorf("change update related change %q does not exist", id)
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func normalizedRelationshipTargetSet(items []string) (map[string]struct{}, error) {
	set := map[string]struct{}{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			return nil, fmt.Errorf("change update relationship edit IDs must not be blank")
		}
		set[trimmed] = struct{}{}
	}
	return set, nil
}

func removeStringValue(items []string, target string) []string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if item == target {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func buildReciprocalRelatedChangeWrites(v *Validator, index *ProjectIndex, writableRoot string, relatedTargets []*ChangeRecord, changeID string, rootRelatedChanges []string) ([]fileRewrite, []FileMutation, error) {
	rootRelatedSet := map[string]struct{}{}
	for _, id := range rootRelatedChanges {
		rootRelatedSet[id] = struct{}{}
	}
	writes := make([]fileRewrite, 0, len(relatedTargets))
	changed := make([]FileMutation, 0, len(relatedTargets))
	for _, target := range relatedTargets {
		targetStatus := cloneMap(index.StatusFiles[target.StatusPath].Data)
		before := uniqueSortedStrings(extractStringList(targetStatus["related_changes"]))
		after := mergeRelatedChangeReciprocal(before, changeID, target.ID, rootRelatedSet)
		if slicesEqualStrings(after, before) {
			continue
		}
		targetStatus["related_changes"] = stringSliceToAny(after)
		write, _, err := buildPrimaryCloseStatusWrite(v, writableRoot, target, targetStatus)
		if err != nil {
			return nil, nil, err
		}
		writes = append(writes, write)
		changed = append(changed, FileMutation{Path: runeContextRelativePath(writableRoot, target.StatusPath), Action: "updated"})
	}
	return writes, changed, nil
}

func mergeRelatedChangeReciprocal(existing []string, changeID string, targetID string, rootRelatedSet map[string]struct{}) []string {
	updated := append([]string(nil), existing...)
	_, shouldLink := rootRelatedSet[targetID]
	hasLink := containsString(updated, changeID)
	if shouldLink && !hasLink {
		updated = append(updated, changeID)
	}
	if !shouldLink && hasLink {
		updated = removeStringValue(updated, changeID)
	}
	return uniqueSortedStrings(updated)
}

func slicesEqualStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func insertOrReplaceStatusWrite(writes *[]fileRewrite, byPath map[string]int, write fileRewrite) {
	if index, ok := byPath[write.Path]; ok {
		(*writes)[index] = write
		return
	}
	byPath[write.Path] = len(*writes)
	*writes = append(*writes, write)
}

func insertOrReplaceChangedFile(items *[]FileMutation, writableRoot, statusPath string) {
	insertOrReplaceChangedFileEntry(items, FileMutation{Path: runeContextRelativePath(writableRoot, statusPath), Action: "updated"})
}

func insertOrReplaceChangedFileEntry(items *[]FileMutation, next FileMutation) {
	for i := range *items {
		if (*items)[i].Path == next.Path {
			(*items)[i] = next
			return
		}
	}
	*items = append(*items, next)
}
