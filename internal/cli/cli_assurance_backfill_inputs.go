package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadBackfillInputs(root string) (assuranceBackfillContext, map[string]any, string, error) {
	if err := requireVerifiedTierForBackfill(root); err != nil {
		return assuranceBackfillContext{}, nil, "", err
	}
	baselinePath := filepath.Join(root, "assurance", "baseline.yaml")
	baselineData, err := os.ReadFile(baselinePath)
	if err != nil {
		return assuranceBackfillContext{}, nil, "", fmt.Errorf("read assurance baseline: %w", err)
	}
	var baseline map[string]any
	if err := yaml.Unmarshal(baselineData, &baseline); err != nil {
		return assuranceBackfillContext{}, nil, "", fmt.Errorf("parse assurance baseline: %w", err)
	}
	adoptionCommit := readBaselineAdoptionCommit(baseline)
	if !isCanonicalLowerHex40(adoptionCommit) {
		return assuranceBackfillContext{}, nil, "", fmt.Errorf("assurance baseline adoption_commit must be a canonical lowercase 40-char hex SHA")
	}
	return assuranceBackfillContext{baselinePath: baselinePath}, baseline, adoptionCommit, nil
}

func requireVerifiedTierForBackfill(root string) error {
	configPath := filepath.Join(root, "runecontext.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read runecontext config: %w", err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("parse runecontext config: %w", err)
	}
	if strings.TrimSpace(fmt.Sprint(config["assurance_tier"])) != "verified" {
		return fmt.Errorf("assurance_tier must be verified before running assurance backfill")
	}
	return nil
}

func readBaselineAdoptionCommit(baseline map[string]any) string {
	if baseline == nil {
		return ""
	}
	value, ok := baseline["value"].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value["adoption_commit"]))
}

func shouldSkipRebuild(root string, baseline map[string]any, adoptionCommit string) (bool, string) {
	relative := filepath.ToSlash(filepath.Join("assurance", "backfill", fmt.Sprintf("imported-git-history-%s.json", adoptionCommit)))
	absPath := filepath.Join(root, filepath.FromSlash(relative))
	if !hasImportedEvidenceEntry(baseline, relative) {
		return false, ""
	}
	if _, err := os.Stat(absPath); err != nil {
		return false, ""
	}
	return true, absPath
}

func hasImportedEvidenceEntry(baseline map[string]any, relativePath string) bool {
	value, ok := baseline["value"].(map[string]any)
	if !ok {
		return false
	}
	entries, ok := value["imported_evidence"].([]any)
	if !ok {
		return false
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		path := strings.TrimSpace(fmt.Sprint(entry["path"]))
		provenance := strings.TrimSpace(fmt.Sprint(entry["provenance"]))
		if path == relativePath && provenance == "imported_git_history" {
			return true
		}
	}
	return false
}

func appendImportedEvidence(baselinePath string, baseline map[string]any, root, historyPath string) (bool, error) {
	relative, err := backfillRelativePathWithinRoot(root, historyPath)
	if err != nil {
		return false, err
	}
	value, ok := baseline["value"].(map[string]any)
	if !ok {
		value = map[string]any{}
		baseline["value"] = value
	}
	if hasImportedEvidenceEntry(baseline, relative) {
		return false, nil
	}
	entries := make([]any, 0)
	if rawEntries, ok := value["imported_evidence"].([]any); ok {
		entries = append(entries, rawEntries...)
	}
	entries = append(entries, map[string]any{"path": relative, "provenance": "imported_git_history"})
	value["imported_evidence"] = entries
	serialized, err := yaml.Marshal(baseline)
	if err != nil {
		return false, fmt.Errorf("render assurance baseline: %w", err)
	}
	if err := writeAtomicFile(baselinePath, serialized, 0o644); err != nil {
		return false, fmt.Errorf("write assurance baseline: %w", err)
	}
	return true, nil
}

func backfillRelativePathWithinRoot(root, target string) (string, error) {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("compute backfill relative path: %w", err)
	}
	clean := filepath.Clean(relative)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
		return "", fmt.Errorf("backfill path escapes repository root: %s", target)
	}
	return filepath.ToSlash(clean), nil
}
