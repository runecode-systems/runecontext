package cli

import "testing"

func TestAdapterReferenceRootForPathSupportsAdaptersFallback(t *testing.T) {
	if got, want := adapterReferenceRootForPath("/tmp/work/adapters"), "adapters"; got != want {
		t.Fatalf("expected adapters fallback root %q, got %q", want, got)
	}
	if got, want := adapterReferenceRootForPath("adapters"), "adapters"; got != want {
		t.Fatalf("expected adapters fallback root %q, got %q", want, got)
	}
}
