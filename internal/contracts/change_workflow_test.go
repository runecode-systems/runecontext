package contracts

import (
	"bytes"
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
	if err := os.MkdirAll(filepath.Join(changesRoot, "CHG-2026-001-a3f2-auth-gateway"), 0o755); err != nil {
		t.Fatalf("mkdir existing change: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(changesRoot, "CHG-2026-002-b4c3-auth-revision"), 0o755); err != nil {
		t.Fatalf("mkdir existing change: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(changesRoot, "CHG-2025-003-c9d1-old-change"), 0o755); err != nil {
		t.Fatalf("mkdir prior year change: %v", err)
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
	raw := map[string]any{
		"status":              "implemented",
		"verification_status": "passed",
		"superseded_by":       []any{"CHG-2026-999-dead-placeholder"},
	}
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
	raw := map[string]any{
		"id":     "CHG-2026-001-a3f2-auth-gateway",
		"status": "implemented",
	}
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
	links, err := BuildSplitChangeGraph(SplitChangePlan{
		UmbrellaID: "CHG-2026-001-a3f2-umbrella",
		SubChanges: []SplitSubChange{
			{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"CHG-2026-001-a3f2-umbrella"}},
			{ID: "CHG-2026-003-c9d1-ui", DependsOn: []string{"CHG-2026-002-b4c3-api"}},
		},
	})
	if err != nil {
		t.Fatalf("build split graph: %v", err)
	}
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
	links, err := BuildSplitChangeGraph(SplitChangePlan{
		UmbrellaID: "CHG-2026-001-a3f2-umbrella",
		SubChanges: []SplitSubChange{{
			ID:        "CHG-2026-002-b4c3-api",
			DependsOn: []string{"CHG-2025-099-dead-external-prereq"},
		}},
	})
	if err != nil {
		t.Fatalf("expected external dependency to be allowed: %v", err)
	}
	if got := links["CHG-2026-002-b4c3-api"].DependsOn; !reflect.DeepEqual(got, []string{"CHG-2025-099-dead-external-prereq"}) {
		t.Fatalf("unexpected dependency set: %#v", got)
	}
}

func TestBuildSplitChangeGraphRejectsInvalidDependencyID(t *testing.T) {
	_, err := BuildSplitChangeGraph(SplitChangePlan{
		UmbrellaID: "CHG-2026-001-a3f2-umbrella",
		SubChanges: []SplitSubChange{{
			ID:        "CHG-2026-002-b4c3-api",
			DependsOn: []string{"external ticket 42"},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "canonical change ID format") {
		t.Fatalf("expected invalid dependency format error, got %v", err)
	}
}

func TestBuildSplitChangeGraphRejectsCycle(t *testing.T) {
	_, err := BuildSplitChangeGraph(SplitChangePlan{
		UmbrellaID: "CHG-2026-001-a3f2-umbrella",
		SubChanges: []SplitSubChange{
			{ID: "CHG-2026-002-b4c3-api", DependsOn: []string{"CHG-2026-003-c9d1-ui"}},
			{ID: "CHG-2026-003-c9d1-ui", DependsOn: []string{"CHG-2026-002-b4c3-api"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "contain a cycle") {
		t.Fatalf("expected cycle detection error, got %v", err)
	}
}

func TestRewriteMarkdownReferenceTargets(t *testing.T) {
	data := []byte("See specs/auth-gateway.md#auth-gateway and decisions/DEC-0001-trust-boundary-model.md#trust-boundary-model.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{
		OldPath:     "specs/auth-gateway.md",
		NewPath:     "specs/security/auth-gateway.md",
		OldFragment: "auth-gateway",
		NewFragment: "gateway-overview",
	}})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	text := string(rewritten)
	if !strings.Contains(text, "specs/security/auth-gateway.md#gateway-overview") {
		t.Fatalf("expected rewritten spec ref, got %q", text)
	}
	if !strings.Contains(text, "decisions/DEC-0001-trust-boundary-model.md#trust-boundary-model") {
		t.Fatalf("expected unrelated ref to stay intact, got %q", text)
	}
}

func TestRewriteMarkdownReferenceTargetsIgnoresFencedCode(t *testing.T) {
	data := []byte("```md\nspecs/auth-gateway.md#auth-gateway\n```\n\nSee specs/auth-gateway.md#auth-gateway.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{
		OldPath:     "specs/auth-gateway.md",
		NewPath:     "specs/security/auth-gateway.md",
		OldFragment: "auth-gateway",
		NewFragment: "gateway-overview",
	}})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	text := string(rewritten)
	if !strings.Contains(text, "```md\nspecs/auth-gateway.md#auth-gateway\n```") {
		t.Fatalf("expected fenced code ref to stay unchanged, got %q", text)
	}
	if !strings.Contains(text, "See specs/security/auth-gateway.md#gateway-overview.") {
		t.Fatalf("expected prose ref to be rewritten, got %q", text)
	}
}

func TestRewriteMarkdownReferenceTargetsIgnoresLongerFenceNesting(t *testing.T) {
	data := []byte("````md\n```\nspecs/auth-gateway.md#auth-gateway\n```\n````\n\nSee specs/auth-gateway.md#auth-gateway.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{{
		OldPath:     "specs/auth-gateway.md",
		NewPath:     "specs/security/auth-gateway.md",
		OldFragment: "auth-gateway",
		NewFragment: "gateway-overview",
	}})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	text := string(rewritten)
	if !strings.Contains(text, "````md\n```\nspecs/auth-gateway.md#auth-gateway\n```\n````") {
		t.Fatalf("expected nested fenced code ref to stay unchanged, got %q", text)
	}
}

func TestRewriteMarkdownReferenceTargetsUsesFirstMatch(t *testing.T) {
	data := []byte("See specs/auth-gateway.md#auth-gateway.\n")
	rewritten, count, err := RewriteMarkdownReferenceTargets(data, []MarkdownReferenceRewrite{
		{OldPath: "specs/auth-gateway.md", NewPath: "specs/security/auth-gateway.md"},
		{OldPath: "specs/auth-gateway.md", OldFragment: "auth-gateway", NewFragment: "gateway-overview"},
	})
	if err != nil {
		t.Fatalf("rewrite markdown refs: %v", err)
	}
	if got, want := count, 1; got != want {
		t.Fatalf("expected %d rewrite, got %d", want, got)
	}
	if got, want := string(rewritten), "See specs/security/auth-gateway.md#auth-gateway.\n"; got != want {
		t.Fatalf("unexpected rewrite result: %q", got)
	}
}

func TestProjectIndexStatusViews(t *testing.T) {
	index := &ProjectIndex{Changes: map[string]*ChangeRecord{
		"CHG-2026-001-a3f2-open":       {Status: StatusProposed},
		"CHG-2026-002-b4c3-closed":     {Status: StatusClosed},
		"CHG-2026-003-c9d1-superseded": {Status: StatusSuperseded},
	}}
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
	if err := os.WriteFile(statusPath, []byte(strings.Join([]string{
		"schema_version: 1",
		"id: CHG-2026-001-a3f2-auth-gateway",
		"title: Add auth gateway",
		"status: superseded",
		"type: feature",
		"size: medium",
		"verification_status: skipped",
		"context_bundles:",
		"  - go-control-plane",
		"related_specs:",
		"  - specs/auth-gateway.md",
		"related_decisions:",
		"  - decisions/DEC-0001-trust-boundary-model.md",
		"related_changes:",
		"  - CHG-2026-002-b4c3-auth-revision",
		"depends_on: []",
		"informed_by: []",
		"supersedes: []",
		"superseded_by:",
		"  - CHG-2026-002-b4c3-auth-revision",
		"created_at: \"2026-03-16\"",
		"closed_at: \"2026-03-18\"",
		"promotion_assessment:",
		"  status: none",
		"  suggested_targets: []",
	}, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write superseded status: %v", err)
	}
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "superseded_by must be bidirectionally consistent") {
		t.Fatalf("expected supersession consistency failure, got %v", err)
	}
}

func TestValidateProjectRejectsInvalidTerminalMetadata(t *testing.T) {
	t.Run("terminal requires closed_at", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
		rewriteFile(t, statusPath, func(text string) string {
			text = strings.Replace(text, "status: proposed", "status: closed", 1)
			text = strings.Replace(text, "verification_status: pending", "verification_status: passed", 1)
			return text
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "requires closed_at") {
			t.Fatalf("expected closed_at failure, got %v", err)
		}
	})

	t.Run("non-terminal must not set closed_at", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
		rewriteFile(t, statusPath, func(text string) string {
			return strings.Replace(text, "closed_at: null", "closed_at: \"2026-03-18\"", 1)
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "must not set closed_at") {
			t.Fatalf("expected non-terminal closed_at failure, got %v", err)
		}
	})

	t.Run("closed must not keep pending verification", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		statusPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "status.yaml")
		rewriteFile(t, statusPath, func(text string) string {
			text = strings.Replace(text, "status: proposed", "status: closed", 1)
			text = strings.Replace(text, "closed_at: null", "closed_at: \"2026-03-18\"", 1)
			return text
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "must not leave verification_status pending") {
			t.Fatalf("expected pending verification failure, got %v", err)
		}
	})
}

func TestValidateProjectRejectsUnmirroredArtifactTraceability(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	specPath := filepath.Join(root, "runecontext", "specs", "auth-gateway.md")
	rewriteFile(t, specPath, func(text string) string {
		return strings.Replace(text, "revised_by_changes:\n  - CHG-2026-002-b4c3-auth-revision", "revised_by_changes: []", 1)
	})
	v := NewValidator(schemaRoot(t))
	_, err := v.ValidateProject(root)
	if err == nil || !strings.Contains(err.Error(), "related_specs entry") {
		t.Fatalf("expected artifact traceability failure, got %v", err)
	}
}

func TestValidateProjectMarkdownDeepRefs(t *testing.T) {
	t.Run("valid heading fragment", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nSee specs/auth-gateway.md#auth-gateway for the stable reference.\n"
		})
		v := NewValidator(schemaRoot(t))
		if _, err := v.ValidateProject(root); err != nil {
			t.Fatalf("expected valid heading-fragment ref: %v", err)
		}
	})

	t.Run("missing heading fragment", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nSee specs/auth-gateway.md#missing-heading for more context.\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "missing heading fragment") {
			t.Fatalf("expected missing-heading failure, got %v", err)
		}
	})

	t.Run("line-number fragment rejected", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nAvoid specs/auth-gateway.md#L10 because line numbers are not durable.\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "must use a heading fragment") {
			t.Fatalf("expected line-number fragment failure, got %v", err)
		}
	})

	t.Run("absolute path rejected", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nBad ref: /specs/auth-gateway.md#auth-gateway\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "must not use an absolute path") {
			t.Fatalf("expected absolute-path failure, got %v", err)
		}
	})

	t.Run("relative traversal rejected", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nBad ref: ../specs/auth-gateway.md#auth-gateway\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "RuneContext-root-relative path") {
			t.Fatalf("expected traversal-path failure, got %v", err)
		}
	})

	t.Run("fenced code refs ignored", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\n```md\n/specs/auth-gateway.md#L10\n```\n"
		})
		v := NewValidator(schemaRoot(t))
		if _, err := v.ValidateProject(root); err != nil {
			t.Fatalf("expected fenced-code ref to be ignored, got %v", err)
		}
	})

	t.Run("blockquote fenced code refs ignored", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\n> ```md\n> /specs/auth-gateway.md#L10\n> ```\n"
		})
		v := NewValidator(schemaRoot(t))
		if _, err := v.ValidateProject(root); err != nil {
			t.Fatalf("expected blockquote fenced-code ref to be ignored, got %v", err)
		}
	})

	t.Run("numeric line fragment rejected", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nAvoid specs/auth-gateway.md#42 because line numbers are not durable.\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || !strings.Contains(err.Error(), "must use a heading fragment") {
			t.Fatalf("expected numeric-line fragment failure, got %v", err)
		}
	})

	t.Run("line range fragment rejected", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nAvoid specs/auth-gateway.md#l10-l20 because line ranges are not durable.\n"
		})
		v := NewValidator(schemaRoot(t))
		_, err := v.ValidateProject(root)
		if err == nil || (!strings.Contains(err.Error(), "must use a heading fragment") && !strings.Contains(err.Error(), "missing heading fragment")) {
			t.Fatalf("expected line-range fragment failure, got %v", err)
		}
	})

	t.Run("external markdown URL ignored", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nReference https://example.com/docs/auth-gateway.md#overview for external context.\n"
		})
		v := NewValidator(schemaRoot(t))
		if _, err := v.ValidateProject(root); err != nil {
			t.Fatalf("expected external markdown URL to be ignored, got %v", err)
		}
	})

	t.Run("project markdown target allowed", func(t *testing.T) {
		root := copyTraceabilityFixtureProject(t, "valid-project")
		projectDir := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway")
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatalf("mkdir change dir: %v", err)
		}
		projectPath := filepath.Join(projectDir, "mission.md")
		if err := os.WriteFile(projectPath, []byte("# Mission\n\nShip stable change workflows.\n"), 0o644); err != nil {
			t.Fatalf("write project markdown: %v", err)
		}
		proposalPath := filepath.Join(root, "runecontext", "changes", "CHG-2026-001-a3f2-auth-gateway", "proposal.md")
		rewriteFile(t, proposalPath, func(text string) string {
			return text + "\nSee changes/CHG-2026-001-a3f2-auth-gateway/mission.md#mission for the high-level intent.\n"
		})
		v := NewValidator(schemaRoot(t))
		if _, err := v.ValidateProject(root); err != nil {
			t.Fatalf("expected indexed change markdown deep ref to validate, got %v", err)
		}
	})
}

func TestExtractMarkdownHeadingFragmentsAvoidsNaturalSuffixCollisions(t *testing.T) {
	headings, err := extractMarkdownHeadingFragments("# Foo\n# Foo\n# Foo 2\n")
	if err != nil {
		t.Fatalf("extract heading fragments: %v", err)
	}
	expected := map[string]string{
		"foo":   "Foo",
		"foo-3": "Foo",
		"foo-2": "Foo 2",
	}
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
