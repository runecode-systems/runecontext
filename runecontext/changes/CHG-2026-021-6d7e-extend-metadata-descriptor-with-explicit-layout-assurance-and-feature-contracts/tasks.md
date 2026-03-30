# Tasks

- Define and document the descriptor `v2` target shape, including explicit top-level `distribution_layouts`, `project_profiles`, `features`, and `canonicalization` surfaces plus expanded `assurance` fields.
- Update the canonical metadata builder to emit the new `v2` descriptor while preserving one semantic authority for release identity, compatibility, layout, assurance, and canonicalization facts.
- Replace the old ambiguous `runtime.layouts` field with explicit distribution-layout reporting and add portable project-profile path metadata derived from the project layout contract.
- Extend assurance metadata to report baseline support and supported verified receipt families using existing portable artifact-family contracts.
- Publish canonicalization and hash-profile tokens for context packs and assurance artifacts from existing machine-contract constants and schemas.
- Add a narrow semantic feature-token surface that captures implemented high-level capabilities without mirroring every command token.
- Update the descriptor schema to a closed `v2` contract and add fail-closed tests for unknown descriptor versions, unknown fields, unknown feature tokens, unknown receipt families, and unknown canonicalization/hash tokens.
- Refresh metadata-derived fixtures and generated docs/reference artifacts, and add CLI/schema/release/docs parity tests so all derived surfaces stay aligned.
- Document the meaning of the new metadata fields in the repo's metadata/docs surface without relocating core semantic authority away from `core/`, `schemas/`, and `internal/contracts/`.
