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
- explicit caller-supplied signed-tag trust inputs for narrow alpha-stage
  validation entrypoints
- deterministic bundle resolution and context-pack hashing
- minimum and lean shaped/full change shapes
- Plain and Verified assurance tiers
- minimal CLI surface
- thin adapters as the primary day-to-day UX
- repo-first releases, reviewable updates, and compatibility documentation
- signed and attested Linux/macOS `runectx` binaries as convenience release
  assets alongside the canonical repo bundles

The MVP does not include RuneCode-specific runtime implementation in this
repository, but it does include the artifacts, semantics, fixtures, and
contracts RuneCode needs in order to integrate cleanly.

## Release Train Summary

| Release | Focus |
| --- | --- |
| `v0.1.0-alpha.1` | Core model, naming, file contracts, schemas, canonical data rules, and validation foundation |
| `v0.1.0-alpha.2` | Source resolution, explicit trust/integrity handling, monorepo support, and deterministic bundle semantics |
| `v0.1.0-alpha.3` | Change workflow, standards linkage, traceability, history preservation, and thin change/status commands |
| `v0.1.0-alpha.4` | Deterministic context packs, generated indexes, and promotion assessment |
| `v0.1.0-alpha.5` | Plain/Verified assurance, baselines, receipts, and backfill |
| `v0.1.0-alpha.6` | Broadened CLI, validation, doctoring, and machine-facing command contracts |
| `v0.1.0-alpha.7` | Generic and tool-specific adapters plus adapter-pack UX |
| `v0.1.0-alpha.8` | Release/install/update hardening and end-to-end MVP readiness fixtures |
| `v0.1.0` | Stabilization, compatibility freeze, and MVP acceptance sign-off |

Signed-tag verification is intentionally part of the MVP and is planned across
`alpha.2` (resolution/integrity implementation using explicit trusted-signer
inputs rather than hidden machine-global trust state), `alpha.4` (context-pack
provenance fields), and `alpha.8` (release/reference-project validation).

## Dogfooding Guidance

- `alpha.3` is the planned point where this repository should be able to start
  dogfooding RuneContext for new work: repo-local project context,
  project-specific standards, active changes, and traceability.
- Post-review `alpha.3` semantics for standards explicitly include strict
  frontmatter validation, path-based standards references in authored change and
  spec docs, warning-level handling for deprecated direct selections, and
  advisory-only `suggested_context_bundles` metadata.
- Final Branch Cut 2 follow-up also clarifies that canonical path references are
  the only supported authored reference form in `alpha.3`, deprecated standards
  without successors warn rather than fail, and copied-body enforcement excludes
  fenced example content.
- The final PR-feedback pass further clarifies that `standards.md` bullets may
  contain other backticked code snippets as description text, but must still
  name exactly one canonical standard path, and that alias metadata is not used
  for runtime reference resolution in `alpha.3`.
- The latest re-review also locks in RuneContext-root-relative diagnostics for
  standards validation and explicitly rejects non-canonical or additional
  `standards/...` references within a single `standards.md` bullet.
- The follow-up hardening pass for Branch Cut 3 also locks in three safety
  behaviors for the thin change/status commands: explicit path arguments remain
  explicit roots even when the caller passes `.`, `change shape` rejects
  terminal changes instead of mutating history, and supersession repair fails
  closed if a reciprocal link would require mutating a terminal successor.
- The same hardening pass also requires optional change-status string fields to
  stay omitted when absent rather than being rewritten as placeholder strings
  such as `<nil>` in summaries or rewritten `status.yaml` files.
- That same rewrite safety rule also preserves the default
  `promotion_assessment.status` of `pending` when the promotion assessment block
  is present but omits an explicit status.
- `alpha.4` is the planned point where this repository should be able to use
  RuneContext as the primary execution-tracking layer for day-to-day feature
  progression, because generated indexes, manifests, and promotion assessment
  complete the basic flow from planned work to durable project knowledge.
- `docs/implementation-plan/` should still remain the home for release-train,
  acceptance, and coverage-accounting documents even after repo-local
  RuneContext dogfooding begins; the goal is not a literal 1:1 migration of
  every planning document into a RuneContext artifact.

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
- Keep one core authored workflow across `plain` and `verified`; higher
  assurance adds evidence rather than alternate source-of-truth files.
- Keep repositories self-sufficient for mixed standalone RuneContext and
  RuneCode teams; RuneCode evidence may be richer but must remain additive
  rather than required for correctness.
- Keep shaped change docs lean: `design.md` and `verification.md` are the
  default shaped artifacts, while `tasks.md` and `references.md` are created
  only when they add real value.
- Keep the release workflow as close as practical to RuneCode's tag-driven
  build/publish split so users can audit one familiar release shape across both
  repositories.
- Keep Nix as the canonical source of unsigned release contents; GitHub Actions
  verifies, signs, attests, and publishes those artifacts rather than
  reassembling them ad hoc in workflow YAML.
- Keep history at stable paths.
- Keep policy, approvals, and runtime capability decisions outside RuneContext.
- Keep normal adapter management local and reviewable; `runectx adapter sync
  <tool>` materializes files from the already-installed RuneContext release and
  must not fetch adapter packs implicitly.
- Keep `runectx` network access limited to explicit `init` and `update` flows.
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
- Add parser and project fixtures for machine-validated heading-fragment refs so
  deep links remain human-readable without relying on brittle line numbers.
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
- Add adapter tests ensuring that mutations to authoritative RuneContext files
  automatically trigger `runectx validate` and surface failures immediately.
- Add RuneCode companion parity fixtures wherever this repository defines a
  contract RuneCode will later consume.
- Do not treat a feature as complete in any alpha until the tests and fixtures
  for that feature land in the same milestone.
