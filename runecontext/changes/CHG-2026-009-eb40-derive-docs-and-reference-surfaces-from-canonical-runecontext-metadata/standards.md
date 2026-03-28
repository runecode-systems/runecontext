## Applicable Standards
- `standards/architecture/derived-artifact-non-authority.md`: Generated reference files should stay derived from canonical metadata rather than becoming the authority themselves.
- `standards/cli/completion-and-metadata-from-canonical-registry.md`: Command and capability docs should derive from the same canonical metadata/registry inputs as completion and adapter surfaces.
- `standards/global/structured-cli-contracts.md`: Reference docs should describe the stable machine-facing contract rather than reinterpret it.
- `standards/release/repo-first-canonical-distribution.md`: Release-facing docs should keep repo-bundle and manifest-based discovery visible.

## Resolution Notes
These standards keep the docs/reference work focused on derivation, anti-drift checks, and replacing stale reference models instead of inventing new semantics.
