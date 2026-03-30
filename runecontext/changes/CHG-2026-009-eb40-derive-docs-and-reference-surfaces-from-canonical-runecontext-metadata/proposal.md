## Summary
Derive docs and reference surfaces from canonical RuneContext metadata

## Problem
Adding a new RuneContext command, capability, or compatibility rule should not require remembering to hand-edit multiple docs and website locations. The repository also still contains stale references to the old `runecontext/operations/` model even though this branch now relies on dogfooded standards plus generated reference material instead.

## Proposed Change
Generate documentation-facing command, capability, compatibility, and layout reference surfaces from the canonical metadata builder, establish a Docus-friendly JSON input artifact plus a single repo sync workflow, and clean up stale assumptions about the in-repo operations reference.

## Why Now
Without a first-class docs/reference derivation path, the new metadata descriptor would still leave website and repo docs vulnerable to drift. This branch is already a good moment to reset those assumptions while the project is actively dogfooding RuneContext.

## Assumptions
- Docus pages can render generated JSON reference data while keeping narrative prose and richer docs-only context hand-authored.
- Generated docs/reference data should be treated as derived artifacts and regenerated from the canonical metadata builder rather than manually edited.
- Any docs-only examples, grouping, callouts, or explanatory overlays belong in the Docus app/content layer rather than in the canonical metadata contract itself.
- The docs/reference work should follow, not replace, the canonical metadata implementation.

## Out of Scope
- Re-defining command semantics outside the CLI registry and contracts layer.
- Inventing website-only capability names or compatibility rules.

## Impact
This change makes the metadata work operationally sustainable by reducing documentation drift and giving downstream docs tooling one canonical machine-readable input.
