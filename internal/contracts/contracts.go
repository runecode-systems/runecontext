package contracts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

var (
	changeIDPattern   = regexp.MustCompile(`^CHG-\d{4}-\d{3}-[a-z0-9]{4,6}-[a-z0-9]+(-[a-z0-9]+)*$`)
	artifactIDPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?(?:/[A-Za-z0-9](?:[A-Za-z0-9._-]*[A-Za-z0-9])?)*$`)
	bundleIDPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
)

type Validator struct {
	schemaRoot string
	cacheMu    sync.RWMutex
	cache      map[string]*jsonschema.Schema
}

type ValidationError struct {
	Path    string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

type MarkdownDocument struct {
	Sections      map[string]string
	Refs          []string
	RefsBySection map[string][]string
}

type FrontmatterDocument struct {
	Frontmatter map[string]any
	Body        string
}

type ProjectIndex struct {
	RootConfigPath        string
	ContentRoot           string
	Resolution            *SourceResolution
	Bundles               *BundleCatalog
	Diagnostics           []ValidationDiagnostic
	AssuranceBaseline     *AssuranceEnvelope
	AssuranceBaselinePath string
	AssuranceBaselineMap  map[string]any
	AssuranceReceipts     map[string]AssuranceReceiptRecord
	ChangeIDs             map[string]struct{}
	Changes               map[string]*ChangeRecord
	MarkdownFiles         map[string]*MarkdownArtifact
	StandardPaths         map[string]struct{}
	Standards             map[string]*StandardRecord
	SpecPaths             map[string]struct{}
	Specs                 map[string]*SpecRecord
	DecisionPaths         map[string]struct{}
	Decisions             map[string]*DecisionRecord
	StatusFiles           map[string]StatusFileRecord
}

type StatusFileRecord struct {
	Data map[string]any
	Raw  []byte
}

type markdownSection struct {
	Heading string
	Body    string
}

func NewValidator(schemaRoot string) *Validator {
	return &Validator{schemaRoot: schemaRoot, cache: map[string]*jsonschema.Schema{}}
}

func (p *ProjectIndex) Close() error {
	if p == nil || p.Resolution == nil {
		return nil
	}
	return p.Resolution.Close()
}

func (p *ProjectIndex) ResolveBundle(id string) (*BundleResolution, error) {
	if p == nil || p.Bundles == nil {
		return nil, fmt.Errorf("bundle catalog is unavailable")
	}
	return p.Bundles.Resolve(id)
}

func (v *Validator) ValidateTraceabilityProject(root string) error {
	index, err := v.ValidateProject(root)
	if err != nil {
		return err
	}
	defer index.Close()
	return nil
}

func (v *Validator) ValidateProject(root string) (*ProjectIndex, error) {
	return v.ValidateProjectWithOptions(root, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
}

func resolveContentRoot(projectRoot string, rootData []byte) (*SourceResolution, error) {
	return resolveSourceFromConfig(filepath.Join(projectRoot, "runecontext.yaml"), projectRoot, rootData, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal})
}

func expectObject(path string, value any, context string) (map[string]any, error) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, &ValidationError{Path: path, Message: fmt.Sprintf("%s must decode to an object", context)}
	}
	return obj, nil
}

func validateChangeIDRefs(path, field string, raw any, known map[string]struct{}) error {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	for _, item := range items {
		id := fmt.Sprint(item)
		if !changeIDPattern.MatchString(id) {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s contains invalid change ID %q", field, id)}
		}
		if _, ok := known[id]; !ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s references missing change %q", field, id)}
		}
	}
	return nil
}

func validatePathRefs(path, field string, raw any, known map[string]struct{}) error {
	if raw == nil {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return &ValidationError{Path: path, Message: fmt.Sprintf("%s must be an array", field)}
	}
	for _, item := range items {
		ref := filepath.ToSlash(fmt.Sprint(item))
		if _, ok := known[ref]; !ok {
			return &ValidationError{Path: path, Message: fmt.Sprintf("%s references missing artifact %q", field, ref)}
		}
	}
	return nil
}

func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
