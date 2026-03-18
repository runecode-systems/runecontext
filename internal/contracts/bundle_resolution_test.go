package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBundleResolutionGoldenFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "bundle-resolution", "valid-project")

	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected valid bundle-resolution fixture to validate: %v", err)
	}
	defer index.Close()

	for _, tc := range []struct {
		name     string
		bundleID string
		golden   string
	}{
		{name: "child reinclude override", bundleID: "child-reinclude", golden: "child-reinclude.yaml"},
		{name: "diamond inheritance precedence", bundleID: "diamond", golden: "diamond.yaml"},
		{name: "exact include exclude", bundleID: "exact-rules", golden: "exact-rules.yaml"},
		{name: "relative glob normalization", bundleID: "relative-glob", golden: "relative-glob.yaml"},
		{name: "diagnostics and empty globs", bundleID: "diagnostics", golden: "diagnostics.yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolution, err := index.ResolveBundle(tc.bundleID)
			if err != nil {
				t.Fatalf("resolve bundle %q: %v", tc.bundleID, err)
			}
			assertBundleResolutionMatchesGolden(t, resolution, fixturePath(t, "bundle-resolution", "golden", tc.golden))
		})
	}

	resolution, err := index.ResolveBundle("diamond")
	if err != nil {
		t.Fatalf("resolve diamond bundle: %v", err)
	}
	count := 0
	for _, id := range resolution.Linearization {
		if id == "base" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected duplicate ancestor collapse to keep base once, got linearization %v", resolution.Linearization)
	}
}

func TestBundleResolutionRejectFixtures(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	for _, tc := range []struct {
		name       string
		fixtureDir string
		contains   string
	}{
		{name: "reject unknown parent", fixtureDir: "reject-unknown-parent", contains: "extends unknown parent"},
		{name: "reject duplicate id", fixtureDir: "reject-duplicate-id", contains: "bundle id \"duplicate\" is duplicated"},
		{name: "reject cycle", fixtureDir: "reject-cycle", contains: "inheritance cycle detected"},
		{name: "reject depth", fixtureDir: "reject-depth", contains: "exceeds maximum of 8"},
		{name: "reject traversal", fixtureDir: "reject-traversal", contains: "must not contain traversal segments"},
		{name: "reject absolute", fixtureDir: "reject-absolute", contains: "must not be absolute or drive-qualified"},
		{name: "reject drive qualified", fixtureDir: "reject-drive-qualified", contains: "must not be absolute or drive-qualified"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := v.ValidateProject(fixturePath(t, "bundle-resolution", tc.fixtureDir))
			if err == nil {
				t.Fatal("expected validation failure")
			}
			if !strings.Contains(err.Error(), tc.contains) {
				t.Fatalf("expected error to contain %q, got %v", tc.contains, err)
			}
		})
	}
}

func TestBundleResolutionWarnsOnDeprecatedStandards(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "bundle-resolution", "valid-project")
	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected valid bundle-resolution fixture to validate: %v", err)
	}
	defer index.Close()

	resolution, err := index.ResolveBundle("child-reinclude")
	if err != nil {
		t.Fatalf("resolve child-reinclude bundle: %v", err)
	}
	found := false
	for _, diagnostic := range resolution.Diagnostics {
		if diagnostic.Code == "deprecated_standard_selected" && len(diagnostic.Matches) == 1 && diagnostic.Matches[0] == "standards/global/legacy.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected deprecated standard diagnostic, got %#v", resolution.Diagnostics)
	}
}

func TestBundleResolutionRejectsDraftStandardSelection(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := copyBundleFixtureProject(t, "valid-project")
	bundlePath := filepath.Join(projectRoot, "runecontext", "bundles", "draft-standard.yaml")
	if err := os.WriteFile(bundlePath, []byte(strings.Join([]string{
		"schema_version: 1",
		"id: draft-standard",
		"includes:",
		"  standards:",
		"    - standards/frontend/example.md",
	}, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write draft bundle: %v", err)
	}

	_, err := v.ValidateProject(projectRoot)
	if err == nil || !strings.Contains(err.Error(), "selects draft standard") {
		t.Fatalf("expected draft standard bundle selection failure, got %v", err)
	}
}

func TestBundleResolutionRejectsAspectEscapeSymlink(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := copyBundleFixtureProject(t, "reject-aspect-escape-symlink")
	if err := os.MkdirAll(filepath.Join(projectRoot, "runecontext", "project"), 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := tryCreateSymlink(filepath.Join("..", "standards", "global", "base.md"), filepath.Join(projectRoot, "runecontext", "project", "escape.md")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	_, err := v.ValidateProject(projectRoot)
	if err == nil || !strings.Contains(err.Error(), "escapes the selected aspect root") {
		t.Fatalf("expected aspect-root symlink escape to fail, got %v", err)
	}
}

func TestBundleResolutionRejectsRootEscapeSymlink(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := copyBundleFixtureProject(t, "reject-root-escape-symlink")
	if err := os.MkdirAll(filepath.Join(projectRoot, "runecontext", "project"), 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	if err := tryCreateSymlink(filepath.Join("..", "..", "outside.md"), filepath.Join(projectRoot, "runecontext", "project", "escape.md")); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	_, err := v.ValidateProject(projectRoot)
	if err == nil || !strings.Contains(err.Error(), "escapes the RuneContext root") {
		t.Fatalf("expected root-escape symlink to fail, got %v", err)
	}
}

func TestBundleResolutionAllowsSymlinkedAspectRootWhenInBounds(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := copyBundleFixtureProject(t, "valid-project")
	contentRoot := filepath.Join(projectRoot, "runecontext")
	original := filepath.Join(contentRoot, "standards")
	target := filepath.Join(contentRoot, "standards-real")
	if err := os.Rename(original, target); err != nil {
		t.Fatalf("rename standards dir: %v", err)
	}
	if err := tryCreateSymlink("standards-real", original); err != nil {
		if strings.Contains(err.Error(), "symlink tests skipped") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected symlinked standards root to validate: %v", err)
	}
	defer index.Close()

	resolution, err := index.ResolveBundle("diamond")
	if err != nil {
		t.Fatalf("resolve bundle with symlinked standards root: %v", err)
	}
	selected := resolution.Aspects[BundleAspectStandards].Selected
	if len(selected) == 0 {
		t.Fatalf("expected standards selections with symlinked standards root")
	}
}

func assertBundleResolutionMatchesGolden(t *testing.T, resolution *BundleResolution, goldenPath string) {
	t.Helper()
	if resolution == nil {
		t.Fatal("expected bundle resolution")
	}
	expected := normalizeResolutionValue(t, mustParseYAML(t, string(readFixture(t, goldenPath))))
	actual := normalizeResolutionValue(t, comparableBundleResolution(resolution))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bundle resolution mismatch\nexpected: %#v\nactual:   %#v", expected, actual)
	}
}

func comparableBundleResolution(resolution *BundleResolution) map[string]any {
	result := map[string]any{
		"id":            resolution.ID,
		"linearization": append([]string(nil), resolution.Linearization...),
		"aspects":       map[string]any{},
	}
	aspects := result["aspects"].(map[string]any)
	for _, aspect := range bundleAspects {
		aspectResolution, ok := resolution.Aspects[aspect]
		if !ok {
			continue
		}
		aspects[string(aspect)] = map[string]any{
			"rules":    comparableBundleRuleEvaluations(aspectResolution.Rules),
			"selected": comparableBundleInventoryEntries(aspectResolution.Selected),
			"excluded": comparableBundleInventoryEntries(aspectResolution.Excluded),
		}
	}
	if len(resolution.Diagnostics) > 0 {
		result["diagnostics"] = comparableBundleDiagnostics(resolution.Diagnostics)
	}
	return result
}

func comparableBundleRuleEvaluations(evaluations []BundleRuleEvaluation) []any {
	result := make([]any, 0, len(evaluations))
	for _, evaluation := range evaluations {
		item := map[string]any{
			"bundle":       evaluation.Bundle,
			"aspect":       string(evaluation.Aspect),
			"rule":         string(evaluation.Rule),
			"pattern":      evaluation.Pattern,
			"pattern_kind": string(evaluation.PatternKind),
			"matches":      append([]string(nil), evaluation.Matches...),
		}
		if len(evaluation.Diagnostics) > 0 {
			item["diagnostics"] = comparableBundleDiagnostics(evaluation.Diagnostics)
		}
		result = append(result, item)
	}
	return result
}

func comparableBundleInventoryEntries(entries []BundleInventoryEntry) []any {
	result := make([]any, 0, len(entries))
	for _, entry := range entries {
		result = append(result, map[string]any{
			"path":       entry.Path,
			"matched_by": comparableBundleRuleReferences(entry.MatchedBy),
			"final_rule": comparableBundleRuleReference(entry.FinalRule),
		})
	}
	return result
}

func comparableBundleRuleReferences(refs []BundleRuleReference) []any {
	result := make([]any, 0, len(refs))
	for _, ref := range refs {
		result = append(result, comparableBundleRuleReference(ref))
	}
	return result
}

func comparableBundleRuleReference(ref BundleRuleReference) map[string]any {
	return map[string]any{
		"bundle":  ref.Bundle,
		"aspect":  string(ref.Aspect),
		"rule":    string(ref.Rule),
		"pattern": ref.Pattern,
		"kind":    string(ref.Kind),
	}
}

func comparableBundleDiagnostics(diagnostics []BundleDiagnostic) []any {
	result := make([]any, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		item := map[string]any{
			"severity": string(diagnostic.Severity),
			"code":     diagnostic.Code,
			"message":  diagnostic.Message,
		}
		if diagnostic.Bundle != "" {
			item["bundle"] = diagnostic.Bundle
		}
		if diagnostic.Aspect != "" {
			item["aspect"] = string(diagnostic.Aspect)
		}
		if diagnostic.Rule != "" {
			item["rule"] = string(diagnostic.Rule)
		}
		if diagnostic.Pattern != "" {
			item["pattern"] = diagnostic.Pattern
		}
		if len(diagnostic.Matches) > 0 {
			item["matches"] = append([]string(nil), diagnostic.Matches...)
		}
		result = append(result, item)
	}
	return result
}

func copyBundleFixtureProject(t *testing.T, name string) string {
	t.Helper()
	src := fixturePath(t, "bundle-resolution", name)
	dst := filepath.Join(t.TempDir(), name)
	copyDirForTest(t, src, dst)
	return dst
}

func TestNormalizeBundlePatternRejectsMismatchedAspectPrefix(t *testing.T) {
	_, _, err := normalizeBundlePattern(BundleAspectProject, "standards/global/base.md")
	if err == nil || !strings.Contains(err.Error(), "must stay within the \"project\" aspect") {
		t.Fatalf("expected mismatched aspect prefix rejection, got %v", err)
	}
}

func TestMatchBundlePattern(t *testing.T) {
	for _, tc := range []struct {
		pattern   string
		candidate string
		want      bool
	}{
		{pattern: "standards/security/**", candidate: "standards/security/heavy-audit.md", want: true},
		{pattern: "standards/*/base.md", candidate: "standards/global/base.md", want: true},
		{pattern: "project/**", candidate: "standards/global/base.md", want: false},
		{pattern: "project/*", candidate: "project/notes/todo.md", want: false},
	} {
		if got := matchBundlePattern(tc.pattern, tc.candidate); got != tc.want {
			t.Fatalf("matchBundlePattern(%q, %q) = %v, want %v", tc.pattern, tc.candidate, got, tc.want)
		}
	}
}

func TestBundleResolutionResolveUnknownBundle(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "bundle-resolution", "valid-project")
	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected valid bundle-resolution fixture to validate: %v", err)
	}
	defer index.Close()

	_, err = index.ResolveBundle("does-not-exist")
	if err == nil || !strings.Contains(err.Error(), "unknown bundle") {
		t.Fatalf("expected unknown bundle error, got %v", err)
	}
}

func TestBundleResolutionReturnsDefensiveCopies(t *testing.T) {
	v := NewValidator(schemaRoot(t))
	projectRoot := fixturePath(t, "bundle-resolution", "valid-project")
	index, err := v.ValidateProject(projectRoot)
	if err != nil {
		t.Fatalf("expected valid bundle-resolution fixture to validate: %v", err)
	}
	defer index.Close()

	first, err := index.ResolveBundle("diamond")
	if err != nil {
		t.Fatalf("resolve first bundle copy: %v", err)
	}
	first.Linearization[0] = "mutated"
	first.Aspects[BundleAspectStandards] = BundleAspectResolution{}

	second, err := index.ResolveBundle("diamond")
	if err != nil {
		t.Fatalf("resolve second bundle copy: %v", err)
	}
	if len(second.Linearization) == 0 || second.Linearization[0] != "base" {
		t.Fatalf("expected cached bundle resolution to remain unchanged, got %v", second.Linearization)
	}
	if len(second.Aspects[BundleAspectStandards].Rules) == 0 {
		t.Fatalf("expected cached standards aspect to remain populated")
	}
}

func ExampleBundleCatalog_Resolve() {
	fmt.Println("bundle resolution is covered by unit tests")
	// Output: bundle resolution is covered by unit tests
}
