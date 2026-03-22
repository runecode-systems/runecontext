package contracts

import (
	"io"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"time"
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

type PromoteOptions struct {
	Accept   bool
	Complete bool
	Targets  []string
}

type FileMutation struct {
	Path   string
	Action string
}

type ChangeOperationResult struct {
	ID                        string
	ChangePath                string
	Mode                      ChangeMode
	RecommendedMode           ChangeMode
	Status                    string
	ClosedAt                  string
	ContextBundles            []string
	ApplicableStandards       []string
	AddedStandards            []string
	ChangedFiles              []FileMutation
	StandardsRefreshAction    string
	ReviewDiffRequired        bool
	Reasons                   []string
	Assumptions               []string
	PromotionAssessmentStatus string
	SuggestedPromotionTargets []string
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

type fileRewrite struct {
	Path string
	Data []byte
}

type fileBackup struct {
	Path string
	Data []byte
	Perm fs.FileMode
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
