package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func UpdateStandard(v *Validator, loaded *LoadedProject, options StandardUpdateOptions) (*StandardMutationResult, error) {
	if err := validateWritableStandardCommand(v, loaded); err != nil {
		return nil, err
	}
	standardPath, writableRoot, absPath, err := prepareStandardUpdateTarget(loaded, options)
	if err != nil {
		return nil, err
	}
	previous, doc, err := loadExistingStandard(v, absPath, standardPath)
	if err != nil {
		return nil, err
	}
	updated, err := buildUpdatedStandardFrontmatter(absPath, standardPath, doc, options)
	if err != nil {
		return nil, err
	}
	newData, err := renderStandardDocument(updated, doc.Body)
	if err != nil {
		return nil, err
	}
	if standardDocumentUnchanged(previous, newData) {
		return &StandardMutationResult{Path: standardPath}, nil
	}
	if _, err := v.ParseStandard(absPath, newData); err != nil {
		return nil, err
	}
	if err := applyStandardWrite(v, loaded, absPath, newData); err != nil {
		return nil, err
	}
	return &StandardMutationResult{
		Path:         standardPath,
		ChangedFiles: []FileMutation{{Path: runeContextRelativePath(writableRoot, absPath), Action: "updated"}},
	}, nil
}

func prepareStandardUpdateTarget(loaded *LoadedProject, options StandardUpdateOptions) (string, string, string, error) {
	standardPath, err := normalizeStandardArtifactPath(options.Path)
	if err != nil {
		return "", "", "", err
	}
	if err := validateStandardUpdateFlags(options); err != nil {
		return "", "", "", err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return "", "", "", err
	}
	absPath := filepath.Join(writableRoot, filepath.FromSlash(standardPath))
	return standardPath, writableRoot, absPath, nil
}

func loadExistingStandard(v *Validator, absPath, standardPath string) ([]byte, *FrontmatterDocument, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("standard %q does not exist", standardPath)
		}
		return nil, nil, err
	}
	doc, err := v.ParseStandard(absPath, data)
	if err != nil {
		return nil, nil, err
	}
	return data, doc, nil
}

func standardDocumentUnchanged(previous, next []byte) bool {
	left := strings.ReplaceAll(string(previous), "\r\n", "\n")
	right := strings.ReplaceAll(string(next), "\r\n", "\n")
	return left == right
}

func validateStandardUpdateFlags(options StandardUpdateOptions) error {
	if options.ReplacedBy != "" && options.ClearReplacedBy {
		return fmt.Errorf("--replaced-by and --clear-replaced-by cannot be used together")
	}
	if hasAnyStandardUpdateMutation(options) {
		return nil
	}
	return fmt.Errorf("standard update requires at least one mutation flag")
}

func hasAnyStandardUpdateMutation(options StandardUpdateOptions) bool {
	if options.ReplaceAliases || options.ReplaceSuggestedContextBundles || options.ClearReplacedBy {
		return true
	}
	for _, value := range []string{options.Title, options.Status, options.ReplacedBy} {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func buildUpdatedStandardFrontmatter(path, standardPath string, doc *FrontmatterDocument, options StandardUpdateOptions) (standardFrontmatter, error) {
	current, err := buildStandardRecord(path, doc)
	if err != nil {
		return standardFrontmatter{}, err
	}
	updated := standardFrontmatterFromRecord(current)
	if err := applyStandardUpdateTextFields(&updated, options); err != nil {
		return standardFrontmatter{}, err
	}
	applyStandardUpdateListFields(&updated, options)
	if err := applyStandardUpdateReplacement(&updated, standardPath, options); err != nil {
		return standardFrontmatter{}, err
	}
	return updated, nil
}

func standardFrontmatterFromRecord(record *StandardRecord) standardFrontmatter {
	return standardFrontmatter{
		SchemaVersion:           1,
		ID:                      record.ID,
		Title:                   record.Title,
		Status:                  record.Status,
		ReplacedBy:              record.ReplacedBy,
		Aliases:                 append([]string(nil), record.Aliases...),
		SuggestedContextBundles: append([]string(nil), record.SuggestedContextBundles...),
	}
}

func applyStandardUpdateTextFields(updated *standardFrontmatter, options StandardUpdateOptions) error {
	if title := strings.TrimSpace(options.Title); title != "" {
		updated.Title = title
	}
	if rawStatus := strings.TrimSpace(options.Status); rawStatus != "" {
		updated.Status = StandardStatus(rawStatus)
		if err := validateStandardStatus(updated.Status); err != nil {
			return err
		}
	}
	return nil
}

func applyStandardUpdateListFields(updated *standardFrontmatter, options StandardUpdateOptions) {
	if options.ReplaceAliases {
		updated.Aliases = dedupeStringsInOrder(options.Aliases)
	}
	if options.ReplaceSuggestedContextBundles {
		updated.SuggestedContextBundles = dedupeStringsInOrder(options.SuggestedContextBundles)
	}
}

func applyStandardUpdateReplacement(updated *standardFrontmatter, standardPath string, options StandardUpdateOptions) error {
	if options.ClearReplacedBy {
		updated.ReplacedBy = ""
	}
	if replacedBy := strings.TrimSpace(options.ReplacedBy); replacedBy != "" {
		normalized, err := normalizeOptionalStandardReplacement(replacedBy)
		if err != nil {
			return err
		}
		updated.ReplacedBy = normalized
	}
	if err := validateStandardReplacementFields(updated.Status, updated.ReplacedBy); err != nil {
		return err
	}
	if updated.ReplacedBy == standardPath {
		return fmt.Errorf("--replaced-by must not reference the standard itself")
	}
	return nil
}
