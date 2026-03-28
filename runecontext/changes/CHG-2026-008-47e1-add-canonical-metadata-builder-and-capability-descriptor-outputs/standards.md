## Applicable Standards
- `standards/architecture/closed-schema-enforcement.md`: The capability descriptor must use closed schemas and reject unknown fields or schema versions.
- `standards/architecture/derived-artifact-non-authority.md`: The descriptor and release-manifest metadata remain derived outputs over authoritative contracts and release metadata.
- `standards/cli/completion-and-metadata-from-canonical-registry.md`: Command and capability reporting should derive from the canonical CLI registry instead of parallel metadata models.
- `standards/global/structured-cli-contracts.md`: `runectx metadata` should be a stable machine-facing contract for scripts and adapters.
- `standards/release/repo-first-canonical-distribution.md`: Release-manifest embedding should support canonical repo-bundle discovery in addition to convenience binaries.
- `standards/source/deterministic-source-resolution.md`: Source modes and verification postures should be reported using the existing deterministic resolution model.

## Resolution Notes
These standards focus the implementation on a fail-closed descriptor, canonical derivation, and release/runtime parity rather than a second semantic authority.
