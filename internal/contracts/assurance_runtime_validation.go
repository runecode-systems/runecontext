package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadProjectAssuranceArtifacts(v *Validator, index *ProjectIndex, projectRoot string, rootConfig map[string]any) error {
	tier := strings.TrimSpace(fmt.Sprint(rootConfig["assurance_tier"]))
	baselinePath := filepath.Join(projectRoot, "runecontext", "assurance", "baseline.yaml")
	if err := loadAssuranceBaselineForTier(v, index, projectRoot, baselinePath, tier); err != nil {
		return err
	}
	if err := ensureAssuranceReceiptsAllowed(projectRoot, tier); err != nil {
		return err
	}
	if err := loadAssuranceReceipts(v, index, projectRoot); err != nil {
		return err
	}
	if err := validateAssuranceLinkage(index, projectRoot); err != nil {
		return err
	}
	return validateAssuranceBackfill(v, index, projectRoot)
}

func loadAssuranceBaselineForTier(v *Validator, index *ProjectIndex, projectRoot, baselinePath, tier string) error {
	baselineExists, err := assuranceFileExists(baselinePath)
	if err != nil {
		return err
	}
	if tier == AssuranceTierVerified && !baselineExists {
		return &ValidationError{Path: baselinePath, Message: "assurance baseline is required when assurance_tier is verified"}
	}
	if !baselineExists {
		return nil
	}
	if tier != AssuranceTierVerified {
		return &ValidationError{Path: baselinePath, Message: "assurance baseline exists but assurance_tier is not verified"}
	}
	return loadAssuranceBaseline(v, index, projectRoot, baselinePath)
}

func ensureAssuranceReceiptsAllowed(projectRoot, tier string) error {
	if tier == AssuranceTierVerified {
		return nil
	}
	for _, family := range assuranceReceiptFamilies {
		receiptsRoot := filepath.Join(projectRoot, "runecontext", "assurance", "receipts", family)
		exists, err := assuranceFileExists(receiptsRoot)
		if err != nil {
			return err
		}
		if exists {
			return &ValidationError{Path: receiptsRoot, Message: "assurance receipts exist but assurance_tier is not verified"}
		}
	}
	return nil
}

func loadAssuranceBaseline(v *Validator, index *ProjectIndex, projectRoot, baselinePath string) error {
	data, err := readProjectFile(projectRoot, baselinePath)
	if err != nil {
		return err
	}
	if err := v.ValidateYAMLFile("assurance-baseline.schema.json", baselinePath, data); err != nil {
		return err
	}
	parsed, err := parseYAML(data)
	if err != nil {
		return &ValidationError{Path: baselinePath, Message: err.Error()}
	}
	obj, err := expectObject(baselinePath, parsed, "assurance baseline")
	if err != nil {
		return err
	}
	var envelope AssuranceEnvelope
	if err := decodeMapIntoStruct(obj, &envelope); err != nil {
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("decode assurance baseline: %v", err)}
	}
	index.AssuranceBaseline = &envelope
	index.AssuranceBaselinePath = filepath.ToSlash(baselinePath)
	index.AssuranceBaselineMap = obj
	return nil
}

func loadAssuranceReceipts(v *Validator, index *ProjectIndex, projectRoot string) error {
	for _, family := range assuranceReceiptFamilies {
		if err := loadAssuranceReceiptFamily(v, index, projectRoot, family); err != nil {
			return err
		}
	}
	return nil
}

func loadAssuranceReceiptFamily(v *Validator, index *ProjectIndex, projectRoot, family string) error {
	receiptsRoot := filepath.Join(projectRoot, "runecontext", "assurance", "receipts", family)
	exists, err := assuranceFileExists(receiptsRoot)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return walkProjectFiles(receiptsRoot, func(path string) error {
		return loadAssuranceReceiptFile(v, index, projectRoot, family, path)
	})
}

func loadAssuranceReceiptFile(v *Validator, index *ProjectIndex, projectRoot, family, path string) error {
	if filepath.Ext(path) != ".json" {
		return nil
	}
	data, err := readProjectFile(projectRoot, path)
	if err != nil {
		return err
	}
	artifact, err := parseAndValidateAssuranceReceipt(v, family, path, data)
	if err != nil {
		return err
	}
	rel := filepath.ToSlash(runeContextRelativePath(projectRoot, path))
	index.AssuranceReceipts[rel] = AssuranceReceiptRecord{Path: rel, Family: family, Artifact: artifact}
	return nil
}

func validateAssuranceLinkage(index *ProjectIndex, projectRoot string) error {
	for _, record := range index.AssuranceReceipts {
		if err := validateAssuranceReceiptLinkageRecord(index, projectRoot, record); err != nil {
			return err
		}
	}
	return nil
}

func validateAssuranceReceiptLinkageRecord(index *ProjectIndex, projectRoot string, record AssuranceReceiptRecord) error {
	path := filepath.Join(projectRoot, filepath.FromSlash(record.Path))
	subject := strings.TrimSpace(record.Artifact.SubjectID)
	switch record.Family {
	case assuranceReceiptFamilyChanges, assuranceReceiptFamilyPromotions, assuranceReceiptFamilyVerifications:
		return validateAssuranceChangeLinkedReceipt(index, path, subject, record.Artifact)
	case assuranceReceiptFamilyContextPacks:
		if strings.HasPrefix(subject, "context-packs/") {
			return nil
		}
		return &ValidationError{Path: path, Message: "assurance context-pack receipt subject_id must be context-packs/<pack-hash>"}
	default:
		return nil
	}
}

func validateAssuranceBackfill(v *Validator, index *ProjectIndex, projectRoot string) error {
	if err := validateAssuranceBackfillArtifacts(v, projectRoot); err != nil {
		return err
	}
	entries, err := baselineImportedEvidenceEntries(index)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := validateImportedEvidenceEntry(index.AssuranceBaselinePath, projectRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func validateAssuranceBackfillArtifacts(v *Validator, projectRoot string) error {
	backfillDir := filepath.Join(projectRoot, "runecontext", "assurance", "backfill")
	exists, err := assuranceFileExists(backfillDir)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return walkProjectFiles(backfillDir, func(path string) error {
		if filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := readProjectFile(projectRoot, path)
		if err != nil {
			return err
		}
		return validateJSONArtifact(v, "assurance-imported-history.schema.json", path, data)
	})
}

func baselineImportedEvidenceEntries(index *ProjectIndex) ([]map[string]any, error) {
	if index.AssuranceBaselineMap == nil {
		return nil, nil
	}
	valueRaw, ok := index.AssuranceBaselineMap["value"]
	if !ok {
		return nil, nil
	}
	value, ok := valueRaw.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: index.AssuranceBaselinePath, Message: "assurance baseline value must be an object"}
	}
	entriesRaw, ok := value["imported_evidence"]
	if !ok || entriesRaw == nil {
		return nil, nil
	}
	entries, ok := entriesRaw.([]any)
	if !ok {
		return nil, &ValidationError{Path: index.AssuranceBaselinePath, Message: "assurance baseline imported_evidence must be a list"}
	}
	result := make([]map[string]any, 0, len(entries))
	for _, rawEntry := range entries {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			return nil, &ValidationError{Path: index.AssuranceBaselinePath, Message: "assurance baseline imported_evidence entries must be objects"}
		}
		result = append(result, entry)
	}
	return result, nil
}

func validateImportedEvidenceEntry(baselinePath, projectRoot string, entry map[string]any) error {
	relPath := strings.TrimSpace(fmt.Sprint(entry["path"]))
	if relPath == "" {
		return &ValidationError{Path: baselinePath, Message: "assurance baseline imported_evidence.path is required"}
	}
	resolved := filepath.Join(projectRoot, filepath.FromSlash(relPath))
	// Resolve symlinks for both the project root and the resolved path to ensure
	// the final target does not escape the repository via symlink traversal.
	realResolved, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("imported_evidence path %q does not exist", relPath)}
		}
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("resolve imported_evidence path %q: %v", relPath, err)}
	}
	realRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("resolve project root: %v", err)}
	}
	rel, err := filepath.Rel(realRoot, realResolved)
	if err != nil {
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("resolve imported_evidence path %q: %v", relPath, err)}
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") {
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("imported_evidence path %q escapes repository root", relPath)}
	}
	if _, err := os.Stat(realResolved); err != nil {
		return &ValidationError{Path: baselinePath, Message: fmt.Sprintf("imported_evidence path %q does not exist", relPath)}
	}
	return nil
}
