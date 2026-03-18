package contracts

import (
	"fmt"
	"path/filepath"
)

type StandardStatus string

const (
	StandardStatusDraft      StandardStatus = "draft"
	StandardStatusActive     StandardStatus = "active"
	StandardStatusDeprecated StandardStatus = "deprecated"
)

type StandardRecord struct {
	Path                    string
	ID                      string
	Title                   string
	Status                  StandardStatus
	ReplacedBy              string
	Aliases                 []string
	SuggestedContextBundles []string
}

func buildStandardRecord(path string, doc *FrontmatterDocument) (*StandardRecord, error) {
	aliases, err := stringSliceField(path, "aliases", doc.Frontmatter["aliases"])
	if err != nil {
		return nil, err
	}
	suggestedBundles, err := stringSliceField(path, "suggested_context_bundles", doc.Frontmatter["suggested_context_bundles"])
	if err != nil {
		return nil, err
	}
	status := StandardStatusActive
	if rawStatus, ok := doc.Frontmatter["status"]; ok && rawStatus != nil {
		status = StandardStatus(fmt.Sprint(rawStatus))
	}
	replacedBy := ""
	if rawReplacedBy, ok := doc.Frontmatter["replaced_by"]; ok && rawReplacedBy != nil {
		replacedBy = fmt.Sprint(rawReplacedBy)
	}
	return &StandardRecord{
		Path:                    path,
		ID:                      fmt.Sprint(doc.Frontmatter["id"]),
		Title:                   fmt.Sprint(doc.Frontmatter["title"]),
		Status:                  status,
		ReplacedBy:              replacedBy,
		Aliases:                 aliases,
		SuggestedContextBundles: suggestedBundles,
	}, nil
}

func validateStandardMetadata(index *ProjectIndex) error {
	idOwners := map[string]string{}
	for _, path := range SortedKeys(index.Standards) {
		record := index.Standards[path]
		if owner, ok := idOwners[record.ID]; ok {
			return &ValidationError{Path: record.Path, Message: fmt.Sprintf("standard id %q is duplicated (already declared in %s)", record.ID, owner)}
		}
		idOwners[record.ID] = record.Path
	}
	aliasOwners := map[string]string{}
	for _, path := range SortedKeys(index.Standards) {
		record := index.Standards[path]
		if record.Status == StandardStatusDeprecated && record.ReplacedBy == "" {
			appendValidationDiagnostic(index, ValidationDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_missing_replacement", Message: fmt.Sprintf("deprecated standard %q does not declare replaced_by guidance", record.Path), Path: record.Path})
		}
		if duplicate, ok := duplicateString(record.Aliases); ok {
			return &ValidationError{Path: record.Path, Message: fmt.Sprintf("aliases contains duplicate value %q", duplicate)}
		}
		if duplicate, ok := duplicateString(record.SuggestedContextBundles); ok {
			return &ValidationError{Path: record.Path, Message: fmt.Sprintf("suggested_context_bundles contains duplicate value %q", duplicate)}
		}
		for _, alias := range record.Aliases {
			if alias == record.ID {
				return &ValidationError{Path: record.Path, Message: "aliases must not repeat the standard's current id"}
			}
			if owner, ok := idOwners[alias]; ok {
				return &ValidationError{Path: record.Path, Message: fmt.Sprintf("alias %q collides with standard id owned by %s", alias, owner)}
			}
			if owner, ok := aliasOwners[alias]; ok {
				return &ValidationError{Path: record.Path, Message: fmt.Sprintf("alias %q is duplicated (already declared in %s)", alias, owner)}
			}
			aliasOwners[alias] = record.Path
		}
		if record.ReplacedBy == "" {
			continue
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
	}
	return nil
}

func validateChangeStandardReferences(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		standardsPath := filepath.Join(record.DirPath, "standards.md")
		for _, ref := range record.ApplicableStandards {
			if err := validateSelectedStandardReference(index, standardsPath, "Applicable Standards", ref); err != nil {
				return err
			}
			if standard, _ := lookupStandard(index, standardsPath, ref); standard != nil && standard.Status == StandardStatusDeprecated {
				message := fmt.Sprintf("standards.md references deprecated standard %q", ref)
				if standard.ReplacedBy != "" {
					message = fmt.Sprintf("%s; consider %q", message, standard.ReplacedBy)
				}
				appendValidationDiagnostic(index, ValidationDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_referenced", Message: message, Path: standardsPath})
			}
		}
		for _, ref := range record.AddedStandards {
			if err := validateSelectedStandardReference(index, standardsPath, "Standards Added Since Last Refresh", ref); err != nil {
				return err
			}
			if standard, _ := lookupStandard(index, standardsPath, ref); standard != nil && standard.Status == StandardStatusDeprecated {
				message := fmt.Sprintf("standards.md references deprecated standard %q", ref)
				if standard.ReplacedBy != "" {
					message = fmt.Sprintf("%s; consider %q", message, standard.ReplacedBy)
				}
				appendValidationDiagnostic(index, ValidationDiagnostic{Severity: DiagnosticSeverityWarning, Code: "deprecated_standard_referenced", Message: message, Path: standardsPath})
			}
		}
		for _, ref := range record.ExcludedStandards {
			if _, err := lookupStandard(index, standardsPath, ref); err != nil {
				return err
			}
		}
	}
	return nil
}

func lookupStandard(index *ProjectIndex, path, ref string) (*StandardRecord, error) {
	standard := index.Standards[ref]
	if standard == nil {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("standards.md references missing standard %q", ref)}
	}
	return standard, nil
}

func validateSelectedStandardReference(index *ProjectIndex, path, section, ref string) error {
	standard, err := lookupStandard(index, path, ref)
	if err != nil {
		return err
	}
	if standard.Status == StandardStatusDraft {
		return &ValidationError{Path: path, Message: fmt.Sprintf("standards.md must not reference draft standard %q in section %q", ref, section)}
	}
	return nil
}

func validateBundleStandardSelections(index *ProjectIndex) error {
	if index == nil || index.Bundles == nil {
		return nil
	}
	for _, bundleID := range SortedKeys(index.Bundles.resolutions) {
		resolution := index.Bundles.resolutions[bundleID]
		aspectResolution, ok := resolution.Aspects[BundleAspectStandards]
		if !ok {
			continue
		}
		for _, entry := range aspectResolution.Selected {
			standard := index.Standards[entry.Path]
			if standard == nil {
				continue
			}
			sourcePath := ""
			if definition := index.Bundles.bundles[entry.FinalRule.Bundle]; definition != nil {
				sourcePath = definition.Path
			}
			switch standard.Status {
			case StandardStatusDraft:
				if sourcePath == "" {
					sourcePath = entry.Path
				}
				return &ValidationError{Path: sourcePath, Message: fmt.Sprintf("bundle %q selects draft standard %q", bundleID, entry.Path)}
			case StandardStatusDeprecated:
				message := fmt.Sprintf("bundle %q selects deprecated standard %q", bundleID, entry.Path)
				if standard.ReplacedBy != "" {
					message = fmt.Sprintf("%s; consider %q", message, standard.ReplacedBy)
				}
				index.Bundles.appendDiagnostic(bundleID, BundleDiagnostic{
					Severity: DiagnosticSeverityWarning,
					Code:     "deprecated_standard_selected",
					Message:  message,
					Bundle:   bundleID,
					Aspect:   BundleAspectStandards,
					Rule:     entry.FinalRule.Rule,
					Pattern:  entry.FinalRule.Pattern,
					Matches:  []string{entry.Path},
				})
			}
		}
	}
	return nil
}

func appendValidationDiagnostic(index *ProjectIndex, diagnostic ValidationDiagnostic) {
	if index == nil {
		return
	}
	index.Diagnostics = append(index.Diagnostics, diagnostic)
}
