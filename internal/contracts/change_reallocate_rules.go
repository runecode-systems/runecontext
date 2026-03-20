package contracts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ensureChangeReallocationIsLocalOnly(index *ProjectIndex, changeID string) error {
	if index == nil {
		return fmt.Errorf("project index is required")
	}
	if err := ensureChangeStatusReferencesAreLocal(index, changeID); err != nil {
		return err
	}
	if err := ensureSpecReferencesAreLocal(index, changeID); err != nil {
		return err
	}
	if err := ensureDecisionReferencesAreLocal(index, changeID); err != nil {
		return err
	}
	return ensureMarkdownReferencesAreLocal(index, changeID)
}

func ensureChangeStatusReferencesAreLocal(index *ProjectIndex, changeID string) error {
	for _, otherID := range SortedKeys(index.Changes) {
		if otherID == changeID {
			continue
		}
		if err := ensureRecordDoesNotReferenceChange(index, index.Changes[otherID], changeID); err != nil {
			return err
		}
	}
	return nil
}

func ensureRecordDoesNotReferenceChange(index *ProjectIndex, record *ChangeRecord, changeID string) error {
	for _, field := range changeReferenceFields(record) {
		if containsString(field.items, changeID) {
			return fmt.Errorf("change %q cannot be reallocated because %s in %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, field.name, runeContextRelativePath(index.ContentRoot, record.StatusPath))
		}
	}
	return nil
}

type changeReferenceField struct {
	name  string
	items []string
}

func changeReferenceFields(record *ChangeRecord) []changeReferenceField {
	return []changeReferenceField{{"related_changes", record.RelatedChanges}, {"depends_on", record.DependsOn}, {"informed_by", record.InformedBy}, {"supersedes", record.Supersedes}, {"superseded_by", record.SupersededBy}}
}

func ensureSpecReferencesAreLocal(index *ProjectIndex, changeID string) error {
	for _, specPath := range SortedKeys(index.Specs) {
		spec := index.Specs[specPath]
		if containsString(spec.OriginatingChanges, changeID) || containsString(spec.RevisedByChanges, changeID) {
			return fmt.Errorf("change %q cannot be reallocated because spec %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, specPath)
		}
	}
	return nil
}

func ensureDecisionReferencesAreLocal(index *ProjectIndex, changeID string) error {
	for _, decisionPath := range SortedKeys(index.Decisions) {
		decision := index.Decisions[decisionPath]
		if containsString(decision.OriginatingChanges, changeID) || containsString(decision.RelatedChanges, changeID) {
			return fmt.Errorf("change %q cannot be reallocated because decision %q references it; alpha.3 reallocation only rewrites local references inside the change", changeID, decisionPath)
		}
	}
	return nil
}

func ensureMarkdownReferencesAreLocal(index *ProjectIndex, changeID string) error {
	changePrefix := changeMarkdownPathPrefix(changeID)
	for _, path := range SortedKeys(index.MarkdownFiles) {
		if strings.HasPrefix(path, changePrefix) {
			continue
		}
		for _, ref := range index.MarkdownFiles[path].Refs {
			if strings.HasPrefix(ref.Path, changePrefix) {
				return fmt.Errorf("change %q cannot be reallocated because markdown deep ref %q in %q points into it; alpha.3 reallocation only rewrites local references inside the change", changeID, ref.Raw, path)
			}
		}
	}
	return nil
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
	if err := renamePath(backupDir, oldChangeDir); err != nil {
		return fmt.Errorf("restore original change %q from backup %q: %v", filepath.ToSlash(oldChangeDir), filepath.ToSlash(backupDir), err)
	}
	return nil
}

func combineReallocationRollbackError(operationErr, rollbackErr error) error {
	if rollbackErr == nil {
		return operationErr
	}
	return fmt.Errorf("%v; rollback also failed and manual recovery may be required: %v", operationErr, rollbackErr)
}
