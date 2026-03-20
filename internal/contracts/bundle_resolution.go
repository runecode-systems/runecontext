package contracts

import "fmt"

const maxBundleInheritanceDepth = 8

var bundleTraversalLimits = snapshotLimits{MaxFiles: 10000, MaxBytes: 0, MaxDepth: 64}

type BundleAspect string

const (
	BundleAspectProject   BundleAspect = "project"
	BundleAspectStandards BundleAspect = "standards"
	BundleAspectSpecs     BundleAspect = "specs"
	BundleAspectDecisions BundleAspect = "decisions"
)

var bundleAspects = []BundleAspect{BundleAspectProject, BundleAspectStandards, BundleAspectSpecs, BundleAspectDecisions}

type BundleRuleKind string

const (
	BundleRuleKindInclude BundleRuleKind = "include"
	BundleRuleKindExclude BundleRuleKind = "exclude"
)

type BundlePatternKind string

const (
	BundlePatternKindExact BundlePatternKind = "exact"
	BundlePatternKindGlob  BundlePatternKind = "glob"
)

type BundleDiagnostic struct {
	Severity DiagnosticSeverity `json:"severity" yaml:"severity"`
	Code     string             `json:"code" yaml:"code"`
	Message  string             `json:"message" yaml:"message"`
	Bundle   string             `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Aspect   BundleAspect       `json:"aspect,omitempty" yaml:"aspect,omitempty"`
	Rule     BundleRuleKind     `json:"rule,omitempty" yaml:"rule,omitempty"`
	Pattern  string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Matches  []string           `json:"matches,omitempty" yaml:"matches,omitempty"`
}

type BundleRuleReference struct {
	Bundle  string            `json:"bundle" yaml:"bundle"`
	Aspect  BundleAspect      `json:"aspect" yaml:"aspect"`
	Rule    BundleRuleKind    `json:"rule" yaml:"rule"`
	Pattern string            `json:"pattern" yaml:"pattern"`
	Kind    BundlePatternKind `json:"kind" yaml:"kind"`
}

type BundleInventoryEntry struct {
	Path      string                `json:"path" yaml:"path"`
	MatchedBy []BundleRuleReference `json:"matched_by" yaml:"matched_by"`
	FinalRule BundleRuleReference   `json:"final_rule" yaml:"final_rule"`
}

type BundleRuleEvaluation struct {
	Bundle      string             `json:"bundle" yaml:"bundle"`
	Aspect      BundleAspect       `json:"aspect" yaml:"aspect"`
	Rule        BundleRuleKind     `json:"rule" yaml:"rule"`
	Pattern     string             `json:"pattern" yaml:"pattern"`
	PatternKind BundlePatternKind  `json:"pattern_kind" yaml:"pattern_kind"`
	Matches     []string           `json:"matches" yaml:"matches"`
	Diagnostics []BundleDiagnostic `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

type BundleAspectResolution struct {
	Rules     []BundleRuleEvaluation `json:"rules" yaml:"rules"`
	Selected  []BundleInventoryEntry `json:"selected" yaml:"selected"`
	Excluded  []BundleInventoryEntry `json:"excluded" yaml:"excluded"`
	Matchable []string               `json:"matchable,omitempty" yaml:"matchable,omitempty"`
}

type BundleResolution struct {
	ID            string                                  `json:"id" yaml:"id"`
	Linearization []string                                `json:"linearization" yaml:"linearization"`
	Aspects       map[BundleAspect]BundleAspectResolution `json:"aspects" yaml:"aspects"`
	Diagnostics   []BundleDiagnostic                      `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}

type bundleRule struct {
	Bundle      string
	Aspect      BundleAspect
	Kind        BundleRuleKind
	Pattern     string
	RawPattern  string
	PatternKind BundlePatternKind
	SourcePath  string
	Index       int
}

type bundleDefinition struct {
	ID       string
	Path     string
	Extends  []string
	Includes map[BundleAspect][]bundleRule
	Excludes map[BundleAspect][]bundleRule
}

type BundleCatalog struct {
	Root        string
	bundles     map[string]*bundleDefinition
	resolutions map[string]*BundleResolution
}

type bundleWalkState struct {
	files int
	depth int
}

func (c *BundleCatalog) Diagnostics() []BundleDiagnostic {
	if c == nil {
		return nil
	}
	items := make([]BundleDiagnostic, 0)
	for _, id := range SortedKeys(c.resolutions) {
		items = append(items, cloneBundleDiagnostics(c.resolutions[id].Diagnostics)...)
	}
	return items
}

func (c *BundleCatalog) appendDiagnostic(bundleID string, diagnostic BundleDiagnostic) {
	if c == nil {
		return
	}
	if resolution := c.resolutions[bundleID]; resolution != nil {
		resolution.Diagnostics = append(resolution.Diagnostics, diagnostic)
	}
}

func (c *BundleCatalog) Resolve(id string) (*BundleResolution, error) {
	if c == nil {
		return nil, fmt.Errorf("bundle catalog is unavailable")
	}
	if resolution, ok := c.resolutions[id]; ok {
		return cloneBundleResolution(resolution), nil
	}
	ordered, err := c.linearize(id)
	if err != nil {
		return nil, err
	}
	resolution, err := c.buildResolution(id, ordered)
	if err != nil {
		return nil, err
	}
	c.resolutions[id] = resolution
	return cloneBundleResolution(resolution), nil
}

func (c *BundleCatalog) buildResolution(id string, ordered []*bundleDefinition) (*BundleResolution, error) {
	resolution := &BundleResolution{ID: id, Linearization: bundleLinearization(ordered), Aspects: map[BundleAspect]BundleAspectResolution{}}
	for _, aspect := range bundleAspects {
		aspectResolution, diagnostics, err := c.resolveAspect(aspect, ordered)
		if err != nil {
			return nil, err
		}
		resolution.Aspects[aspect] = aspectResolution
		resolution.Diagnostics = append(resolution.Diagnostics, diagnostics...)
	}
	return resolution, nil
}

func bundleLinearization(ordered []*bundleDefinition) []string {
	items := make([]string, 0, len(ordered))
	for _, bundle := range ordered {
		items = append(items, bundle.ID)
	}
	return items
}
