package contracts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestBuildContextPackMatchesGoldenFixture(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	assertContextPackValidAgainstSchema(t, v, pack)
	assertContextPackMatchesGolden(t, pack, fixturePath(t, "context-packs", "golden", "child-reinclude.yaml"))
}

func TestBuildContextPackSupportsOrderedTopLevelBundleRequests(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"left", "right"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	assertContextPackValidAgainstSchema(t, v, pack)
	assertContextPackMatchesGolden(t, pack, fixturePath(t, "context-packs", "golden", "left-right.yaml"))
}

func TestBuildContextPackHashExcludesGeneratedAt(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	first, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build first context pack: %v", err)
	}
	second, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 21, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build second context pack: %v", err)
	}
	if first.GeneratedAt == second.GeneratedAt {
		t.Fatal("expected generated_at values to differ")
	}
	if first.PackHash != second.PackHash {
		t.Fatalf("expected generated_at to stay outside canonical hash input, got %q and %q", first.PackHash, second.PackHash)
	}
	if !reflect.DeepEqual(comparableContextPackWithoutGeneratedAt(first), comparableContextPackWithoutGeneratedAt(second)) {
		t.Fatal("expected identical pack content aside from generated_at")
	}
}

func TestBuildContextPackDeterministicAcrossFreshProjectCopies(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	rootA := t.TempDir()
	rootB := t.TempDir()
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), rootA)
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), rootB)
	indexA, err := v.ValidateProject(rootA)
	if err != nil {
		t.Fatalf("validate project copy A: %v", err)
	}
	defer indexA.Close()
	indexB, err := v.ValidateProject(rootB)
	if err != nil {
		t.Fatalf("validate project copy B: %v", err)
	}
	defer indexB.Close()
	wantTime := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	packA, err := indexA.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: wantTime})
	if err != nil {
		t.Fatalf("build context pack A: %v", err)
	}
	packB, err := indexB.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: wantTime})
	if err != nil {
		t.Fatalf("build context pack B: %v", err)
	}
	if !reflect.DeepEqual(comparableContextPack(packA), comparableContextPack(packB)) {
		t.Fatalf("expected deterministic packs across fresh project copies\npackA: %#v\npackB: %#v", comparableContextPack(packA), comparableContextPack(packB))
	}
}

func TestBuildContextPackCapturesSignedTagMetadata(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	repoDir, details := createSignedContextPackRepo(t)
	projectRoot := writeRootConfigProject(t, fmt.Sprintf("schema_version: 1\nrunecontext_version: 0.1.0-alpha.4\nassurance_tier: plain\nsource:\n  type: git\n  url: %s\n  signed_tag: %s\n  expect_commit: %s\n  subdir: runecontext\n", repoDir, details.SignedTagName, details.Commit))
	loaded, err := v.LoadProject(projectRoot, ResolveOptions{ConfigDiscovery: ConfigDiscoveryExplicitRoot, ExecutionMode: ExecutionModeLocal, GitTrust: GitTrustInputs{SignedTagVerifier: newSSHAllowedSignersVerifierForTest(t, details.AllowedSigners)}})
	if err != nil {
		t.Fatalf("load signed-tag project: %v", err)
	}
	defer loaded.Close()
	index, err := v.ValidateLoadedProject(loaded)
	if err != nil {
		t.Fatalf("validate signed-tag project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"base"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build signed-tag context pack: %v", err)
	}
	assertContextPackValidAgainstSchema(t, v, pack)
	if pack.ResolvedFrom.SourceMode != SourceModeGit {
		t.Fatalf("expected git source mode, got %q", pack.ResolvedFrom.SourceMode)
	}
	if pack.ResolvedFrom.SourceRef != details.SignedTagName {
		t.Fatalf("expected signed tag source_ref %q, got %q", details.SignedTagName, pack.ResolvedFrom.SourceRef)
	}
	if pack.ResolvedFrom.SourceCommit != details.Commit {
		t.Fatalf("expected source commit %q, got %q", details.Commit, pack.ResolvedFrom.SourceCommit)
	}
	if pack.ResolvedFrom.SourceVerification != VerificationPostureVerifiedSignedTag {
		t.Fatalf("expected verified signed tag posture, got %q", pack.ResolvedFrom.SourceVerification)
	}
	if pack.ResolvedFrom.VerifiedSignerIdentity != details.SignerIdentity {
		t.Fatalf("expected signer identity %q, got %q", details.SignerIdentity, pack.ResolvedFrom.VerifiedSignerIdentity)
	}
	if pack.ResolvedFrom.VerifiedSignerFingerprint != details.SignerFingerprint {
		t.Fatalf("expected signer fingerprint %q, got %q", details.SignerFingerprint, pack.ResolvedFrom.VerifiedSignerFingerprint)
	}
}

func TestBuildContextPackRejectsMissingGeneratedAt(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	_, err = index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}})
	if err == nil || !strings.Contains(err.Error(), "explicit generated_at") {
		t.Fatalf("expected missing generated_at error, got %v", err)
	}
}

func TestBuildContextPackRejectsSubSecondGeneratedAt(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	_, err = index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 123, time.UTC)})
	if err == nil || !strings.Contains(err.Error(), "rounded to whole seconds") {
		t.Fatalf("expected whole-second generated_at error, got %v", err)
	}
}

func TestBuildContextPackRejectsInvalidRequestedBundleIDs(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	tests := []struct {
		name      string
		bundleIDs []string
		want      string
	}{
		{name: "empty", bundleIDs: nil, want: "at least one requested bundle ID is required"},
		{name: "whitespace", bundleIDs: []string{"  "}, want: "requested bundle IDs must not be empty"},
		{name: "duplicate", bundleIDs: []string{"left", "left"}, want: "must not contain duplicates"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := index.BuildContextPack(ContextPackOptions{BundleIDs: tc.bundleIDs, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestBuildContextPackNormalizesTextLineEndingsForHashing(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	rootA := t.TempDir()
	rootB := t.TempDir()
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), rootA)
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), rootB)
	missionPath := filepath.Join(rootB, "runecontext", "project", "mission.md")
	data, err := os.ReadFile(missionPath)
	if err != nil {
		t.Fatalf("read mission file: %v", err)
	}
	crlf := strings.ReplaceAll(string(data), "\n", "\r\n")
	if err := os.WriteFile(missionPath, []byte(crlf), 0o644); err != nil {
		t.Fatalf("rewrite mission file with CRLF: %v", err)
	}
	indexA, err := v.ValidateProject(rootA)
	if err != nil {
		t.Fatalf("validate LF project: %v", err)
	}
	defer indexA.Close()
	indexB, err := v.ValidateProject(rootB)
	if err != nil {
		t.Fatalf("validate CRLF project: %v", err)
	}
	defer indexB.Close()
	generatedAt := time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)
	packA, err := indexA.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: generatedAt})
	if err != nil {
		t.Fatalf("build LF pack: %v", err)
	}
	packB, err := indexB.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: generatedAt})
	if err != nil {
		t.Fatalf("build CRLF pack: %v", err)
	}
	if !reflect.DeepEqual(comparableContextPack(packA), comparableContextPack(packB)) {
		t.Fatalf("expected CRLF-normalized pack parity\npackA: %#v\npackB: %#v", comparableContextPack(packA), comparableContextPack(packB))
	}
}

func TestBuildContextPackRejectsSelectedEntriesWithoutProvenance(t *testing.T) {
	_, err := buildContextPackSelectedFiles(filepath.Join(fixturePath(t, "bundle-resolution", "valid-project"), "runecontext"), []BundleInventoryEntry{{Path: "project/mission.md"}})
	if err == nil || !strings.Contains(err.Error(), "missing selector provenance") {
		t.Fatalf("expected selector provenance error, got %v", err)
	}
}

func TestBuildContextPackRejectsNonPortableLocalSourceRef(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{name: "relative portable", source: "local-runecontext", wantErr: false},
		{name: "nested portable", source: "subdir/local-runecontext", wantErr: false},
		{name: "absolute unix", source: "/tmp/local-runecontext", wantErr: true},
		{name: "drive qualified", source: `C:\local-runecontext`, wantErr: true},
		{name: "unc", source: `\\server\share`, wantErr: true},
		{name: "backslash relative", source: `local\runecontext`, wantErr: true},
		{name: "dot segment", source: "./local-runecontext", wantErr: true},
		{name: "dotdot segment", source: "../local-runecontext", wantErr: true},
		{name: "embedded dotdot", source: "local/../runecontext", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildContextPackResolvedFrom(&SourceResolution{SourceMode: SourceModePath, SourceRef: tc.source, VerificationPosture: VerificationPostureUnverifiedLocal}, []string{"base"})
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "portable source_ref") {
					t.Fatalf("expected portability error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected portable source_ref to pass, got %v", err)
			}
		})
	}
}

func TestBuildContextPackReportsMissingSelectedFilePath(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	root := t.TempDir()
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project"), root)
	index, err := v.ValidateProject(root)
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	original := contextPackReadProjectFile
	contextPackReadProjectFile = func(boundaryPath, path string) ([]byte, error) {
		if strings.HasSuffix(filepath.ToSlash(path), "project/mission.md") {
			return nil, fmt.Errorf("synthetic read failure")
		}
		return original(boundaryPath, path)
	}
	defer func() { contextPackReadProjectFile = original }()
	_, err = index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err == nil || !strings.Contains(err.Error(), `hash context-pack file "project/mission.md"`) {
		t.Fatalf("expected wrapped hashing error, got %v", err)
	}
}

func TestResolveRequestReturnsDefensiveClone(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	first, err := index.Bundles.ResolveRequest([]string{"left", "right"})
	if err != nil {
		t.Fatalf("resolve first request: %v", err)
	}
	first.Linearization[0] = "mutated"
	project := first.Aspects[BundleAspectProject]
	project.Selected[0].Path = "mutated.md"
	first.Aspects[BundleAspectProject] = project
	second, err := index.Bundles.ResolveRequest([]string{"left", "right"})
	if err != nil {
		t.Fatalf("resolve second request: %v", err)
	}
	if second.Linearization[0] != "base" {
		t.Fatalf("expected cloned linearization, got %#v", second.Linearization)
	}
	if second.Aspects[BundleAspectProject].Selected[0].Path != "project/mission.md" {
		t.Fatalf("expected cloned aspect inventory, got %#v", second.Aspects[BundleAspectProject].Selected)
	}
}

func TestCanonicalContextPackHashInputIgnoresGeneratedAtAndPackHash(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"child-reinclude"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	first, err := canonicalContextPackHashInput(pack)
	if err != nil {
		t.Fatalf("canonical hash input: %v", err)
	}
	copyPack := *pack
	copyPack.GeneratedAt = "2026-03-21T12:00:00Z"
	copyPack.PackHash = strings.Repeat("f", 64)
	second, err := canonicalContextPackHashInput(&copyPack)
	if err != nil {
		t.Fatalf("canonical hash input copy: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("expected canonical input to ignore generated_at and pack_hash\nfirst:  %s\nsecond: %s", string(first), string(second))
	}
}

func TestBuildContextPackPrimaryIDMatchesFirstRequestedBundleID(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	index, err := v.ValidateProject(fixturePath(t, "bundle-resolution", "valid-project"))
	if err != nil {
		t.Fatalf("validate fixture project: %v", err)
	}
	defer index.Close()
	pack, err := index.BuildContextPack(ContextPackOptions{BundleIDs: []string{"left", "right"}, GeneratedAt: time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("build context pack: %v", err)
	}
	if pack.ID != pack.RequestedBundleIDs[0] {
		t.Fatalf("expected pack id %q to equal first requested bundle ID %q", pack.ID, pack.RequestedBundleIDs[0])
	}
}

func TestMarshalCanonicalJSONSortsKeysAndDisablesHTMLEscaping(t *testing.T) {
	encoded, err := marshalCanonicalJSON(map[string]any{"z": "a<b>&c", "a": "snowman ☃", "n": "line1\nline2"})
	if err != nil {
		t.Fatalf("marshal canonical JSON: %v", err)
	}
	want := `{"a":"snowman ☃","n":"line1\nline2","z":"a<b>&c"}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, string(encoded))
	}
	if strings.Contains(string(encoded), `\u003c`) || strings.Contains(string(encoded), `\u003e`) || strings.Contains(string(encoded), `\u0026`) {
		t.Fatalf("expected HTML characters to remain unescaped, got %s", string(encoded))
	}
}

func TestMarshalCanonicalJSONSupportsUnsignedIntegers(t *testing.T) {
	encoded, err := marshalCanonicalJSON(map[string]any{"count": uint32(7)})
	if err != nil {
		t.Fatalf("marshal canonical JSON: %v", err)
	}
	if string(encoded) != `{"count":7}` {
		t.Fatalf("expected unsigned integer support, got %s", string(encoded))
	}
}

func TestMarshalCanonicalJSONRejectsFloatValues(t *testing.T) {
	_, err := marshalCanonicalJSON(map[string]any{"ratio": 1.5})
	if err == nil || !strings.Contains(err.Error(), "unsupported canonical JSON value float64") {
		t.Fatalf("expected float rejection, got %v", err)
	}
}

func TestContextPackSchemaConstantsMatchMachineContracts(t *testing.T) {
	data := readFixture(t, filepath.Join(schemaRoot(t), "context-pack.schema.json"))
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse context-pack schema: %v", err)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties object in context-pack schema")
	}
	assertSchemaConstValue(t, properties, "canonicalization", contextPackCanonicalization)
	assertSchemaConstValue(t, properties, "pack_hash_alg", contextPackHashAlgorithm)
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("expected required array in context-pack schema")
	}
	for _, field := range []string{"requested_bundle_ids", "excluded", "generated_at"} {
		if !schemaRequiredField(required, field) {
			t.Fatalf("expected context-pack schema required fields to include %q", field)
		}
	}
	requestedBundles, ok := properties["requested_bundle_ids"].(map[string]any)
	if !ok {
		t.Fatal("expected requested_bundle_ids schema property")
	}
	if unique, ok := requestedBundles["uniqueItems"].(bool); !ok || !unique {
		t.Fatalf("expected requested_bundle_ids to require unique items, got %#v", requestedBundles["uniqueItems"])
	}
	assertSchemaPatternValue(t, properties, "id", `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	items, ok := requestedBundles["items"].(map[string]any)
	if !ok {
		t.Fatal("expected requested_bundle_ids items schema")
	}
	if pattern, ok := items["pattern"].(string); !ok || pattern != `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$` {
		t.Fatalf("expected requested_bundle_ids item pattern to match bundle ID grammar, got %#v", items["pattern"])
	}
}

func assertContextPackValidAgainstSchema(t *testing.T, v *Validator, pack *ContextPack) {
	t.Helper()
	data, err := yaml.Marshal(pack)
	if err != nil {
		t.Fatalf("marshal context pack: %v", err)
	}
	if err := v.ValidateYAMLFile("context-pack.schema.json", "generated-context-pack.yaml", data); err != nil {
		t.Fatalf("expected generated context pack to satisfy schema: %v\n%s", err, string(data))
	}
}

func assertContextPackMatchesGolden(t *testing.T, pack *ContextPack, goldenPath string) {
	t.Helper()
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			data, marshalErr := yaml.Marshal(pack)
			if marshalErr != nil {
				t.Fatalf("missing golden %s and marshal failed: %v", goldenPath, marshalErr)
			}
			t.Fatalf("missing golden fixture %s\n%s", goldenPath, string(data))
		}
		t.Fatalf("read golden fixture %s: %v", goldenPath, err)
	}
	expected := normalizeResolutionValue(t, mustParseYAML(t, string(goldenData)))
	actual := normalizeResolutionValue(t, contextPackDocumentValue(pack))
	if !reflect.DeepEqual(actual, expected) {
		data, _ := yaml.Marshal(pack)
		t.Fatalf("context pack mismatch\nexpected: %#v\nactual:   %#v\nactual_yaml:\n%s", expected, actual, string(data))
	}
}

func comparableContextPack(pack *ContextPack) map[string]any {
	if pack == nil {
		return nil
	}
	return contextPackDocumentValue(pack)
}

func comparableContextPackWithoutGeneratedAt(pack *ContextPack) map[string]any {
	result := comparableContextPack(pack)
	delete(result, "generated_at")
	return result
}

func comparableSelectedFiles(items []ContextPackSelectedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":        item.Path,
			"sha256":      item.SHA256,
			"selected_by": comparableRuleReferences(item.SelectedBy),
		}
	}
	return result
}

func comparableExcludedFiles(items []ContextPackExcludedFile) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"path":      item.Path,
			"last_rule": comparableRuleReference(item.LastRule),
		}
	}
	return result
}

func comparableRuleReferences(items []ContextPackRuleReference) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = comparableRuleReference(item)
	}
	return result
}

func comparableRuleReference(item ContextPackRuleReference) map[string]any {
	return map[string]any{
		"bundle":  item.Bundle,
		"aspect":  string(item.Aspect),
		"rule":    string(item.Rule),
		"pattern": item.Pattern,
		"kind":    string(item.Kind),
	}
}

func createSignedContextPackRepo(t *testing.T) (string, signedGitSourceDetails) {
	t.Helper()
	requireToolForContractsTests(t, "git")
	requireToolForContractsTests(t, "ssh-keygen")
	repoDir := t.TempDir()
	runGitTest(t, repoDir, "init", "--initial-branch=main")
	copyDirForTest(t, fixturePath(t, "bundle-resolution", "valid-project", "runecontext"), filepath.Join(repoDir, "runecontext"))
	runGitTest(t, repoDir, "add", ".")
	runGitTest(t, repoDir, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "commit", "-m", "initial runecontext")
	commit := strings.TrimSpace(gitOutputForTest(t, repoDir, "rev-parse", "HEAD"))
	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "signer")
	runCommandForTest(t, repoDir, sanitizedGitEnv(), "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", keyPath)
	publicKey := strings.TrimSpace(string(readFixture(t, keyPath+".pub")))
	allowedSigners := []byte(fmt.Sprintf("alice@example.com %s\n", publicKey))
	signedTagName := "v1.0.0-signed"
	runGitTest(t, repoDir, "-c", "gpg.format=ssh", "-c", "user.signingkey="+keyPath, "-c", "user.name=RuneContext Tests", "-c", "user.email=tests@example.com", "tag", "-s", "-m", "signed tag", signedTagName)
	verifier := newSSHAllowedSignersVerifierForTest(t, allowedSigners)
	verification, err := verifier.VerifySignedTag(repoDir, signedTagName)
	if err != nil {
		t.Fatalf("verify signed context-pack tag: %v", err)
	}
	return repoDir, signedGitSourceDetails{
		Commit:            commit,
		SignedTagName:     signedTagName,
		SignerIdentity:    verification.SignerIdentity,
		SignerFingerprint: verification.SignerFingerprint,
		AllowedSigners:    allowedSigners,
	}
}

func assertSchemaConstValue(t *testing.T, properties map[string]any, field, want string) {
	t.Helper()
	property, ok := properties[field].(map[string]any)
	if !ok {
		t.Fatalf("expected schema property %q", field)
	}
	got, ok := property["const"].(string)
	if !ok {
		t.Fatalf("expected schema property %q to define string const", field)
	}
	if got != want {
		t.Fatalf("expected schema property %q const %q, got %q", field, want, got)
	}
}

func assertSchemaPatternValue(t *testing.T, properties map[string]any, field, want string) {
	t.Helper()
	property, ok := properties[field].(map[string]any)
	if !ok {
		t.Fatalf("expected schema property %q", field)
	}
	got, ok := property["pattern"].(string)
	if !ok {
		t.Fatalf("expected schema property %q to define string pattern", field)
	}
	if got != want {
		t.Fatalf("expected schema property %q pattern %q, got %q", field, want, got)
	}
}

func schemaRequiredField(required []any, want string) bool {
	for _, field := range required {
		if field == want {
			return true
		}
	}
	return false
}
