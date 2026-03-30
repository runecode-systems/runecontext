## Summary
Add staged per-hop upgrade execution and validation

## Problem
Even with a valid migration path, the current apply flow only rewrites `runecontext_version` once and validates once. It has no way to represent per-hop migration logic, no way to verify intermediate staged states, and no way to safely chain multiple migration steps before replacing real project files.

## Proposed Change
Execute approved upgrade paths as staged multi-hop migrations: apply each hop against a copied project tree, validate and verify after every hop, then replace real files only after the full migration chain succeeds.

## Why Now
Once multi-hop planning exists, apply-time behavior must honor the same explicit migration semantics. Without staged per-hop execution, approved paths would still collapse into a single version rewrite and would not be safe for future hops that need real file migrations.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Some hops will need only a plain version rewrite, while others may need dedicated file-migration logic and hop-specific verification.

## Out of Scope
- Replacing preview/path planning contracts.
- Mutating real project files incrementally during intermediate hops.
- Skipping staged validation or hop-specific verification in the name of convenience.

## Impact
The change makes upgrade apply trustworthy for future multi-version migrations by giving each declared hop a typed execution path, staged verification, and fail-closed rollback before the real project tree is touched.
