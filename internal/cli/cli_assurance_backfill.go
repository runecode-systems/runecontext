package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type assuranceBackfillResult struct {
	baselinePath   string
	historyPath    string
	adoptionCommit string
	commitCount    int
	importedAdded  bool
}

func executeAssuranceBackfill(root string) (assuranceBackfillResult, error) {
	context, baselineMap, adoptionCommit, err := loadBackfillInputs(root)
	if err != nil {
		return assuranceBackfillResult{}, err
	}

	// Avoid rebuilding the history if the baseline already references an
	// existing artifact and the file is present on disk. This keeps backfill
	// reruns idempotent and prevents trivial updates to generated_at.
	if skip, existing := shouldSkipRebuild(root, baselineMap, adoptionCommit); skip {
		return assuranceBackfillResult{
			baselinePath:   context.baselinePath,
			historyPath:    existing,
			adoptionCommit: adoptionCommit,
			commitCount:    0,
			importedAdded:  false,
		}, nil
	}
	// Otherwise build and write the history as before.
	historyPath, commitCount, err := buildAndWriteImportedHistory(root, adoptionCommit)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	baselineUpdated, err := appendImportedEvidenceAndWriteBaseline(root, context.baselinePath, baselineMap, historyPath)
	if err != nil {
		return assuranceBackfillResult{}, err
	}
	return assuranceBackfillResult{
		baselinePath:   context.baselinePath,
		historyPath:    historyPath,
		adoptionCommit: adoptionCommit,
		commitCount:    commitCount,
		importedAdded:  baselineUpdated,
	}, nil
}

func buildAndWriteImportedHistory(root, adoptionCommit string) (string, int, error) {
	history, err := buildImportedGitHistory(root, adoptionCommit)
	if err != nil {
		return "", 0, err
	}
	historyPath, err := writeImportedGitHistory(root, adoptionCommit, history)
	if err != nil {
		return "", 0, err
	}
	return historyPath, len(history), nil
}

func appendImportedEvidenceAndWriteBaseline(root, baselinePath string, baselineMap map[string]any, historyPath string) (bool, error) {
	relativeHistoryPath, err := backfillRelativePathWithinRoot(root, historyPath)
	if err != nil {
		return false, err
	}
	baselineUpdated, err := appendImportedEvidenceToBaseline(baselineMap, relativeHistoryPath)
	if err != nil {
		return false, err
	}
	if !baselineUpdated {
		return false, nil
	}
	updatedBaseline, err := yaml.Marshal(baselineMap)
	if err != nil {
		return false, fmt.Errorf("marshal updated baseline: %w", err)
	}
	if err := writeAtomicFile(baselinePath, updatedBaseline, 0o644); err != nil {
		return false, fmt.Errorf("write updated baseline: %w", err)
	}
	return true, nil
}

func loadAssuranceBaseline(path string) (contracts.AssuranceEnvelope, map[string]any, error) {
	baselineData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("assurance baseline not found: %s", path)
		}
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("read assurance baseline: %w", err)
	}
	var envelope contracts.AssuranceEnvelope
	if err := yaml.Unmarshal(baselineData, &envelope); err != nil {
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("parse assurance baseline: %w", err)
	}
	var baselineMap map[string]any
	if err := yaml.Unmarshal(baselineData, &baselineMap); err != nil {
		return contracts.AssuranceEnvelope{}, nil, fmt.Errorf("parse baseline object: %w", err)
	}
	return envelope, baselineMap, nil
}

func baselineAdoptionCommit(envelope contracts.AssuranceEnvelope) (string, error) {
	value, ok := envelope.Value.(map[string]any)
	if !ok {
		return "", fmt.Errorf("assurance baseline value must be an object")
	}
	adoptionCommit := readOptionalString(value, "adoption_commit")
	if adoptionCommit == "" {
		return "", fmt.Errorf("assurance baseline adoption_commit is required for backfill")
	}
	// Require a canonical lowercase 40-char hex SHA to avoid path-traversal
	// and ambiguity; this matches existing git-source validation elsewhere.
	if !isCanonicalLowerHex40(adoptionCommit) {
		return "", fmt.Errorf("assurance baseline adoption_commit must be a canonical lowercase 40-char hex SHA")
	}
	return adoptionCommit, nil
}

func appendImportedEvidenceToBaseline(baseline map[string]any, historyPath string) (bool, error) {
	valueRaw, ok := baseline["value"]
	if !ok || valueRaw == nil {
		valueRaw = map[string]any{}
	}
	value, ok := valueRaw.(map[string]any)
	if !ok {
		return false, fmt.Errorf("assurance baseline value must be an object")
	}
	evidenceRaw, ok := value["imported_evidence"]
	if !ok || evidenceRaw == nil {
		evidenceRaw = make([]any, 0)
	}
	evidence, ok := evidenceRaw.([]any)
	if !ok {
		return false, fmt.Errorf("assurance baseline imported_evidence must be a list")
	}
	if importedEvidenceExists(evidence, historyPath) {
		return false, nil
	}
	evidence = append(evidence, map[string]any{
		"provenance": "imported_git_history",
		"path":       filepath.ToSlash(historyPath),
	})
	value["imported_evidence"] = evidence
	baseline["value"] = value
	return true, nil
}

func importedEvidenceExists(evidence []any, historyPath string) bool {
	for _, raw := range evidence {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if readOptionalString(entry, "provenance") != "imported_git_history" {
			continue
		}
		if filepath.ToSlash(readOptionalString(entry, "path")) == filepath.ToSlash(historyPath) {
			return true
		}
	}
	return false
}

func isCanonicalLowerHex40(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') || ('a' <= r && r <= 'f')) {
			return false
		}
	}
	return true
}

// loadBackfillInputs loads and validates the baseline, adoption commit, and
// assurance context needed for backfill operations.
func loadBackfillInputs(root string) (*assuranceEnableContext, map[string]any, string, error) {
	context, err := newAssuranceEnableContext(root)
	if err != nil {
		return nil, nil, "", err
	}
	if fmt.Sprint(context.rootCfg["assurance_tier"]) != "verified" {
		return nil, nil, "", fmt.Errorf("assurance_tier must be verified before running backfill")
	}
	baselineEnvelope, baselineMap, err := loadAssuranceBaseline(context.baselinePath)
	if err != nil {
		return nil, nil, "", err
	}
	adoptionCommit, err := baselineAdoptionCommit(baselineEnvelope)
	if err != nil {
		return nil, nil, "", err
	}
	return context, baselineMap, adoptionCommit, nil
}

func backfillRelativePathWithinRoot(root, targetPath string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("backfill path resolution requires a repository root")
	}
	if strings.TrimSpace(targetPath) == "" {
		return "", fmt.Errorf("backfill path resolution requires a history path")
	}
	rel, err := filepath.Rel(root, targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve relative history path: %w", err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "", fmt.Errorf("resolve relative history path for %q: empty relative output", targetPath)
	}
	if strings.HasPrefix(rel, "../") || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") {
		return "", fmt.Errorf("path %q escapes repository root", targetPath)
	}
	return rel, nil
}

// shouldSkipRebuild inspects the baseline map for an imported_evidence entry
// that references the history file for adoptionCommit and returns (true,
// path) when the file exists on disk and the rebuild can be safely skipped.
func shouldSkipRebuild(root string, baselineMap map[string]any, adoptionCommit string) (bool, string) {
	intendedHistoryPath := filepath.Join(root, "assurance", "backfill", fmt.Sprintf("imported-git-history-%s.json", adoptionCommit))
	relPath, err := backfillRelativePathWithinRoot(root, intendedHistoryPath)
	if err != nil {
		return false, ""
	}
	valueRaw, _ := baselineMap["value"]
	value, _ := valueRaw.(map[string]any)
	evidenceRaw, _ := value["imported_evidence"]
	evidence, _ := evidenceRaw.([]any)
	if importedEvidenceExists(evidence, relPath) {
		if _, err := os.Stat(intendedHistoryPath); err == nil {
			return true, intendedHistoryPath
		}
	}
	return false, ""
}
