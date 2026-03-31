# Design

## Overview
Implement apply-time upgrade execution as a staged executor that works only against a copied project tree until the final target state succeeds. Some upgrades may require zero migration hops and only a final pinned-version rewrite. Others may require one or more explicit migration hops represented by apply and verify phases. The executor should copy the project tree to a temporary stage root, apply each migration-required hop in order when present, validate and verify after each hop, then rewrite the final pinned project version and refresh managed artifacts against the staged final tree. Only after all required work and the final staged validation succeed should the real project files be atomically replaced. If any stage fails, no real project files should be changed.

## Migration-Hop Model
- Each hop should be represented by a typed migration unit with explicit apply and verify phases.
- Explicit hops should exist only when real migration logic is required.
- Hops that need file rewrites, managed-asset refreshes, or layout changes should encapsulate those behaviors explicitly instead of hiding them behind a generic final rewrite.

## Execution Rules
- Apply must operate only on a staged project copy until the entire hop chain succeeds.
- Version-bump-only upgrades must still use the staged project copy and final validation flow even when no hop runs.
- After each migration-required hop, the staged tree must pass generic project validation and hop-specific verification before the next hop begins.
- Final replacement of real files should happen only after the full staged chain, final pinned-version rewrite, and final validation succeed.
- Any hop failure, verification failure, or final validation failure should abort the apply and leave the real project tree unchanged.

## Transaction Direction
- The existing upgrade transaction and snapshot model should remain the outer safety boundary for replacing real files.
- Adapter refresh should be integrated with staged execution so managed artifacts are updated from the final planned state rather than incrementally mutating the live tree between hops.
- Project-newer-than-cli failures should be rejected before staged execution begins.
