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
	idOwners, err := collectStandardIDOwners(index)
	if err != nil {
		return err
	}
	return validateStandardMetadataRecords(index, idOwners)
}

func validateChangeStandardReferences(index *ProjectIndex) error {
	for _, id := range SortedKeys(index.Changes) {
		record := index.Changes[id]
		standardsPath, err := changeArtifactRelativePath(index, record, "standards.md")
		if err != nil {
			return err
		}
		if err := validateChangeStandardSection(index, standardsPath, "Applicable Standards", record.ApplicableStandards); err != nil {
			return err
		}
		if err := validateChangeStandardSection(index, standardsPath, "Standards Added Since Last Refresh", record.AddedStandards); err != nil {
			return err
		}
		if err := validateExcludedStandardReferences(index, standardsPath, record.ExcludedStandards); err != nil {
			return err
		}
	}
	return nil
}

func changeArtifactRelativePath(index *ProjectIndex, record *ChangeRecord, name string) (string, error) {
	if index == nil || record == nil {
		return "", fmt.Errorf("change artifact path requires index and record")
	}
	rel, err := filepath.Rel(index.ContentRoot, filepath.Join(record.DirPath, name))
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
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
		if err := validateBundleResolutionStandards(index, bundleID); err != nil {
			return err
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
