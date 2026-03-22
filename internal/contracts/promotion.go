package contracts

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

type promoteOperation string

type promoteChangePrepared struct {
	record       *ChangeRecord
	writableRoot string
	updated      map[string]any
	writes       []fileRewrite
	changedFiles []FileMutation
	changeID     string
}

const (
	promoteOperationAccept   promoteOperation = "accept"
	promoteOperationComplete promoteOperation = "complete"
)

func PromoteChange(v *Validator, loaded *LoadedProject, changeID string, options PromoteOptions) (*ChangeOperationResult, error) {
	if err := validateWritableChangeCommand(v, loaded); err != nil {
		return nil, err
	}
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		return nil, err
	}
	defer index.Close()

	prepared, err := preparePromoteChange(v, index, loaded, changeID, options)
	if err != nil {
		return nil, err
	}
	if err := applyPromoteChangeWrites(v, loaded, prepared.writes); err != nil {
		return nil, err
	}
	return buildPromoteChangeResult(prepared), nil
}

func preparePromoteChange(v *Validator, index *ProjectIndex, loaded *LoadedProject, changeID string, options PromoteOptions) (*promoteChangePrepared, error) {
	record := index.Changes[changeID]
	if record == nil {
		return nil, fmt.Errorf("change %q does not exist", changeID)
	}
	if err := validatePromoteTargets(options.Targets); err != nil {
		return nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	updated := cloneMap(index.StatusFiles[record.StatusPath].Data)
	changed, err := applyPromoteStatus(updated, options)
	if err != nil {
		return nil, err
	}
	writes, changedFiles, err := promoteStatusWrites(v, writableRoot, record, updated, changed)
	if err != nil {
		return nil, err
	}
	return &promoteChangePrepared{
		record:       record,
		writableRoot: writableRoot,
		updated:      updated,
		writes:       writes,
		changedFiles: changedFiles,
		changeID:     changeID,
	}, nil
}

func promoteStatusWrites(v *Validator, writableRoot string, record *ChangeRecord, updated map[string]any, changed bool) ([]fileRewrite, []FileMutation, error) {
	if !changed {
		return nil, nil, nil
	}
	write, changedFiles, err := buildPrimaryCloseStatusWrite(v, writableRoot, record, updated)
	if err != nil {
		return nil, nil, err
	}
	return []fileRewrite{write}, changedFiles, nil
}

func applyPromoteChangeWrites(v *Validator, loaded *LoadedProject, writes []fileRewrite) error {
	return applyFileRewritesTransaction(writes, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	})
}

func buildPromoteChangeResult(prepared *promoteChangePrepared) *ChangeOperationResult {
	sortFileMutations(prepared.changedFiles)
	changeDir := filepath.Join(prepared.writableRoot, "changes", prepared.changeID)
	promotionStatus, promotionTargets := closePromotionAssessmentDetails(prepared.updated)
	return &ChangeOperationResult{
		ID:                        prepared.changeID,
		ChangePath:                runeContextRelativePath(prepared.writableRoot, changeDir),
		Mode:                      inferChangeMode(changeDir),
		RecommendedMode:           inferChangeMode(changeDir),
		Status:                    fmt.Sprint(prepared.updated["status"]),
		ContextBundles:            append([]string(nil), prepared.record.ContextBundles...),
		ApplicableStandards:       append([]string(nil), prepared.record.ApplicableStandards...),
		ChangedFiles:              prepared.changedFiles,
		ReviewDiffRequired:        false,
		PromotionAssessmentStatus: promotionStatus,
		SuggestedPromotionTargets: promotionTargets,
	}
}

func applyPromoteStatus(updated map[string]any, options PromoteOptions) (bool, error) {
	operation, err := resolvePromoteOperation(options)
	if err != nil {
		return false, err
	}
	promotion := ensurePromotionAssessmentMap(updated)
	status, err := promotionAssessmentStatusValue(promotion["status"])
	if err != nil {
		return false, err
	}

	statusChanged, err := applyStatusTransition(promotion, status, operation)
	if err != nil {
		return false, err
	}
	targetChanged := false
	if len(options.Targets) > 0 {
		targetChanged = applyPromotionTargets(promotion, options.Targets)
	}
	return statusChanged || targetChanged, nil
}

func applyStatusTransition(promotion map[string]any, status string, operation promoteOperation) (bool, error) {
	switch operation {
	case promoteOperationAccept:
		switch status {
		case "accepted", "completed":
			return false, nil
		case "suggested":
			promotion["status"] = "accepted"
			return true, nil
		default:
			return false, fmt.Errorf("promotion accept requires current promotion_assessment.status to be \"suggested\"")
		}
	case promoteOperationComplete:
		switch status {
		case "completed":
			return false, nil
		case "accepted":
			promotion["status"] = "completed"
			return true, nil
		default:
			return false, fmt.Errorf("promotion complete requires current promotion_assessment.status to be \"accepted\"")
		}
	default:
		return false, fmt.Errorf("unsupported promote operation %q", operation)
	}
}

func applyPromotionTargets(promotion map[string]any, targets []string) bool {
	if len(targets) == 0 {
		return false
	}
	newTargets := make([]any, 0, len(targets))
	for _, rawTarget := range targets {
		trimmed := strings.TrimSpace(rawTarget)
		if trimmed == "" {
			continue
		}
		targetType, targetPath, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		typeTrim := strings.TrimSpace(targetType)
		newTargets = append(newTargets, map[string]any{
			"target_type": typeTrim,
			"target_path": strings.TrimSpace(targetPath),
			"summary":     defaultPromotionTargetSummary(typeTrim),
		})
	}
	if len(newTargets) == 0 {
		return false
	}
	existingTargets := extractAnySlice(promotion["suggested_targets"])
	if reflect.DeepEqual(existingTargets, newTargets) {
		return false
	}
	promotion["suggested_targets"] = newTargets
	return true
}

func resolvePromoteOperation(options PromoteOptions) (promoteOperation, error) {
	if options.Accept && options.Complete {
		return "", fmt.Errorf("--accept and --complete cannot be used together")
	}
	if options.Complete {
		return promoteOperationComplete, nil
	}
	return promoteOperationAccept, nil
}

func ensurePromotionAssessmentMap(updated map[string]any) map[string]any {
	promotion, ok := updated["promotion_assessment"].(map[string]any)
	if !ok {
		promotion = map[string]any{}
		updated["promotion_assessment"] = promotion
	}
	if _, ok := promotion["suggested_targets"]; !ok {
		promotion["suggested_targets"] = []any{}
	}
	return promotion
}

var allowedPromotionTargetTypes = map[string]struct{}{
	"spec":     {},
	"standard": {},
	"decision": {},
}

const allowedPromotionTargetTypeMessage = "spec, standard, decision"

func validatePromoteTargets(targets []string) error {
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			return fmt.Errorf("promotion target cannot be empty")
		}
		targetType, targetPath, ok := strings.Cut(target, ":")
		if !ok || strings.TrimSpace(targetType) == "" || strings.TrimSpace(targetPath) == "" {
			return fmt.Errorf("invalid promotion target %q: expected TYPE:PATH", target)
		}
		typeTrim := strings.TrimSpace(targetType)
		if !allowedPromotionTargetType(typeTrim) {
			return fmt.Errorf("invalid promotion target %q: unknown target type %q (allowed: %s)", target, typeTrim, allowedPromotionTargetTypeMessage)
		}
	}
	return nil
}

func allowedPromotionTargetType(targetType string) bool {
	_, ok := allowedPromotionTargetTypes[targetType]
	return ok
}
