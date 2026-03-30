package contracts

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func UpdateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (*ChangeOperationResult, error) {
	updated, statusWrites, context, err := prepareUpdateChange(v, loaded, changeID, options)
	if err != nil {
		return nil, err
	}
	if err := applyUpdateChangeWrites(v, loaded, statusWrites); err != nil {
		return nil, err
	}
	return buildUpdateChangeResult(updated, context), nil
}

type updateChangeContext struct {
	changeID             string
	writableRoot         string
	record               *ChangeRecord
	status               LifecycleStatus
	changedFiles         []FileMutation
	recursive            bool
	recursiveTargetIDs   []string
	recursiveTargetCount int
}

func prepareUpdateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (map[string]any, []fileRewrite, updateChangeContext, error) {
	index, record, targets, nextStatus, err := validateUpdateChangePreconditions(v, loaded, changeID, options)
	if err != nil {
		return nil, nil, updateChangeContext{}, err
	}
	defer index.Close()

	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, nil, updateChangeContext{}, err
	}
	updated, statusWrites, changedFiles, err := buildUpdateTargetWrites(v, index, writableRoot, targets, changeID, nextStatus, options.VerificationStatus)
	if err != nil {
		return nil, nil, updateChangeContext{}, err
	}
	recursiveTargetIDs := collectRecursiveTargetIDs(targets, changeID)
	return updated, statusWrites, updateChangeContext{
		changeID:             changeID,
		writableRoot:         writableRoot,
		record:               record,
		status:               nextStatus,
		changedFiles:         changedFiles,
		recursive:            options.Recursive,
		recursiveTargetIDs:   recursiveTargetIDs,
		recursiveTargetCount: len(recursiveTargetIDs),
	}, nil
}

func buildUpdateTargetWrites(v *Validator, index *ProjectIndex, writableRoot string, targets []*ChangeRecord, changeID string, nextStatus LifecycleStatus, requestedVerificationStatus string) (map[string]any, []fileRewrite, []FileMutation, error) {
	verificationStatus := strings.TrimSpace(requestedVerificationStatus)
	statusWrites := make([]fileRewrite, 0, len(targets))
	changedFiles := make([]FileMutation, 0, len(targets))
	updated := map[string]any{}
	for _, target := range targets {
		targetStatus := cloneMap(index.StatusFiles[target.StatusPath].Data)
		targetStatus["status"] = string(nextStatus)
		if verificationStatus != "" {
			targetStatus["verification_status"] = verificationStatus
		}
		statusWrite, _, err := buildPrimaryCloseStatusWrite(v, writableRoot, target, targetStatus)
		if err != nil {
			return nil, nil, nil, err
		}
		statusWrites = append(statusWrites, statusWrite)
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, target.StatusPath), Action: "updated"})
		if target.ID == changeID {
			updated = targetStatus
		}
	}
	sortFileMutations(changedFiles)
	return updated, statusWrites, changedFiles, nil
}

func collectRecursiveTargetIDs(targets []*ChangeRecord, rootID string) []string {
	recursiveTargetIDs := make([]string, 0, len(targets)-1)
	for _, target := range targets {
		if target.ID != rootID {
			recursiveTargetIDs = append(recursiveTargetIDs, target.ID)
		}
	}
	sort.Strings(recursiveTargetIDs)
	return recursiveTargetIDs
}

func validateUpdateChangePreconditions(v *Validator, loaded *LoadedProject, changeID string, options ChangeUpdateOptions) (*ProjectIndex, *ChangeRecord, []*ChangeRecord, LifecycleStatus, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, nil, nil, "", err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, nil, nil, "", err
	}
	record := index.Changes[changeID]
	if record == nil {
		index.Close()
		return nil, nil, nil, "", fmt.Errorf("change %q does not exist", changeID)
	}
	targets, err := resolveUpdateTargets(index, record, options.Recursive)
	if err != nil {
		index.Close()
		return nil, nil, nil, "", err
	}
	nextStatus, err := validateUpdateLifecycleStatus(options.Status)
	if err != nil {
		index.Close()
		return nil, nil, nil, "", err
	}
	if err := validateUpdateTargetsForTransition(targets, nextStatus, options.VerificationStatus); err != nil {
		index.Close()
		return nil, nil, nil, "", err
	}
	return index, record, targets, nextStatus, nil
}

func resolveUpdateTargets(index *ProjectIndex, record *ChangeRecord, recursive bool) ([]*ChangeRecord, error) {
	targets := []*ChangeRecord{record}
	if !recursive {
		return targets, nil
	}
	recursiveTargets, err := resolveRecursiveFeatureSubChangeTargets(index, record)
	if err != nil {
		return nil, err
	}
	return append(targets, recursiveTargets...), nil
}

func validateUpdateTargetsForTransition(targets []*ChangeRecord, nextStatus LifecycleStatus, verificationStatus string) error {
	for _, target := range targets {
		if err := validateUpdateTargetForTransition(target, nextStatus, verificationStatus); err != nil {
			return err
		}
	}
	return nil
}

func validateUpdateTargetForTransition(target *ChangeRecord, nextStatus LifecycleStatus, verificationStatus string) error {
	if isTerminalLifecycleStatus(target.Status) {
		return fmt.Errorf("change %q is already in terminal status %q and cannot be updated", target.ID, target.Status)
	}
	if err := ValidateLifecycleTransition(string(target.Status), string(nextStatus)); err != nil {
		return fmt.Errorf("change %q: %w", target.ID, err)
	}
	if err := validateUpdateVerificationStatus(target, nextStatus, verificationStatus); err != nil {
		return fmt.Errorf("change %q: %w", target.ID, err)
	}
	return nil
}

func applyUpdateChangeWrites(v *Validator, loaded *LoadedProject, statusWrites []fileRewrite) error {
	return applyFileRewritesTransaction(statusWrites, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	})
}

func buildUpdateChangeResult(updated map[string]any, context updateChangeContext) *ChangeOperationResult {
	promotionStatus, promotionTargets := closePromotionAssessmentDetails(updated)
	changeDir := changeDirectoryPath(context.writableRoot, context.changeID)
	return &ChangeOperationResult{
		ID:                        context.changeID,
		ChangePath:                runeContextRelativePath(context.writableRoot, changeDir),
		Mode:                      inferChangeMode(changeDir),
		RecommendedMode:           inferChangeMode(changeDir),
		Status:                    string(context.status),
		ContextBundles:            append([]string(nil), context.record.ContextBundles...),
		ApplicableStandards:       append([]string(nil), context.record.ApplicableStandards...),
		ChangedFiles:              append([]FileMutation(nil), context.changedFiles...),
		ReviewDiffRequired:        false,
		PromotionAssessmentStatus: promotionStatus,
		SuggestedPromotionTargets: promotionTargets,
		Recursive:                 context.recursive,
		RecursiveTargetCount:      context.recursiveTargetCount,
		RecursiveTargetIDs:        append([]string(nil), context.recursiveTargetIDs...),
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

func validateUpdateVerificationStatus(record *ChangeRecord, nextStatus LifecycleStatus, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		if nextStatus != StatusVerified {
			return fmt.Errorf("change update --verification-status is only supported when --status verified")
		}
		if requested == "pending" {
			return fmt.Errorf("change update must not set verification_status to pending")
		}
		if !isSupportedVerificationStatus(requested) {
			return fmt.Errorf("unsupported verification_status %q", requested)
		}
		return nil
	}
	if nextStatus == StatusVerified {
		existing := strings.TrimSpace(record.VerificationStatus)
		if !isSupportedVerificationStatus(existing) {
			return fmt.Errorf("verified changes must record a completed verification_status")
		}
	}
	return nil
}
