package contracts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type reallocationPlan struct {
	record       *ChangeRecord
	writableRoot string
	changeID     string
	newID        string
	oldChangeDir string
	newChangeDir string
	backupDir    string
	stagedDir    string
}

func ReallocateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeReallocateOptions) (*ChangeReallocationResult, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	plan, err := prepareReallocationPlan(index, loaded, changeID, options)
	if err != nil {
		return nil, err
	}
	changedFiles, rewrittenRefs, err := executeChangeReallocation(v, loaded, plan)
	if err != nil {
		return nil, err
	}
	return finalizeChangeReallocation(plan, changedFiles, rewrittenRefs), nil
}

func prepareReallocationPlan(index *ProjectIndex, loaded *LoadedProject, changeID string, options ChangeReallocateOptions) (reallocationPlan, error) {
	record, writableRoot, err := validateReallocationRecord(index, loaded, changeID)
	if err != nil {
		return reallocationPlan{}, err
	}
	newID, err := allocateReallocatedChangeID(writableRoot, changeID, record.Title, options.Entropy)
	if err != nil {
		return reallocationPlan{}, err
	}
	plan := buildReallocationPlan(writableRoot, changeID, newID, record)
	if err := validateReallocationPaths(plan); err != nil {
		return reallocationPlan{}, err
	}
	return plan, nil
}

func validateReallocationRecord(index *ProjectIndex, loaded *LoadedProject, changeID string) (*ChangeRecord, string, error) {
	record := index.Changes[changeID]
	if record == nil {
		return nil, "", fmt.Errorf("change %q does not exist", changeID)
	}
	if isTerminalLifecycleStatus(record.Status) {
		return nil, "", fmt.Errorf("change %q is already in terminal status %q and cannot be reallocated", changeID, record.Status)
	}
	if err := ensureChangeReallocationIsLocalOnly(index, changeID); err != nil {
		return nil, "", err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, "", err
	}
	return record, writableRoot, nil
}

func allocateReallocatedChangeID(writableRoot, changeID, title string, entropy io.Reader) (string, error) {
	originalYear, _, _, _, err := parseChangeID(changeID)
	if err != nil {
		return "", err
	}
	allocationTime := time.Date(originalYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	return AllocateChangeID(writableRoot, allocationTime, title, entropy)
}

func buildReallocationPlan(writableRoot, changeID, newID string, record *ChangeRecord) reallocationPlan {
	changesRoot := filepath.Join(writableRoot, "changes")
	return reallocationPlan{
		record:       record,
		writableRoot: writableRoot,
		changeID:     changeID,
		newID:        newID,
		oldChangeDir: filepath.Join(changesRoot, changeID),
		newChangeDir: filepath.Join(changesRoot, newID),
		backupDir:    filepath.Join(writableRoot, ".reallocate-"+changeID+"-backup"),
		stagedDir:    "",
	}
}

func validateReallocationPaths(plan reallocationPlan) error {
	for _, path := range []string{filepath.Join(plan.writableRoot, "changes"), plan.oldChangeDir, plan.newChangeDir, plan.backupDir} {
		if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
			return err
		}
	}
	if err := ensureMissingReallocationPath(plan.newChangeDir, "reallocated change path"); err != nil {
		return err
	}
	return ensureMissingReallocationBackup(plan.writableRoot, plan.changeID, plan.backupDir)
}

func ensureMissingReallocationPath(path, description string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s %q already exists", description, runeContextRelativePath(filepath.Dir(filepath.Dir(path)), path))
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ensureMissingReallocationBackup(writableRoot, changeID, backupDir string) error {
	if _, err := os.Lstat(backupDir); err == nil {
		return fmt.Errorf("cannot reallocate change %q because backup path %q already exists; inspect or remove the leftover backup before retrying", changeID, runeContextRelativePath(writableRoot, backupDir))
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func executeChangeReallocation(v *Validator, loaded *LoadedProject, plan reallocationPlan) ([]FileMutation, int, error) {
	stagedDir, changedFiles, rewrittenRefs, err := stageChangeReallocation(plan)
	if err != nil {
		return nil, 0, err
	}
	plan.stagedDir = stagedDir
	if err := commitReallocatedChange(v, loaded, plan); err != nil {
		return nil, 0, err
	}
	return changedFiles, rewrittenRefs, nil
}

func stageChangeReallocation(plan reallocationPlan) (string, []FileMutation, int, error) {
	stagedDir, err := mkdirTempDir(plan.writableRoot, ".reallocate-"+plan.newID+"-stage-")
	if err != nil {
		return "", nil, 0, err
	}
	if err := ensurePathAndParentAreNotSymlinks(stagedDir); err != nil {
		_ = removeAllPath(stagedDir)
		return "", nil, 0, err
	}
	changedFiles, rewrittenRefs, err := stageReallocatedChange(plan.oldChangeDir, stagedDir, plan.changeID, plan.newID)
	if err != nil {
		_ = removeAllPath(stagedDir)
		return "", nil, 0, err
	}
	return stagedDir, changedFiles, rewrittenRefs, nil
}

func commitReallocatedChange(v *Validator, loaded *LoadedProject, plan reallocationPlan) error {
	defer func() { _ = removeAllPath(plan.stagedDir) }()
	if err := renamePath(plan.oldChangeDir, plan.backupDir); err != nil {
		return err
	}
	if err := renamePath(plan.stagedDir, plan.newChangeDir); err != nil {
		rollbackErr := restoreOriginalChangeFromBackup(plan.backupDir, plan.oldChangeDir)
		return combineReallocationRollbackError(err, rollbackErr)
	}
	if err := validateChangeMutation(v, loaded.Resolution.ProjectRoot); err != nil {
		rollbackErr := rollbackCommittedReallocatedChange(plan.newChangeDir, plan.backupDir, plan.oldChangeDir)
		return combineReallocationRollbackError(err, rollbackErr)
	}
	return nil
}

func finalizeChangeReallocation(plan reallocationPlan, changedFiles []FileMutation, rewrittenRefs int) *ChangeReallocationResult {
	warnings := cleanupReallocationBackup(plan.writableRoot, plan.backupDir)
	sortFileMutations(changedFiles)
	return &ChangeReallocationResult{
		OldID:                   plan.changeID,
		ID:                      plan.newID,
		OldChangePath:           runeContextRelativePath(plan.writableRoot, plan.oldChangeDir),
		ChangePath:              runeContextRelativePath(plan.writableRoot, plan.newChangeDir),
		RewrittenReferenceCount: rewrittenRefs,
		ChangedFiles:            changedFiles,
		Warnings:                warnings,
	}
}

func cleanupReallocationBackup(writableRoot, backupDir string) []string {
	if err := removeAllPath(backupDir); err != nil {
		warning := fmt.Sprintf("reallocation succeeded but could not remove backup path %q: %v", runeContextRelativePath(writableRoot, backupDir), err)
		return []string{warning}
	}
	return nil
}
