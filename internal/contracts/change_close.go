package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func CloseChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeCloseOptions) (*ChangeOperationResult, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	record, writableRoot, updated, writes, changedFiles, err := prepareCloseChange(v, index, loaded, changeID, options)
	if err != nil {
		return nil, err
	}
	if isVerifiedAssuranceTier(loaded) {
		writes, changedFiles, err = appendCloseChangeReceiptWrite(writes, changedFiles, loaded.Resolution.ProjectRoot, changeID, updated, options)
		if err != nil {
			return nil, err
		}
	}
	if err := applyCloseChangeWrites(v, loaded, writes); err != nil {
		return nil, err
	}
	return buildCloseChangeResult(record, writableRoot, changeID, updated, changedFiles), nil
}

func appendCloseChangeReceiptWrite(rewrites []fileRewrite, changedFiles []FileMutation, projectRoot, changeID string, updated map[string]any, options ChangeCloseOptions) ([]fileRewrite, []FileMutation, error) {
	createdAt := options.ClosedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC().Truncate(time.Second)
	}
	value := map[string]any{
		"receipt_family":      assuranceReceiptFamilyChanges,
		"change_id":           changeID,
		"change_status":       fmt.Sprint(updated["status"]),
		"verification_status": fmt.Sprint(updated["verification_status"]),
	}
	return appendCapturedVerifiedReceiptRewrite(
		rewrites,
		changedFiles,
		projectRoot,
		assuranceReceiptFamilyChanges,
		"changes/"+changeID,
		value,
		createdAt.Unix(),
	)
}

func prepareCloseChange(v *Validator, index *ProjectIndex, loaded *LoadedProject, changeID string, options ChangeCloseOptions) (*ChangeRecord, string, map[string]any, []fileRewrite, []FileMutation, error) {
	record := index.Changes[changeID]
	if record == nil {
		return nil, "", nil, nil, nil, fmt.Errorf("change %q does not exist", changeID)
	}
	if err := validateCloseVerificationStatus(record.VerificationStatus, options.VerificationStatus); err != nil {
		return nil, "", nil, nil, nil, err
	}
	if err := validateCloseSuccessors(index, changeID, options.SupersededBy); err != nil {
		return nil, "", nil, nil, nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}
	updated, err := buildClosedStatusMap(index, record, options)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}
	writes, changedFiles, err := buildCloseStatusWrites(v, index, writableRoot, record, changeID, updated, options.SupersededBy)
	if err != nil {
		return nil, "", nil, nil, nil, err
	}
	return record, writableRoot, updated, writes, changedFiles, nil
}

func validateCloseSuccessors(index *ProjectIndex, changeID string, successorIDs []string) error {
	for _, successorID := range successorIDs {
		if successorID == changeID {
			return fmt.Errorf("superseded_by must not reference the change itself")
		}
		if _, ok := index.Changes[successorID]; !ok {
			return fmt.Errorf("superseded_by references missing change %q", successorID)
		}
	}
	return nil
}

func buildClosedStatusMap(index *ProjectIndex, record *ChangeRecord, options ChangeCloseOptions) (map[string]any, error) {
	updated := cloneMap(index.StatusFiles[record.StatusPath].Data)
	if strings.TrimSpace(options.VerificationStatus) != "" {
		updated["verification_status"] = strings.TrimSpace(options.VerificationStatus)
	}
	closed, err := CloseChangeStatus(updated, CloseChangeOptions{ClosedAt: options.ClosedAt, SupersededBy: options.SupersededBy})
	if err != nil {
		return nil, err
	}
	applyClosePromotionAssessment(closed, record)
	return closed, nil
}

func buildCloseStatusWrites(v *Validator, index *ProjectIndex, writableRoot string, record *ChangeRecord, changeID string, updated map[string]any, successorIDs []string) ([]fileRewrite, []FileMutation, error) {
	mainWrite, changedFiles, err := buildPrimaryCloseStatusWrite(v, writableRoot, record, updated)
	if err != nil {
		return nil, nil, err
	}
	writes := []fileRewrite{mainWrite}
	successorWrites, successorFiles, err := buildSuccessorCloseStatusWrites(v, index, writableRoot, changeID, successorIDs)
	if err != nil {
		return nil, nil, err
	}
	writes = append(writes, successorWrites...)
	changedFiles = append(changedFiles, successorFiles...)
	return writes, changedFiles, nil
}

func buildPrimaryCloseStatusWrite(v *Validator, writableRoot string, record *ChangeRecord, updated map[string]any) (fileRewrite, []FileMutation, error) {
	if err := ensurePathAndParentAreNotSymlinks(record.StatusPath); err != nil {
		return fileRewrite{}, nil, err
	}
	data, err := prepareStatusRewrite(v, record.StatusPath, updated)
	if err != nil {
		return fileRewrite{}, nil, err
	}
	changed := []FileMutation{{Path: runeContextRelativePath(writableRoot, record.StatusPath), Action: "updated"}}
	return fileRewrite{Path: record.StatusPath, Data: data}, changed, nil
}

func buildSuccessorCloseStatusWrites(v *Validator, index *ProjectIndex, writableRoot, changeID string, successorIDs []string) ([]fileRewrite, []FileMutation, error) {
	writes := make([]fileRewrite, 0, len(successorIDs))
	changedFiles := make([]FileMutation, 0, len(successorIDs))
	for _, successorID := range successorIDs {
		write, changed, err := reciprocalSupersedesWrite(v, index, writableRoot, changeID, successorID)
		if err != nil {
			return nil, nil, err
		}
		if changed {
			writes = append(writes, write)
			changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, write.Path), Action: "updated"})
		}
	}
	return writes, changedFiles, nil
}

func reciprocalSupersedesWrite(v *Validator, index *ProjectIndex, writableRoot, changeID, successorID string) (fileRewrite, bool, error) {
	successor := index.Changes[successorID]
	successorStatus := cloneMap(index.StatusFiles[successor.StatusPath].Data)
	supersedes := extractStringList(successorStatus["supersedes"])
	if containsString(supersedes, changeID) {
		return fileRewrite{}, false, nil
	}
	if isTerminalLifecycleStatus(successor.Status) {
		return fileRewrite{}, false, fmt.Errorf("successor change %q is already in terminal status %q and cannot be updated with a reciprocal supersedes link", successorID, successor.Status)
	}
	if err := ensurePathAndParentAreNotSymlinks(successor.StatusPath); err != nil {
		return fileRewrite{}, false, err
	}
	successorStatus["supersedes"] = stringSliceToAny(uniqueSortedStrings(append(supersedes, changeID)))
	data, err := prepareStatusRewrite(v, successor.StatusPath, successorStatus)
	if err != nil {
		return fileRewrite{}, false, err
	}
	return fileRewrite{Path: successor.StatusPath, Data: data}, true, nil
}

func applyCloseChangeWrites(v *Validator, loaded *LoadedProject, writes []fileRewrite) error {
	return applyFileRewritesTransaction(writes, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	})
}

func buildCloseChangeResult(record *ChangeRecord, writableRoot, changeID string, updated map[string]any, changedFiles []FileMutation) *ChangeOperationResult {
	changeDir := filepath.Join(writableRoot, "changes", changeID)
	sortFileMutations(changedFiles)
	closedAt, _ := updated["closed_at"].(string)
	status := fmt.Sprint(updated["status"])
	promotionStatus, promotionTargets := closePromotionAssessmentDetails(updated)
	return &ChangeOperationResult{
		ID:                        changeID,
		ChangePath:                runeContextRelativePath(writableRoot, changeDir),
		Mode:                      inferChangeMode(changeDir),
		RecommendedMode:           inferChangeMode(changeDir),
		Status:                    status,
		ClosedAt:                  closedAt,
		ContextBundles:            append([]string(nil), record.ContextBundles...),
		ApplicableStandards:       append([]string(nil), record.ApplicableStandards...),
		ChangedFiles:              changedFiles,
		ReviewDiffRequired:        false,
		PromotionAssessmentStatus: promotionStatus,
		SuggestedPromotionTargets: promotionTargets,
	}
}
