package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateStandard(v *Validator, loaded *LoadedProject, options StandardCreateOptions) (*StandardMutationResult, error) {
	if err := validateWritableStandardCommand(v, loaded); err != nil {
		return nil, err
	}
	prepared, err := prepareCreateStandardOptions(options)
	if err != nil {
		return nil, err
	}
	writableRoot, err := writableContentRoot(loaded)
	if err != nil {
		return nil, err
	}
	absPath, err := createStandardAbsolutePath(writableRoot, prepared.path)
	if err != nil {
		return nil, err
	}
	data, err := buildCreateStandardData(v, absPath, prepared, options.Body)
	if err != nil {
		return nil, err
	}
	if err := applyStandardWrite(v, loaded, absPath, data); err != nil {
		return nil, err
	}
	return &StandardMutationResult{
		Path:         prepared.path,
		ChangedFiles: []FileMutation{{Path: runeContextRelativePath(writableRoot, absPath), Action: "created"}},
	}, nil
}

func prepareCreateStandardOptions(options StandardCreateOptions) (preparedCreateStandardOptions, error) {
	title := strings.TrimSpace(options.Title)
	if title == "" {
		return preparedCreateStandardOptions{}, fmt.Errorf("standard create requires --title")
	}
	path, err := normalizeStandardArtifactPath(options.Path)
	if err != nil {
		return preparedCreateStandardOptions{}, err
	}
	status, err := resolvedCreateStandardStatus(options.Status)
	if err != nil {
		return preparedCreateStandardOptions{}, err
	}
	id := resolvedCreateStandardID(options.ID, path)
	if id == "" {
		return preparedCreateStandardOptions{}, fmt.Errorf("standard create could not infer id from %q", path)
	}
	replacedBy, err := normalizeOptionalStandardReplacement(options.ReplacedBy)
	if err != nil {
		return preparedCreateStandardOptions{}, err
	}
	if err := validateStandardReplacementFields(status, replacedBy); err != nil {
		return preparedCreateStandardOptions{}, err
	}
	if replacedBy == path {
		return preparedCreateStandardOptions{}, fmt.Errorf("--replaced-by must not reference the standard itself")
	}
	return preparedCreateStandardOptions{
		path:                    path,
		id:                      id,
		title:                   title,
		status:                  status,
		replacedBy:              replacedBy,
		aliases:                 dedupeStringsInOrder(options.Aliases),
		suggestedContextBundles: dedupeStringsInOrder(options.SuggestedContextBundles),
	}, nil
}

func resolvedCreateStandardStatus(status StandardStatus) (StandardStatus, error) {
	if status == "" {
		status = StandardStatusDraft
	}
	if err := validateStandardStatus(status); err != nil {
		return "", err
	}
	return status, nil
}

func resolvedCreateStandardID(id, path string) string {
	trimmed := strings.TrimSpace(id)
	if trimmed != "" {
		return trimmed
	}
	return standardIDFromPath(path)
}

func createStandardAbsolutePath(writableRoot, standardPath string) (string, error) {
	absPath := filepath.Join(writableRoot, filepath.FromSlash(standardPath))
	if _, err := os.Stat(absPath); err == nil {
		return "", fmt.Errorf("standard %q already exists", standardPath)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return absPath, nil
}

func buildCreateStandardData(v *Validator, absPath string, prepared preparedCreateStandardOptions, body string) ([]byte, error) {
	frontmatter := standardFrontmatter{
		SchemaVersion:           1,
		ID:                      prepared.id,
		Title:                   prepared.title,
		Status:                  prepared.status,
		ReplacedBy:              prepared.replacedBy,
		Aliases:                 prepared.aliases,
		SuggestedContextBundles: prepared.suggestedContextBundles,
	}
	data, err := renderStandardDocument(frontmatter, body)
	if err != nil {
		return nil, err
	}
	if _, err := v.ParseStandard(absPath, data); err != nil {
		return nil, err
	}
	return data, nil
}

func applyStandardWrite(v *Validator, loaded *LoadedProject, absPath string, data []byte) error {
	rewrite := fileRewrite{Path: absPath, Data: data, Perm: 0o644}
	return applyFileRewritesTransaction([]fileRewrite{rewrite}, func() error {
		return validateChangeMutation(v, loaded.Resolution.ProjectRoot)
	})
}
