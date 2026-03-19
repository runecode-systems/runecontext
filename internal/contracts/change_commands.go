package contracts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

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

func CreateChange(v *Validator, loaded *LoadedProject, options ChangeCreateOptions) (*ChangeOperationResult, error) {
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
	if err := os.MkdirAll(changesRoot, 0o755); err != nil {
		return nil, err
	}
	id, changeDir, err := createUniqueChangeDir(writableRoot, now, options.Title, options.Entropy)
	if err != nil {
		return nil, err
	}
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
		if err := os.WriteFile(file.path, file.data, 0o644); err != nil {
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
	validated, err := v.ValidateProjectWithOptions(loaded.Resolution.ProjectRoot, ResolveOptions{
		ConfigDiscovery: ConfigDiscoveryExplicitRoot,
		ExecutionMode:   ExecutionModeLocal,
	})
	if err != nil {
		return nil, err
	}
	_ = validated.Close()
	result := &ChangeOperationResult{
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
	changedFiles := make([]FileMutation, 0, 1+len(options.SupersededBy))
	if err := writeStatusMap(record.StatusPath, updated); err != nil {
		return nil, err
	}
	changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, record.StatusPath), Action: "updated"})
	for _, successorID := range options.SupersededBy {
		successor := index.Changes[successorID]
		successorStatus := cloneMap(index.StatusFiles[successor.StatusPath].Data)
		supersedes := extractStringList(successorStatus["supersedes"])
		if !containsString(supersedes, changeID) {
			if isTerminalLifecycleStatus(successor.Status) {
				return nil, fmt.Errorf("successor change %q is already in terminal status %q and cannot be updated with a reciprocal supersedes link", successorID, successor.Status)
			}
			supersedes = append(supersedes, changeID)
			successorStatus["supersedes"] = stringSliceToAny(uniqueSortedStrings(supersedes))
			if err := writeStatusMap(successor.StatusPath, successorStatus); err != nil {
				return nil, err
			}
			changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(writableRoot, successor.StatusPath), Action: "updated"})
		}
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
		if err := os.WriteFile(path, file.data, 0o644); err != nil {
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
