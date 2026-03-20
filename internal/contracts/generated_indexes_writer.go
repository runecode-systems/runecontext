package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (p *ProjectIndex) WriteGeneratedIndexes() error {
	if p == nil {
		return fmt.Errorf("project index is required")
	}
	if strings.TrimSpace(p.ContentRoot) == "" {
		return fmt.Errorf("project index content root is required")
	}
	manifest, err := p.BuildGeneratedManifest()
	if err != nil {
		return err
	}
	changesByStatus, err := p.BuildGeneratedChangesByStatusIndex()
	if err != nil {
		return err
	}
	bundles, err := p.BuildGeneratedBundlesIndex()
	if err != nil {
		return err
	}
	indexDir := filepath.Join(p.ContentRoot, generatedIndexesDirectoryRelative)
	if err := os.MkdirAll(indexDir, 0o755); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedChangesIndexRelativePath)), changesByStatus); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedBundlesIndexRelativePath)), bundles); err != nil {
		return err
	}
	if err := writeGeneratedYAML(filepath.Join(p.ContentRoot, filepath.FromSlash(generatedManifestRelativePath)), manifest); err != nil {
		return err
	}
	return nil
}
