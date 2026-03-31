package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type assuranceLayoutAlpha13Migration struct{}

func (m assuranceLayoutAlpha13Migration) Apply(ctx upgradeMigrationContext, hop upgradeHop) error {
	legacyRoot := filepath.Join(ctx.Root, "assurance")
	canonicalRoot := filepath.Join(ctx.Root, "runecontext", "assurance")
	legacyExists, err := pathExists(legacyRoot)
	if err != nil {
		return err
	}
	canonicalExists, err := pathExists(canonicalRoot)
	if err != nil {
		return err
	}
	if legacyExists && canonicalExists {
		return fmt.Errorf("conflicting assurance layouts: both %s and %s exist", legacyRoot, canonicalRoot)
	}
	if legacyExists {
		if err := os.MkdirAll(filepath.Dir(canonicalRoot), 0o755); err != nil {
			return fmt.Errorf("prepare canonical assurance parent: %w", err)
		}
		if err := os.Rename(legacyRoot, canonicalRoot); err != nil {
			return fmt.Errorf("move assurance layout to canonical path: %w", err)
		}
	}
	if err := rewriteImportedEvidenceBackfillPaths(filepath.Join(canonicalRoot, "baseline.yaml")); err != nil {
		return err
	}
	return nil
}

func (m assuranceLayoutAlpha13Migration) Verify(ctx upgradeMigrationContext, hop upgradeHop) error {
	legacyRoot := filepath.Join(ctx.Root, "assurance")
	if exists, err := pathExists(legacyRoot); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("legacy assurance path still exists: %s", legacyRoot)
	}
	return verifyImportedEvidenceBackfillPaths(filepath.Join(ctx.Root, "runecontext", "assurance", "baseline.yaml"))
}

func rewriteImportedEvidenceBackfillPaths(baselinePath string) error {
	exists, err := pathExists(baselinePath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return fmt.Errorf("read assurance baseline for migration: %w", err)
	}
	root := map[string]any{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse assurance baseline for migration: %w", err)
	}
	if !rewriteImportedEvidenceEntries(root) {
		return nil
	}
	serialized, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("render assurance baseline for migration: %w", err)
	}
	if err := writeAtomicFile(baselinePath, serialized, 0o644); err != nil {
		return fmt.Errorf("write assurance baseline for migration: %w", err)
	}
	return nil
}

func rewriteImportedEvidenceEntries(root map[string]any) bool {
	value, ok := root["value"].(map[string]any)
	if !ok {
		return false
	}
	entries, ok := value["imported_evidence"].([]any)
	if !ok {
		return false
	}
	changed := false
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		path := strings.TrimSpace(fmt.Sprint(entry["path"]))
		if strings.HasPrefix(path, "assurance/backfill/") {
			entry["path"] = "runecontext/" + path
			changed = true
		}
	}
	return changed
}

func verifyImportedEvidenceBackfillPaths(baselinePath string) error {
	entries, err := loadImportedEvidenceEntriesForMigrationVerification(baselinePath)
	if err != nil {
		return err
	}
	for _, path := range entries {
		if strings.HasPrefix(path, "assurance/backfill/") {
			return fmt.Errorf("legacy imported_evidence path remains in baseline: %s", path)
		}
	}
	return nil
}

func loadImportedEvidenceEntriesForMigrationVerification(baselinePath string) ([]string, error) {
	exists, err := pathExists(baselinePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return nil, fmt.Errorf("read assurance baseline for verification: %w", err)
	}
	root := map[string]any{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse assurance baseline for verification: %w", err)
	}
	value, ok := root["value"].(map[string]any)
	if !ok {
		return nil, nil
	}
	rawEntries, ok := value["imported_evidence"].([]any)
	if !ok {
		return nil, nil
	}
	entries := make([]string, 0, len(rawEntries))
	for _, raw := range rawEntries {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		entries = append(entries, strings.TrimSpace(fmt.Sprint(entry["path"])))
	}
	return entries, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
