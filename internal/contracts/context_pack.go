package contracts

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	contextPackSchemaVersion    = 1
	contextPackCanonicalization = "runecontext-canonical-json-v1"
	contextPackHashAlgorithm    = "sha256"
)

var (
	contextPackReadProjectFileMu sync.RWMutex
	contextPackReadProjectFile   = readProjectFile
)

type ContextPack struct {
	SchemaVersion      int                          `json:"schema_version" yaml:"schema_version"`
	Canonicalization   string                       `json:"canonicalization" yaml:"canonicalization"`
	PackHashAlg        string                       `json:"pack_hash_alg" yaml:"pack_hash_alg"`
	PackHash           string                       `json:"pack_hash" yaml:"pack_hash"`
	ID                 string                       `json:"id" yaml:"id"`
	RequestedBundleIDs []string                     `json:"requested_bundle_ids" yaml:"requested_bundle_ids"`
	ResolvedFrom       ContextPackResolvedFrom      `json:"resolved_from" yaml:"resolved_from"`
	Selected           ContextPackAspectSet         `json:"selected" yaml:"selected"`
	Excluded           ContextPackExcludedAspectSet `json:"excluded" yaml:"excluded"`
	GeneratedAt        string                       `json:"generated_at" yaml:"generated_at"`
}

type ContextPackResolvedFrom struct {
	SourceMode                SourceMode          `json:"source_mode" yaml:"source_mode"`
	SourceRef                 string              `json:"source_ref" yaml:"source_ref"`
	SourceCommit              string              `json:"source_commit,omitempty" yaml:"source_commit,omitempty"`
	SourceVerification        VerificationPosture `json:"source_verification" yaml:"source_verification"`
	VerifiedSignerIdentity    string              `json:"verified_signer_identity,omitempty" yaml:"verified_signer_identity,omitempty"`
	VerifiedSignerFingerprint string              `json:"verified_signer_fingerprint,omitempty" yaml:"verified_signer_fingerprint,omitempty"`
	ContextBundleIDs          []string            `json:"context_bundle_ids" yaml:"context_bundle_ids"`
}

type ContextPackAspectSet struct {
	Project   []ContextPackSelectedFile `json:"project" yaml:"project"`
	Standards []ContextPackSelectedFile `json:"standards" yaml:"standards"`
	Specs     []ContextPackSelectedFile `json:"specs" yaml:"specs"`
	Decisions []ContextPackSelectedFile `json:"decisions" yaml:"decisions"`
}

type ContextPackExcludedAspectSet struct {
	Project   []ContextPackExcludedFile `json:"project" yaml:"project"`
	Standards []ContextPackExcludedFile `json:"standards" yaml:"standards"`
	Specs     []ContextPackExcludedFile `json:"specs" yaml:"specs"`
	Decisions []ContextPackExcludedFile `json:"decisions" yaml:"decisions"`
}

type ContextPackSelectedFile struct {
	Path       string                     `json:"path" yaml:"path"`
	SHA256     string                     `json:"sha256" yaml:"sha256"`
	SelectedBy []ContextPackRuleReference `json:"selected_by" yaml:"selected_by"`
}

type ContextPackExcludedFile struct {
	Path     string                   `json:"path" yaml:"path"`
	LastRule ContextPackRuleReference `json:"last_rule" yaml:"last_rule"`
}

type ContextPackRuleReference struct {
	Bundle  string            `json:"bundle" yaml:"bundle"`
	Aspect  BundleAspect      `json:"aspect" yaml:"aspect"`
	Rule    BundleRuleKind    `json:"rule" yaml:"rule"`
	Pattern string            `json:"pattern" yaml:"pattern"`
	Kind    BundlePatternKind `json:"kind" yaml:"kind"`
}

type ContextPackOptions struct {
	BundleIDs   []string
	GeneratedAt time.Time
}

func (p *ProjectIndex) BuildContextPack(options ContextPackOptions) (*ContextPack, error) {
	pack, _, err := p.buildStableContextPack(options)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

type contextPackInputs struct {
	requested   []string
	resolved    ContextPackResolvedFrom
	selected    ContextPackAspectSet
	excluded    ContextPackExcludedAspectSet
	generatedAt string
}

func validateContextPackProjectIndex(index *ProjectIndex) error {
	if index == nil {
		return fmt.Errorf("project index is required")
	}
	if index.Bundles == nil {
		return fmt.Errorf("bundle catalog is unavailable")
	}
	if index.Resolution == nil {
		return fmt.Errorf("source resolution is unavailable")
	}
	return nil
}

func newContextPack(inputs contextPackInputs) *ContextPack {
	return &ContextPack{
		SchemaVersion:      contextPackSchemaVersion,
		Canonicalization:   contextPackCanonicalization,
		PackHashAlg:        contextPackHashAlgorithm,
		ID:                 inputs.requested[0],
		RequestedBundleIDs: append([]string(nil), inputs.requested...),
		ResolvedFrom:       inputs.resolved,
		Selected:           inputs.selected,
		Excluded:           inputs.excluded,
		GeneratedAt:        inputs.generatedAt,
	}
}

func validateContextPackIdentity(pack *ContextPack) error {
	if pack.ID != pack.RequestedBundleIDs[0] {
		return fmt.Errorf("context-pack id %q must match first requested bundle ID %q", pack.ID, pack.RequestedBundleIDs[0])
	}
	return nil
}

func readContextPackProjectFile(boundaryPath, path string) ([]byte, error) {
	return currentContextPackReadProjectFile()(boundaryPath, path)
}

func currentContextPackReadProjectFile() func(boundaryPath, path string) ([]byte, error) {
	contextPackReadProjectFileMu.RLock()
	reader := contextPackReadProjectFile
	contextPackReadProjectFileMu.RUnlock()
	if reader == nil {
		return readProjectFile
	}
	return reader
}

func setContextPackReadProjectFileHookForTest(hook func(boundaryPath, path string) ([]byte, error)) func() {
	contextPackReadProjectFileMu.Lock()
	previous := contextPackReadProjectFile
	contextPackReadProjectFile = hook
	contextPackReadProjectFileMu.Unlock()
	return func() {
		contextPackReadProjectFileMu.Lock()
		contextPackReadProjectFile = previous
		contextPackReadProjectFileMu.Unlock()
	}
}

func normalizeContextPackBundleIDs(bundleIDs []string) ([]string, error) {
	if len(bundleIDs) == 0 {
		return nil, fmt.Errorf("at least one requested bundle ID is required")
	}
	result := make([]string, 0, len(bundleIDs))
	seen := map[string]struct{}{}
	for _, raw := range bundleIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, fmt.Errorf("requested bundle IDs must not be empty")
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("requested bundle IDs must not contain duplicates: %q", id)
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func buildContextPackResolvedFrom(resolution *SourceResolution, linearization []string) (ContextPackResolvedFrom, error) {
	if resolution == nil {
		return ContextPackResolvedFrom{}, fmt.Errorf("source resolution is required")
	}
	if resolution.SourceMode == SourceModePath && !isPortableLocalSourceRef(resolution.SourceRef) {
		return ContextPackResolvedFrom{}, fmt.Errorf("context packs require a portable source_ref; local path sources must use a relative forward-slash form without drive-qualified, UNC, or traversal segments")
	}
	result := ContextPackResolvedFrom{
		SourceMode:         resolution.SourceMode,
		SourceRef:          filepath.ToSlash(resolution.SourceRef),
		SourceVerification: resolution.VerificationPosture,
		ContextBundleIDs:   append([]string(nil), linearization...),
	}
	if resolution.ResolvedCommit != "" {
		result.SourceCommit = resolution.ResolvedCommit
	}
	if resolution.VerifiedSignerIdentity != "" {
		result.VerifiedSignerIdentity = resolution.VerifiedSignerIdentity
	}
	if resolution.VerifiedSignerFingerprint != "" {
		result.VerifiedSignerFingerprint = resolution.VerifiedSignerFingerprint
	}
	return result, nil
}

// generated_at uses whole-second UTC RFC3339 output to keep emitted packs
// stable and merge-friendly while avoiding hidden wall-clock defaults.
func formatContextPackGeneratedAt(value time.Time) (string, error) {
	if value.IsZero() {
		return "", fmt.Errorf("context packs require explicit generated_at; core builder does not default wall-clock time")
	}
	if value.Nanosecond() != 0 {
		return "", fmt.Errorf("context packs require generated_at values rounded to whole seconds")
	}
	return value.UTC().Format(time.RFC3339), nil
}
