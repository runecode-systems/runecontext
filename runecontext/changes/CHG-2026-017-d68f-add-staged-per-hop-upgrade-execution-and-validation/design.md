# Design

## Overview
Implement apply-time migration as a staged multi-hop executor that works only against a copied project tree until every hop succeeds. Each hop should be represented by explicit migration logic with an apply step and a verify step. The executor should copy the project tree to a temporary stage root, apply hop 1, validate the staged tree, run hop-specific verification, then continue through the remaining hops in order. Only after all hops and the final staged validation succeed should the real project files be atomically replaced. If any hop fails, no real project files should be changed. Adapter refresh should remain transactional and should run against the staged final tree or final planned state rather than mutating the real tree incrementally.

## Migration-Hop Model
- Each hop should be represented by a typed migration unit with explicit apply and verify phases.
- Hops that need no custom rewrite logic should still use the same interface through a default version-rewrite migration.
- Hops that need file rewrites, managed-asset refreshes, or layout changes should encapsulate those behaviors explicitly instead of hiding them behind a generic final rewrite.

## Execution Rules
- Apply must operate only on a staged project copy until the entire hop chain succeeds.
- After each hop, the staged tree must pass generic project validation and hop-specific verification before the next hop begins.
- Final replacement of real files should happen only after the full staged chain and final validation succeed.
- Any hop failure, verification failure, or final validation failure should abort the apply and leave the real project tree unchanged.

## Transaction Direction
- The existing upgrade transaction and snapshot model should remain the outer safety boundary for replacing real files.
- Adapter refresh should be integrated with staged execution so managed artifacts are updated from the final planned state rather than incrementally mutating the live tree between hops.
