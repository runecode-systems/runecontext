package contracts

import "fmt"

func (p *ProjectIndex) BuildGeneratedBundlesIndex() (*GeneratedBundlesIndex, error) {
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	index := &GeneratedBundlesIndex{SchemaVersion: generatedBundleIndexSchemaVersion, Bundles: []GeneratedBundleEntry{}}
	if p.Bundles == nil {
		return index, nil
	}
	for _, bundleID := range SortedKeys(p.Bundles.bundles) {
		bundle := p.Bundles.bundles[bundleID]
		if bundle == nil {
			continue
		}
		resolution, err := p.Bundles.Resolve(bundleID)
		if err != nil {
			return nil, err
		}
		bundlePath, err := generatedRelativeArtifactPath(p.ContentRoot, bundle.Path)
		if err != nil {
			return nil, fmt.Errorf("build generated bundles index: %w", err)
		}
		entry := GeneratedBundleEntry{
			ID:              bundle.ID,
			Path:            bundlePath,
			Extends:         append([]string(nil), bundle.Extends...),
			ResolvedParents: resolvedBundleParents(resolution, bundleID),
			ReferencedPatterns: GeneratedBundlePatternAspectSet{
				Project:   generatedBundleAspectPatterns(bundle, BundleAspectProject),
				Standards: generatedBundleAspectPatterns(bundle, BundleAspectStandards),
				Specs:     generatedBundleAspectPatterns(bundle, BundleAspectSpecs),
				Decisions: generatedBundleAspectPatterns(bundle, BundleAspectDecisions),
			},
		}
		index.Bundles = append(index.Bundles, entry)
	}
	return index, nil
}
