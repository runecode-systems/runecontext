package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ShapeChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeShapeOptions) (*ChangeOperationResult, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	record, assessment, writableRoot, changeDir, err := prepareShapeChange(index, loaded, changeID)
	if err != nil {
		return nil, err
	}
	changedFiles, err := materializeShapeFiles(changeDir, writableRoot, loaded.Resolution.ProjectRoot, record.Title, assessment, options)
	if err != nil {
		return nil, err
	}
	result, err := refreshShapedChangeStandards(v, loaded, index, record, writableRoot, changeDir, assessment, changedFiles)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func prepareShapeChange(index *ProjectIndex, loaded *LoadedProject, changeID string) (*ChangeRecord, changeIntakeAssessment, string, string, error) {
	record := index.Changes[changeID]
	if record == nil {
		return nil, changeIntakeAssessment{}, "", "", fmt.Errorf("change %q does not exist", changeID)
	}
	if isTerminalLifecycleStatus(record.Status) {
		return nil, changeIntakeAssessment{}, "", "", fmt.Errorf("change %q is already in terminal status %q and cannot be shaped", changeID, record.Status)
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, changeIntakeAssessment{}, "", "", err
	}
	assessment := assessShapeChange(record)
	return record, assessment, writableRoot, filepath.Join(writableRoot, "changes", changeID), nil
}

func assessShapeChange(record *ChangeRecord) changeIntakeAssessment {
	assessment := assessChangeIntake(record.Title, record.Type, record.Size, "")
	if assessment.RecommendedMode == ChangeModeMinimum {
		assessment.Reasons = append([]string{"Full mode was requested explicitly to deepen the change."}, assessment.Reasons...)
	}
	return assessment
}

func refreshShapedChangeStandards(v *Validator, loaded *LoadedProject, index *ProjectIndex, record *ChangeRecord, writableRoot, changeDir string, assessment changeIntakeAssessment, changedFiles []FileMutation) (*ChangeOperationResult, error) {
	applicableStandards, standardAssumptions, addedStandards, standardsAction, changedFiles, err := updateShapedChangeStandards(v, index, record, writableRoot, changeDir, changedFiles)
	if err != nil {
		return nil, err
	}
	if err := validateChangeMutation(v, loaded.Resolution.ProjectRoot); err != nil {
		return nil, err
	}
	sortFileMutations(changedFiles)
	return &ChangeOperationResult{
		ID:                     record.ID,
		ChangePath:             runeContextRelativePath(writableRoot, changeDir),
		Mode:                   ChangeModeFull,
		RecommendedMode:        assessment.RecommendedMode,
		Status:                 string(record.Status),
		ContextBundles:         append([]string(nil), record.ContextBundles...),
		ApplicableStandards:    append([]string(nil), applicableStandards...),
		AddedStandards:         append([]string(nil), addedStandards...),
		ChangedFiles:           changedFiles,
		StandardsRefreshAction: standardsAction,
		ReviewDiffRequired:     standardsAction != "unchanged",
		Reasons:                append([]string(nil), assessment.Reasons...),
		Assumptions:            append([]string(nil), standardAssumptions...),
	}, nil
}

func updateShapedChangeStandards(v *Validator, index *ProjectIndex, record *ChangeRecord, writableRoot, changeDir string, changedFiles []FileMutation) ([]string, []string, []string, string, []FileMutation, error) {
	applicableStandards, standardAssumptions, err := resolveApplicableStandards(index, record.ContextBundles)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}
	standardsPath := filepath.Join(changeDir, "standards.md")
	previousData, err := os.ReadFile(standardsPath)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}
	addedStandards := sliceDifference(applicableStandards, record.ApplicableStandards)
	updated, err := refreshStandardsMarkdown(v, standardsPath, previousData, applicableStandards, addedStandards)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}
	if updated.changed {
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, standardsPath), Action: "updated"})
	}
	return applicableStandards, standardAssumptions, addedStandards, updated.action, changedFiles, nil
}

type standardsRefreshResult struct {
	action  string
	changed bool
}

func refreshStandardsMarkdown(v *Validator, standardsPath string, previousData []byte, applicable, added []string) (standardsRefreshResult, error) {
	preservedSections, err := preservedStandardsSections(previousData)
	if err != nil {
		return standardsRefreshResult{}, err
	}
	newStandards := renderStandardsMarkdown(previousData, applicable, added, preservedSections, false)
	if string(newStandards) == strings.ReplaceAll(string(previousData), "\r\n", "\n") {
		return standardsRefreshResult{action: "unchanged"}, nil
	}
	if err := v.ValidateStandardsMarkdown(standardsPath, newStandards); err != nil {
		return standardsRefreshResult{}, err
	}
	if err := os.WriteFile(standardsPath, newStandards, 0o644); err != nil {
		return standardsRefreshResult{}, err
	}
	return standardsRefreshResult{action: "updated", changed: true}, nil
}
