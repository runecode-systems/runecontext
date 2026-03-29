package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func ValidateLifecycleTransition(from, to string) error {
	fromStatus := LifecycleStatus(from)
	toStatus := LifecycleStatus(to)
	if _, ok := lifecycleOrder[fromStatus]; !ok {
		return fmt.Errorf("unknown lifecycle state %q", from)
	}
	if _, ok := lifecycleOrder[toStatus]; !ok {
		return fmt.Errorf("unknown lifecycle state %q", to)
	}
	if fromStatus == toStatus {
		return nil
	}
	if isTerminalLifecycleStatus(fromStatus) {
		return fmt.Errorf("cannot transition from terminal lifecycle state %q", from)
	}
	if lifecycleOrder[toStatus] < lifecycleOrder[fromStatus] {
		return fmt.Errorf("cannot transition backward from %q to %q", from, to)
	}
	return nil
}

func CloseChangeStatus(raw map[string]any, options CloseChangeOptions) (map[string]any, error) {
	if raw == nil {
		return nil, fmt.Errorf("status data is required")
	}
	statusValue, ok := raw["status"].(string)
	if !ok || strings.TrimSpace(statusValue) == "" {
		return nil, fmt.Errorf("status data must include a valid string status")
	}
	currentID, _ := raw["id"].(string)
	if err := validateSuccessorChangeIDs(options.SupersededBy, currentID); err != nil {
		return nil, err
	}
	nextStatus := string(StatusClosed)
	if len(options.SupersededBy) > 0 {
		nextStatus = string(StatusSuperseded)
	}
	if err := ValidateLifecycleTransition(statusValue, nextStatus); err != nil {
		return nil, err
	}
	return closedChangeStatusMap(raw, options, nextStatus), nil
}

func closedChangeStatusMap(raw map[string]any, options CloseChangeOptions, nextStatus string) map[string]any {
	closedAt := options.ClosedAt
	if closedAt.IsZero() {
		closedAt = time.Now().UTC()
	}
	updated := make(map[string]any, len(raw))
	for key, value := range raw {
		updated[key] = cloneTopLevelValue(value)
	}
	updated["closed_at"] = closedAt.Format("2006-01-02")
	updated["status"] = nextStatus
	if len(options.SupersededBy) > 0 {
		updated["superseded_by"] = stringSliceToAny(options.SupersededBy)
		return updated
	}
	updated["superseded_by"] = []any{}
	return updated
}

func buildChangeRecord(changeDir, statusPath string, data map[string]any) (*ChangeRecord, error) {
	id := fmt.Sprint(data["id"])
	if filepath.Base(changeDir) != id {
		return nil, &ValidationError{Path: statusPath, Message: fmt.Sprintf("change folder %q must match status id %q", filepath.Base(changeDir), id)}
	}
	fields, err := loadChangeRecordStringFields(statusPath, data)
	if err != nil {
		return nil, err
	}
	if err := validateChangeRecordSlices(statusPath, id, fields); err != nil {
		return nil, err
	}
	record := &ChangeRecord{
		ID:                 id,
		DirPath:            changeDir,
		StatusPath:         statusPath,
		Title:              requiredStringValue(data["title"]),
		Status:             LifecycleStatus(requiredStringValue(data["status"])),
		Type:               requiredStringValue(data["type"]),
		Size:               optionalStringValue(data["size"]),
		VerificationStatus: requiredStringValue(data["verification_status"]),
		CreatedAt:          optionalStringValue(data["created_at"]),
		ContextBundles:     extractStringList(data["context_bundles"]),
		RelatedSpecs:       fields.relatedSpecs,
		RelatedDecisions:   fields.relatedDecisions,
		RelatedChanges:     fields.relatedChanges,
		DependsOn:          fields.dependsOn,
		InformedBy:         fields.informedBy,
		Supersedes:         fields.supersedes,
		SupersededBy:       fields.supersededBy,
		Data:               data,
	}
	if closedAt, ok := data["closed_at"]; ok && closedAt != nil {
		record.ClosedAt = optionalStringValue(closedAt)
		record.HasClosedAt = true
	}
	return record, nil
}

type changeRecordStringFields struct {
	relatedSpecs     []string
	relatedDecisions []string
	relatedChanges   []string
	dependsOn        []string
	informedBy       []string
	supersedes       []string
	supersededBy     []string
}

func loadChangeRecordStringFields(statusPath string, data map[string]any) (changeRecordStringFields, error) {
	fieldNames := []struct {
		name   string
		target *[]string
	}{
		{"related_specs", nil},
	}
	_ = fieldNames
	relatedSpecs, err := stringSliceField(statusPath, "related_specs", data["related_specs"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	relatedDecisions, err := stringSliceField(statusPath, "related_decisions", data["related_decisions"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	relatedChanges, err := stringSliceField(statusPath, "related_changes", data["related_changes"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	dependsOn, err := stringSliceField(statusPath, "depends_on", data["depends_on"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	informedBy, err := stringSliceField(statusPath, "informed_by", data["informed_by"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	supersedes, err := stringSliceField(statusPath, "supersedes", data["supersedes"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	supersededBy, err := stringSliceField(statusPath, "superseded_by", data["superseded_by"])
	if err != nil {
		return changeRecordStringFields{}, err
	}
	return changeRecordStringFields{relatedSpecs, relatedDecisions, relatedChanges, dependsOn, informedBy, supersedes, supersededBy}, nil
}

func validateChangeRecordSlices(statusPath, id string, fields changeRecordStringFields) error {
	for _, field := range []struct {
		name  string
		items []string
	}{{"related_specs", fields.relatedSpecs}, {"related_decisions", fields.relatedDecisions}, {"related_changes", fields.relatedChanges}, {"depends_on", fields.dependsOn}, {"informed_by", fields.informedBy}, {"supersedes", fields.supersedes}, {"superseded_by", fields.supersededBy}} {
		if duplicate, ok := duplicateString(field.items); ok {
			return &ValidationError{Path: statusPath, Message: fmt.Sprintf("%s contains duplicate value %q", field.name, duplicate)}
		}
	}
	for _, field := range []struct {
		name  string
		items []string
	}{{"related_changes", fields.relatedChanges}, {"depends_on", fields.dependsOn}, {"informed_by", fields.informedBy}, {"supersedes", fields.supersedes}, {"superseded_by", fields.supersededBy}} {
		if containsString(field.items, id) {
			return &ValidationError{Path: statusPath, Message: fmt.Sprintf("%s must not reference the change itself", field.name)}
		}
	}
	return nil
}

func buildSpecRecord(path string, doc *FrontmatterDocument) (*SpecRecord, error) {
	originating, err := stringSliceField(path, "originating_changes", doc.Frontmatter["originating_changes"])
	if err != nil {
		return nil, err
	}
	revisedBy, err := stringSliceField(path, "revised_by_changes", doc.Frontmatter["revised_by_changes"])
	if err != nil {
		return nil, err
	}
	return &SpecRecord{OriginatingChanges: originating, RevisedByChanges: revisedBy}, nil
}

func buildDecisionRecord(path string, doc *FrontmatterDocument) (*DecisionRecord, error) {
	originating, err := stringSliceField(path, "originating_changes", doc.Frontmatter["originating_changes"])
	if err != nil {
		return nil, err
	}
	related, err := stringSliceField(path, "related_changes", doc.Frontmatter["related_changes"])
	if err != nil {
		return nil, err
	}
	return &DecisionRecord{OriginatingChanges: originating, RelatedChanges: related}, nil
}

func stringSliceField(path, field string, raw any) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, filepath.ToSlash(fmt.Sprint(item)))
	}
	return result, nil
}

func validateSuccessorChangeIDs(successorIDs []string, currentID string) error {
	if len(successorIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(successorIDs))
	for _, successorID := range successorIDs {
		if err := validateChangeIDValue(successorID, fmt.Sprintf("superseded_by entry %q", successorID)); err != nil {
			return err
		}
		if currentID != "" && successorID == currentID {
			return fmt.Errorf("superseded_by must not reference the change itself")
		}
		if _, ok := seen[successorID]; ok {
			return fmt.Errorf("superseded_by contains duplicate value %q", successorID)
		}
		seen[successorID] = struct{}{}
	}
	return nil
}

func validateChangeIDValue(id, label string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if !changeIDPattern.MatchString(id) {
		return fmt.Errorf("%s must match the canonical change ID format", label)
	}
	return nil
}
