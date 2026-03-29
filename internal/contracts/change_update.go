package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
)

func UpdateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (*ChangeOperationResult, error) {
	updated, statusWrite, context, err := prepareUpdateChange(v, loaded, changeID, options)
	if err != nil {
		return nil, err
	}
	if err := applyUpdateChangeWrite(v, loaded, statusWrite); err != nil {
		return nil, err
	}
	return buildUpdateChangeResult(updated, context), nil
}

type updateChangeContext struct {
	changeID     string
	writableRoot string
	record       *ChangeRecord
	status       LifecycleStatus
}

func prepareUpdateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (map[string]any, fileRewrite, updateChangeContext, error) {
	index, record, nextStatus, err := validateUpdateChangePreconditions(v, loaded, changeID, options)
	if err != nil {
		return nil, fileRewrite{}, updateChangeContext{}, err
	}
	defer index.Close()

	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, fileRewrite{}, updateChangeContext{}, err
	}
	updated := cloneMap(index.StatusFiles[record.StatusPath].Data)
	updated["status"] = string(nextStatus)
	statusWrite, _, err := buildPrimaryCloseStatusWrite(v, writableRoot, record, updated)
	if err != nil {
		return nil, fileRewrite{}, updateChangeContext{}, err
	}
	return updated, statusWrite, updateChangeContext{
		changeID:     changeID,
		writableRoot: writableRoot,
		record:       record,
		status:       nextStatus,
	}, nil
}

func validateUpdateChangePreconditions(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (*ProjectIndex, *ChangeRecord, LifecycleStatus, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, nil, "", err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, nil, "", err
	}
	record := index.Changes[changeID]
	if record == nil {
		index.Close()
		return nil, nil, "", fmt.Errorf("change %q does not exist", changeID)
	}
	nextStatus, err := validateUpdateLifecycleStatus(options.Status)
	if err != nil {
		index.Close()
		return nil, nil, "", err
	}
	if isTerminalLifecycleStatus(record.Status) {
		index.Close()
		return nil, nil, "", fmt.Errorf("change %q is already in terminal status %q and cannot be updated", changeID, record.Status)
	}
	if err := ValidateLifecycleTransition(string(record.Status), string(nextStatus)); err != nil {
		index.Close()
		return nil, nil, "", err
	}
	return index, record, nextStatus, nil
}

func applyUpdateChangeWrite(v *Validator, loaded *LoadedProject, statusWrite fileRewrite) error {
	return applyFileRewritesTransaction([]fileRewrite{statusWrite}, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	})
}

func buildUpdateChangeResult(updated map[string]any, context updateChangeContext) *ChangeOperationResult {
	promotionStatus, promotionTargets := closePromotionAssessmentDetails(updated)
	changedFiles := []FileMutation{{
		Path:   runeContextRelativePath(context.writableRoot, context.record.StatusPath),
		Action: "updated",
	}}
	changeDir := changeDirectoryPath(context.writableRoot, context.changeID)
	return &ChangeOperationResult{
		ID:                        context.changeID,
		ChangePath:                runeContextRelativePath(context.writableRoot, changeDir),
		Mode:                      inferChangeMode(changeDir),
		RecommendedMode:           inferChangeMode(changeDir),
		Status:                    string(context.status),
		ContextBundles:            append([]string(nil), context.record.ContextBundles...),
		ApplicableStandards:       append([]string(nil), context.record.ApplicableStandards...),
		ChangedFiles:              changedFiles,
		ReviewDiffRequired:        false,
		PromotionAssessmentStatus: promotionStatus,
		SuggestedPromotionTargets: promotionTargets,
	}
}

func changeDirectoryPath(writableRoot, changeID string) string {
	return filepath.Join(writableRoot, "changes", changeID)
}

func validateUpdateLifecycleStatus(status string) (LifecycleStatus, error) {
	trimmed := strings.TrimSpace(status)
	if trimmed == "" {
		return "", fmt.Errorf("change update requires --status")
	}
	switch LifecycleStatus(trimmed) {
	case StatusPlanned, StatusImplemented, StatusVerified:
		return LifecycleStatus(trimmed), nil
	default:
		return "", fmt.Errorf("change update --status must be one of planned, implemented, or verified")
	}
}
