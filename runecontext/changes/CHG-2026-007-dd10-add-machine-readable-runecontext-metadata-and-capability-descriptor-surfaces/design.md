# Design

## Overview
Capture the metadata capability work as an umbrella over descriptor design, canonical machine-readable outputs, and generated reference surfaces. The umbrella should keep the contract shape, implementation sequencing, and documentation strategy aligned while the repository is already dogfooding RuneContext for its own planning.

## Canonical Maintenance Model
- `nix/release/metadata.nix` is the single human-edited source for release identity used by metadata-derived artifacts.
- `runectx metadata`, release-manifest `metadata_descriptor`, and docs/reference JSON all derive from the same canonical metadata builder.
- Checked-in generated artifacts remain reviewable, but they are refreshed through one repo-owned sync path rather than by independent hand edits.
- The repo-maintainer workflow for release-identity changes is: update `nix/release/metadata.nix`, run `just sync-metadata`, then run `just ci`.

## Planned Sub-Changes
- `CHG-2026-008-47e1-add-canonical-metadata-builder-and-capability-descriptor-outputs` owns the descriptor schema, canonical builder, `runectx metadata`, release-manifest embedding, and parity tests.
- `CHG-2026-009-eb40-derive-docs-and-reference-surfaces-from-canonical-runecontext-metadata` owns Docus-facing generated reference data, stale-reference cleanup, and anti-drift checks for documentation-facing outputs.
- `CHG-2026-021-6d7e-extend-metadata-descriptor-with-explicit-layout-assurance-and-feature-contracts` owns the descriptor `v2` extension for explicit distribution layouts, portable project profiles, assurance artifact-family metadata, canonicalization/hash profile tokens, and semantic feature tokens.

## Branch-Specific Constraints
- Supported-project-version reporting must reflect the current alpha-line compatibility rules, not only the explicit upgrade-edge registry.
- Runtime/layout reporting must include both repo-bundle and installed share-layout discovery because binary archives now ship `schemas/` and `adapters/` under `share/runecontext`.
- Documentation work must replace stale references that still assume `runecontext/operations/` is the canonical in-repo operations reference.
- Documentation/reference generation should emit a canonical JSON artifact for Docus consumption; docs-only prose, examples, grouping, and other richer site context remain outside the canonical metadata contract.

## Relationship Model
- Use reciprocal `related_changes` links between the umbrella and both sub-changes for navigability.
- Reserve `depends_on` for real sequencing only; documentation/reference derivation should depend on canonical metadata outputs rather than on umbrella semantics alone.
