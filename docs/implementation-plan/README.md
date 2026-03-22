# RuneContext Implementation Plan

This folder turns `docs/project_idea.md` into a concrete delivery plan for the
`runecontext` repository.

`docs/project_idea.md` remains the original design baseline and product
rationale. It is historical by default and should not be edited during normal
feature work.

When implementation details evolve (for example, hardening changes across alpha
cuts), treat the following as the current normative contract surfaces:

- `core/` and `schemas/` for machine-readable semantics
- `docs/implementation-plan/` for milestone-scoped implementation decisions
- tests and fixtures in `internal/` and `fixtures/` for executable behavior

If a narrowly scoped historical correction to `docs/project_idea.md` is ever
needed, record the rationale in this file so reviewers can distinguish
historical cleanup from new feature design work.

## Planning Boundaries

- This plan covers RuneContext Core, adapters, CLI, release/install/upgrade, and
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

## Historical Corrections Log

- 2026-03-20: Added explicit guidance that `docs/project_idea.md` is a
  historical baseline and that alpha-stage contract refinements should be
  captured in `core/`, `schemas/`, and implementation-plan docs instead of
  rewriting historical idea text. This addresses review feedback where legacy
  examples (for example, earlier canonicalization labels) may diverge from the
  current alpha.4 contract.
- 2026-03-22: Renamed the planned repo upgrade flow from `runectx update` to
  `runectx upgrade` and locked in the alpha.8 contract: preview-first planning
  via `runectx upgrade`, explicit mutation via `runectx upgrade apply`,
  transactional staging with validate-before-replace and automatic in-flight
  rollback, no hidden migrations in read-only commands, and externally managed
  handling for `type: path` sources.
- 2026-03-22: Recorded that the old `runecontext/commands/` wording in
  `docs/project_idea.md` is stale historical text. The current normative path
  for the canonical in-project reference is `runecontext/operations/`.
- 2026-03-22: Clarified adapter terminology across the plan: `adapter` means the
  tool-specific UX layer, `adapter pack` means the packaged release payload for
  an adapter, and `runectx adapter sync <tool>` is the local materialization
  command for those packaged contents.
- 2026-03-22: Recorded the adapter UX refinement for the remaining MVP work:
  conversational tool-native flows for `change new`, `change shape`,
  `standard discover`, and `promote` belong in adapters as thin UX over
  explicit core operations, with any discovery scope/focus inputs exposed in
  the underlying operation contract rather than hidden in prompt-only state.
- 2026-03-22: Recorded two post-MVP planning boundaries from enhancement
  review: migration from other spec-driven systems is a separate adoption/
  import surface rather than part of `runectx upgrade`, and future portable
  instruction-module work should compile into tool-native skills/prompts/
  instructions without turning capability-bearing tool config into RuneContext
  core semantics.

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
- conversational adapter UX for selected authoring, discovery, and promotion
  flows built as thin layers over explicit core operations
- repo-first releases, reviewable upgrades, and compatibility documentation
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
| `v0.1.0-alpha.4` | Deterministic context packs, stable generated indexes, and reviewable promotion assessment |
| `v0.1.0-alpha.5` | Broadened CLI, `init` scaffolding, promotion/resolve flows, validation, doctoring, and machine-facing command contracts |
| `v0.1.0-alpha.6` | Plain/Verified assurance, baselines, receipts, and backfill |
| `v0.1.0-alpha.7` | Generic and tool-specific adapters, conversational adapter UX, completion UX, and local adapter sync |
| `v0.1.0-alpha.8` | Release/install/upgrade hardening, networked `init`/`upgrade` flows, and end-to-end MVP readiness fixtures |
| `v0.1.0` | Stabilization, compatibility freeze, and MVP acceptance sign-off |

Signed-tag verification is intentionally part of the MVP and is planned across
`alpha.2` (resolution/integrity implementation using explicit trusted-signer
inputs rather than hidden machine-global trust state), `alpha.4` (context-pack
provenance fields), and `alpha.8` (release/reference-project validation).

## Alpha.1 Foundation Recap

The old one-off Epic 2 implementation summary has been folded back into the main
plan documents. The enduring alpha.1 baseline is:

- four core machine-readable schemas: `runecontext.yaml`, `bundles/*.yaml`,
  `changes/*/status.yaml`, and `context-pack.yaml`
- closed-schema defaults with explicit opt-in `extensions` for authored files
  and no extensions in generated artifacts
- JSON Schema Draft 2020-12 as the standard dialect for shipped contracts
- the restricted machine-readable YAML profile plus markdown structure contracts
  for `proposal.md` and `standards.md`
- shipped schema/profile fixtures covering standalone validation,
  project-level extension policy, and YAML-profile rejection cases

Later review hardening for alpha.2-alpha.4 is also tracked directly in the main
plan now rather than in per-epic recap files, so milestone history, acceptance,
and coverage stay in one place.

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
- The final Branch Cut 4 hardening pass extends that same fail-closed posture to
  `runectx change reallocate CHANGE_ID [--path PATH]`: terminal changes cannot
  be rewritten, reallocation only rewrites local change-path references inside
  the change, unchanged markdown keeps its original bytes, rewrite token
  boundaries stay UTF-8-safe, staging happens outside the live `changes/` tree,
  and leftover backup-cleanup problems surface as warnings instead of ambiguous
  post-success failures.
- The latest follow-up hardening pass applies the same fail-closed expectation
  to `change close` and `change new`: failed close operations roll back status
  rewrites instead of leaving partial history mutations behind, failed creates
  clean up their transient change directories, mutation paths reject symlinked
  targets across create/close/reallocate, reallocate also rejects symlinked
  rename roots before directory swaps, transactional rewrites preserve file
  permissions, successful markdown path rewrites keep the original file newline
  style, and atomic file replacement now has a Windows-safe fallback when the
  destination already exists.
- The same hardening pass also requires optional change-status string fields to
  stay omitted when absent rather than being rewritten as placeholder strings
  such as `<nil>` in summaries or rewritten `status.yaml` files.
- That same rewrite safety rule now normalizes close-time promotion assessment
  outcomes to `none` or `suggested` when the promotion block is missing or
  empty, while preserving explicitly advanced promotion states (`accepted`,
  `completed`) for later dedicated promotion workflows.
- The latest alpha.3 traceability hardening pass also requires markdown deep-ref
  tokenization to stay UTF-8-safe during validation, so surrounding Unicode
  punctuation such as smart quotes terminates local ref tokens cleanly instead
  of producing false missing-artifact errors, while machine-readable heading
  fragments remain ASCII-bounded even when adjacent prose is non-ASCII.
- The same alpha.3 hardening pass also aligns terminal lifecycle validation with
  close-time behavior: both `closed` and `superseded` changes must carry a
  completed `verification_status`, and missing spec/decision reciprocity now
  points reviewers back to the referenced change `status.yaml` instead of
  repeating the same change ID twice.
- A subsequent PR-follow-up hardening pass also aligns duplicate markdown
  heading fragments with deterministic markdown-anchor numbering (`foo`,
  `foo-1`, `foo-2`, ...) while still skipping already occupied suffixed forms,
  makes broken bundle symlink targets fail closed consistently once traversal
  begins, rejects another long flag token as the missing value for thin
  `change` command string flags, and propagates YAML encoder close failures
  during `status.yaml` rewrites instead of silently discarding them.
- That same follow-up keeps the current alpha.2 git-source contract intact:
  validation rejects option-like and remote-helper forms, but local repository
  paths remain intentionally allowed for now rather than narrowing `type: git`
  to remote-only URLs mid-train.
- `alpha.4` is the planned point where this repository should be able to use
  RuneContext as the primary execution-tracking layer for day-to-day feature
  progression, because generated indexes, manifests, and promotion assessment
  complete the basic flow from planned work to durable project knowledge.
- The latest alpha.4 planning pass also locks in five implementation details to
  avoid later refactors: required `generated_at` stays outside canonical
  `pack_hash` input, persisted pack provenance retains `pattern` and `kind`,
  context-pack request identity uses a hybrid authored-composite plus ordered
  `requested_bundle_ids` model, generated indexes land at fixed optional paths
  with closed schemas, and close-time promotion assessment writes only `none` or
  `suggested` while explicit later workflows own `accepted` and `completed`.
- A Branch Cut 4 follow-up hardening pass also requires generated-index path
  emission to fail closed if an artifact path escapes the RuneContext content
  root, requires unknown lifecycle statuses to fail loudly during
  `changes-by-status` generation, and locks bundle-index determinism plus
  closed-schema unknown-field rejection under dedicated tests.
- The same Branch Cut 4 hardening also tightens manifest and bundle-index path
  patterns to reject traversal (`..`), hidden (`.`-prefixed), and empty
  segments so external tooling can validate generated artifacts against a
  stricter fail-closed path contract.
- That same Branch Cut 4 follow-up keeps the generated `changes-by-status`
  schema aligned with `change-status.schema.json` so the custom `x-` type
  allowance stays deterministic with the existing lifecycle semantics.
- That same follow-up also orders generated writes so the change and bundle
  indexes land before `manifest.yaml`, preventing partial manifest updates
  from leaving downstream tooling with inconsistent state.
- Branch Cut 3 follow-up hardening further specifies that close-time promotion
  assessment must not regress existing `accepted`/`completed` states, and that
  close behavior for both `closed` and `superseded` lifecycle outcomes remains
  deterministic and reviewable.
- The same Branch Cut 3 hardening also requires close-time suggested
  `target_path` values to be emitted from already-normalized traceability
  records, so platform-specific separators cannot leak into
  `promotion_assessment.suggested_targets` output.
- Branch Cut 1 hardening also clarifies that the core pack builder requires an
  explicit whole-second `generated_at`, path-mode `source_ref` values must stay
  portable, LF/CRLF text checkouts hash identically, and the emitted pack
  canonicalization token is RuneContext-owned rather than an overclaimed full
  RFC 8785 label.
- Branch Cut 2 hardening also clarifies that machine-readable pack reports carry
  their own explicit schema version and schema file, non-transient stability
  check read errors must surface directly instead of being masked as generic
  rebuild noise, and pack-only versus enriched report-building flows should stay
  separable even when they share the same fail-closed rebuild logic.
- That same hardening also documents two narrower Branch Cut 2 boundaries: a
  zero-valued advisory-threshold struct means "use defaults" while explicit
  field zeros remain meaningful once any field is set, and rebuild stability is
  evaluated against the loaded project snapshot rather than hot-reloading bundle
  definitions from disk mid-attempt.
- The advisory-threshold defaults themselves should also be exposed as copy-
  returning values rather than mutable exported global structs so tests and
  callers cannot silently rewrite process-wide defaults.
- Branch Cut 2 review cleanup additionally tightens report warning schema fields
  to non-negative counters and keeps test-only read-hook plumbing nil-safe via a
  fallback to the real file reader.
- The recommended alpha.4 review order is pack engine and determinism fixtures,
  then pack explanation and limits, then promotion assessment, and finally
  generated indexes/manifests.
- `alpha.5` is the planned point where this repository should be able to
  broaden that dogfooding across additional repositories and automation-heavy
  workflows, because the broader `runectx` CLI surface adds `init`,
  `bundle resolve`, `promote`, stable machine-facing flags, and the non-thin
  command contracts needed for cross-repo day-to-day use.
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
- Keep deterministic hashes tied to canonical resolved content rather than
  regeneration-only metadata such as `generated_at` timestamps.
- Keep portable generated artifacts free of host-specific absolute paths;
  persisted path fields should use stable RuneContext-relative or equivalently
  portable identifiers.
- Keep the normal authored context-selection model centered on one top-level
  bundle or authored composite bundles, while still leaving generated pack
  metadata room to record ordered runtime bundle requests for RuneCode and other
  automation consumers.
- Keep one core authored workflow across `plain` and `verified`; higher
  assurance adds evidence rather than alternate source-of-truth files.
- Keep repositories self-sufficient for mixed standalone RuneContext and
  RuneCode teams; RuneCode evidence may be richer but must remain additive
  rather than required for correctness.
- Keep one shared machine-facing CLI envelope and failure taxonomy across
  commands; broaden thin wrappers rather than letting command-specific contracts
  drift apart.
- Keep invalid-output path fields consistent across commands: `root` is the
  project root and `error_path` is the specific failing artifact path when
  present.
- Keep CLI command boundaries crisp: `status` is workflow summary, `validate` is
  authoritative contract enforcement, and `doctor` is environment/install/
  source-posture diagnosis.
- Keep `runectx standard discover` advisory-only and `runectx promote` as the
  only durable promotion-mutation surface; interactive handoff must use
  explicit candidate data rather than hidden session state.
- Keep `runectx init` local-first and local-only; it should scaffold from the
  already-installed RuneContext release contents rather than fetching project
  files over the network.
- Keep completion and autocomplete metadata derived from the same stable CLI
  command/flag/value definitions rather than maintaining a second hand-authored
  command model.
- Keep one typed internal command/operation registry as the canonical source for
  CLI metadata; human-readable operations docs, shell completion scripts,
  machine-readable completion metadata, and adapter-native suggestion surfaces
  should all be derived from that same registry.
- Keep adapter-layer features implemented as thin UX over explicit core
  operations and stable candidate data; adapters may be more conversational,
  but they must not invent hidden semantics, alternate mutation paths, or a
  second source of truth.
- Keep conversational adapter UX focused on authoring/discovery/promotion flows
  such as `change new`, `change shape`, `standard discover`, and `promote`;
  prefer normal host conversation turns plus reviewable outputs over disruptive
  questionnaire-style widgets when the host can support that pattern.
- Keep discovery scoping and user-supplied focus inputs explicit in the
  underlying operation and stable CLI contract whenever adapters expose them,
  so those semantics are not trapped inside prompt text.
- Keep repo-aware suggestions read-only, nearest-root-aware, and resilient when
  the current directory is not a RuneContext project.
- Keep adapter-triggered validation narrowly scoped to authored authoritative
  RuneContext files rather than generated artifacts, adapter-managed files, or
  unrelated repository code.
- Keep adapter sync ownership explicit: tool-managed files live in a namespaced
  managed subtree, user-owned config boundaries stay reviewable, and synced
  manifests remain convenience metadata rather than correctness-critical state.
- Keep adapter terminology crisp: `adapter` names the tool UX layer, `adapter
  pack` names the packaged release payload for an adapter, and
  `runectx adapter sync <tool>` names the local materialization surface.
- Keep the `generic` adapter as a thin host-agnostic baseline pack focused on
  docs, examples, and manual/CLI-assisted workflows; dynamic suggestions and
  tool-native automation belong to shared CLI or tool-specific layers instead.
- Keep adapter compatibility mode explicit and capability-based: weaker hosts may
  lose convenience features such as prompts, hooks, or dynamic suggestions, but
  they must not change core RuneContext semantics.
- Keep future project/company instruction assets as a separate portable source
  family compiled by adapters into tool-native skills, prompts, or instruction
  files; do not treat those generated tool-native files as the authoritative
  source of truth.
- Keep capability-bearing tool configuration outside RuneContext core semantics,
  including permissions, execution rules, hooks with side effects, MCP trust or
  credential config, model-selection policy, and any other setting that would
  widen runtime authority.
- Keep write-command `--dry-run` behavior centered on simulating planned
  mutations and validating the resulting would-be project state rather than
  emitting prose-only intent.
- Keep deployment-specific evidence discovery, service locators, tenancy/auth,
  and checkpoint-routing metadata outside RuneContext core semantics; RuneCode-
  owned metadata may reference RuneContext outputs without redefining them.
- Keep assurance baseline and receipt families aligned around one portable
  artifact envelope: explicit artifact kind, stable subject identity,
  deterministic hashing/canonicalization metadata where applicable, and visible
  provenance classes so future audit consumers do not need a format refactor.
- Keep context packs generally on-demand or ephemeral, and keep high-frequency
  runtime evidence out of the committed RuneContext tree; baselines and minimal
  portable receipts may be committed when assurance requires them.
- Keep `runectx bundle resolve` read-only in every assurance tier; portable
  context-pack receipts come from an explicit verified capture surface that
  builds the pack and receipt from the same validated snapshot rather than from
  hidden side effects during resolve.
- Keep assurance validation repo-local and self-contained: schema, integrity,
  and linkage checks should not depend on external services, home-directory
  caches, or replaying historical operations.
- Keep backfill additive-only and bounded to pre-adoption history; imported
  evidence attaches to the adoption baseline and never rewrites native captured
  verified receipts.
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
- Keep alpha.7 adapter sync focused on local materialization from installed or
  pinned release contents; alpha.8 hardens release packaging and broader sync/
  update behavior without changing the local-first sync model.
- Keep `runectx` network access limited to an explicit, narrow `upgrade` flow;
  routine project initialization, adapter sync, validation, and other project
  file operations should use already-installed local release contents.
- Keep `runectx upgrade` explicit and preview-first: `runectx upgrade` reports
  the reviewable plan, and `runectx upgrade apply` is the only durable mutation
  surface for source upgrades and migrations.
- Keep migration from other spec-driven systems separate from `runectx upgrade`;
  external adoption/import flows are future dedicated surfaces, not hidden
  upgrade behavior.
- Keep `runectx upgrade` state explicit and fail-closed: project state should be
  classified as current, upgradeable, unsupported, mixed/stale, or conflicted
  before apply is allowed.
- Keep project upgrade planning centered on project-level `runecontext_version`
  transitions, with file-level `schema_version` checks and explicit migration
  markers acting as subordinate gates for individual transforms.
- Keep source upgrades transactional: stage work in tool-owned temporary space,
  validate the staged result before replacing live files, and roll back
  automatically on any in-flight failure.
- Keep successful rollback in normal project history rather than a hidden
  RuneContext rollback store or other second source of truth.
- Keep embedded upgrade conflict handling fail-closed: if user-modified managed
  files are detected, preview should emit a reviewable conflict set and
  `upgrade apply` should refuse to proceed rather than auto-merging or
  overwriting.
- Keep `type: path` sources externally managed for upgrades: surface the owning
  source path and instructions, but never mutate files outside the selected
  project root.
- Keep mixed-version trees after merge/rebase invalid but repairable through an
  explicit rerun of `runectx upgrade`; `validate` and `doctor` should detect the
  stale-file state rather than silently tolerating it.
- Keep Windows MVP support focused on repo-bundle usability and portability
  validation; convenience binary/distribution parity beyond Linux/macOS remains
  post-MVP work.
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
- When a generated artifact carries both deterministic content and emitted audit
  metadata, test hash stability against the canonical content contract rather
  than assuming every persisted field belongs in the canonical hash input.
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
- Add golden and behavior tests for the shared machine-facing CLI envelope so
  JSON/failure contracts do not drift command by command.
- Add discovery-versus-promotion tests proving that `standard discover` stays
  advisory, interactive confirmation is required before any promote handoff, and
  `--non-interactive` discovery exits without mutation.
- Before full `--json` lands, narrow early CLI commands may use stable
  line-oriented machine output if that contract is explicitly documented and
  tested.
- Add adapter smoke tests and reference-project tests so UX layers stay aligned
  with the same core semantics.
- Add upgrade tests covering preview/apply behavior, transactional rollback,
  `type: path` refusal, and merge/rebase recovery through idempotent reruns.
- Add adapter tests ensuring that mutations to authoritative RuneContext files
  automatically trigger `runectx validate` and surface failures immediately.
- Add golden tests for generated Bash, Zsh, and Fish completion scripts and
  parity tests proving the completion metadata matches the actual CLI surface.
- Add fixture tests for repo-aware suggestions so dynamic completion stays
  correct across embedded, linked, monorepo, and out-of-project cases.
- Add clean-machine and no-hidden-state parity tests showing that portable
  outputs stay correct when home-directory state, caches, or other non-declared
  local tool state are absent.
- Add RuneCode companion parity fixtures wherever this repository defines a
  contract RuneCode will later consume.
- Do not treat a feature as complete in any alpha until the tests and fixtures
  for that feature land in the same milestone.
