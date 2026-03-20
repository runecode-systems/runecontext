package contracts

import "fmt"

func collectStandardIDOwners(index *ProjectIndex) (map[string]string, error) {
	idOwners := map[string]string{}
	for _, path := range SortedKeys(index.Standards) {
		record := index.Standards[path]
		if owner, ok := idOwners[record.ID]; ok {
			return nil, &ValidationError{Path: record.Path, Message: fmt.Sprintf("standard id %q is duplicated (already declared in %s)", record.ID, owner)}
		}
		idOwners[record.ID] = record.Path
	}
	return idOwners, nil
}

func validateStandardMetadataRecords(index *ProjectIndex, idOwners map[string]string) error {
	aliasOwners := map[string]string{}
	for _, path := range SortedKeys(index.Standards) {
		if err := validateStandardMetadataRecord(index, index.Standards[path], idOwners, aliasOwners); err != nil {
			return err
		}
	}
	return nil
}

func validateStandardMetadataRecord(index *ProjectIndex, record *StandardRecord, idOwners, aliasOwners map[string]string) error {
	appendDeprecatedStandardWarning(index, record)
	if err := validateStandardMetadataDuplicates(record); err != nil {
		return err
	}
	if err := validateStandardAliases(record, idOwners, aliasOwners); err != nil {
		return err
	}
	return validateStandardReplacement(index, record)
}

func appendDeprecatedStandardWarning(index *ProjectIndex, record *StandardRecord) {
	if record.Status == StandardStatusDeprecated && record.ReplacedBy == "" {
		appendValidationDiagnostic(index, ValidationDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_missing_replacement", Message: fmt.Sprintf("deprecated standard %q does not declare replaced_by guidance", record.Path), Path: record.Path})
	}
}

func validateStandardMetadataDuplicates(record *StandardRecord) error {
	if duplicate, ok := duplicateString(record.Aliases); ok {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("aliases contains duplicate value %q", duplicate)}
	}
	if duplicate, ok := duplicateString(record.SuggestedContextBundles); ok {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("suggested_context_bundles contains duplicate value %q", duplicate)}
	}
	return nil
}

func validateStandardAliases(record *StandardRecord, idOwners, aliasOwners map[string]string) error {
	for _, alias := range record.Aliases {
		if err := validateStandardAlias(record, alias, idOwners, aliasOwners); err != nil {
			return err
		}
		aliasOwners[alias] = record.Path
	}
	return nil
}

func validateStandardAlias(record *StandardRecord, alias string, idOwners, aliasOwners map[string]string) error {
	if alias == record.ID {
		return &ValidationError{Path: record.Path, Message: "aliases must not repeat the standard's current id"}
	}
	if owner, ok := idOwners[alias]; ok {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("alias %q collides with standard id owned by %s", alias, owner)}
	}
	if owner, ok := aliasOwners[alias]; ok {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("alias %q is duplicated (already declared in %s)", alias, owner)}
	}
	return nil
}

func validateStandardReplacement(index *ProjectIndex, record *StandardRecord) error {
	if record.ReplacedBy == "" {
		return nil
	}
	if record.Status != StandardStatusDeprecated {
		return &ValidationError{Path: record.Path, Message: "only deprecated standards may set replaced_by"}
	}
	if !isCanonicalStandardPathRef(record.ReplacedBy) {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("replaced_by %q must use the canonical standards/<path>.md reference form", record.ReplacedBy)}
	}
	if record.ReplacedBy == record.Path {
		return &ValidationError{Path: record.Path, Message: "replaced_by must not reference the standard itself"}
	}
	if _, ok := index.StandardPaths[record.ReplacedBy]; !ok {
		return &ValidationError{Path: record.Path, Message: fmt.Sprintf("replaced_by references missing standard %q", record.ReplacedBy)}
	}
	return nil
}

func validateChangeStandardSection(index *ProjectIndex, standardsPath, section string, refs []string) error {
	for _, ref := range refs {
		if err := validateSelectedStandardReference(index, standardsPath, section, ref); err != nil {
			return err
		}
		appendDeprecatedStandardReferenceWarning(index, standardsPath, ref)
	}
	return nil
}

func appendDeprecatedStandardReferenceWarning(index *ProjectIndex, standardsPath, ref string) {
	standard, _ := lookupStandard(index, standardsPath, ref)
	if standard == nil || standard.Status != StandardStatusDeprecated {
		return
	}
	message := fmt.Sprintf("standards.md references deprecated standard %q", ref)
	if standard.ReplacedBy != "" {
		message = fmt.Sprintf("%s; consider %q", message, standard.ReplacedBy)
	}
	appendValidationDiagnostic(index, ValidationDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_referenced", Message: message, Path: standardsPath})
}

func validateExcludedStandardReferences(index *ProjectIndex, standardsPath string, refs []string) error {
	for _, ref := range refs {
		if _, err := lookupStandard(index, standardsPath, ref); err != nil {
			return err
		}
	}
	return nil
}

func validateBundleResolutionStandards(index *ProjectIndex, bundleID string) error {
	resolution := index.Bundles.resolutions[bundleID]
	aspectResolution, ok := resolution.Aspects[BundleAspectStandards]
	if !ok {
		return nil
	}
	for _, entry := range aspectResolution.Selected {
		if err := validateBundleSelectedStandard(index, bundleID, entry); err != nil {
			return err
		}
	}
	return nil
}

func validateBundleSelectedStandard(index *ProjectIndex, bundleID string, entry BundleInventoryEntry) error {
	standard := index.Standards[entry.Path]
	if standard == nil {
		return nil
	}
	sourcePath := bundleSelectionSourcePath(index, entry)
	if standard.Status == StandardStatusDraft {
		if sourcePath == "" {
			sourcePath = entry.Path
		}
		return &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q selects draft standard %q", bundleID, entry.Path)}
	}
	if standard.Status == StandardStatusDeprecated {
		appendDeprecatedBundleSelectionWarning(index, bundleID, entry, standard)
	}
	return nil
}

func bundleSelectionSourcePath(index *ProjectIndex, entry BundleInventoryEntry) string {
	if definition := index.Bundles.bundles[entry.FinalRule.Bundle]; definition != nil {
		return definition.Path
	}
	return ""
}

func appendDeprecatedBundleSelectionWarning(index *ProjectIndex, bundleID string, entry BundleInventoryEntry, standard *StandardRecord) {
	message := fmt.Sprintf("bundle %q selects deprecated standard %q", bundleID, entry.Path)
	if standard.ReplacedBy != "" {
		message = fmt.Sprintf("%s; consider %q", message, standard.ReplacedBy)
	}
	index.Bundles.appendDiagnostic(bundleID, BundleDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_selected", Message: message, Bundle: bundleID, Aspect: BundleAspectStandards, Rule: entry.FinalRule.Rule, Pattern: entry.FinalRule.Pattern, Matches: []string{entry.Path}})
}
