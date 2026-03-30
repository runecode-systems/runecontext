## Applicable Standards
- `standards/architecture/closed-schema-enforcement.md`: The metadata descriptor must remain schema-validated, closed by default, and fail closed on unknown schema versions, fields, and token values.
- `standards/architecture/derived-artifact-non-authority.md`: `runectx metadata`, release-manifest embedding, and generated docs/reference artifacts remain derived views over the authoritative contracts and docs.
- `standards/cli/completion-and-metadata-from-canonical-registry.md`: Command and capability surfaces must continue to derive from the canonical CLI registry rather than a parallel hand-maintained metadata model.
- `standards/context-packs/explicit-generated-at-and-stable-hashing.md`: Context-pack canonicalization and hashing tokens published by metadata must match the existing stable machine contract.
- `standards/assurance/fail-closed-verified-boundary.md`: Verified-mode artifact reporting must preserve explicit fail-closed boundaries instead of implying soft optional semantics.
- `standards/assurance/portable-receipt-linkage.md`: Receipt-family metadata must reflect the portable verified artifact families RuneContext actually emits and validates.
- `standards/global/compatibility-is-not-upgradeability.md`: Metadata must keep broad project-version compatibility distinct from explicit upgrade edges.
- `standards/global/structured-cli-contracts.md`: The metadata command must stay deterministic, machine-friendly, and stable for scripts and adapters.
- `standards/release/repo-first-canonical-distribution.md`: Distribution-layout reporting must describe canonical repo-bundle and installed-share release lanes without elevating convenience packaging into the sole model.

## Resolution Notes
This change focuses on making the existing metadata surface more explicit and downstream-useful while keeping all core RuneContext semantics anchored in the authoritative schemas, contracts, fixtures, and operations/docs surfaces.
