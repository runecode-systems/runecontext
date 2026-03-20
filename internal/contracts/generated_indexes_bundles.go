package contracts

type GeneratedBundlesIndex struct {
	SchemaVersion int                    `json:"schema_version" yaml:"schema_version"`
	Bundles       []GeneratedBundleEntry `json:"bundles" yaml:"bundles"`
}

type GeneratedBundleEntry struct {
	ID                 string                          `json:"id" yaml:"id"`
	Path               string                          `json:"path" yaml:"path"`
	Extends            []string                        `json:"extends" yaml:"extends"`
	ResolvedParents    []string                        `json:"resolved_parents" yaml:"resolved_parents"`
	ReferencedPatterns GeneratedBundlePatternAspectSet `json:"referenced_patterns" yaml:"referenced_patterns"`
}

type GeneratedBundlePatternAspectSet struct {
	Project   GeneratedBundleAspectPatterns `json:"project" yaml:"project"`
	Standards GeneratedBundleAspectPatterns `json:"standards" yaml:"standards"`
	Specs     GeneratedBundleAspectPatterns `json:"specs" yaml:"specs"`
	Decisions GeneratedBundleAspectPatterns `json:"decisions" yaml:"decisions"`
}

type GeneratedBundleAspectPatterns struct {
	Includes []GeneratedBundlePattern `json:"includes" yaml:"includes"`
	Excludes []GeneratedBundlePattern `json:"excludes" yaml:"excludes"`
}

type GeneratedBundlePattern struct {
	Pattern string            `json:"pattern" yaml:"pattern"`
	Kind    BundlePatternKind `json:"kind" yaml:"kind"`
}
