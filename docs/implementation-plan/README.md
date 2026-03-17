# RuneContext Implementation Plan

This folder turns `docs/project_idea.md` into a concrete delivery plan for the
`runecontext` repository.

`docs/project_idea.md` remains the reference document and source of truth. The
documents in this folder are implementation planning artifacts only and must not
replace or rewrite the original idea document.

## Planning Boundaries

- This plan covers RuneContext Core, adapters, CLI, release/install/update, and
  the contracts needed for future RuneCode integration.
- This plan does not schedule RuneCode runtime implementation work inside this
  repository.
- RuneCode companion-track checkpoints are called out in each alpha so the
  `runecode` repository can test parity, compatibility, and audit-readiness as
  RuneContext matures.

## Implementation Expectation

- The RuneContext core library and `runectx` CLI are expected to be implemented
  in Go.
- The planning documents remain language-neutral where possible, but release,
  test, and integration planning assume a Go implementation in this repository.

## MVP Definition

`v0.1.0` is the RuneContext MVP for this repository.

The MVP includes every v1 RuneContext feature described in
`docs/project_idea.md` for this repository, including:

- portable markdown/yaml/json-first source artifacts
- embedded and linked source modes
- signed-tag verification support for linked sources
- deterministic bundle resolution and context-pack hashing
- minimum and full change shapes
- Plain and Verified assurance tiers
- minimal CLI surface
- thin adapters as the primary day-to-day UX
- repo-first releases, reviewable updates, and compatibility documentation

The MVP does not include RuneCode-specific runtime implementation in this
repository, but it does include the artifacts, semantics, fixtures, and
contracts RuneCode needs in order to integrate cleanly.

## Release Train Summary

| Release | Focus |
| --- | --- |
| `v0.1.0-alpha.1` | Core model, naming, file contracts, schemas, canonical data rules, and validation foundation |
| `v0.1.0-alpha.2` | Source resolution, storage modes, monorepo support, and bundle semantics |
| `v0.1.0-alpha.3` | Change workflow, standards linkage, traceability, and history preservation |
| `v0.1.0-alpha.4` | Deterministic context packs, generated indexes, and promotion assessment |
| `v0.1.0-alpha.5` | Plain/Verified assurance, baselines, receipts, and backfill |
| `v0.1.0-alpha.6` | Minimal CLI, validation, doctoring, and machine-facing command contracts |
| `v0.1.0-alpha.7` | Generic and tool-specific adapters plus adapter-pack UX |
| `v0.1.0-alpha.8` | Release/install/update hardening and end-to-end MVP readiness fixtures |
| `v0.1.0` | Stabilization, compatibility freeze, and MVP acceptance sign-off |

Signed-tag verification is intentionally part of the MVP and is planned across
`alpha.2` (resolution/integrity implementation), `alpha.4` (context-pack
provenance fields), and `alpha.8` (release/reference-project validation).

## Document Index

- `docs/implementation-plan/milestone-breakdown.md`
  - full alpha-by-alpha milestone, epic, and issue breakdown
- `docs/implementation-plan/mvp-acceptance.md`
  - final MVP acceptance checklist for `v0.1.0`
- `docs/implementation-plan/coverage-matrix.md`
  - section-by-section mapping from `docs/project_idea.md` into the plan
- `docs/implementation-plan/post-mvp.md`
  - grouped post-MVP work after `v0.1.0`

## Planning Principles

- Keep the on-disk model, schemas, and resolution semantics canonical.
- Treat adapters as UX layers, not alternate sources of truth.
- Keep generated artifacts derived and reviewable.
- Keep history at stable paths.
- Keep policy, approvals, and runtime capability decisions outside RuneContext.
- Treat RuneContext content as untrusted LLM input as well as untrusted policy
  input; rely on typed boundaries, review, and isolation rather than trusting
  the text itself.
- Keep every alpha shippable and useful on its own.
- Require excellent automated test coverage for every new semantic surface.
- Ensure each alpha adds at least one concrete RuneCode companion-track test or
  parity checkpoint.

## Testing Strategy

Excellent test coverage is a release requirement, not cleanup work deferred
until the end.

- Add unit tests for every new core rule: schema validation, markdown contract
  parsing, source resolution, bundle precedence, lifecycle invariants,
  promotion-state transitions, and assurance behavior.
- Add golden fixtures for deterministic outputs: resolved bundles, context
  packs, pack hashes, manifests, baselines, receipts, and machine-readable CLI
  output.
- Add parser and project fixtures for markdown contracts and traceability rules,
  including `proposal.md`, `standards.md`, `specs/*.md`, and `decisions/*.md`.
- Make whole-project validation exercise the same alpha-stage contracts the docs
  claim are enforced; do not leave parser-only checks unwired.
- Validate against the project's declared content root instead of assuming a
  fixed embedded directory name when alpha-stage source settings allow variation.
- Keep release metadata, module metadata, and executable validation behavior in
  sync with the documented alpha train so foundational tooling does not drift.
- Add CLI integration tests for write flows, non-interactive behavior, dry-run
  behavior, explain output, and failure classes.
- Before full `--json` lands, narrow early CLI commands may use stable
  line-oriented machine output if that contract is explicitly documented and
  tested.
- Add adapter smoke tests and reference-project tests so UX layers stay aligned
  with the same core semantics.
- Add RuneCode companion parity fixtures wherever this repository defines a
  contract RuneCode will later consume.
- Do not treat a feature as complete in any alpha until the tests and fixtures
  for that feature land in the same milestone.
