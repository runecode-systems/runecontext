package contracts

import "strings"

func ListStandards(v *Validator, loaded *LoadedProject, options StandardListOptions) (*StandardListResult, error) {
	if err := validateChangeCommandInputs(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	normalizedScopePaths, err := normalizeStandardScopePaths(options.ScopePaths)
	if err != nil {
		return nil, err
	}
	normalizedStatuses, err := normalizeStandardStatuses(options.Statuses)
	if err != nil {
		return nil, err
	}
	focus := strings.ToLower(strings.TrimSpace(options.Focus))
	standards := listFilteredStandards(index, normalizedScopePaths, focus, normalizedStatuses)
	return &StandardListResult{
		ScopePaths: normalizedScopePaths,
		Focus:      focus,
		Statuses:   normalizedStatuses,
		Standards:  standards,
	}, nil
}

func listFilteredStandards(index *ProjectIndex, scopePaths []string, focus string, statuses []StandardStatus) []StandardRecord {
	if index == nil {
		return nil
	}
	allowedStatuses := makeAllowedStatusSet(statuses)
	paths := SortedKeys(index.Standards)
	out := make([]StandardRecord, 0, len(paths))
	for _, path := range paths {
		record := index.Standards[path]
		if shouldSkipStandardListRecord(record, allowedStatuses, path, scopePaths, focus) {
			continue
		}
		out = append(out, cloneStandardListRecord(record))
	}
	return out
}

func makeAllowedStatusSet(statuses []StandardStatus) map[StandardStatus]struct{} {
	allowed := map[StandardStatus]struct{}{}
	for _, status := range statuses {
		allowed[status] = struct{}{}
	}
	return allowed
}

func shouldSkipStandardListRecord(record *StandardRecord, allowedStatuses map[StandardStatus]struct{}, path string, scopePaths []string, focus string) bool {
	if record == nil {
		return true
	}
	if len(allowedStatuses) > 0 {
		if _, ok := allowedStatuses[record.Status]; !ok {
			return true
		}
	}
	if !standardPathMatchesScope(path, scopePaths) {
		return true
	}
	return !standardMatchesFocus(path, record, focus)
}

func standardPathMatchesScope(path string, scopes []string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, scope := range scopes {
		if path == scope || strings.HasPrefix(path, scope+"/") {
			return true
		}
	}
	return false
}

func standardMatchesFocus(path string, record *StandardRecord, focus string) bool {
	if focus == "" {
		return true
	}
	for _, candidate := range []string{path, record.ID, record.Title} {
		if strings.Contains(strings.ToLower(candidate), focus) {
			return true
		}
	}
	return false
}

func cloneStandardListRecord(record *StandardRecord) StandardRecord {
	if record == nil {
		return StandardRecord{}
	}
	return StandardRecord{
		Path:                    record.Path,
		ID:                      record.ID,
		Title:                   record.Title,
		Status:                  record.Status,
		ReplacedBy:              record.ReplacedBy,
		Aliases:                 append([]string(nil), record.Aliases...),
		SuggestedContextBundles: append([]string(nil), record.SuggestedContextBundles...),
	}
}
