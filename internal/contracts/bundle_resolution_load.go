package contracts

import (
	"fmt"
	"path/filepath"
	"sort"
)

func loadBundleCatalog(v *Validator, rootConfigPath string, rootData []byte, contentRoot string) (*BundleCatalog, error) {
	resolvedRoot, err := filepath.EvalSymlinks(contentRoot)
	if err != nil {
		return nil, &ValidationError{Path: contentRoot, Message: err.Error()}
	}
	catalog := &BundleCatalog{Root: filepath.Clean(resolvedRoot), bundles: map[string]*bundleDefinition{}, resolutions: map[string]*BundleResolution{}}
	bundlePaths, err := discoverBundlePaths(contentRoot)
	if err != nil {
		return nil, err
	}
	if err := loadBundleDefinitions(v, catalog, rootConfigPath, rootData, filepath.Join(contentRoot, "bundles"), bundlePaths); err != nil {
		return nil, err
	}
	if err := validateBundleParents(catalog); err != nil {
		return nil, err
	}
	if err := preResolveBundles(catalog); err != nil {
		return nil, err
	}
	return catalog, nil
}

func discoverBundlePaths(contentRoot string) ([]string, error) {
	bundlePaths := make([]string, 0)
	bundlesRoot := filepath.Join(contentRoot, "bundles")
	err := walkProjectFiles(bundlesRoot, func(path string) error {
		if filepath.Ext(path) == ".yaml" {
			bundlePaths = append(bundlePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(bundlePaths)
	return bundlePaths, nil
}

func loadBundleDefinitions(v *Validator, catalog *BundleCatalog, rootConfigPath string, rootData []byte, bundlesRoot string, bundlePaths []string) error {
	for _, bundlePath := range bundlePaths {
		bundle, err := loadBundleDefinition(v, rootConfigPath, rootData, bundlesRoot, bundlePath)
		if err != nil {
			return err
		}
		if existing, ok := catalog.bundles[bundle.ID]; ok {
			return &ValidationError{Path: bundlePath, Message: fmt.Sprintf("bundle id %q is duplicated (already declared in %s)", bundle.ID, existing.Path)}
		}
		catalog.bundles[bundle.ID] = bundle
	}
	return nil
}

func loadBundleDefinition(v *Validator, rootConfigPath string, rootData []byte, bundlesRoot, bundlePath string) (*bundleDefinition, error) {
	data, err := readProjectFile(bundlesRoot, bundlePath)
	if err != nil {
		return nil, err
	}
	if err := v.ValidateYAMLFile("bundle.schema.json", bundlePath, data); err != nil {
		return nil, err
	}
	if err := v.ValidateExtensionOptIn(rootConfigPath, rootData, bundlePath, data); err != nil {
		return nil, err
	}
	return parseBundleDefinition(bundlePath, data)
}

func validateBundleParents(catalog *BundleCatalog) error {
	for _, id := range SortedKeys(catalog.bundles) {
		bundle := catalog.bundles[id]
		for _, parentID := range bundle.Extends {
			if _, ok := catalog.bundles[parentID]; !ok {
				return &ValidationError{Path: bundle.Path, Message: fmt.Sprintf("bundle %q extends unknown parent %q", bundle.ID, parentID)}
			}
		}
	}
	return nil
}

func preResolveBundles(catalog *BundleCatalog) error {
	for _, id := range SortedKeys(catalog.bundles) {
		if _, err := catalog.Resolve(id); err != nil {
			return err
		}
	}
	return nil
}

func parseBundleDefinition(path string, data []byte) (*bundleDefinition, error) {
	parsed, err := parseYAML(data)
	if err != nil {
		return nil, &ValidationError{Path: path, Message: err.Error()}
	}
	obj, err := expectObject(path, parsed, "bundle")
	if err != nil {
		return nil, err
	}
	bundle := &bundleDefinition{ID: fmt.Sprint(obj["id"]), Path: path, Extends: extractStringList(obj["extends"]), Includes: map[BundleAspect][]bundleRule{}, Excludes: map[BundleAspect][]bundleRule{}}
	bundle.Includes, err = extractBundleRules(path, bundle.ID, BundleRuleKindInclude, obj["includes"])
	if err != nil {
		return nil, err
	}
	bundle.Excludes, err = extractBundleRules(path, bundle.ID, BundleRuleKindExclude, obj["excludes"])
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func extractStringList(raw any) []string {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, fmt.Sprint(item))
	}
	return result
}
