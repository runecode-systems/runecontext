# Design

## Overview
Use the canonical metadata builder as the only structured source for documentation-facing command and capability reference data. Generate a Docus-friendly JSON reference input from that builder, keep narrative documentation hand-authored, and add parity checks so command surfaces, capability tokens, and compatibility facts are never manually copied into multiple locations.

## Shape Rationale
- Large, ambiguous, or high-risk feature work should move to full mode early.

## Documentation Rules
- Generated reference data comes from the canonical metadata builder, not hand-maintained markdown copies.
- Narrative docs remain hand-authored and may explain workflow, rationale, and examples.
- The generated artifact should be stable machine-readable JSON intended for Docus ingestion; richer docs-only content such as examples, grouping, callouts, ordering, or migration notes should live in the Docus consumer layer rather than in canonical metadata fields.
- Metadata-derived docs/reference artifacts should refresh through one repo-owned sync command so release-version bumps do not require multiple manual edits.
- Stale references to `runecontext/operations/` should be replaced with the current dogfooded standards and generated reference surfaces.
