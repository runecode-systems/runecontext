## Summary
Add machine-readable RuneContext metadata and capability descriptor surfaces

## Problem
Downstream tooling needs a stable machine-readable way to discover which RuneContext release it is talking to, which project versions that release supports, which runtime and distribution layouts it ships, and which operational, assurance, and source-verification surfaces are available. Today those facts are split across CLI registry data, contract constants, release metadata, tests, and prose docs, which makes exact feature probing and documentation sync drift-prone.

## Proposed Change
Track the metadata capability initiative as an umbrella over two linked deliverables:

- `CHG-2026-008-47e1-add-canonical-metadata-builder-and-capability-descriptor-outputs` for the descriptor schema, canonical builder, CLI surface, and release-manifest embedding.
- `CHG-2026-009-eb40-derive-docs-and-reference-surfaces-from-canonical-runecontext-metadata` for generated reference data, Docus-friendly outputs, and stale-reference cleanup.

## Why Now
Recent branch work expanded alpha-line compatibility handling, added installed runtime assets under `share/runecontext` in binary archives, and removed the old in-repo `runecontext/operations/` reference. The metadata plan needs to reflect that current reality before more integrations or docs grow around stale assumptions.

## Assumptions
- Version-range compatibility remains the primary gate; the capability descriptor stays supplemental for exact feature probing and profile/layout detection.
- `runectx metadata` and the release manifest should be derived from the same canonical builder rather than parallel implementations.
- Docus-facing reference pages can consume generated JSON or YAML inputs while narrative documentation remains hand-authored.

## Out of Scope
- Tool-specific approval, permission, or policy semantics.
- Replacing the authoritative schemas, contracts, fixtures, or tests with the descriptor.
- Defining downstream-product-specific behavior in the descriptor itself.

## Impact
The umbrella keeps contract design, implementation sequencing, and anti-drift documentation work explicitly linked while preserving one semantic authority for RuneContext behavior.
