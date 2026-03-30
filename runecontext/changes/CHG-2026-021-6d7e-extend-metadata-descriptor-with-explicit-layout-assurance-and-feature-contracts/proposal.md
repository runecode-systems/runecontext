## Summary
Extend the metadata descriptor with explicit layout, assurance, canonicalization, and semantic feature contracts

## Problem
`runectx metadata` already exposes a solid machine-readable descriptor, but downstream consumers still have to infer too much from ambiguous field names, raw version strings, command presence, or prose documentation. In particular, the current `runtime.layouts` shape mixes install/distribution discovery with project on-disk shape, assurance metadata stops at tier support instead of naming the verified artifact families consumers can rely on, and the descriptor does not publish the canonicalization/hash profile tokens or higher-level semantic capability families that downstream tools need for stable startup gating.

## Proposed Change
Deliver a deliberate descriptor `v2` that keeps RuneContext's existing authority boundaries intact while making the implemented metadata surface more explicit. The new version should distinguish distribution layouts from project profiles, add verified-assurance artifact-family reporting, publish canonicalization/hash profile tokens for context packs and assurance artifacts, and expose a small generic semantic feature-token surface alongside the existing exact command list.

## Why Now
RuneCode is already validating the descriptor as an integration contract and the current metadata implementation is still tracked under an active metadata umbrella project. This is the right moment to fix the remaining ambiguity cleanly before more downstream consumers normalize the current `v1` shape or start scraping documentation for facts the CLI can publish directly.

## Assumptions
- The capability descriptor remains a derived capability/profile description, not a second semantic authority over core contracts, schemas, or operations docs.
- The current `runtime.layouts` name is ambiguous enough that a clean `v2` rename is preferable to silently stretching `v1` semantics.
- The descriptor schema should remain closed and fail closed on unknown schema versions, unknown fields, and unknown token values.
- Generic downstream consumers need profile IDs and stable tokens more than prose descriptions or tool-specific behavior.
- Project profile paths should be published as portable project-root-relative paths matching `core/layout.md` rather than implementation-specific absolute or content-root-only internals.

## Out of Scope
- RuneCode-only workflow assumptions or startup policy semantics.
- Encoding permissions, approvals, or runtime authority decisions in metadata.
- Introducing environment-derived or host-state-derived metadata.
- Replacing the authoritative contracts in `core/`, `schemas/`, `fixtures/`, or operations/docs with the metadata descriptor.

## Impact
This change turns the existing descriptor into a stronger generic integration contract. Downstream tools will be able to answer release identity, supported project versions, project-profile shape, distribution layout, assurance artifact families, canonicalization/hash tokens, and high-level semantic feature support from `runectx metadata` alone, while the repo keeps one semantic authority and explicit fail-closed versioning.
