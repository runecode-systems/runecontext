package cli

import "testing"

func TestPlanMigrationEdgesWithinIntervalSelectsSparseRealEdges(t *testing.T) {
	t.Parallel()
	registry := defaultUpgradePlannerRegistry()
	registry.registerEdge("0.1.0-alpha.12", "0.1.0-alpha.13")

	hops, planned, err := registry.planMigrationEdgesWithinInterval("0.1.0-alpha.10", "0.1.0-alpha.13")
	if err != nil {
		t.Fatalf("planMigrationEdgesWithinInterval returned error: %v", err)
	}
	if !planned {
		t.Fatalf("expected interval planning to be active")
	}
	if len(hops) != 1 {
		t.Fatalf("expected one migration edge, got %d", len(hops))
	}
	if got, want := hops[0].From, "0.1.0-alpha.12"; got != want {
		t.Fatalf("expected hop from %q, got %q", want, got)
	}
	if got, want := hops[0].To, "0.1.0-alpha.13"; got != want {
		t.Fatalf("expected hop to %q, got %q", want, got)
	}
}

func TestPlanMigrationEdgesWithinIntervalReturnsZeroHopsWhenNonePresent(t *testing.T) {
	t.Parallel()
	registry := defaultUpgradePlannerRegistry()
	hops, planned, err := registry.planMigrationEdgesWithinInterval("0.1.0-alpha.10", "0.1.0-alpha.13")
	if err != nil {
		t.Fatalf("planMigrationEdgesWithinInterval returned error: %v", err)
	}
	if !planned {
		t.Fatalf("expected interval planning to be active")
	}
	if len(hops) != 0 {
		t.Fatalf("expected zero migration edges, got %d", len(hops))
	}
}

func TestPlanMigrationEdgesWithinIntervalRejectsOverlappingEdges(t *testing.T) {
	t.Parallel()
	registry := defaultUpgradePlannerRegistry()
	registry.registerEdge("0.1.0-alpha.10", "0.1.0-alpha.12")
	registry.registerEdge("0.1.0-alpha.11", "0.1.0-alpha.13")

	_, planned, err := registry.planMigrationEdgesWithinInterval("0.1.0-alpha.10", "0.1.0-alpha.13")
	if !planned {
		t.Fatalf("expected interval planning to be active")
	}
	if err == nil {
		t.Fatalf("expected overlap error")
	}
}

func TestPlanMigrationEdgesWithinIntervalNotPlannedForNonForwardInterval(t *testing.T) {
	t.Parallel()
	registry := defaultUpgradePlannerRegistry()
	hops, planned, err := registry.planMigrationEdgesWithinInterval("0.1.0-alpha.13", "0.1.0-alpha.13")
	if err != nil {
		t.Fatalf("planMigrationEdgesWithinInterval returned error: %v", err)
	}
	if planned {
		t.Fatalf("expected interval planning to be disabled for non-forward interval")
	}
	if len(hops) != 0 {
		t.Fatalf("expected zero hops, got %d", len(hops))
	}
}
