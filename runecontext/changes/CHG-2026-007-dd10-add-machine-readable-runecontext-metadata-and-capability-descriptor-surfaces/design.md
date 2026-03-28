# Design

## Overview
Capture the metadata capability work as an umbrella over descriptor design, canonical machine-readable outputs, and generated reference surfaces. The umbrella should keep the contract shape, implementation sequencing, and documentation strategy aligned while the repository is already dogfooding RuneContext for its own planning.

## Planned Sub-Changes
- `CHG-2026-008-47e1-add-canonical-metadata-builder-and-capability-descriptor-outputs` owns the descriptor schema, canonical builder, `runectx metadata`, release-manifest embedding, and parity tests.
- `CHG-2026-009-eb40-derive-docs-and-reference-surfaces-from-canonical-runecontext-metadata` owns Docus-facing generated reference data, stale-reference cleanup, and anti-drift checks for documentation-facing outputs.

## Branch-Specific Constraints
- Supported-project-version reporting must reflect the current alpha-line compatibility rules, not only the explicit upgrade-edge registry.
- Runtime/layout reporting must include both repo-bundle and installed share-layout discovery because binary archives now ship `schemas/` and `adapters/` under `share/runecontext`.
- Documentation work must replace stale references that still assume `runecontext/operations/` is the canonical in-repo operations reference.

## Relationship Model
- Use reciprocal `related_changes` links between the umbrella and both sub-changes for navigability.
- Reserve `depends_on` for real sequencing only; documentation/reference derivation should depend on canonical metadata outputs rather than on umbrella semantics alone.
