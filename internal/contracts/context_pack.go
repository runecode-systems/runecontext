package contracts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	contextPackSchemaVersion    = 1
	contextPackCanonicalization = "runecontext-canonical-json-v1"
	contextPackHashAlgorithm    = "sha256"
)

var contextPackReadProjectFile = readProjectFile

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
	if p == nil {
		return nil, fmt.Errorf("project index is required")
	}
	if p.Bundles == nil {
		return nil, fmt.Errorf("bundle catalog is unavailable")
	}
	if p.Resolution == nil {
		return nil, fmt.Errorf("source resolution is unavailable")
	}
	requested, err := normalizeContextPackBundleIDs(options.BundleIDs)
	if err != nil {
		return nil, err
	}
	resolution, err := p.Bundles.ResolveRequest(requested)
	if err != nil {
		return nil, err
	}
	resolvedFrom, err := buildContextPackResolvedFrom(p.Resolution, resolution.Linearization)
	if err != nil {
		return nil, err
	}
	generatedAt, err := formatContextPackGeneratedAt(options.GeneratedAt)
	if err != nil {
		return nil, err
	}
	selected, excluded, err := buildContextPackInventories(p.ContentRoot, resolution)
	if err != nil {
		return nil, err
	}
	pack := &ContextPack{
		SchemaVersion:      contextPackSchemaVersion,
		Canonicalization:   contextPackCanonicalization,
		PackHashAlg:        contextPackHashAlgorithm,
		ID:                 requested[0],
		RequestedBundleIDs: append([]string(nil), requested...),
		ResolvedFrom:       resolvedFrom,
		Selected:           selected,
		Excluded:           excluded,
		GeneratedAt:        generatedAt,
	}
	if pack.ID != pack.RequestedBundleIDs[0] {
		return nil, fmt.Errorf("context-pack id %q must match first requested bundle ID %q", pack.ID, pack.RequestedBundleIDs[0])
	}
	hash, err := pack.computePackHash()
	if err != nil {
		return nil, err
	}
	pack.PackHash = hash
	return pack, nil
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

func buildContextPackInventories(contentRoot string, resolution *BundleResolution) (ContextPackAspectSet, ContextPackExcludedAspectSet, error) {
	selected := newContextPackAspectSet()
	excluded := newContextPackExcludedAspectSet()
	if resolution == nil {
		return selected, excluded, fmt.Errorf("bundle resolution is required")
	}
	for _, aspect := range bundleAspects {
		aspectResolution := resolution.Aspects[aspect]
		selectedItems, err := buildContextPackSelectedFiles(contentRoot, aspectResolution.Selected)
		if err != nil {
			return ContextPackAspectSet{}, ContextPackExcludedAspectSet{}, err
		}
		excludedItems := buildContextPackExcludedFiles(aspectResolution.Excluded)
		assignContextPackSelectedAspect(&selected, aspect, selectedItems)
		assignContextPackExcludedAspect(&excluded, aspect, excludedItems)
	}
	return selected, excluded, nil
}

func newContextPackAspectSet() ContextPackAspectSet {
	return ContextPackAspectSet{
		Project:   []ContextPackSelectedFile{},
		Standards: []ContextPackSelectedFile{},
		Specs:     []ContextPackSelectedFile{},
		Decisions: []ContextPackSelectedFile{},
	}
}

func newContextPackExcludedAspectSet() ContextPackExcludedAspectSet {
	return ContextPackExcludedAspectSet{
		Project:   []ContextPackExcludedFile{},
		Standards: []ContextPackExcludedFile{},
		Specs:     []ContextPackExcludedFile{},
		Decisions: []ContextPackExcludedFile{},
	}
}

func assignContextPackSelectedAspect(target *ContextPackAspectSet, aspect BundleAspect, items []ContextPackSelectedFile) {
	switch aspect {
	case BundleAspectProject:
		target.Project = items
	case BundleAspectStandards:
		target.Standards = items
	case BundleAspectSpecs:
		target.Specs = items
	case BundleAspectDecisions:
		target.Decisions = items
	}
}

func assignContextPackExcludedAspect(target *ContextPackExcludedAspectSet, aspect BundleAspect, items []ContextPackExcludedFile) {
	switch aspect {
	case BundleAspectProject:
		target.Project = items
	case BundleAspectStandards:
		target.Standards = items
	case BundleAspectSpecs:
		target.Specs = items
	case BundleAspectDecisions:
		target.Decisions = items
	}
}

func buildContextPackSelectedFiles(contentRoot string, entries []BundleInventoryEntry) ([]ContextPackSelectedFile, error) {
	result := make([]ContextPackSelectedFile, 0, len(entries))
	for _, entry := range entries {
		if len(entry.MatchedBy) == 0 {
			return nil, fmt.Errorf("selected context-pack file %q is missing selector provenance", entry.Path)
		}
		hash, err := hashContextPackFile(contentRoot, entry.Path)
		if err != nil {
			return nil, err
		}
		result = append(result, ContextPackSelectedFile{
			Path:       entry.Path,
			SHA256:     hash,
			SelectedBy: contextPackRuleReferences(entry.MatchedBy),
		})
	}
	return result, nil
}

func buildContextPackExcludedFiles(entries []BundleInventoryEntry) []ContextPackExcludedFile {
	result := make([]ContextPackExcludedFile, 0, len(entries))
	for _, entry := range entries {
		result = append(result, ContextPackExcludedFile{
			Path:     entry.Path,
			LastRule: contextPackRuleReference(entry.FinalRule),
		})
	}
	return result
}

func contextPackRuleReferences(items []BundleRuleReference) []ContextPackRuleReference {
	result := make([]ContextPackRuleReference, len(items))
	for i, item := range items {
		result[i] = contextPackRuleReference(item)
	}
	return result
}

func contextPackRuleReference(item BundleRuleReference) ContextPackRuleReference {
	return ContextPackRuleReference{
		Bundle:  item.Bundle,
		Aspect:  item.Aspect,
		Rule:    item.Rule,
		Pattern: item.Pattern,
		Kind:    item.Kind,
	}
}

func hashContextPackFile(contentRoot, relativePath string) (string, error) {
	fullPath := filepath.Join(contentRoot, filepath.FromSlash(relativePath))
	data, err := contextPackReadProjectFile(contentRoot, fullPath)
	if err != nil {
		return "", fmt.Errorf("hash context-pack file %q: %w", relativePath, err)
	}
	data = normalizeContextPackFileContent(data)
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:]), nil
}

func isPortableLocalSourceRef(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, `\\`) || strings.HasPrefix(trimmed, "//") {
		return false
	}
	if strings.Contains(trimmed, `\`) {
		return false
	}
	if len(trimmed) >= 2 && trimmed[1] == ':' {
		prefix := trimmed[0]
		if (prefix >= 'A' && prefix <= 'Z') || (prefix >= 'a' && prefix <= 'z') {
			return false
		}
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(trimmed, "/./") || strings.Contains(trimmed, "/../") || strings.HasPrefix(trimmed, "./") || strings.HasPrefix(trimmed, "../") || strings.HasSuffix(trimmed, "/.") || strings.HasSuffix(trimmed, "/..") {
		return false
	}
	return true
}

func normalizeContextPackFileContent(data []byte) []byte {
	if !looksLikePortableText(data) || !bytes.Contains(data, []byte{'\r'}) {
		return data
	}
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(normalized, []byte{'\r'}, []byte{'\n'})
}

func looksLikePortableText(data []byte) bool {
	return utf8.Valid(data) && !bytes.Contains(data, []byte{0})
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

func (p *ContextPack) computePackHash() (string, error) {
	canonical, err := canonicalContextPackHashInput(p)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return fmt.Sprintf("%x", sum[:]), nil
}

func canonicalContextPackHashInput(pack *ContextPack) ([]byte, error) {
	if pack == nil {
		return nil, fmt.Errorf("context pack is required")
	}
	return marshalCanonicalJSON(contextPackCanonicalValue(pack))
}

func contextPackCanonicalValue(pack *ContextPack) map[string]any {
	value := contextPackDocumentValue(pack)
	delete(value, "pack_hash")
	delete(value, "generated_at")
	return value
}

func contextPackDocumentValue(pack *ContextPack) map[string]any {
	if pack == nil {
		return nil
	}
	return map[string]any{
		"schema_version":       pack.SchemaVersion,
		"canonicalization":     pack.Canonicalization,
		"pack_hash_alg":        pack.PackHashAlg,
		"pack_hash":            pack.PackHash,
		"id":                   pack.ID,
		"requested_bundle_ids": append([]string(nil), pack.RequestedBundleIDs...),
		"resolved_from":        contextPackResolvedFromValue(pack.ResolvedFrom),
		"selected":             contextPackSelectedValue(pack.Selected),
		"excluded":             contextPackExcludedValue(pack.Excluded),
		"generated_at":         pack.GeneratedAt,
	}
}

func contextPackResolvedFromValue(value ContextPackResolvedFrom) map[string]any {
	result := map[string]any{
		"source_mode":         string(value.SourceMode),
		"source_ref":          value.SourceRef,
		"source_verification": string(value.SourceVerification),
		"context_bundle_ids":  append([]string(nil), value.ContextBundleIDs...),
	}
	if value.SourceCommit != "" {
		result["source_commit"] = value.SourceCommit
	}
	if value.VerifiedSignerIdentity != "" {
		result["verified_signer_identity"] = value.VerifiedSignerIdentity
	}
	if value.VerifiedSignerFingerprint != "" {
		result["verified_signer_fingerprint"] = value.VerifiedSignerFingerprint
	}
	return result
}

func contextPackSelectedValue(value ContextPackAspectSet) map[string]any {
	return map[string]any{
		"project":   contextPackSelectedFilesValue(value.Project),
		"standards": contextPackSelectedFilesValue(value.Standards),
		"specs":     contextPackSelectedFilesValue(value.Specs),
		"decisions": contextPackSelectedFilesValue(value.Decisions),
	}
}

func contextPackExcludedValue(value ContextPackExcludedAspectSet) map[string]any {
	return map[string]any{
		"project":   contextPackExcludedFilesValue(value.Project),
		"standards": contextPackExcludedFilesValue(value.Standards),
		"specs":     contextPackExcludedFilesValue(value.Specs),
		"decisions": contextPackExcludedFilesValue(value.Decisions),
	}
}

func contextPackSelectedFilesValue(items []ContextPackSelectedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":        item.Path,
			"sha256":      item.SHA256,
			"selected_by": contextPackRuleReferencesValue(item.SelectedBy),
		}
	}
	return result
}

func contextPackExcludedFilesValue(items []ContextPackExcludedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":      item.Path,
			"last_rule": contextPackRuleReferenceValue(item.LastRule),
		}
	}
	return result
}

func contextPackRuleReferencesValue(items []ContextPackRuleReference) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = contextPackRuleReferenceValue(item)
	}
	return result
}

func contextPackRuleReferenceValue(item ContextPackRuleReference) map[string]any {
	return map[string]any{
		"bundle":  item.Bundle,
		"aspect":  string(item.Aspect),
		"rule":    string(item.Rule),
		"pattern": item.Pattern,
		"kind":    string(item.Kind),
	}
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeCanonicalJSON(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeCanonicalJSON intentionally supports only the concrete value shapes that
// context-pack canonical hash inputs emit: maps with string keys, arrays,
// strings, booleans, nil, and integral numbers. This keeps the implementation
// narrowly correct for emitted pack data while still following RFC 8785-style
// ordering and string escaping rules for those values.
func writeCanonicalJSON(buf *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		buf.WriteString("null")
		return nil
	case string:
		writeCanonicalJSONString(buf, typed)
		return nil
	case bool:
		if typed {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil
	case []string:
		buf.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonicalJSON(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	case []any:
		buf.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonicalJSON(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
		return nil
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeCanonicalJSON(buf, key); err != nil {
				return err
			}
			buf.WriteByte(':')
			if err := writeCanonicalJSON(buf, typed[key]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
		return nil
	default:
		if writeCanonicalInteger(buf, value) {
			return nil
		}
		return fmt.Errorf("unsupported canonical JSON value %T", value)
	}
}

func writeCanonicalJSONString(buf *bytes.Buffer, value string) {
	buf.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\', '"':
			buf.WriteByte('\\')
			buf.WriteRune(r)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r >= 0 && r < 0x20 {
				buf.WriteString(fmt.Sprintf(`\u%04x`, r))
				continue
			}
			buf.WriteRune(r)
		}
	}
	buf.WriteByte('"')
}

func writeCanonicalInteger(buf *bytes.Buffer, value any) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(rv.Int(), 10))
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return true
	default:
		return false
	}
}
