package contracts

import (
	"fmt"
	"strings"
)

func validateAssuranceChangeLinkedReceipt(index *ProjectIndex, path, subject string, artifact ReceiptArtifact) error {
	changeID, ok := strings.CutPrefix(subject, "changes/")
	if !ok || strings.TrimSpace(changeID) == "" {
		return &ValidationError{Path: path, Message: "assurance receipt subject_id must be changes/<change-id>"}
	}
	if _, exists := index.ChangeIDs[changeID]; exists {
		// Verify the artifact value contains a matching change_id for change-linked receipts
		valueMap, ok := artifact.Value.(map[string]any)
		if !ok {
			return &ValidationError{Path: path, Message: "assurance receipt value must be an object"}
		}
		valChangeID := strings.TrimSpace(fmt.Sprint(valueMap["change_id"]))
		if valChangeID == "" {
			return &ValidationError{Path: path, Message: "assurance receipt value.change_id is required for change-linked receipts"}
		}
		if valChangeID != changeID {
			return &ValidationError{Path: path, Message: fmt.Sprintf("assurance receipt change_id mismatch: value.change_id=%q subject change=%q", valChangeID, changeID)}
		}
		return nil
	}
	return &ValidationError{Path: path, Message: fmt.Sprintf("assurance receipt subject references missing change %q", changeID)}
}
