package contracts

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildProjectStatusSummaryIncludesRelationshipVerificationAndRecencyMetadata(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v, loaded := mustLoadWorkflowProject(t, root)
	defer loaded.Close()

	summary, err := BuildProjectStatusSummary(v, loaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}

	if len(summary.Active) != 2 || len(summary.Closed) != 0 || len(summary.Superseded) != 0 {
		t.Fatalf("unexpected status grouping: active=%d closed=%d superseded=%d", len(summary.Active), len(summary.Closed), len(summary.Superseded))
	}
	assertActiveStatusRelationshipEntry(t, mustFindStatusEntryByID(t, summary.Active, "CHG-2026-001-a3f2-auth-gateway"))
	assertActiveStatusDependencyEntry(t, mustFindStatusEntryByID(t, summary.Active, "CHG-2026-002-b4c3-auth-revision"))
}

func assertActiveStatusRelationshipEntry(t *testing.T, first ChangeStatusEntry) {
	t.Helper()
	if got, want := first.VerificationStatus, "pending"; got != want {
		t.Fatalf("expected first verification_status %q, got %q", want, got)
	}
	if got, want := first.RelatedChanges, []string{"CHG-2026-002-b4c3-auth-revision"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected first related_changes %#v, got %#v", want, got)
	}
	if len(first.DependsOn) != 0 || len(first.Supersedes) != 0 || len(first.SupersededBy) != 0 {
		t.Fatalf("expected first dependency/supersession lists empty, got depends_on=%#v supersedes=%#v superseded_by=%#v", first.DependsOn, first.Supersedes, first.SupersededBy)
	}
	if got, want := first.CreatedAt, "2026-03-16"; got != want {
		t.Fatalf("expected first created_at %q, got %q", want, got)
	}
	if first.ClosedAt != "" {
		t.Fatalf("expected first closed_at empty, got %q", first.ClosedAt)
	}
}

func assertActiveStatusDependencyEntry(t *testing.T, second ChangeStatusEntry) {
	t.Helper()
	if got, want := second.VerificationStatus, "pending"; got != want {
		t.Fatalf("expected second verification_status %q, got %q", want, got)
	}
	if got, want := second.DependsOn, []string{"CHG-2026-001-a3f2-auth-gateway"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected second depends_on %#v, got %#v", want, got)
	}
	if got, want := second.RelatedChanges, []string{"CHG-2026-001-a3f2-auth-gateway"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected second related_changes %#v, got %#v", want, got)
	}
	if got, want := second.CreatedAt, "2026-03-17"; got != want {
		t.Fatalf("expected second created_at %q, got %q", want, got)
	}
	if second.ClosedAt != "" {
		t.Fatalf("expected second closed_at empty, got %q", second.ClosedAt)
	}
}

func TestBuildProjectStatusSummaryIncludesTerminalRecencyAndSupersessionMetadata(t *testing.T) {
	root := copyTraceabilityFixtureProject(t, "valid-project")
	v, loaded := mustLoadWorkflowProject(t, root)
	if _, err := CloseChange(v, loaded, "CHG-2026-001-a3f2-auth-gateway", ChangeCloseOptions{
		VerificationStatus: "skipped",
		ClosedAt:           time.Date(2026, time.March, 18, 0, 0, 0, 0, time.UTC),
		SupersededBy:       []string{"CHG-2026-002-b4c3-auth-revision"},
	}); err != nil {
		loaded.Close()
		t.Fatalf("supersede change: %v", err)
	}
	loaded.Close()

	reloaded := mustReloadWorkflowProject(t, v, root)
	defer reloaded.Close()
	summary, err := BuildProjectStatusSummary(v, reloaded)
	if err != nil {
		t.Fatalf("build status summary: %v", err)
	}

	if len(summary.Active) != 1 || len(summary.Closed) != 0 || len(summary.Superseded) != 1 {
		t.Fatalf("unexpected status grouping: active=%d closed=%d superseded=%d", len(summary.Active), len(summary.Closed), len(summary.Superseded))
	}
	assertSupersededStatusEntry(t, mustFindStatusEntryByID(t, summary.Superseded, "CHG-2026-001-a3f2-auth-gateway"))
}

func assertSupersededStatusEntry(t *testing.T, entry ChangeStatusEntry) {
	t.Helper()
	if got, want := entry.VerificationStatus, "skipped"; got != want {
		t.Fatalf("expected verification_status %q, got %q", want, got)
	}
	if got, want := entry.SupersededBy, []string{"CHG-2026-002-b4c3-auth-revision"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected superseded_by %#v, got %#v", want, got)
	}
	if got, want := entry.RelatedChanges, []string{"CHG-2026-002-b4c3-auth-revision"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expected related_changes %#v, got %#v", want, got)
	}
	if got, want := entry.CreatedAt, "2026-03-16"; got != want {
		t.Fatalf("expected created_at %q, got %q", want, got)
	}
	if got, want := entry.ClosedAt, "2026-03-18"; got != want {
		t.Fatalf("expected closed_at %q, got %q", want, got)
	}
}

func mustFindStatusEntryByID(t *testing.T, entries []ChangeStatusEntry, id string) ChangeStatusEntry {
	t.Helper()
	for _, entry := range entries {
		if entry.ID == id {
			return entry
		}
	}
	t.Fatalf("missing status entry %q in %#v", id, entries)
	return ChangeStatusEntry{}
}
