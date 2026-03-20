package contracts

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type createChangeState struct {
	id                  string
	title               string
	changeType          string
	now                 time.Time
	changeDir           string
	writableRoot        string
	projectRoot         string
	assessment          changeIntakeAssessment
	selectedMode        ChangeMode
	contextBundles      []string
	applicableStandards []string
	assumptions         []string
}

func CreateChange(v *Validator, loaded *LoadedProject, options ChangeCreateOptions) (result *ChangeOperationResult, err error) {
	if err := validateCreateChangeInputs(v, loaded, options); err != nil {
		return nil, err
	}
	state, cleanup, err := prepareCreateChange(v, loaded, options)
	if err != nil {
		return nil, err
	}
	defer cleanup(&err)
	changedFiles, err := writeCreateChangeFiles(v, loaded, state, options)
	if err != nil {
		return nil, err
	}
	if err := validateChangeMutation(v, loaded.Resolution.ProjectRoot); err != nil {
		return nil, err
	}
	cleanup = disableChangeDirCleanup
	return buildCreateChangeResult(state, changedFiles), nil
}

func validateCreateChangeInputs(v *Validator, loaded *LoadedProject, options ChangeCreateOptions) error {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return err
	}
	if strings.TrimSpace(options.Title) == "" {
		return fmt.Errorf("change title is required")
	}
	if err := validateChangeTypeValue(strings.TrimSpace(options.Type)); err != nil {
		return err
	}
	return validateRequestedMode(options.RequestedMode)
}

func prepareCreateChange(v *Validator, loaded *LoadedProject, options ChangeCreateOptions) (createChangeState, func(*error), error) {
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return createChangeState{}, nil, err
	}
	defer index.Close()
	state, err := buildCreateChangeState(index, loaded, options)
	if err != nil {
		return createChangeState{}, nil, err
	}
	return state, changeDirCleanup(state.changeDir), nil
}

func buildCreateChangeState(index *ProjectIndex, loaded *LoadedProject, options ChangeCreateOptions) (createChangeState, error) {
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return createChangeState{}, err
	}
	assessment, selectedMode, assumptions, contextBundles, standards, err := resolveCreateChangeContext(index, loaded, options)
	if err != nil {
		return createChangeState{}, err
	}
	changeType := strings.TrimSpace(options.Type)
	now := normalizedChangeTime(options.Now)
	id, changeDir, err := allocateCreateChangeDir(writableRoot, now, options.Title, options.Entropy)
	if err != nil {
		return createChangeState{}, err
	}
	return createChangeState{
		id:                  id,
		title:               options.Title,
		changeType:          changeType,
		now:                 now,
		changeDir:           changeDir,
		writableRoot:        writableRoot,
		projectRoot:         loaded.Resolution.ProjectRoot,
		assessment:          assessment,
		selectedMode:        selectedMode,
		contextBundles:      contextBundles,
		applicableStandards: standards,
		assumptions:         assumptions,
	}, nil
}

func resolveCreateChangeContext(index *ProjectIndex, loaded *LoadedProject, options ChangeCreateOptions) (changeIntakeAssessment, ChangeMode, []string, []string, []string, error) {
	assessment := assessChangeIntake(options.Title, strings.TrimSpace(options.Type), options.Size, options.Description)
	assessment.VerificationCmds, _ = inferVerificationCommands(loaded.Resolution.ProjectRoot)
	selectedMode := assessment.RecommendedMode
	if options.RequestedMode != "" {
		selectedMode = options.RequestedMode
	}
	contextBundles, contextAssumptions, err := resolveContextBundlesForChange(index, options.ContextBundles)
	if err != nil {
		return changeIntakeAssessment{}, "", nil, nil, nil, err
	}
	standards, standardAssumptions, err := resolveApplicableStandards(index, contextBundles)
	if err != nil {
		return changeIntakeAssessment{}, "", nil, nil, nil, err
	}
	assumptions := uniqueStringsInOrder(append(append([]string{}, assessment.Assumptions...), contextAssumptions...))
	assumptions = uniqueStringsInOrder(append(assumptions, standardAssumptions...))
	if note := verificationAssumption(loaded.Resolution.ProjectRoot); note != "" {
		assumptions = append(assumptions, note)
	}
	return assessment, selectedMode, assumptions, contextBundles, standards, nil
}

func allocateCreateChangeDir(writableRoot string, now time.Time, title string, entropy io.Reader) (string, string, error) {
	changesRoot := filepath.Join(writableRoot, "changes")
	if err := ensurePathAndParentAreNotSymlinks(changesRoot); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(changesRoot, 0o755); err != nil {
		return "", "", err
	}
	id, changeDir, err := createUniqueChangeDir(writableRoot, now, title, entropy)
	if err != nil {
		return "", "", err
	}
	if err := ensurePathAndParentAreNotSymlinks(changeDir); err != nil {
		_ = removeAllPath(changeDir)
		return "", "", err
	}
	return id, changeDir, nil
}

func writeCreateChangeFiles(v *Validator, loaded *LoadedProject, state createChangeState, options ChangeCreateOptions) ([]FileMutation, error) {
	files, err := createChangeFiles(v, state, options)
	if err != nil {
		return nil, err
	}
	changedFiles, err := writeCreatedChangeFiles(state.writableRoot, files)
	if err != nil {
		return nil, err
	}
	if state.selectedMode != ChangeModeFull {
		return changedFiles, nil
	}
	shapeFiles, err := materializeShapeFiles(state.changeDir, state.writableRoot, state.projectRoot, state.title, state.assessment, ChangeShapeOptions{
		Design:       options.Design,
		Verification: options.Verification,
		Tasks:        options.Tasks,
		References:   options.References,
	})
	if err != nil {
		return nil, err
	}
	return append(changedFiles, shapeFiles...), nil
}

type createChangeFile struct {
	path string
	data []byte
}

func createChangeFiles(v *Validator, state createChangeState, options ChangeCreateOptions) ([]createChangeFile, error) {
	statusPath := filepath.Join(state.changeDir, "status.yaml")
	statusData, err := validatedCreateStatusData(v, state, statusPath)
	if err != nil {
		return nil, err
	}
	proposalPath := filepath.Join(state.changeDir, "proposal.md")
	proposalData, err := validatedProposalData(v, state, proposalPath, options.Description)
	if err != nil {
		return nil, err
	}
	standardsPath := filepath.Join(state.changeDir, "standards.md")
	standardsData, err := validatedStandardsData(v, state, standardsPath)
	if err != nil {
		return nil, err
	}
	return []createChangeFile{{statusPath, statusData}, {proposalPath, proposalData}, {standardsPath, standardsData}}, nil
}

func validatedCreateStatusData(v *Validator, state createChangeState, statusPath string) ([]byte, error) {
	statusRaw := newStatusMap(state.id, state.title, state.changeType, state.assessment.Size, state.contextBundles, state.now)
	statusData, err := renderStatusYAML(statusRaw)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateYAMLFile("change-status.schema.json", statusPath, statusData); err != nil {
		return nil, err
	}
	return statusData, nil
}

func validatedProposalData(v *Validator, state createChangeState, proposalPath, description string) ([]byte, error) {
	proposalData := renderProposalMarkdown(state.title, description, state.selectedMode, state.assessment.Reasons, state.assumptions)
	if err := v.ValidateProposalMarkdown(proposalPath, proposalData); err != nil {
		return nil, err
	}
	return proposalData, nil
}

func validatedStandardsData(v *Validator, state createChangeState, standardsPath string) ([]byte, error) {
	standardsData := renderStandardsMarkdown(nil, state.applicableStandards, nil, nil, true)
	if err := v.ValidateStandardsMarkdown(standardsPath, standardsData); err != nil {
		return nil, err
	}
	return standardsData, nil
}

func writeCreatedChangeFiles(writableRoot string, files []createChangeFile) ([]FileMutation, error) {
	changedFiles := make([]FileMutation, 0, len(files))
	for _, file := range files {
		if err := writeFileAtomically(file.path, file.data, 0o644); err != nil {
			return nil, err
		}
		changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, file.path), Action: "created"})
	}
	return changedFiles, nil
}

func buildCreateChangeResult(state createChangeState, changedFiles []FileMutation) *ChangeOperationResult {
	sortFileMutations(changedFiles)
	return &ChangeOperationResult{
		ID:                     state.id,
		ChangePath:             runeContextRelativePath(state.writableRoot, state.changeDir),
		Mode:                   state.selectedMode,
		RecommendedMode:        state.assessment.RecommendedMode,
		Status:                 string(StatusProposed),
		ContextBundles:         append([]string(nil), state.contextBundles...),
		ApplicableStandards:    append([]string(nil), state.applicableStandards...),
		ChangedFiles:           changedFiles,
		StandardsRefreshAction: "created",
		ReviewDiffRequired:     true,
		Reasons:                append([]string(nil), state.assessment.Reasons...),
		Assumptions:            append([]string(nil), state.assumptions...),
	}
}
