package contracts

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAllocateChangeID(t *testing.T) {
	contentRoot := t.TempDir()
	changesRoot := filepath.Join(contentRoot, "changes")
	for _, path := range []string{
		filepath.Join(changesRoot, "CHG-2026-001-a3f2-auth-gateway"),
		filepath.Join(changesRoot, "CHG-2026-002-b4c3-auth-revision"),
		filepath.Join(changesRoot, "CHG-2025-003-c9d1-old-change"),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("mkdir existing change: %v", err)
		}
	}
	id, err := AllocateChangeID(contentRoot, time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), "Add Auth Gateway", bytes.NewReader([]byte{0xaa, 0xbb}))
	if err != nil {
		t.Fatalf("allocate change ID: %v", err)
	}
	if got, want := id, "CHG-2026-003-aabb-add-auth-gateway"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAllocateChangeIDSlugifiesToASCII(t *testing.T) {
	contentRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(contentRoot, "changes"), 0o755); err != nil {
		t.Fatalf("mkdir changes root: %v", err)
	}
	title := "Caf" + string(rune(0x00e9)) + " Auth / R" + string(rune(0x00e9)) + "vision"
	id, err := AllocateChangeID(contentRoot, time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC), title, bytes.NewReader([]byte{0xaa, 0xbb}))
	if err != nil {
		t.Fatalf("allocate change ID: %v", err)
	}
	if got, want := id, "CHG-2026-001-aabb-caf-auth-rvision"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCloseChangeStatus(t *testing.T) {
	raw := map[string]any{"status": "implemented", "verification_status": "passed", "superseded_by": []any{"CHG-2026-999-dead-placeholder"}}
	closedAt := time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC)
	closed, err := CloseChangeStatus(raw, CloseChangeOptions{ClosedAt: closedAt})
	if err != nil {
		t.Fatalf("close change: %v", err)
	}
	if got, want := closed["status"], "closed"; got != want {
		t.Fatalf("expected status %q, got %#v", want, got)
	}
	if got, want := closed["closed_at"], "2026-03-18"; got != want {
		t.Fatalf("expected closed_at %q, got %#v", want, got)
	}
	if got := closed["superseded_by"]; !reflect.DeepEqual(got, []any{}) {
		t.Fatalf("expected superseded_by to be cleared, got %#v", got)
	}
	superseded, err := CloseChangeStatus(raw, CloseChangeOptions{ClosedAt: closedAt, SupersededBy: []string{"CHG-2026-010-cafe-successor"}})
	if err != nil {
		t.Fatalf("supersede change: %v", err)
	}
	if got, want := superseded["status"], "superseded"; got != want {
		t.Fatalf("expected status %q, got %#v", want, got)
	}
	if got := superseded["superseded_by"]; !reflect.DeepEqual(got, []any{"CHG-2026-010-cafe-successor"}) {
		t.Fatalf("expected superseded_by to be updated, got %#v", got)
	}
}

func TestCloseChangeStatusRejectsMissingStatus(t *testing.T) {
	_, err := CloseChangeStatus(map[string]any{"verification_status": "passed"}, CloseChangeOptions{})
	if err == nil || !strings.Contains(err.Error(), "valid string status") {
		t.Fatalf("expected missing-status error, got %v", err)
	}
}

func TestCloseChangeStatusRejectsInvalidSupersededBy(t *testing.T) {
	raw := map[string]any{"id": "CHG-2026-001-a3f2-auth-gateway", "status": "implemented"}
	for _, tc := range []struct {
		name         string
		supersededBy []string
		want         string
	}{
		{name: "invalid format", supersededBy: []string{"not-a-change"}, want: "canonical change ID format"},
		{name: "self reference", supersededBy: []string{"CHG-2026-001-a3f2-auth-gateway"}, want: "must not reference the change itself"},
		{name: "duplicate", supersededBy: []string{"CHG-2026-010-cafe-successor", "CHG-2026-010-cafe-successor"}, want: "duplicate value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CloseChangeStatus(raw, CloseChangeOptions{SupersededBy: tc.supersededBy})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestCloneTopLevelValueClonesTypedCollections(t *testing.T) {
	original := map[string]any{"context_bundles": []string{"base", "security"}, "metadata": map[string]string{"owner": "platform"}, "matrix": [2]string{"a", "b"}}
	cloned := cloneMap(original)
	assertClonedCollectionsIndependent(t, original, cloned)
}

func assertClonedCollectionsIndependent(t *testing.T, original, cloned map[string]any) {
	t.Helper()
	clonedBundles := cloned["context_bundles"].([]string)
	clonedBundles[0] = "mutated"
	cloned["context_bundles"] = clonedBundles
	clonedMetadata := cloned["metadata"].(map[string]string)
	clonedMetadata["owner"] = "security"
	cloned["metadata"] = clonedMetadata
	clonedMatrix := cloned["matrix"].([2]string)
	clonedMatrix[0] = "z"
	cloned["matrix"] = clonedMatrix
	if !reflect.DeepEqual(original["context_bundles"].([]string), []string{"base", "security"}) {
		t.Fatalf("expected original bundles to remain unchanged, got %#v", original["context_bundles"])
	}
	if !reflect.DeepEqual(original["metadata"].(map[string]string), map[string]string{"owner": "platform"}) {
		t.Fatalf("expected original metadata to remain unchanged, got %#v", original["metadata"])
	}
	if !reflect.DeepEqual(original["matrix"].([2]string), [2]string{"a", "b"}) {
		t.Fatalf("expected original array to remain unchanged, got %#v", original["matrix"])
	}
}

func TestBuildChangeRecordUsesEmptyStringForMissingOptionalFields(t *testing.T) {
	changeDir := filepath.Join(t.TempDir(), "CHG-2026-001-a3f2-auth-gateway")
	statusPath := filepath.Join(changeDir, "status.yaml")
	record, err := buildChangeRecord(changeDir, statusPath, map[string]any{"id": "CHG-2026-001-a3f2-auth-gateway", "title": "Add auth gateway", "status": "proposed", "type": "feature", "verification_status": "pending", "context_bundles": []any{"base"}, "related_specs": []any{}, "related_decisions": []any{}, "related_changes": []any{}, "depends_on": []any{}, "informed_by": []any{}, "supersedes": []any{}, "superseded_by": []any{}, "closed_at": nil})
	if err != nil {
		t.Fatalf("build change record: %v", err)
	}
	if got := record.Size; got != "" {
		t.Fatalf("expected missing size to remain empty, got %q", got)
	}
	if got := record.Title; got != "Add auth gateway" {
		t.Fatalf("expected title to be preserved, got %q", got)
	}
}

func TestValidateLifecycleTransition(t *testing.T) {
	if err := ValidateLifecycleTransition("planned", "verified"); err != nil {
		t.Fatalf("expected forward transition to succeed: %v", err)
	}
	if err := ValidateLifecycleTransition("verified", "planned"); err == nil {
		t.Fatal("expected backward transition to fail")
	}
	if err := ValidateLifecycleTransition("closed", "superseded"); err == nil {
		t.Fatal("expected terminal transition to fail")
	}
}

func TestBuildSplitChangeGraph(t *testing.T) {
	links, err := BuildSplitChangeGraph(SplitChangePlan{UmbrellaID: "CHG-2026-001-a3f2-umbrella", SubChanges: []SplitSubChange{{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"CHG-2026-001-a3f2-umbrella"}}, {ID: "CHG-2026-003-c9d1-ui", DependsOn: []string{"CHG-2026-002-b4c3-api"}}}})
	if err != nil {
		t.Fatalf("build split graph: %v", err)
	}
	assertSplitGraphLinks(t, links)
}

func assertSplitGraphLinks(t *testing.T, links map[string]ChangeGraphLinks) {
	t.Helper()
	if got := links["CHG-2026-001-a3f2-umbrella"].RelatedChanges; !reflect.DeepEqual(got, []string{"CHG-2026-002-b4c3-api", "CHG-2026-003-c9d1-ui"}) {
		t.Fatalf("unexpected umbrella related changes: %#v", got)
	}
	if got := links["CHG-2026-002-b4c3-api"]; !reflect.DeepEqual(got.RelatedChanges, []string{"CHG-2026-001-a3f2-umbrella", "CHG-2026-003-c9d1-ui"}) || !reflect.DeepEqual(got.DependsOn, []string{"CHG-2026-001-a3f2-umbrella"}) {
		t.Fatalf("unexpected sub-change graph for api: %#v", got)
	}
	if got := links["CHG-2026-003-c9d1-ui"]; !reflect.DeepEqual(got.RelatedChanges, []string{"CHG-2026-001-a3f2-umbrella", "CHG-2026-002-b4c3-api"}) || !reflect.DeepEqual(got.DependsOn, []string{"CHG-2026-002-b4c3-api"}) {
		t.Fatalf("unexpected sub-change graph for ui: %#v", got)
	}
}

func TestBuildSplitChangeGraphAllowsExternalDependency(t *testing.T) {
	links, err := BuildSplitChangeGraph(SplitChangePlan{UmbrellaID: "CHG-2026-001-a3f2-umbrella", SubChanges: []SplitSubChange{{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"CHG-2025-099-dead-external-prereq"}}}})
	if err != nil {
		t.Fatalf("expected external dependency to be allowed: %v", err)
	}
	if got := links["CHG-2026-002-b4c3-api"].DependsOn; !reflect.DeepEqual(got, []string{"CHG-2025-099-dead-external-prereq"}) {
		t.Fatalf("unexpected dependency set: %#v", got)
	}
}

func TestBuildSplitChangeGraphRejectsInvalidDependencyID(t *testing.T) {
	_, err := BuildSplitChangeGraph(SplitChangePlan{UmbrellaID: "CHG-2026-001-a3f2-umbrella", SubChanges: []SplitSubChange{{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"external ticket 42"}}}})
	if err == nil || !strings.Contains(err.Error(), "canonical change ID format") {
		t.Fatalf("expected invalid dependency format error, got %v", err)
	}
}

func TestBuildSplitChangeGraphRejectsCycle(t *testing.T) {
	_, err := BuildSplitChangeGraph(SplitChangePlan{UmbrellaID: "CHG-2026-001-a3f2-umbrella", SubChanges: []SplitSubChange{{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"CHG-2026-003-c9d1-ui"}}, {ID: "CHG-2026-003-c9d1-ui", DependsOn: []string{"CHG-2026-002-b4c3-api"}}}})
	if err == nil || !strings.Contains(err.Error(), "contain a cycle") {
		t.Fatalf("expected cycle detection error, got %v", err)
	}
}

func TestProjectIndexStatusViews(t *testing.T) {
	index := &ProjectIndex{Changes: map[string]*ChangeRecord{"CHG-2026-001-a3f2-open": {Status: StatusProposed}, "CHG-2026-002-b4c3-closed": {Status: StatusClosed}, "CHG-2026-003-c9d1-superseded": {Status: StatusSuperseded}}}
	if got := index.OpenChangeIDs(); !reflect.DeepEqual(got, []string{"CHG-2026-001-a3f2-open"}) {
		t.Fatalf("unexpected open change IDs: %#v", got)
	}
	if got := index.ClosedChangeIDs(); !reflect.DeepEqual(got, []string{"CHG-2026-002-b4c3-closed"}) {
		t.Fatalf("unexpected closed change IDs: %#v", got)
	}
	if got := index.SupersededChangeIDs(); !reflect.DeepEqual(got, []string{"CHG-2026-003-c9d1-superseded"}) {
		t.Fatalf("unexpected superseded change IDs: %#v", got)
	}
}

func TestValidateProjectRejectsNonReciprocalRelatedChanges(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml")
	rewriteFile(t, statusPath, func(text string) string {
		return strings.Replace(text, "related_changes:\n  - CHG-2026-001-a3f2-auth-gateway", "related_changes: []", 1)
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "related_changes must be reciprocal") {
		t.Fatalf("expected reciprocal related_changes failure, got %v", err)
	}
}

func TestValidateProjectRejectsSupersessionInconsistency(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
	if err := os.WriteFile(statusPath, []byte(strings.Join([]string{"schema_version: 1", "id: CHG-2026-001-a3f2-auth-gateway", "title: Add auth gateway", "status: superseded", "type: feature", "size: medium", "verification_status: skipped", "context_bundles:", "  - go-control-plane", "related_specs:", "  - specs/auth-gateway.md", "related_decisions:", "  - decisions/DEC-0001-trust-boundary-model.md", "related_changes:", "  - CHG-2026-002-b4c3-auth-revision", "depends_on: []", "informed_by: []", "supersedes: []", "superseded_by:", "  - CHG-2026-002-b4c3-auth-revision", "created_at: \"2026-03-16\"", "closed_at: \"2026-03-18\"", "promotion_assessment:", "  status: none", "  suggested_targets: []"}, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write superseded status: %v", err)
	}
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "superseded_by must be bidirectionally consistent") {
		t.Fatalf("expected supersession consistency failure, got %v", err)
	}
}

func TestValidateProjectRejectsUnmirroredArtifactTraceability(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	specPath := filepath.Join(root, "runecontext", "specs", "auth-gateway.md")
	rewriteFile(t, specPath, func(text string) string {
		return strings.Replace(text, "revised_by_changes:\n  - CHG-2026-002-b4c3-auth-revision", "revised_by_changes: []", 1)
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	assertTraceabilityValidationError(t, err, "related_specs", filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml"))
}

func TestValidateProjectRejectsUnmirroredDecisionTraceability(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	decisionPath := filepath.Join(root, "runecontext", "decisions", "DEC-0001-trust-boundary-model.md")
	rewriteFile(t, decisionPath, func(text string) string {
		return strings.Replace(text, "related_changes:\n  - CHG-2026-002-b4c3-auth-revision", "related_changes: []", 1)
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	assertTraceabilityValidationError(t, err, "related_decisions", filepath.Join(root, "runecontext", "changes", "CHG-2026-002-b4c3-auth-revision", "status.yaml"))
}

func assertTraceabilityValidationError(t *testing.T, err error, contains, wantPath string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), contains) || !strings.Contains(err.Error(), "status.yaml") {
		t.Fatalf("expected artifact traceability failure, got %v", err)
	}
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %T", err)
	}
	if got := filepath.ToSlash(validationErr.Path); got != filepath.ToSlash(wantPath) {
		t.Fatalf("expected path %q, got %q", wantPath, got)
	}
}

func TestExtractMarkdownHeadingFragmentsAvoidsNaturalSuffixCollisions(t *testing.T) {
	headings, err := extractMarkdownHeadingFragments("# Foo\n# Foo\n# Foo 2\n")
	if err != nil {
		t.Fatalf("extract heading fragments: %v", err)
	}
	expected := map[string]string{"foo": "Foo", "foo-1": "Foo", "foo-2": "Foo 2"}
	if !reflect.DeepEqual(headings, expected) {
		t.Fatalf("unexpected heading fragments: %#v", headings)
	}
}

func TestExtractMarkdownHeadingFragmentsSkipsOccupiedSuffixes(t *testing.T) {
	headings, err := extractMarkdownHeadingFragments("# Foo\n# Foo 1\n# Foo\n")
	if err != nil {
		t.Fatalf("extract heading fragments: %v", err)
	}
	expected := map[string]string{"foo": "Foo", "foo-1": "Foo 1", "foo-2": "Foo"}
	if !reflect.DeepEqual(headings, expected) {
		t.Fatalf("unexpected heading fragments: %#v", headings)
	}
}

func TestExtractMarkdownHeadingFragmentsSlugifiesToASCII(t *testing.T) {
	heading := "# Caf" + string(rune(0x00e9)) + " R" + string(rune(0x00e9)) + "vision\n"
	headings, err := extractMarkdownHeadingFragments(heading)
	if err != nil {
		t.Fatalf("extract heading fragments: %v", err)
	}
	expected := map[string]string{"caf-rvision": "Caf" + string(rune(0x00e9)) + " R" + string(rune(0x00e9)) + "vision"}
	if !reflect.DeepEqual(headings, expected) {
		t.Fatalf("unexpected heading fragments: %#v", headings)
	}
}

func copyTraceabilityFixtureProject(t *testing.T, name string) string {
	t.Helper()
	src := fixturePath(t, "traceability", name)
	dst := t.TempDir()
	copyDirTree(t, src, dst)
	return dst
}

func copyDirTree(t *testing.T, srcRoot, dstRoot string) {
	t.Helper()
	if err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstRoot, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture tree: %v", err)
	}
}

func rewriteFile(t *testing.T, path string, transform func(string) string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	updated := transform(normalized)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
