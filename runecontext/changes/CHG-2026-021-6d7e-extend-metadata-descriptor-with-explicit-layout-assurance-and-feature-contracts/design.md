# Design

## Overview
Extend the canonical metadata builder to emit a deliberate descriptor `v2` shape instead of stretching the original `v1` contract. The refined `v2` descriptor should keep the existing release, command, assurance, canonicalization, and resolution surfaces, but replace ambiguous compatibility and versioning fields with clearer machine-readable contracts, keep explicit distribution-layout reporting, add explicit project-profile metadata, and retain a narrow semantic feature-token surface for high-level capability probing.

## Versioning Approach
- Use descriptor `v2` with one authoritative `schema_version` field rather than keeping redundant version fields with different types.
- Treat the refined `v2` descriptor as the only valid contract for the updated compatibility and feature semantics.
- Reject legacy/ambiguous mixed payloads rather than attempting to support dual interpretation in one schema.

## Shape Rules
- Replace `compatibility.supported_project_versions` with `compatibility.directly_supported_project_versions` and `compatibility.upgradeable_from_project_versions` so direct compatibility and upgrade-only entrypoints are machine-distinct.
- Replace ambiguous `runtime.layouts` with top-level `distribution_layouts` that describe how a RuneContext release is packaged or installed.
- Add top-level `project_profiles` that describe supported on-disk project shapes using portable project-root-relative paths.
- Keep `capabilities.commands` as the exact command surface and add a separate top-level `features` token list for coarse semantic capability detection.
- Extend `assurance` to report tier support plus baseline support and supported receipt families.
- Add a top-level `canonicalization` object with explicit context-pack and assurance-artifact token families.

## Authority Boundaries
- Project profile paths and meaning must come from the authoritative project layout contract in `core/layout.md` and related generated-artifact constants, not from ad hoc CLI guesses.
- Assurance receipt families must reflect the existing portable receipt-family contract in `internal/contracts/assurance_runtime_shared.go` and `schemas/assurance-receipt.schema.json`.
- Canonicalization/hash tokens must come from the existing machine-contract constants and schemas for context packs and assurance artifacts.
- Semantic feature tokens should describe implemented capability families already present in core behavior; they must not become a shortcut for inventing new policy, shadow command presence, or substitute for version/schema validation.

## Distribution Layout Surface
- `distribution_layouts` should cover the repo-bundle layout and the installed share layout currently emitted by the metadata builder.
- The shape should stay small: profile ID plus `schema_path` and `adapters_path` are sufficient for the known install/distribution cases.
- Distribution-layout metadata answers packaging/discovery questions only; it must not be stretched to describe managed project content.

## Project Profile Surface
- Start with one explicit portable project profile reflecting the current RuneContext project shape.
- The profile should expose stable IDs and the minimum durable paths needed by downstream tooling: `root_config`, `content_root`, `assurance_path`, `manifest_path`, and `indexes_root`.
- These paths should remain portable and relative to the project root so linked/path/embedded resolution modes do not change the published profile contract.

## Assurance And Canonicalization Surface
- `assurance.tiers` remains the assurance mode list.
- `assurance.baseline_supported` should state whether the implementation supports the baseline artifact family at all.
- `assurance.receipt_families` should enumerate the portable verified receipt families RuneContext supports today.
- `canonicalization.context_pack` and `canonicalization.assurance_artifacts` should separately expose profile tokens and hash algorithms even when they currently share the same values.

## Semantic Feature Tokens
- Keep feature tokens generic, flat, and intentionally narrower than the command registry.
- Include tokens only when they describe implemented semantic capability families that a downstream consumer would otherwise have to infer from multiple lower-level fields or prose documentation.
- Candidate tokens for the refined `v2` surface include signed-tag verification, mutable-ref opt-in, monorepo nearest-root discovery, context-pack capture, verified assurance, completion registry, dynamic suggestions, generated indexes, promotion workflow, upgrade planning, staged upgrade execution, and mixed-or-stale-tree detection.
- Do not duplicate every CLI command as a feature token.

## Documentation And Derived Artifacts
- Update the generated docs/reference artifact to mirror the refined `v2` descriptor contract.
- Keep narrative docs hand-authored and let generated reference artifacts stay a derived machine-readable surface.
- Document the new field meanings in the metadata/docs surface without moving semantic authority away from the core layout, schema, and contracts documents.
