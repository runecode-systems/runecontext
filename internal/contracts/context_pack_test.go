package contracts

import (
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
	rewriteMissionFilesForLineEndingParityTest(t, rootA, missionPath, data)
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

func rewriteMissionFilesForLineEndingParityTest(t *testing.T, rootA, missionPath string, data []byte) {
	t.Helper()
	lf := strings.ReplaceAll(string(data), "\r\n", "\n")
	lf = strings.ReplaceAll(lf, "\r", "\n")
	crlf := strings.ReplaceAll(lf, "\n", "\r\n")
	if err := os.WriteFile(filepath.Join(rootA, "runecontext", "project", "mission.md"), []byte(lf), 0o644); err != nil {
		t.Fatalf("rewrite mission file with LF: %v", err)
	}
	if err := os.WriteFile(missionPath, []byte(crlf), 0o644); err != nil {
		t.Fatalf("rewrite mission file with CRLF: %v", err)
	}
}

func TestBuildContextPackRejectsSelectedEntriesWithoutProvenance(t *testing.T) {
	_, _, err := buildContextPackSelectedFiles(filepath.Join(fixturePath(t, "bundle-resolution", "valid-project"), "runecontext"), []BundleInventoryEntry{{Path: "project/mission.md"}})
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
			assertContextPackLocalSourceRefResult(t, tc.source, tc.wantErr)
		})
	}
}

func assertContextPackLocalSourceRefResult(t *testing.T, source string, wantErr bool) {
	t.Helper()
	_, err := buildContextPackResolvedFrom(&SourceResolution{SourceMode: SourceModePath, SourceRef: source, VerificationPosture: VerificationPostureUnverifiedLocal}, []string{"base"})
	if wantErr {
		if err == nil || !strings.Contains(err.Error(), "portable source_ref") {
			t.Fatalf("expected portability error, got %v", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("expected portable source_ref to pass, got %v", err)
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
	original := currentContextPackReadProjectFile()
	restore := setContextPackReadProjectFileHookForTest(func(boundaryPath, path string) ([]byte, error) {
		if strings.HasSuffix(filepath.ToSlash(path), "project/mission.md") {
			return nil, fmt.Errorf("synthetic read failure")
		}
		return original(boundaryPath, path)
	})
	defer restore()
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
