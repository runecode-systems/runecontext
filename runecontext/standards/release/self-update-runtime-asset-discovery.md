---
schema_version: 1
id: release/self-update-runtime-asset-discovery
title: Self-Update Runtime Asset Discovery
status: active
tags:
  - release
  - install
  - self-update
  - runtime-assets
  - cli
---

# Self-Update Runtime Asset Discovery

## Intent

Keep CLI self-update behavior anchored to shipped runtime assets so installed `runectx` releases do not depend on repository-local files, ambient working-directory state, or untrusted discovery paths.

## Requirements

- CLI self-update runtime discovery must anchor to the installed `runectx` executable location, not the current working directory.
- Default self-update flows must not search repository-local paths or arbitrary ancestor directories for runtime assets.
- Self-update metadata such as the default `latest` release hint must come from shipped runtime assets, such as the installed release manifest or an explicit release lookup, not from repo-only authoring files.
- Installer scripts invoked by CLI self-update must come from the shipped runtime layout for the resolved release or another explicitly trusted distribution source.
- Runtime discovery must fail closed when required shipped assets are missing or malformed.

## Rationale

`runectx upgrade cli` and `runectx upgrade cli apply` are operational commands for installed releases, not repository-maintainer-only helpers. If they discover runtime assets from the working directory or repo-only files such as `nix/release/metadata.nix`, they become sensitive to unrelated checkout state and can drift from the actual installed release contents. Executable-anchored shipped assets keep self-update behavior predictable, reviewable, and safer.

## Implementation Notes

- Prefer installed runtime layouts such as `share/runecontext/` as the default source of manifests, schemas, adapters, and installer scripts.
- If an explicit developer override is ever needed, keep it opt-in and clearly separate from default end-user behavior.
- Treat missing installer scripts or malformed shipped manifests as hard failures rather than silently falling back to repo-local heuristics.
