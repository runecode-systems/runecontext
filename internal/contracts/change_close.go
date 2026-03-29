package contracts

import (
	"fmt"
	"path/filepath"
	"sort"
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
	record, writableRoot, updated, writes, changedFiles, recursiveTargetIDs, err := prepareCloseChange(v, index, loaded, changeID, options)
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
	return buildCloseChangeResult(record, writableRoot, changeID, updated, changedFiles, options.Recursive, recursiveTargetIDs), nil
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

func prepareCloseChange(v *Validator, index *ProjectIndex, loaded *LoadedProject, changeID string, options ChangeCloseOptions) (*ChangeRecord, string, map[string]any, []fileRewrite, []FileMutation, []string, error) {
	record := index.Changes[changeID]
	if record == nil {
		return nil, "", nil, nil, nil, nil, fmt.Errorf("change %q does not exist", changeID)
	}
	targets, err := resolveCloseTargets(index, record, options.Recursive)
	if err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	if err := validateCloseSuccessors(index, changeID, options.SupersededBy, targets); err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	if err := validateCloseTargetsVerificationStatus(targets, options.VerificationStatus); err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	updated, err := buildClosedStatusMap(index, record, options)
	if err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	writes, changedFiles, err := buildCloseStatusWrites(v, index, writableRoot, targets, changeID, options, updated)
	if err != nil {
		return nil, "", nil, nil, nil, nil, err
	}
	recursiveTargetIDs := collectCloseRecursiveTargetIDs(targets, changeID)
	return record, writableRoot, updated, writes, changedFiles, recursiveTargetIDs, nil
}

func resolveCloseTargets(index *ProjectIndex, record *ChangeRecord, recursive bool) ([]*ChangeRecord, error) {
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

func validateCloseTargetsVerificationStatus(targets []*ChangeRecord, requested string) error {
	for _, target := range targets {
		if err := validateCloseVerificationStatus(target.VerificationStatus, requested); err != nil {
			return fmt.Errorf("change %q: %w", target.ID, err)
		}
	}
	return nil
}

func collectCloseRecursiveTargetIDs(targets []*ChangeRecord, rootID string) []string {
	recursiveTargetIDs := make([]string, 0, len(targets)-1)
	for _, target := range targets {
		if target.ID != rootID {
			recursiveTargetIDs = append(recursiveTargetIDs, target.ID)
		}
	}
	sort.Strings(recursiveTargetIDs)
	return recursiveTargetIDs
}

func validateCloseSuccessors(index *ProjectIndex, changeID string, successorIDs []string, targets []*ChangeRecord) error {
	// Build a quick set of recursive targets to detect unsafe self-supersession.
	targetSet := make(map[string]bool, len(targets))
	for _, t := range targets {
		targetSet[t.ID] = true
	}
	for _, successorID := range successorIDs {
		if successorID == changeID {
			return fmt.Errorf("superseded_by must not reference the change itself")
		}
		if _, ok := index.Changes[successorID]; !ok {
			return fmt.Errorf("superseded_by references missing change %q", successorID)
		}
		if targetSet[successorID] {
			return fmt.Errorf("superseded_by references change %q which is also a recursive close target; this would create a self-supersession", successorID)
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

func buildCloseStatusWrites(v *Validator, index *ProjectIndex, writableRoot string, targets []*ChangeRecord, changeID string, options ChangeCloseOptions, rootUpdated map[string]any) ([]fileRewrite, []FileMutation, error) {
	writes := make([]fileRewrite, 0, len(targets)+len(options.SupersededBy))
	changedFiles := make([]FileMutation, 0, len(targets)+len(options.SupersededBy))
	supersededTargetIDs := make([]string, 0, len(targets))
	for _, target := range targets {
		supersededTargetIDs = append(supersededTargetIDs, target.ID)
		updated := rootUpdated
		if target.ID != changeID {
			var err error
			updated, err = buildClosedStatusMap(index, target, options)
			if err != nil {
				return nil, nil, fmt.Errorf("change %q: %w", target.ID, err)
			}
		}
		write, _, err := buildPrimaryCloseStatusWrite(v, writableRoot, target, updated)
		if err != nil {
			return nil, nil, err
		}
		writes = append(writes, write)
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, target.StatusPath), Action: "updated"})
	}
	successorWrites, successorFiles, err := buildSuccessorCloseStatusWrites(v, index, writableRoot, supersededTargetIDs, options.SupersededBy)
	if err != nil {
		return nil, nil, err
	}
	writes = append(writes, successorWrites...)
	changedFiles = append(changedFiles, successorFiles...)
	sortFileMutations(changedFiles)
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

func buildSuccessorCloseStatusWrites(v *Validator, index *ProjectIndex, writableRoot string, supersededTargetIDs, successorIDs []string) ([]fileRewrite, []FileMutation, error) {
	writes := make([]fileRewrite, 0, len(successorIDs))
	changedFiles := make([]FileMutation, 0, len(successorIDs))
	for _, successorID := range successorIDs {
		write, changed, err := reciprocalSupersedesWrite(v, index, writableRoot, supersededTargetIDs, successorID)
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

func reciprocalSupersedesWrite(v *Validator, index *ProjectIndex, writableRoot string, supersededTargetIDs []string, successorID string) (fileRewrite, bool, error) {
	successor := index.Changes[successorID]
	successorStatus := cloneMap(index.StatusFiles[successor.StatusPath].Data)
	supersedes := extractStringList(successorStatus["supersedes"])
	mergedSupersedes := uniqueSortedStrings(append(append([]string(nil), supersedes...), supersededTargetIDs...))
	if len(mergedSupersedes) == len(supersedes) {
		return fileRewrite{}, false, nil
	}
	if isTerminalLifecycleStatus(successor.Status) {
		return fileRewrite{}, false, fmt.Errorf("successor change %q is already in terminal status %q and cannot be updated with a reciprocal supersedes link", successorID, successor.Status)
	}
	if err := ensurePathAndParentAreNotSymlinks(successor.StatusPath); err != nil {
		return fileRewrite{}, false, err
	}
	successorStatus["supersedes"] = stringSliceToAny(mergedSupersedes)
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

func buildCloseChangeResult(record *ChangeRecord, writableRoot, changeID string, updated map[string]any, changedFiles []FileMutation, recursive bool, recursiveTargetIDs []string) *ChangeOperationResult {
	changeDir := filepath.Join(writableRoot, "changes", changeID)
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
		Recursive:                 recursive,
		RecursiveTargetCount:      len(recursiveTargetIDs),
		RecursiveTargetIDs:        append([]string(nil), recursiveTargetIDs...),
	}
}
