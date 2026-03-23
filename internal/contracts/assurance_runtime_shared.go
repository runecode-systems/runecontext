package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	AssuranceTierVerified = "verified"

	assuranceReceiptFamilyContextPacks  = "context-packs"
	assuranceReceiptFamilyChanges       = "changes"
	assuranceReceiptFamilyPromotions    = "promotions"
	assuranceReceiptFamilyVerifications = "verifications"
)

var assuranceReceiptFamilies = []string{
	assuranceReceiptFamilyContextPacks,
	assuranceReceiptFamilyChanges,
	assuranceReceiptFamilyPromotions,
	assuranceReceiptFamilyVerifications,
}

type AssuranceReceiptRecord struct {
	Path     string
	Family   string
	Artifact ReceiptArtifact
}

func isVerifiedAssuranceTier(loaded *LoadedProject) bool {
	if loaded == nil {
		return false
	}
	return strings.TrimSpace(fmt.Sprint(loaded.RootConfig["assurance_tier"])) == AssuranceTierVerified
}

func IsVerifiedAssuranceTierForCLI(loaded *LoadedProject) bool {
	return isVerifiedAssuranceTier(loaded)
}

func appendCapturedVerifiedReceiptRewrite(rewrites []fileRewrite, changedFiles []FileMutation, projectRoot, family, subjectID string, value map[string]any, createdAt int64) ([]fileRewrite, []FileMutation, error) {
	family = strings.TrimSpace(family)
	if family == "" {
		return nil, nil, fmt.Errorf("receipt family is required")
	}
	artifact, filename, err := BuildCapturedVerifiedReceipt(family, subjectID, value, createdAt)
	if err != nil {
		return nil, nil, err
	}
	receiptData, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal assurance receipt: %w", err)
	}
	receiptData = append(receiptData, '\n')
	receiptPath := filepath.Join(projectRoot, "assurance", "receipts", family, filename)
	rewrites = append(rewrites, fileRewrite{Path: receiptPath, Data: receiptData, Perm: 0o644})
	changedFiles = append(changedFiles, FileMutation{Path: runeContextRelativePath(projectRoot, receiptPath), Action: "created"})
	return rewrites, changedFiles, nil
}

func assuranceFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, &ValidationError{Path: path, Message: err.Error()}
}

func decodeMapIntoStruct(value any, target any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func parseJSONObject(path string, data []byte) (map[string]any, error) {
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("parse json: %v", err)}
	}
	if obj == nil {
		return nil, &ValidationError{Path: path, Message: "JSON document must be an object"}
	}
	return obj, nil
}

func validateJSONArtifact(v *Validator, schemaName, path string, data []byte) error {
	obj, err := parseJSONObject(path, data)
	if err != nil {
		return err
	}
	return v.ValidateValue(schemaName, path, obj)
}

func parseAndValidateAssuranceReceipt(v *Validator, family, path string, data []byte) (ReceiptArtifact, error) {
	if err := validateJSONArtifact(v, "assurance-receipt.schema.json", path, data); err != nil {
		return ReceiptArtifact{}, err
	}
	obj, err := parseJSONObject(path, data)
	if err != nil {
		return ReceiptArtifact{}, err
	}
	var artifact ReceiptArtifact
	if err := decodeMapIntoStruct(obj, &artifact); err != nil {
		return ReceiptArtifact{}, &ValidationError{Path: path, Message: fmt.Sprintf("decode assurance receipt: %v", err)}
	}
	if err := validateAssuranceReceiptPathAndFamily(path, family, artifact); err != nil {
		return ReceiptArtifact{}, err
	}
	hash, err := ComputeReceiptHash(artifact)
	if err != nil {
		return ReceiptArtifact{}, &ValidationError{Path: path, Message: err.Error()}
	}
	if hash != artifact.ReceiptHash {
		return ReceiptArtifact{}, &ValidationError{Path: path, Message: "assurance receipt_hash does not match canonical receipt content"}
	}
	return artifact, nil
}

func validateAssuranceReceiptPathAndFamily(path, family string, artifact ReceiptArtifact) error {
	valueMap, ok := artifact.Value.(map[string]any)
	if !ok {
		return &ValidationError{Path: path, Message: "assurance receipt value must be an object"}
	}
	receiptFamily := strings.TrimSpace(fmt.Sprint(valueMap["receipt_family"]))
	if receiptFamily == "" {
		return &ValidationError{Path: path, Message: "assurance receipt value.receipt_family is required"}
	}
	if receiptFamily != family {
		return &ValidationError{Path: path, Message: fmt.Sprintf("assurance receipt family mismatch: value.receipt_family=%q path-family=%q", receiptFamily, family)}
	}
	if artifact.Provenance != "captured_verified" {
		return &ValidationError{Path: path, Message: "assurance receipt provenance must be captured_verified"}
	}
	return nil
}
