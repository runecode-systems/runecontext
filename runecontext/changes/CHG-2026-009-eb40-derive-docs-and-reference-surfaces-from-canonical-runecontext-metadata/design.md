# Design

## Overview
Use the canonical metadata builder as the only structured source for documentation-facing command and capability reference data. Generate Docus-friendly JSON or YAML reference inputs from that builder, keep narrative documentation hand-authored, and add parity checks so command surfaces, capability tokens, and compatibility facts are never manually copied into multiple locations.

## Shape Rationale
- Large, ambiguous, or high-risk feature work should move to full mode early.

## Documentation Rules
- Generated reference data comes from the canonical metadata builder, not hand-maintained markdown copies.
- Narrative docs remain hand-authored and may explain workflow, rationale, and examples.
- Stale references to `runecontext/operations/` should be replaced with the current dogfooded standards and generated reference surfaces.
