## Applicable Standards
- `standards/architecture/closed-schema-enforcement.md`: The descriptor and generated reference inputs must fail closed on unknown schema versions and fields.
- `standards/architecture/derived-artifact-non-authority.md`: `runectx metadata`, release-manifest embedding, and generated docs inputs remain derived views over the authoritative contracts and registry.
- `standards/cli/completion-and-metadata-from-canonical-registry.md`: Command and capability metadata should derive from one canonical registry instead of parallel hand-maintained models.
- `standards/global/structured-cli-contracts.md`: The new metadata surface must stay stable and machine-friendly for direct CLI use, scripts, and adapters.
- `standards/release/repo-first-canonical-distribution.md`: Release discovery should cover repo-bundle and release-manifest lanes without treating convenience binaries as the only integration surface.
- `standards/release/tag-driven-signed-release-process.md`: Release metadata should stay aligned with tag-driven, signed release provenance.

## Resolution Notes
This umbrella focuses the work on derived machine-readable metadata, stable CLI contracts, and release/runtime discovery without creating a second semantic authority.
