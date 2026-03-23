package contracts

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type ContextPackAssuranceCaptureOptions struct {
	BundleIDs   []string
	GeneratedAt time.Time
}

type AssuranceCaptureResult struct {
	ReceiptPath  string
	ReceiptID    string
	PackHash     string
	ChangedFiles []FileMutation
}

func CaptureContextPackAssurance(v *Validator, loaded *LoadedProject, index *ProjectIndex, options ContextPackAssuranceCaptureOptions) (*AssuranceCaptureResult, error) {
	if err := validateContextPackAssuranceCaptureInputs(v, loaded, index, options); err != nil {
		return nil, err
	}
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: append([]string(nil), options.BundleIDs...), GeneratedAt: options.GeneratedAt})
	if err != nil {
		return nil, err
	}
	projectRoot := strings.TrimSpace(loaded.Resolution.ProjectRoot)
	rewrite, changed, err := buildContextPackAssuranceCaptureWrite(projectRoot, pack, options.GeneratedAt.Unix())
	if err != nil {
		return nil, err
	}
	if err := applyAssuranceCaptureValidationTransaction(v, projectRoot, rewrite); err != nil {
		return nil, err
	}
	receiptPath := runeContextRelativePath(projectRoot, rewrite.Path)
	receipt := ReceiptArtifact{}
	if err := json.Unmarshal(rewrite.Data, &receipt); err != nil {
		return nil, fmt.Errorf("decode generated assurance receipt: %w", err)
	}
	sortFileMutations(changed)
	return &AssuranceCaptureResult{
		ReceiptPath:  filepath.ToSlash(receiptPath),
		ReceiptID:    receipt.ReceiptID,
		PackHash:     pack.PackHash,
		ChangedFiles: changed,
	}, nil
}

func validateContextPackAssuranceCaptureInputs(v *Validator, loaded *LoadedProject, index *ProjectIndex, options ContextPackAssuranceCaptureOptions) error {
	if v == nil {
		return fmt.Errorf("validator is required")
	}
	if loaded == nil || loaded.Resolution == nil {
		return fmt.Errorf("loaded project is required")
	}
	if index == nil {
		return fmt.Errorf("project index is required")
	}
	if !isVerifiedAssuranceTier(loaded) {
		return fmt.Errorf("assurance_tier must be verified before capturing assurance receipts")
	}
	if options.GeneratedAt.IsZero() {
		return fmt.Errorf("generated_at is required for assurance capture")
	}
	if strings.TrimSpace(loaded.Resolution.ProjectRoot) == "" {
		return fmt.Errorf("project root is unavailable")
	}
	return nil
}

func buildContextPackAssuranceCaptureWrite(projectRoot string, pack *ContextPack, createdAt int64) (fileRewrite, []FileMutation, error) {
	if pack == nil {
		return fileRewrite{}, nil, fmt.Errorf("context-pack is required")
	}
	value := map[string]any{
		"receipt_family":       assuranceReceiptFamilyContextPacks,
		"pack_hash":            pack.PackHash,
		"requested_bundle_ids": append([]string(nil), pack.RequestedBundleIDs...),
		"source_mode":          string(pack.ResolvedFrom.SourceMode),
		"source_ref":           pack.ResolvedFrom.SourceRef,
	}
	rewrites, changed, err := appendCapturedVerifiedReceiptRewrite(
		nil,
		nil,
		projectRoot,
		assuranceReceiptFamilyContextPacks,
		"context-packs/"+pack.PackHash,
		value,
		createdAt,
	)
	if err != nil {
		return fileRewrite{}, nil, err
	}
	if len(rewrites) != 1 {
		return fileRewrite{}, nil, fmt.Errorf("assurance capture produced no receipt writes")
	}
	return rewrites[0], changed, nil
}

func applyAssuranceCaptureValidationTransaction(v *Validator, projectRoot string, rewrite fileRewrite) error {
	return applyFileRewritesTransaction([]fileRewrite{rewrite}, func() error {
		validated, err := v.ValidateProjectWithOptions(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
		if err != nil {
			return err
		}
		_ = validated.Close()
		return nil
	})
}
