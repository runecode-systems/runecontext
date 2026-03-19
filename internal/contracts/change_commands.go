package contracts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type ChangeMode string

const (
	ChangeModeMinimum ChangeMode = "minimum"
	ChangeModeFull    ChangeMode = "full"
)

type ChangeCreateOptions struct {
	Title          string
	Type           string
	Size           string
	Description    string
	ContextBundles []string
	RequestedMode  ChangeMode
	Now            time.Time
	Entropy        io.Reader
	Design         string
	Verification   string
	Tasks          []string
	References     []string
}

type ChangeShapeOptions struct {
	Design       string
	Verification string
	Tasks        []string
	References   []string
}

type ChangeCloseOptions struct {
	VerificationStatus string
	ClosedAt           time.Time
	SupersededBy       []string
}

type ChangeReallocateOptions struct {
	Entropy io.Reader
}

type FileMutation struct {
	Path   string
	Action string
}

type ChangeOperationResult struct {
	ID                     string
	ChangePath             string
	Mode                   ChangeMode
	RecommendedMode        ChangeMode
	Status                 string
	ClosedAt               string
	ContextBundles         []string
	ApplicableStandards    []string
	AddedStandards         []string
	ChangedFiles           []FileMutation
	StandardsRefreshAction string
	ReviewDiffRequired     bool
	Reasons                []string
	Assumptions            []string
}

type ChangeReallocationResult struct {
	OldID                   string
	ID                      string
	OldChangePath           string
	ChangePath              string
	RewrittenReferenceCount int
	ChangedFiles            []FileMutation
	Warnings                []string
}

type ProjectStatusSummary struct {
	Root               string
	SelectedConfigPath string
	RuneContextVersion string
	AssuranceTier      string
	Active             []ChangeStatusEntry
	Closed             []ChangeStatusEntry
	Superseded         []ChangeStatusEntry
	BundleIDs          []string
}

type ChangeStatusEntry struct {
	ID     string
	Title  string
	Status string
	Type   string
	Size   string
	Path   string
}

type changeIntakeAssessment struct {
	Size             string
	RecommendedMode  ChangeMode
	Reasons          []string
	Assumptions      []string
	ChecklistTitle   string
	ChecklistItems   []string
	FollowUpPrompts  []string
	VerificationCmds []string
	VerificationNote string
}

type promotionAssessmentDocument struct {
	Status           string                    `yaml:"status"`
	SuggestedTargets []promotionTargetDocument `yaml:"suggested_targets"`
}

type promotionTargetDocument struct {
	TargetType string `yaml:"target_type"`
	TargetPath string `yaml:"target_path"`
	Summary    string `yaml:"summary"`
}

type statusDocument struct {
	SchemaVersion       int                         `yaml:"schema_version"`
	ID                  string                      `yaml:"id"`
	Title               string                      `yaml:"title"`
	Status              string                      `yaml:"status"`
	Type                string                      `yaml:"type"`
	Size                string                      `yaml:"size,omitempty"`
	VerificationStatus  string                      `yaml:"verification_status"`
	ContextBundles      []string                    `yaml:"context_bundles"`
	RelatedSpecs        []string                    `yaml:"related_specs"`
	RelatedDecisions    []string                    `yaml:"related_decisions"`
	RelatedChanges      []string                    `yaml:"related_changes"`
	DependsOn           []string                    `yaml:"depends_on"`
	InformedBy          []string                    `yaml:"informed_by"`
	Supersedes          []string                    `yaml:"supersedes"`
	SupersededBy        []string                    `yaml:"superseded_by"`
	CreatedAt           string                      `yaml:"created_at,omitempty"`
	ClosedAt            any                         `yaml:"closed_at"`
	PromotionAssessment promotionAssessmentDocument `yaml:"promotion_assessment"`
	Extensions          map[string]any              `yaml:"extensions,omitempty"`
}

const maxCreateChangeDirAttempts = 8

var justfileTestTargetPattern = regexp.MustCompile(`(?m)^test\s*:`)

var allowedPromotionAssessmentStatuses = map[string]struct{}{
	"pending":   {},
	"none":      {},
	"suggested": {},
	"accepted":  {},
	"completed": {},
}

var (
	renamePath                         = os.Rename
	removeAllPath                      = os.RemoveAll
	mkdirTempDir                       = os.MkdirTemp
	createTempFilePath                 = os.CreateTemp
	writeFilePath                      = os.WriteFile
	chmodPath                          = os.Chmod
	lstatPath                          = os.Lstat
	atomicReplaceNeedsFallback         = runtime.GOOS == "windows"
	validateProjectAfterChangeMutation = func(v *Validator, projectRoot string) (*ProjectIndex, error) {
		return v.ValidateProjectWithOptions(projectRoot, ResolveOptions{
			ConfigDiscovery: ConfigDiscoveryExplicitRoot,
			ExecutionMode:   ExecutionModeLocal,
		})
	}
)

func CreateChange(v *Validator, loaded *LoadedProject, options ChangeCreateOptions) (result *ChangeOperationResult, err error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	if strings.TrimSpace(options.Title) == "" {
		return nil, fmt.Errorf("change title is required")
	}
	changeType := strings.TrimSpace(options.Type)
	if err := validateChangeTypeValue(changeType); err != nil {
		return nil, err
	}
	if err := validateRequestedMode(options.RequestedMode); err != nil {
		return nil, err
	}
	if err := requireWritableChangeSource(loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	assessment := assessChangeIntake(options.Title, changeType, options.Size, options.Description)
	selectedMode := assessment.RecommendedMode
	if options.RequestedMode != "" {
		selectedMode = options.RequestedMode
	}
	contextBundles, contextAssumptions, err := resolveContextBundlesForChange(index, options.ContextBundles)
	if err != nil {
		return nil, err
	}
	applicableStandards, standardAssumptions, err := resolveApplicableStandards(index, contextBundles)
	if err != nil {
		return nil, err
	}
	now := options.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	changesRoot := filepath.Join(writableRoot, "changes")
	if err := ensurePathAndParentAreNotSymlinks(changesRoot); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(changesRoot, 0o755); err != nil {
		return nil, err
	}
	id, changeDir, err := createUniqueChangeDir(writableRoot, now, options.Title, options.Entropy)
	if err != nil {
		return nil, err
	}
	if err := ensurePathAndParentAreNotSymlinks(changeDir); err != nil {
		_ = removeAllPath(changeDir)
		return nil, err
	}
	cleanupChangeDir := true
	defer func() {
		if !cleanupChangeDir {
			return
		}
		if cleanupErr := removeAllPath(changeDir); cleanupErr != nil {
			if err == nil {
				err = cleanupErr
				return
			}
			err = fmt.Errorf("%v; cleanup also failed and manual removal may be required: %v", err, cleanupErr)
		}
	}()
	verificationCommands, verificationAssumption := inferVerificationCommands(loaded.Resolution.ProjectRoot)
	assessment.VerificationCmds = verificationCommands
	assumptions := uniqueStringsInOrder(append(append(append([]string{}, assessment.Assumptions...), contextAssumptions...), standardAssumptions...))
	if verificationAssumption != "" {
		assumptions = append(assumptions, verificationAssumption)
	}
	statusRaw := newStatusMap(id, options.Title, changeType, assessment.Size, contextBundles, now)
	statusData, err := renderStatusYAML(statusRaw)
	if err != nil {
		return nil, err
	}
	statusPath := filepath.Join(changeDir, "status.yaml")
	if err := v.ValidateYAMLFile("change-status.schema.json", statusPath, statusData); err != nil {
		return nil, err
	}
	proposalData := renderProposalMarkdown(options.Title, options.Description, selectedMode, assessment.Reasons, assumptions)
	proposalPath := filepath.Join(changeDir, "proposal.md")
	if err := v.ValidateProposalMarkdown(proposalPath, proposalData); err != nil {
		return nil, err
	}
	standardsData := renderStandardsMarkdown(nil, applicableStandards, nil, nil, true)
	standardsPath := filepath.Join(changeDir, "standards.md")
	if err := v.ValidateStandardsMarkdown(standardsPath, standardsData); err != nil {
		return nil, err
	}
	changedFiles := make([]FileMutation, 0, 5)
	for _, file := range []struct {
		path string
		data []byte
	}{
		{path: statusPath, data: statusData},
		{path: proposalPath, data: proposalData},
		{path: standardsPath, data: standardsData},
	} {
		if err := writeFileAtomically(file.path, file.data, 0o644); err != nil {
			return nil, err
		}
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, file.path), Action: "created"})
	}
	if selectedMode == ChangeModeFull {
		shapeResult, err := materializeShapeFiles(changeDir, writableRoot, loaded.Resolution.ProjectRoot, options.Title, assessment, ChangeShapeOptions{
			Design:       options.Design,
			Verification: options.Verification,
			Tasks:        options.Tasks,
			References:   options.References,
		})
		if err != nil {
			return nil, err
		}
		changedFiles = append(changedFiles, shapeResult...)
	}
	sortFileMutations(changedFiles)
	validated, err := validateProjectAfterChangeMutation(v, loaded.Resolution.ProjectRoot)
	if err != nil {
		return nil, err
	}
	_ = validated.Close()
	cleanupChangeDir = false
	result = &ChangeOperationResult{
		ID:                     id,
		ChangePath:             runeContextRelativePath(writableRoot, changeDir),
		Mode:                   selectedMode,
		RecommendedMode:        assessment.RecommendedMode,
		Status:                 string(StatusProposed),
		ContextBundles:         append([]string(nil), contextBundles...),
		ApplicableStandards:    append([]string(nil), applicableStandards...),
		ChangedFiles:           changedFiles,
		StandardsRefreshAction: "created",
		ReviewDiffRequired:     true,
		Reasons:                append([]string(nil), assessment.Reasons...),
		Assumptions:            assumptions,
	}
	if selectedMode == ChangeModeFull {
		result.AddedStandards = nil
	}
	return result, nil
}

func ShapeChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeShapeOptions) (*ChangeOperationResult, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	if err := requireWritableChangeSource(loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	record := index.Changes[changeID]
	if record == nil {
		return nil, fmt.Errorf("change %q does not exist", changeID)
	}
	if isTerminalLifecycleStatus(record.Status) {
		return nil, fmt.Errorf("change %q is already in terminal status %q and cannot be shaped", changeID, record.Status)
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	assessment := assessChangeIntake(record.Title, record.Type, record.Size, "")
	if assessment.RecommendedMode == ChangeModeMinimum {
		assessment.Reasons = append([]string{"Full mode was requested explicitly to deepen the change."}, assessment.Reasons...)
	}
	changeDir := filepath.Join(writableRoot, "changes", changeID)
	changedFiles, err := materializeShapeFiles(changeDir, writableRoot, loaded.Resolution.ProjectRoot, record.Title, assessment, options)
	if err != nil {
		return nil, err
	}
	applicableStandards, standardAssumptions, err := resolveApplicableStandards(index, record.ContextBundles)
	if err != nil {
		return nil, err
	}
	standardsPath := filepath.Join(changeDir, "standards.md")
	previousData, err := os.ReadFile(standardsPath)
	if err != nil {
		return nil, err
	}
	preservedSections, err := preservedStandardsSections(previousData)
	if err != nil {
		return nil, err
	}
	addedStandards := sliceDifference(applicableStandards, record.ApplicableStandards)
	newStandards := renderStandardsMarkdown(previousData, applicableStandards, addedStandards, preservedSections, false)
	standardsAction := "unchanged"
	if string(newStandards) != strings.ReplaceAll(string(previousData), "\r\n", "\n") {
		if err := v.ValidateStandardsMarkdown(standardsPath, newStandards); err != nil {
			return nil, err
		}
		if err := os.WriteFile(standardsPath, newStandards, 0o644); err != nil {
			return nil, err
		}
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, standardsPath), Action: "updated"})
		standardsAction = "updated"
	}
	sortFileMutations(changedFiles)
	validated, err := v.ValidateProjectWithOptions(loaded.Resolution.ProjectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		return nil, err
	}
	_ = validated.Close()
	return &ChangeOperationResult{
		ID:                     changeID,
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

func CloseChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeCloseOptions) (*ChangeOperationResult, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	if err := requireWritableChangeSource(loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	record := index.Changes[changeID]
	if record == nil {
		return nil, fmt.Errorf("change %q does not exist", changeID)
	}
	if err := validateCloseVerificationStatus(record.VerificationStatus, options.VerificationStatus); err != nil {
		return nil, err
	}
	for _, successorID := range options.SupersededBy {
		if _, ok := index.Changes[successorID]; !ok {
			return nil, fmt.Errorf("superseded_by references missing change %q", successorID)
		}
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	statusRecord := index.StatusFiles[record.StatusPath]
	updated := cloneMap(statusRecord.Data)
	if strings.TrimSpace(options.VerificationStatus) != "" {
		updated["verification_status"] = strings.TrimSpace(options.VerificationStatus)
	}
	updated, err = CloseChangeStatus(updated, CloseChangeOptions{ClosedAt: options.ClosedAt, SupersededBy: options.SupersededBy})
	if err != nil {
		return nil, err
	}
	statusWrites := make([]fileRewrite, 0, 1+len(options.SupersededBy))
	if err := ensurePathAndParentAreNotSymlinks(record.StatusPath); err != nil {
		return nil, err
	}
	mainStatusData, err := prepareStatusRewrite(v, record.StatusPath, updated)
	if err != nil {
		return nil, err
	}
	statusWrites = append(statusWrites, fileRewrite{Path: record.StatusPath, Data: mainStatusData})
	changedFiles := make([]FileMutation, 0, 1+len(options.SupersededBy))
	changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, record.StatusPath), Action: "updated"})
	for _, successorID := range options.SupersededBy {
		successor := index.Changes[successorID]
		successorStatus := cloneMap(index.StatusFiles[successor.StatusPath].Data)
		supersedes := extractStringList(successorStatus["supersedes"])
		if !containsString(supersedes, changeID) {
			if isTerminalLifecycleStatus(successor.Status) {
				return nil, fmt.Errorf("successor change %q is already in terminal status %q and cannot be updated with a reciprocal supersedes link", successorID, successor.Status)
			}
			if err := ensurePathAndParentAreNotSymlinks(successor.StatusPath); err != nil {
				return nil, err
			}
			supersedes = append(supersedes, changeID)
			successorStatus["supersedes"] = stringSliceToAny(uniqueSortedStrings(supersedes))
			successorStatusData, err := prepareStatusRewrite(v, successor.StatusPath, successorStatus)
			if err != nil {
				return nil, err
			}
			statusWrites = append(statusWrites, fileRewrite{Path: successor.StatusPath, Data: successorStatusData})
			changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, successor.StatusPath), Action: "updated"})
		}
	}
	if err := applyFileRewritesTransaction(statusWrites, func() error {
		validated, err := validateProjectAfterChangeMutation(v, loaded.Resolution.ProjectRoot)
		if err != nil {
			return err
		}
		_ = validated.Close()
		return nil
	}); err != nil {
		return nil, err
	}
	sortFileMutations(changedFiles)
	closedAt, _ := updated["closed_at"].(string)
	status := fmt.Sprint(updated["status"])
	return &ChangeOperationResult{
		ID:                  changeID,
		ChangePath:          runeContextRelativePath(writableRoot, filepath.Join(writableRoot, "changes", changeID)),
		Mode:                inferChangeMode(filepath.Join(writableRoot, "changes", changeID)),
		RecommendedMode:     inferChangeMode(filepath.Join(writableRoot, "changes", changeID)),
		Status:              status,
		ClosedAt:            closedAt,
		ContextBundles:      append([]string(nil), record.ContextBundles...),
		ApplicableStandards: append([]string(nil), record.ApplicableStandards...),
		ChangedFiles:        changedFiles,
		ReviewDiffRequired:  false,
	}, nil
}

type fileRewrite struct {
	Path string
	Data []byte
}

type fileBackup struct {
	Path string
	Data []byte
	Perm fs.FileMode
}

func prepareStatusRewrite(v *Validator, path string, raw map[string]any) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	data, err := renderStatusYAML(raw)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateYAMLFile("change-status.schema.json", path, data); err != nil {
		return nil, err
	}
	return data, nil
}

func applyFileRewritesTransaction(rewrites []fileRewrite, postWriteValidate func() error) error {
	if len(rewrites) == 0 {
		if postWriteValidate == nil {
			return nil
		}
		return postWriteValidate()
	}
	backups := make([]fileBackup, 0, len(rewrites))
	for _, rewrite := range rewrites {
		if err := ensurePathAndParentAreNotSymlinks(rewrite.Path); err != nil {
			return err
		}
		info, err := os.Stat(rewrite.Path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(rewrite.Path)
		if err != nil {
			return err
		}
		backups = append(backups, fileBackup{Path: rewrite.Path, Data: data, Perm: info.Mode().Perm()})
	}
	written := 0
	for _, rewrite := range rewrites {
		if err := writeFileAtomically(rewrite.Path, rewrite.Data, backups[written].Perm); err != nil {
			rollbackErr := restoreFileBackups(backups[:written])
			return combineFileRewriteRollbackError(err, rollbackErr)
		}
		written++
	}
	if postWriteValidate == nil {
		return nil
	}
	if err := postWriteValidate(); err != nil {
		rollbackErr := restoreFileBackups(backups)
		return combineFileRewriteRollbackError(err, rollbackErr)
	}
	return nil
}

func restoreFileBackups(backups []fileBackup) error {
	errMessages := make([]string, 0)
	for i := len(backups) - 1; i >= 0; i-- {
		backup := backups[i]
		if err := writeFileAtomically(backup.Path, backup.Data, backup.Perm); err != nil {
			errMessages = append(errMessages, fmt.Sprintf("restore %q: %v", filepath.ToSlash(backup.Path), err))
		}
	}
	if len(errMessages) == 0 {
		return nil
	}
	return errors.New(strings.Join(errMessages, "; "))
}

func combineFileRewriteRollbackError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%v; rollback also failed and manual recovery may be required: %v", operationErr, rollbackErr)
}

func ensurePathAndParentAreNotSymlinks(path string) error {
	cleanPath := filepath.Clean(path)
	for _, candidate := range []string{filepath.Dir(cleanPath), cleanPath} {
		info, err := lstatPath(candidate)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("mutation does not support symlinked targets: %s", filepath.ToSlash(candidate))
		}
	}
	return nil
}

func writeFileAtomically(path string, data []byte, perm fs.FileMode) error {
	if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
		return err
	}
	tempFile, err := createTempFilePath(filepath.Dir(path), ".mutation-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = removeAllPath(tempPath)
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = removeAllPath(tempPath)
		}
	}()
	if err := writeFilePath(tempPath, data, perm); err != nil {
		return err
	}
	if err := chmodPath(tempPath, perm); err != nil {
		return err
	}
	if err := replacePathAtomically(tempPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func replacePathAtomically(tempPath, targetPath string) error {
	err := renamePath(tempPath, targetPath)
	if err == nil {
		return nil
	}
	if !atomicReplaceNeedsFallback {
		return err
	}
	if fallbackErr := ensurePathAndParentAreNotSymlinks(targetPath); fallbackErr != nil {
		return fallbackErr
	}
	if _, statErr := os.Stat(targetPath); statErr != nil {
		return err
	}
	backupFile, backupErr := createTempFilePath(filepath.Dir(targetPath), ".replace-backup-*")
	if backupErr != nil {
		return err
	}
	backupPath := backupFile.Name()
	if closeErr := backupFile.Close(); closeErr != nil {
		_ = removeAllPath(backupPath)
		return closeErr
	}
	if removeErr := removeAllPath(backupPath); removeErr != nil {
		return removeErr
	}
	cleanupBackup := true
	defer func() {
		if cleanupBackup {
			_ = removeAllPath(backupPath)
		}
	}()
	if renameBackupErr := renamePath(targetPath, backupPath); renameBackupErr != nil {
		return err
	}
	if renameTempErr := renamePath(tempPath, targetPath); renameTempErr != nil {
		rollbackErr := renamePath(backupPath, targetPath)
		return combineFileRewriteRollbackError(renameTempErr, rollbackErr)
	}
	cleanupBackup = false
	_ = removeAllPath(backupPath)
	return nil
}

func ReallocateChange(v *Validator, loaded *LoadedProject, changeID string, options ChangeReallocateOptions) (*ChangeReallocationResult, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	if err := requireWritableChangeSource(loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	record := index.Changes[changeID]
	if record == nil {
		return nil, fmt.Errorf("change %q does not exist", changeID)
	}
	if isTerminalLifecycleStatus(record.Status) {
		return nil, fmt.Errorf("change %q is already in terminal status %q and cannot be reallocated", changeID, record.Status)
	}
	if err := ensureChangeReallocationIsLocalOnly(index, changeID); err != nil {
		return nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	originalYear, _, _, _, err := parseChangeID(changeID)
	if err != nil {
		return nil, err
	}
	allocationTime := time.Date(originalYear, time.January, 1, 0, 0, 0, 0, time.UTC)
	newID, err := AllocateChangeID(writableRoot, allocationTime, record.Title, options.Entropy)
	if err != nil {
		return nil, err
	}
	changesRoot := filepath.Join(writableRoot, "changes")
	oldChangeDir := filepath.Join(changesRoot, changeID)
	newChangeDir := filepath.Join(changesRoot, newID)
	backupDir := filepath.Join(writableRoot, ".reallocate-"+changeID+"-backup")
	for _, path := range []string{changesRoot, oldChangeDir, newChangeDir, backupDir} {
		if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
			return nil, err
		}
	}
	if _, err := os.Stat(newChangeDir); err == nil {
		return nil, fmt.Errorf("reallocated change path %q already exists", runeContextRelativePath(writableRoot, newChangeDir))
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Lstat(backupDir); err == nil {
		return nil, fmt.Errorf("cannot reallocate change %q because backup path %q already exists; inspect or remove the leftover backup before retrying", changeID, runeContextRelativePath(writableRoot, backupDir))
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	stagedDir, err := mkdirTempDir(writableRoot, ".reallocate-"+newID+"-stage-")
	if err != nil {
		return nil, err
	}
	if err := ensurePathAndParentAreNotSymlinks(stagedDir); err != nil {
		_ = removeAllPath(stagedDir)
		return nil, err
	}
	changedFiles, rewrittenRefs, err := stageReallocatedChange(oldChangeDir, stagedDir, changeID, newID)
	if err != nil {
		_ = removeAllPath(stagedDir)
		return nil, err
	}
	cleanupStaged := true
	defer func() {
		if cleanupStaged {
			_ = removeAllPath(stagedDir)
		}
	}()
	for _, path := range []string{changesRoot, oldChangeDir, backupDir, stagedDir, newChangeDir} {
		if err := ensurePathAndParentAreNotSymlinks(path); err != nil {
			return nil, err
		}
	}
	if err := renamePath(oldChangeDir, backupDir); err != nil {
		return nil, err
	}
	if err := renamePath(stagedDir, newChangeDir); err != nil {
		rollbackErr := restoreOriginalChangeFromBackup(backupDir, oldChangeDir)
		return nil, combineReallocationRollbackError(err, rollbackErr)
	}
	cleanupStaged = false
	validated, err := validateProjectAfterChangeMutation(v, loaded.Resolution.ProjectRoot)
	if err != nil {
		rollbackErr := rollbackCommittedReallocatedChange(newChangeDir, backupDir, oldChangeDir)
		return nil, combineReallocationRollbackError(err, rollbackErr)
	}
	_ = validated.Close()
	warnings := make([]string, 0, 1)
	if err := removeAllPath(backupDir); err != nil {
		warnings = append(warnings, fmt.Sprintf("reallocation succeeded but could not remove backup path %q: %v", runeContextRelativePath(writableRoot, backupDir), err))
	}
	sortFileMutations(changedFiles)
	return &ChangeReallocationResult{
		OldID:                   changeID,
		ID:                      newID,
		OldChangePath:           runeContextRelativePath(writableRoot, oldChangeDir),
		ChangePath:              runeContextRelativePath(writableRoot, newChangeDir),
		RewrittenReferenceCount: rewrittenRefs,
		ChangedFiles:            changedFiles,
		Warnings:                warnings,
	}, nil
}

func BuildProjectStatusSummary(v *Validator, loaded *LoadedProject) (*ProjectStatusSummary, error) {
	if v == nil {
		return nil, fmt.Errorf("validator is required")
	}
	if loaded == nil {
		return nil, fmt.Errorf("loaded project is required")
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()
	summary := &ProjectStatusSummary{
		Root:               loaded.Resolution.ProjectRoot,
		SelectedConfigPath: loaded.Resolution.SelectedConfigPath,
		RuneContextVersion: fmt.Sprint(loaded.RootConfig["runecontext_version"]),
		AssuranceTier:      fmt.Sprint(loaded.RootConfig["assurance_tier"]),
		Active:             make([]ChangeStatusEntry, 0),
		Closed:             make([]ChangeStatusEntry, 0),
		Superseded:         make([]ChangeStatusEntry, 0),
	}
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		entry := ChangeStatusEntry{
			ID:     record.ID,
			Title:  record.Title,
			Status: string(record.Status),
			Type:   record.Type,
			Size:   record.Size,
			Path:   runeContextRelativePath(index.ContentRoot, filepath.Join(record.DirPath, "status.yaml")),
		}
		switch record.Status {
		case StatusClosed:
			summary.Closed = append(summary.Closed, entry)
		case StatusSuperseded:
			summary.Superseded = append(summary.Superseded, entry)
		default:
			summary.Active = append(summary.Active, entry)
		}
	}
	if index.Bundles != nil {
		summary.BundleIDs = SortedKeys(index.Bundles.bundles)
	}
	return summary, nil
}

func ensureChangeReallocationIsLocalOnly(index *ProjectIndex, changeID string) error {
	if index == nil {
		return fmt.Errorf("project index is required")
	}
	for _, otherID := range SortedKeys(index.Changes) {
		if otherID == changeID {
			continue
		}
		record := index.Changes[otherID]
		for _, field := range []struct {
			name  string
			items []string
		}{
			{name: "related_changes", items: record.RelatedChanges},
			{name: "depends_on", items: record.DependsOn},
			{name: "informed_by", items: record.InformedBy},
			{name: "supersedes", items: record.Supersedes},
			{name: "superseded_by", items: record.SupersededBy},
		} {
			if containsString(field.items, changeID) {
				return fmt.Errorf("change %q cannot be reallocated because %s in %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, field.name, runeContextRelativePath(index.ContentRoot, record.StatusPath))
			}
		}
	}
	for _, specPath := range SortedKeys(index.Specs) {
		spec := index.Specs[specPath]
		if containsString(spec.OriginatingChanges, changeID) || containsString(spec.RevisedByChanges, changeID) {
			return fmt.Errorf("change %q cannot be reallocated because spec %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, specPath)
		}
	}
	for _, decisionPath := range SortedKeys(index.Decisions) {
		decision := index.Decisions[decisionPath]
		if containsString(decision.OriginatingChanges, changeID) || containsString(decision.RelatedChanges, changeID) {
			return fmt.Errorf("change %q cannot be reallocated because decision %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, decisionPath)
		}
	}
	changePrefix := changeMarkdownPathPrefix(changeID)
	for _, path := range SortedKeys(index.MarkdownFiles) {
		if strings.HasPrefix(path, changePrefix) {
			continue
		}
		artifact := index.MarkdownFiles[path]
		for _, ref := range artifact.Refs {
			if strings.HasPrefix(ref.Path, changePrefix) {
				return fmt.Errorf("change %q cannot be reallocated because markdown deep ref %q in %q points into it; alpha.3 reallocation only rewrites local references inside the change", changeID, ref.Raw, path)
			}
		}
	}
	return nil
}

func stageReallocatedChange(oldChangeDir, stagedDir, oldID, newID string) ([]FileMutation, int, error) {
	info, err := os.Stat(oldChangeDir)
	if err != nil {
		return nil, 0, err
	}
	if err := ensureNoSymlinksInTree(oldChangeDir, filepath.ToSlash(filepath.Join("changes", oldID))); err != nil {
		return nil, 0, err
	}
	oldRoot := filepath.ToSlash(filepath.Join("changes", oldID))
	newRoot := filepath.ToSlash(filepath.Join("changes", newID))
	totalRewritten := 0
	changedFiles := make([]FileMutation, 0)
	err = filepath.WalkDir(oldChangeDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(oldChangeDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		targetPath := filepath.Join(stagedDir, rel)
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode().Perm()); err != nil {
				return err
			}
			return chmodPath(targetPath, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		action := "moved"
		switch {
		case rel == "status.yaml":
			parsed, err := parseYAML(data)
			if err != nil {
				return err
			}
			statusMap, err := expectObject(path, parsed, "status file")
			if err != nil {
				return err
			}
			statusMap = cloneMap(statusMap)
			statusMap["id"] = newID
			data, err = renderStatusYAML(statusMap)
			if err != nil {
				return err
			}
			action = "updated"
		case filepath.Ext(rel) == ".md":
			rewritten, count, err := rewriteMarkdownChangePathMentions(data, oldRoot, newRoot)
			if err != nil {
				return err
			}
			data = rewritten
			totalRewritten += count
			if count > 0 {
				action = "updated"
			}
		}
		if err := os.WriteFile(targetPath, data, info.Mode().Perm()); err != nil {
			return err
		}
		if err := chmodPath(targetPath, info.Mode().Perm()); err != nil {
			return err
		}
		changedFiles = append(changedFiles, FileMutation{Path: filepath.ToSlash(filepath.Join("changes", newID, rel)), Action: action})
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	if err := chmodPath(stagedDir, info.Mode().Perm()); err != nil {
		return nil, 0, err
	}
	return changedFiles, totalRewritten, nil
}

func changeMarkdownPathPrefix(changeID string) string {
	return filepath.ToSlash(filepath.Join("changes", changeID)) + "/"
}

func ensureNoSymlinksInTree(rootPath, relativeRoot string) error {
	return filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == rootPath {
			return nil
		}
		if entry.Type()&os.ModeSymlink == 0 {
			return nil
		}
		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		return fmt.Errorf("reallocation does not support symlinks in change directories: %s", filepath.ToSlash(filepath.Join(relativeRoot, rel)))
	})
}

func rewriteMarkdownChangePathMentions(data []byte, oldRoot, newRoot string) ([]byte, int, error) {
	newline := detectPreferredNewline(data)
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	segments := markdownTextSegments(text)
	var out strings.Builder
	total := 0
	for _, segment := range segments {
		if segment.fenced {
			out.WriteString(segment.text)
			continue
		}
		rewritten, count := rewriteLiteralPathRootInText(segment.text, oldRoot, newRoot)
		out.WriteString(rewritten)
		total += count
	}
	if total == 0 {
		return append([]byte(nil), data...), 0, nil
	}
	result := out.String()
	if newline == "\r\n" {
		result = strings.ReplaceAll(result, "\n", "\r\n")
	}
	return []byte(result), total, nil
}

func rewriteLiteralPathRootInText(text, oldRoot, newRoot string) (string, int) {
	if oldRoot == "" || oldRoot == newRoot {
		return text, 0
	}
	var out strings.Builder
	i := 0
	count := 0
	for i < len(text) {
		idx := strings.Index(text[i:], oldRoot)
		if idx < 0 {
			out.WriteString(text[i:])
			break
		}
		idx += i
		prevOK := idx == 0 || !isMarkdownPathChar(previousRune(text, idx))
		nextPos := idx + len(oldRoot)
		nextOK := nextPos == len(text) || text[nextPos] == '/' || !isMarkdownPathChar(nextRune(text, nextPos))
		if !prevOK || !nextOK {
			out.WriteString(text[i:nextPos])
			i = nextPos
			continue
		}
		out.WriteString(text[i:idx])
		out.WriteString(newRoot)
		count++
		i = nextPos
	}
	return out.String(), count
}

func previousRune(text string, index int) rune {
	if index <= 0 || index > len(text) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeLastRuneInString(text[:index])
	return r
}

func nextRune(text string, index int) rune {
	if index < 0 || index >= len(text) {
		return utf8.RuneError
	}
	r, _ := utf8.DecodeRuneInString(text[index:])
	return r
}

func detectPreferredNewline(data []byte) string {
	text := string(data)
	if strings.Contains(text, "\r\n") && !strings.Contains(strings.ReplaceAll(text, "\r\n", ""), "\n") {
		return "\r\n"
	}
	return "\n"
}

func rollbackCommittedReallocatedChange(newChangeDir, backupDir, oldChangeDir string) error {
	errMessages := make([]string, 0, 2)
	if err := removeAllPath(newChangeDir); err != nil && !os.IsNotExist(err) {
		errMessages = append(errMessages, fmt.Sprintf("remove reallocated change %q: %v", filepath.ToSlash(newChangeDir), err))
	}
	if err := restoreOriginalChangeFromBackup(backupDir, oldChangeDir); err != nil {
		errMessages = append(errMessages, err.Error())
	}
	if len(errMessages) == 0 {
		return nil
	}
	return errors.New(strings.Join(errMessages, "; "))
}

func restoreOriginalChangeFromBackup(backupDir, oldChangeDir string) error {
	errMessages := make([]string, 0, 1)
	if err := renamePath(backupDir, oldChangeDir); err != nil {
		errMessages = append(errMessages, fmt.Sprintf("restore original change %q from backup %q: %v", filepath.ToSlash(oldChangeDir), filepath.ToSlash(backupDir), err))
	}
	if len(errMessages) == 0 {
		return nil
	}
	return errors.New(strings.Join(errMessages, "; "))
}

func combineReallocationRollbackError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%v; rollback also failed and manual recovery may be required: %v", operationErr, rollbackErr)
}

func requireWritableChangeSource(loaded *LoadedProject) error {
	if loaded == nil || loaded.Resolution == nil {
		return fmt.Errorf("loaded project resolution is required")
	}
	switch loaded.Resolution.SourceMode {
	case SourceModeEmbedded, SourceModePath:
		return nil
	default:
		return fmt.Errorf("change write operations are only supported for embedded and local path sources in alpha.3")
	}
}

func writableContentRoot(loaded *LoadedProject) (string, error) {
	if err := requireWritableChangeSource(loaded); err != nil {
		return "", err
	}
	if loaded.Resolution.SourceMode == SourceModeEmbedded {
		return loaded.Resolution.MaterializedRoot(), nil
	}
	if filepath.IsAbs(loaded.Resolution.SourceRoot) {
		return filepath.Clean(loaded.Resolution.SourceRoot), nil
	}
	return filepath.Clean(filepath.Join(loaded.Resolution.ProjectRoot, loaded.Resolution.SourceRoot)), nil
}

func validateRequestedMode(mode ChangeMode) error {
	if mode == "" || mode == ChangeModeMinimum || mode == ChangeModeFull {
		return nil
	}
	return fmt.Errorf("change mode must be %q or %q", ChangeModeMinimum, ChangeModeFull)
}

func validateChangeTypeValue(changeType string) error {
	if strings.TrimSpace(changeType) == "" {
		return fmt.Errorf("change type is required")
	}
	if strings.HasPrefix(changeType, "x-") {
		return nil
	}
	switch changeType {
	case "project", "feature", "bug", "standard", "chore":
		return nil
	default:
		return fmt.Errorf("unsupported change type %q", changeType)
	}
}

func assessChangeIntake(title, changeType, requestedSize, description string) changeIntakeAssessment {
	size := strings.TrimSpace(requestedSize)
	assumptions := make([]string, 0)
	if size == "" {
		size = defaultChangeSize(changeType)
		assumptions = append(assumptions, fmt.Sprintf("Inferred size %q from the change type because no explicit size was provided.", size))
	}
	recommendedMode := ChangeModeMinimum
	reasons := make([]string, 0)
	followUps := make([]string, 0)
	checklistTitle := ""
	checklistItems := []string(nil)
	verificationNote := "Use the repository's standard verification flow before closing this change."
	risky := containsHeuristicKeyword(title+" "+description, []string{"security", "schema", "api", "migration", "rollout", "deploy", "auth", "permission", "secret"})
	ambiguous := containsHeuristicKeyword(title+" "+description, []string{"unclear", "unknown", "investigate", "spike", "explore", "ambiguous"})
	switch changeType {
	case "project":
		recommendedMode = ChangeModeFull
		reasons = append(reasons, "Project work uses deeper intake because bad defaults compound.")
		checklistTitle = "Project Intake Checklist"
		checklistItems = []string{
			"Mission and target users.",
			"Stack and runtime constraints.",
			"Deployment and security constraints.",
			"Success criteria.",
			"Non-goals.",
		}
		followUps = append(followUps, checklistItems...)
	case "feature":
		if size == "large" || risky || ambiguous {
			recommendedMode = ChangeModeFull
			reasons = append(reasons, "Large, ambiguous, or high-risk feature work should move to full mode early.")
		}
	case "bug":
		if size == "large" || risky || ambiguous {
			recommendedMode = ChangeModeFull
			reasons = append(reasons, "Bugs with unclear root cause, ambiguity, or security/schema/API impact should be shaped in full mode.")
			checklistTitle = "Bug Escalation Checklist"
			checklistItems = []string{
				"Clarify the current behavior and the expected behavior.",
				"Confirm whether security, schema, or API surfaces are affected.",
				"Record the root-cause hypothesis and any open uncertainties.",
			}
			followUps = append(followUps,
				"User-facing behavior that materially changes the fix.",
				"API or interface changes introduced by the fix.",
				"Verification and acceptance criteria for the repaired behavior.",
			)
		}
	case "standard", "chore":
		if size == "large" || containsHeuristicKeyword(title+" "+description, []string{"deprecate", "rename", "migration", "rollout"}) {
			recommendedMode = ChangeModeFull
			reasons = append(reasons, "Broad standards or chore changes should be shaped so future impact stays reviewable.")
		}
	default:
		if size == "large" || risky || ambiguous {
			recommendedMode = ChangeModeFull
			reasons = append(reasons, "The requested change looks large, ambiguous, or risky enough to justify full mode.")
		}
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Minimum mode is sufficient for the current size and risk signal.")
	}
	return changeIntakeAssessment{
		Size:             size,
		RecommendedMode:  recommendedMode,
		Reasons:          reasons,
		Assumptions:      assumptions,
		ChecklistTitle:   checklistTitle,
		ChecklistItems:   checklistItems,
		FollowUpPrompts:  followUps,
		VerificationNote: verificationNote,
	}
}

func defaultChangeSize(changeType string) string {
	switch changeType {
	case "project":
		return "large"
	case "feature":
		return "medium"
	case "bug", "standard", "chore":
		return "small"
	default:
		return "medium"
	}
}

func resolveContextBundlesForChange(index *ProjectIndex, requested []string) ([]string, []string, error) {
	bundles := uniqueSortedStrings(requested)
	if index == nil || index.Bundles == nil || len(index.Bundles.bundles) == 0 {
		return bundles, nil, nil
	}
	for _, id := range bundles {
		if _, ok := index.Bundles.bundles[id]; !ok {
			return nil, nil, fmt.Errorf("context bundle %q does not exist", id)
		}
	}
	if len(bundles) > 0 {
		return bundles, nil, nil
	}
	if len(index.Bundles.bundles) == 1 {
		inferred := SortedKeys(index.Bundles.bundles)
		return inferred, []string{fmt.Sprintf("Inferred context bundle %q because it is the only bundle in the project.", inferred[0])}, nil
	}
	for _, candidate := range []string{"default", "base", "core"} {
		if _, ok := index.Bundles.bundles[candidate]; ok {
			return []string{candidate}, []string{fmt.Sprintf("Inferred context bundle %q from the repository defaults.", candidate)}, nil
		}
	}
	return nil, []string{"No context bundle was selected automatically; standards fall back to all non-draft standards."}, nil
}

func resolveApplicableStandards(index *ProjectIndex, contextBundles []string) ([]string, []string, error) {
	if index == nil {
		return nil, nil, fmt.Errorf("project index is required")
	}
	selected := make([]string, 0)
	for _, bundleID := range uniqueSortedStrings(contextBundles) {
		resolution, err := index.ResolveBundle(bundleID)
		if err != nil {
			return nil, nil, err
		}
		if aspect, ok := resolution.Aspects[BundleAspectStandards]; ok {
			for _, entry := range aspect.Selected {
				selected = append(selected, entry.Path)
			}
		}
	}
	selected = uniqueSortedStrings(selected)
	if len(selected) > 0 {
		return selected, nil, nil
	}
	fallback := make([]string, 0)
	for _, path := range SortedKeys(index.Standards) {
		if index.Standards[path].Status == StandardStatusDraft {
			continue
		}
		fallback = append(fallback, path)
	}
	if len(fallback) == 0 {
		return nil, nil, fmt.Errorf("cannot infer applicable standards because the project has no selectable standards")
	}
	return fallback, []string{"Used all non-draft standards as a conservative fallback because no standards were selected through context bundles."}, nil
}

func newStatusMap(id, title, changeType, size string, contextBundles []string, now time.Time) map[string]any {
	return map[string]any{
		"schema_version":       1,
		"id":                   id,
		"title":                title,
		"status":               string(StatusProposed),
		"type":                 changeType,
		"size":                 size,
		"verification_status":  "pending",
		"context_bundles":      stringSliceToAny(contextBundles),
		"related_specs":        []any{},
		"related_decisions":    []any{},
		"related_changes":      []any{},
		"depends_on":           []any{},
		"informed_by":          []any{},
		"supersedes":           []any{},
		"superseded_by":        []any{},
		"created_at":           now.Format("2006-01-02"),
		"closed_at":            nil,
		"promotion_assessment": map[string]any{"status": "pending", "suggested_targets": []any{}},
	}
}

func renderStatusYAML(raw map[string]any) ([]byte, error) {
	doc, err := statusDocumentFromMap(raw)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	err = encoder.Encode(doc)
	_ = encoder.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func statusDocumentFromMap(raw map[string]any) (statusDocument, error) {
	doc := statusDocument{
		SchemaVersion:      intValue(raw["schema_version"], 1),
		ID:                 requiredStringValue(raw["id"]),
		Title:              requiredStringValue(raw["title"]),
		Status:             requiredStringValue(raw["status"]),
		Type:               requiredStringValue(raw["type"]),
		Size:               optionalStringValue(raw["size"]),
		VerificationStatus: requiredStringValue(raw["verification_status"]),
		ContextBundles:     nonNilStrings(extractStringList(raw["context_bundles"])),
		RelatedSpecs:       nonNilStrings(extractStringList(raw["related_specs"])),
		RelatedDecisions:   nonNilStrings(extractStringList(raw["related_decisions"])),
		RelatedChanges:     nonNilStrings(extractStringList(raw["related_changes"])),
		DependsOn:          nonNilStrings(extractStringList(raw["depends_on"])),
		InformedBy:         nonNilStrings(extractStringList(raw["informed_by"])),
		Supersedes:         nonNilStrings(extractStringList(raw["supersedes"])),
		SupersededBy:       nonNilStrings(extractStringList(raw["superseded_by"])),
		CreatedAt:          optionalStringValue(raw["created_at"]),
		ClosedAt:           raw["closed_at"],
		PromotionAssessment: promotionAssessmentDocument{
			Status:           "pending",
			SuggestedTargets: []promotionTargetDocument{},
		},
	}
	if promotionRaw, ok := raw["promotion_assessment"].(map[string]any); ok {
		status, err := promotionAssessmentStatusValue(promotionRaw["status"])
		if err != nil {
			return statusDocument{}, err
		}
		doc.PromotionAssessment.Status = status
		for _, targetRaw := range extractAnySlice(promotionRaw["suggested_targets"]) {
			targetMap, ok := targetRaw.(map[string]any)
			if !ok {
				continue
			}
			doc.PromotionAssessment.SuggestedTargets = append(doc.PromotionAssessment.SuggestedTargets, promotionTargetDocument{
				TargetType: fmt.Sprint(targetMap["target_type"]),
				TargetPath: fmt.Sprint(targetMap["target_path"]),
				Summary:    fmt.Sprint(targetMap["summary"]),
			})
		}
	}
	if extensions, ok := raw["extensions"].(map[string]any); ok && len(extensions) > 0 {
		doc.Extensions = cloneMap(extensions)
	}
	return doc, nil
}

func writeStatusMap(path string, raw map[string]any) error {
	data, err := renderStatusYAML(raw)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func renderProposalMarkdown(title, description string, mode ChangeMode, reasons, assumptions []string) []byte {
	summary := strings.TrimSpace(title)
	problem := "The repository needs a reviewable RuneContext change record for this work."
	proposed := fmt.Sprintf("Track %s through the %s RuneContext change artifacts.", strings.TrimSpace(title), mode)
	whyNow := "The work needs stable intent, standards linkage, and verification planning before it moves further."
	outOfScope := "Work outside the scoped change tracked here."
	impact := "The change keeps intent, assumptions, and standards linkage reviewable."
	if trimmed := strings.TrimSpace(description); trimmed != "" {
		problem = trimmed
		proposed = fmt.Sprintf("Track and deliver %s while keeping the intent and standards linkage reviewable.", strings.TrimSpace(title))
	}
	assumptionsBody := "N/A"
	if len(assumptions) > 0 {
		assumptionsBody = renderBulletList(assumptions)
	}
	return []byte(strings.Join([]string{
		"## Summary",
		summary,
		"",
		"## Problem",
		problem,
		"",
		"## Proposed Change",
		proposed,
		"",
		"## Why Now",
		whyNow,
		"",
		"## Assumptions",
		assumptionsBody,
		"",
		"## Out of Scope",
		outOfScope,
		"",
		"## Impact",
		impact,
		"",
	}, "\n"))
}

func renderStandardsMarkdown(existing []byte, applicable, added []string, preserved []markdownSection, creating bool) []byte {
	sections := make([]string, 0, 8)
	sections = append(sections,
		"## Applicable Standards",
		renderStandardBullets(applicable, "Selected from the current context bundles."),
	)
	if len(added) > 0 {
		sections = append(sections,
			"",
			"## Standards Added Since Last Refresh",
			renderStandardBullets(added, "Newly selected during standards refresh."),
		)
	}
	hasResolutionNotes := false
	for _, section := range preserved {
		switch section.Heading {
		case "Applicable Standards", "Standards Added Since Last Refresh":
			continue
		case "Resolution Notes":
			hasResolutionNotes = true
		}
		sections = append(sections, "", "## "+section.Heading, section.Body)
	}
	if creating && !hasResolutionNotes {
		sections = append(sections,
			"",
			"## Resolution Notes",
			"Generated from the current context bundle selection; review any automatic refresh before committing.",
		)
	}
	return []byte(strings.Join(sections, "\n") + "\n")
}

func preservedStandardsSections(data []byte) ([]markdownSection, error) {
	sections, err := parseLevel2Sections("standards.md", data)
	if err != nil {
		return nil, err
	}
	preserved := make([]markdownSection, 0, len(sections))
	for _, section := range sections {
		switch section.Heading {
		case "Applicable Standards", "Standards Added Since Last Refresh":
			continue
		default:
			preserved = append(preserved, section)
		}
	}
	return preserved, nil
}

func renderStandardBullets(paths []string, description string) string {
	if len(paths) == 0 {
		return "- `standards/placeholder.md`: Replace this placeholder once the project defines a selectable standard."
	}
	lines := make([]string, 0, len(paths))
	for _, path := range paths {
		lines = append(lines, fmt.Sprintf("- `%s`: %s", path, description))
	}
	return strings.Join(lines, "\n")
}

func materializeShapeFiles(changeDir, writableRoot, projectRoot, title string, assessment changeIntakeAssessment, options ChangeShapeOptions) ([]FileMutation, error) {
	verificationCommands := assessment.VerificationCmds
	if len(verificationCommands) == 0 {
		verificationCommands, _ = inferVerificationCommands(projectRoot)
	}
	files := []struct {
		name string
		data []byte
		ok   bool
	}{
		{name: "design.md", data: renderDesignMarkdown(title, assessment, options.Design), ok: true},
		{name: "verification.md", data: renderVerificationMarkdown(title, verificationCommands, assessment.VerificationNote, options.Verification), ok: true},
		{name: "tasks.md", data: renderSupplementalMarkdown("Tasks", options.Tasks), ok: len(options.Tasks) > 0},
		{name: "references.md", data: renderSupplementalMarkdown("References", options.References), ok: len(options.References) > 0},
	}
	changed := make([]FileMutation, 0, len(files))
	for _, file := range files {
		if !file.ok {
			continue
		}
		path := filepath.Join(changeDir, file.name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if err := writeFileAtomically(path, file.data, 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, FileMutation{Path: runeContextRelativePath(writableRoot, path), Action: "created"})
	}
	return changed, nil
}

func renderDesignMarkdown(title string, assessment changeIntakeAssessment, design string) []byte {
	overview := strings.TrimSpace(design)
	if overview == "" {
		overview = fmt.Sprintf("Shape %s before implementation so scope, standards linkage, and verification stay reviewable.", strings.TrimSpace(title))
	}
	lines := []string{"# Design", "", "## Overview", overview}
	if len(assessment.Reasons) > 0 {
		lines = append(lines, "", "## Shape Rationale", renderBulletList(assessment.Reasons))
	}
	if len(assessment.ChecklistItems) > 0 {
		lines = append(lines, "", "## "+assessment.ChecklistTitle, renderBulletList(assessment.ChecklistItems))
	}
	if len(assessment.FollowUpPrompts) > 0 {
		lines = append(lines, "", "## Ask More When", renderBulletList(assessment.FollowUpPrompts))
	}
	lines = append(lines, "")
	return []byte(strings.Join(lines, "\n"))
}

func renderVerificationMarkdown(title string, commands []string, note, verification string) []byte {
	lines := []string{"# Verification", ""}
	if strings.TrimSpace(verification) != "" {
		lines = append(lines, "## Notes", strings.TrimSpace(verification), "")
	}
	planned := make([]string, 0)
	for _, command := range commands {
		planned = append(planned, fmt.Sprintf("`%s`", command))
	}
	if len(planned) == 0 {
		planned = append(planned, "Define the repository verification commands before closing this change.")
	}
	lines = append(lines, "## Planned Checks", renderBulletList(planned))
	if strings.TrimSpace(note) != "" {
		lines = append(lines, "", "## Close Gate", note)
	}
	lines = append(lines, "")
	return []byte(strings.Join(lines, "\n"))
}

func renderSupplementalMarkdown(title string, items []string) []byte {
	lines := []string{"# " + title, "", renderBulletList(items), ""}
	return []byte(strings.Join(lines, "\n"))
}

func renderBulletList(items []string) string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, "- "+strings.TrimSpace(item))
	}
	return strings.Join(lines, "\n")
}

func inferVerificationCommands(projectRoot string) ([]string, string) {
	justfilePath := filepath.Join(projectRoot, "justfile")
	if data, err := os.ReadFile(justfilePath); err == nil && justfileTestTargetPattern.Match(data) {
		return []string{"just test"}, "Inferred `just test` from the repository's justfile test target."
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
		return []string{"go test ./..."}, "Inferred `go test ./...` from the repository's Go module layout."
	}
	if _, err := os.Stat(filepath.Join(projectRoot, "package.json")); err == nil {
		return []string{"npm test"}, "Inferred `npm test` from the repository's package.json."
	}
	return nil, ""
}

func promotionAssessmentStatusValue(raw any) (string, error) {
	status := optionalStringValue(raw)
	if status == "" {
		return "pending", nil
	}
	if _, ok := allowedPromotionAssessmentStatuses[status]; !ok {
		return "", fmt.Errorf("promotion_assessment.status must be one of pending, none, suggested, accepted, or completed")
	}
	return status, nil
}

func validateCloseVerificationStatus(current, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		if current == "pending" {
			return fmt.Errorf("change close requires --verification-status when the current verification_status is pending")
		}
		return nil
	}
	if requested == "pending" {
		return fmt.Errorf("change close must not set verification_status to pending")
	}
	switch requested {
	case "passed", "failed", "skipped":
		return nil
	default:
		return fmt.Errorf("unsupported verification_status %q", requested)
	}
}

func inferChangeMode(changeDir string) ChangeMode {
	for _, name := range []string{"design.md", "verification.md"} {
		if _, err := os.Stat(filepath.Join(changeDir, name)); err == nil {
			return ChangeModeFull
		}
	}
	return ChangeModeMinimum
}

func containsHeuristicKeyword(value string, keywords []string) bool {
	value = strings.ToLower(value)
	for _, keyword := range keywords {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func sliceDifference(items, base []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	for _, item := range base {
		seen[item] = struct{}{}
	}
	result := make([]string, 0)
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		result = append(result, item)
	}
	return uniqueSortedStrings(result)
}

func uniqueStringsInOrder(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = cloneTopLevelValue(value)
	}
	return result
}

func nonNilStrings(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

func extractAnySlice(raw any) []any {
	items, _ := raw.([]any)
	if items == nil {
		return []any{}
	}
	return items
}

func intValue(raw any, fallback int) int {
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return fallback
	}
}

func sortFileMutations(items []FileMutation) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Action < items[j].Action
		}
		return items[i].Path < items[j].Path
	})
}

func createUniqueChangeDir(contentRoot string, now time.Time, title string, entropy io.Reader) (string, string, error) {
	changesRoot := filepath.Join(contentRoot, "changes")
	for attempt := 0; attempt < maxCreateChangeDirAttempts; attempt++ {
		id, err := AllocateChangeID(contentRoot, now, title, entropy)
		if err != nil {
			return "", "", err
		}
		changeDir := filepath.Join(changesRoot, id)
		if err := os.Mkdir(changeDir, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", err
		}
		return id, changeDir, nil
	}
	return "", "", fmt.Errorf("could not allocate a unique change directory after %d attempts", maxCreateChangeDirAttempts)
}
