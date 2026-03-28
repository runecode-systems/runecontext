## Summary
Add canonical metadata builder and capability descriptor outputs

## Problem
RuneContext needs one generic machine-readable descriptor that downstream tools can trust for exact feature probing. Right now the relevant facts are spread across the CLI command registry, contract constants, compatibility logic, schema assets, and release metadata, which makes drift between CLI output, release artifacts, and integration assumptions too easy.

## Proposed Change
Add the core descriptor implementation: a closed versioned schema, a typed canonical metadata builder, a dedicated `runectx metadata` command, and release-manifest embedding of the same payload.

## Why Now
The current branch already changed supported-version logic and binary runtime layout. The canonical metadata output needs to describe those facts before more downstream tooling depends on stale or partial assumptions.

## Assumptions
- Supported project versions and explicit upgrade edges are related but distinct compatibility facts and should be reported separately.
- Runtime discovery should cover repo-bundle and installed share-layout installs because binary archives now ship `schemas/` and `adapters/` under `share/runecontext`.
- `runectx version --json` should remain backward-compatible even if the new descriptor lives in a dedicated command.

## Out of Scope
- Docus page composition and broader docs/reference generation strategy.
- Downstream-specific policy or approval semantics.

## Impact
This change provides the canonical machine-readable metadata contract that later docs/reference surfaces can consume without re-modeling RuneContext behavior.
