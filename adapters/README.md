# Adapters

Canonical adapter definitions live under `adapters/source/`.

- `adapters/source/shared/` defines shared flow metadata.
- `adapters/source/tools/` defines per-tool capabilities and generation inputs.
- `adapters/source/packs/` contains passthrough adapter content consumed by generation.

Rendered adapter packs are build-generated at `build/generated/adapters/` via `just sync-adapters`.
Those generated outputs are ephemeral and must not be committed.
