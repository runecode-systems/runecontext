package contracts

import (
	"fmt"
)

func (p *ProjectIndex) BuildGeneratedManifest() (*GeneratedManifest, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	manifest := &GeneratedManifest{
		SchemaVersion: generatedManifestSchemaVersion,
		Indexes: GeneratedManifestIndexes{
			ChangesByStatus: generatedChangesIndexRelativePath,
			Bundles:         generatedBundlesIndexRelativePath,
		},
		Standards: SortedKeys(p.Standards),
		Bundles:   sortedBundleIDs(p),
		Changes:   SortedKeys(p.Changes),
		Specs:     SortedKeys(p.Specs),
		Decisions: SortedKeys(p.Decisions),
	}
	manifest.Counts = GeneratedManifestCounts{
		Standards: len(manifest.Standards),
		Bundles:   len(manifest.Bundles),
		Changes:   len(manifest.Changes),
		Specs:     len(manifest.Specs),
		Decisions: len(manifest.Decisions),
	}
	return manifest, nil
}
