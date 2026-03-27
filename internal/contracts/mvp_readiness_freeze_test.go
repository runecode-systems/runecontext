package contracts

import (
	"reflect"
	"sort"
	"testing"
)

func TestMVPReadinessFreezeLifecycleStatusSet(t *testing.T) {
	got := make([]LifecycleStatus, 0, len(lifecycleOrder))
	for status := range lifecycleOrder {
		got = append(got, status)
	}
	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})
	want := []LifecycleStatus{
		StatusClosed,
		StatusImplemented,
		StatusPlanned,
		StatusProposed,
		StatusSuperseded,
		StatusVerified,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected frozen lifecycle status set %v, got %v", want, got)
	}
}

func TestMVPReadinessFreezeGeneratedArtifactStandardPaths(t *testing.T) {
	if got, want := generatedManifestRelativePath, "manifest.yaml"; got != want {
		t.Fatalf("expected generated manifest path %q, got %q", want, got)
	}
	if got, want := generatedChangesIndexRelativePath, "indexes/changes-by-status.yaml"; got != want {
		t.Fatalf("expected generated changes index path %q, got %q", want, got)
	}
	if got, want := generatedBundlesIndexRelativePath, "indexes/bundles.yaml"; got != want {
		t.Fatalf("expected generated bundles index path %q, got %q", want, got)
	}
	if got, want := generatedIndexesDirectoryRelative, "indexes"; got != want {
		t.Fatalf("expected generated indexes directory %q, got %q", want, got)
	}
}
