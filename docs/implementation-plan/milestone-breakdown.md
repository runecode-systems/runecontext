# Milestone Breakdown

This document breaks the RuneContext MVP into alpha-sized milestones, then
decomposes each milestone into epics and issue-sized work items.

## Planning Notes

- Every alpha should be releasable.
- Every issue should be reviewable on its own.
- RuneCode companion-track checkpoints are not work items for this repository,
  but they should be possible by the end of the listed alpha.
- The milestone sequence assumes core semantics land before CLI/adapter UX and
  release hardening.

## `v0.1.0-alpha.1` - Core Model And Contracts

Primary outcome: freeze the portable RuneContext source model and the contracts
that all later CLI, adapters, and RuneCode integration must share.

### Epic 1: Core terminology, boundaries, and ownership

- [x] Issue: publish the normative terminology and disambiguation set for
  `standard`, `project context`, `context bundle`, `context pack`, `change`,
  `spec`, and `decision`.
- [x] Issue: document the three-layer boundary between RuneContext Core,
  adapters, and RuneCode integration, including explicit non-responsibilities.
- [x] Issue: finalize the authoritative on-disk layout, including which files
  are hand-authored, generated, optional, or review-only.
- [x] Issue: codify the policy-neutrality rule so RuneContext text never
  becomes runtime authority.
- [x] Issue: codify the LLM input trust-boundary rule so RuneContext text is
  treated as untrusted model input, not just untrusted policy input.

### Epic 2: Schema and file-contract baseline

- [x] Issue: author the schema and validation contract for root
  `runecontext.yaml`, including `schema_version`, `runecontext_version`,
  `assurance_tier`, `source`, and `allow_extensions` opt-in.
- [x] Issue: author the schema and validation contract for `bundles/*.yaml`
  with closed-schema defaults and optional `extensions` object.
- [x] Issue: author the schema and validation contract for
  `changes/*/status.yaml` with closed-schema defaults and optional `extensions`.
- [x] Issue: codify the base `status.yaml` `type` enum plus the `x-` prefix rule
  for custom type values.
- [x] Issue: codify the `verification_status` enum values `pending`, `passed`,
  `failed`, and `skipped`.
- [x] Issue: define the normative `source_verification` enum values: `pinned_commit`,
  `verified_signed_tag`, `unverified_mutable_ref`, `unverified_local_source`, `embedded`.
- [x] Issue: author the schema and validation contract for generated context
  packs (fully closed, no extensions in v1).
- [x] Issue: define the schema inventory for assurance baseline and receipt
  artifacts (closed schemas, deferred implementation to alpha.5).
- [x] Issue: codify the restricted machine-readable YAML profile with no anchors,
  aliases, duplicate keys, or custom tags; UTF-8 only; normalized formatting.
- [x] Issue: define the canonical JSON data model for hashing derived from YAML.
- [x] Issue: codify RFC 8785 JCS as the canonical hashing serialization; SHA256 for hash
  algorithm.
- [x] Issue: standardize on JSON Schema Draft 2020-12 so conditional variants can remain closed without reopening the core contracts.
- [x] Issue: define unknown-field behavior: closed schemas by default; unknown
  `schema_version` fails closed; optional `extensions` object (owner.name namespaces,
  non-authoritative) allowed only when `runecontext.yaml` sets `allow_extensions: true`;
  no extensions in generated artifacts.

### Epic 3: Markdown contract enforcement

- [x] Issue: define and validate the strict section ordering for `proposal.md`:
  Summary, Problem, Proposed Change, Why Now, Assumptions, Out of Scope, Impact
  (in exact order, level-2 headings, each required or explicit N/A).
- [x] Issue: define the normalized structure and regeneration behavior for
  `standards.md`: Applicable Standards, Standards Added Since Last Refresh (optional),
  Standards Considered But Excluded (optional), Resolution Notes (optional).
  Auto-maintained by tooling; always present; reviewable diffs required.
- [ ] Issue: define traceability metadata conventions for `specs/` and
  `decisions/`.

### Epic 4: Canonical data rules

Completed as part of Epic 2 (consolidated with schema contracts for better audit coverage).

### Epic 5: Testing foundation

- [ ] Issue: define repository-wide test layers and coverage expectations for
  unit tests, golden fixtures, CLI integration tests, adapter smoke tests, and
  reference-project tests.
- [ ] Issue: define fixture taxonomy and storage conventions for schemas,
  markdown contracts, source resolution, context packs, assurance artifacts,
  and CLI JSON output.
- [ ] Issue: establish the rule that new semantics cannot land without tests or
  fixtures in the same milestone.

### Exit Criteria

- All v1 JSON schemas are authored and versioned: `runecontext.yaml`, `bundles/*.yaml`,
  `changes/*/status.yaml`, `context-pack.yaml`.
- The restricted YAML profile and canonical JSON/JCS hashing model are frozen.
- The schema dialect is frozen at JSON Schema Draft 2020-12 for all v1 contracts.
- Unknown-field behavior is explicit: closed schemas by default; optional `extensions`
  object with namespaced keys; no extensions in generated artifacts.
- Context-pack serialization shape is explicit enough to hash deterministically across implementations, including stable handling of empty aspect inventories.
- Alpha-stage contract refinements must update generators, fixtures, and docs together before downstream consumers depend on the hash or schema shape.
- Schema version 1 files must fail closed on unknown `schema_version` and unknown enum
  values.
- Markdown contracts for `proposal.md` and `standards.md` are frozen with section ordering
  and regeneration rules.
- The `allow_extensions: true` opt-in flag is defined in `runecontext.yaml` with visible
  warning behavior.
- The core file model is frozen enough for fixture generation.
- The policy-neutrality rule is explicit and testable.
- The LLM-input trust rule is explicit and testable.
- The test strategy and fixture taxonomy are defined early enough to shape every
  later alpha.
- Future alphas can build without reopening naming or ownership decisions.

### RuneCode Companion-Track Checkpoints

- RuneCode can start version-gating against the root `runecontext.yaml`
  contract.
- RuneCode can parse `proposal.md` into a compact intent summary without
  granting it runtime authority.
- RuneCode can begin fixture-based validation of policy-neutrality assumptions.

## `v0.1.0-alpha.2` - Source Resolution And Bundle Engine

Primary outcome: make storage modes and bundle semantics deterministic,
auditable, and safe for future local/remote parity.

### Epic 1: Source modes and discovery

- [ ] Issue: implement embedded-mode RuneContext resolution.
- [ ] Issue: implement linked git source resolution by pinned commit SHA.
- [ ] Issue: implement linked git source resolution by signed tag, including
  trusted-signer verification, resolved signer identity capture, `expect_commit`
  validation, and fail-closed mismatch behavior.
- [ ] Issue: implement linked git source resolution by mutable ref with
  required `allow_mutable_ref` opt-in and visible warnings.
- [ ] Issue: implement local path source resolution with `unverified_local_source`
  posture, bounded symlink handling, and snapshot-before-hash behavior.
- [ ] Issue: implement monorepo nearest-ancestor discovery and selected-config
  reporting.

### Epic 2: Context bundle semantics

- [ ] Issue: implement bundle loading, `id` uniqueness checks, and unknown
  parent rejection.
- [ ] Issue: implement depth-first, left-to-right parent linearization with
  duplicate ancestor collapse.
- [ ] Issue: implement inheritance cycle rejection and maximum depth `8`
  enforcement.
- [ ] Issue: implement ordered include/exclude rule evaluation with
  last-matching-rule-wins semantics per aspect family.
- [ ] Issue: implement exact-path, glob, and authoring-time diagnostics for
  missing paths and changed match sets.

### Epic 3: Path and integrity guardrails

- [ ] Issue: reject path traversal segments, absolute paths, and drive-qualified
  paths in bundle rules.
- [ ] Issue: reject files that escape the RuneContext root or selected aspect
  roots through traversal or symlink resolution.
- [ ] Issue: record resolved source metadata, including source mode, resolved
  commit, verification posture, and signed-tag signer details when present.
- [ ] Issue: define remote/CI invalidity rules for `type: path` sources.

### Epic 4: Resolution tests and fixtures

- [ ] Issue: add unit tests for embedded, linked-by-commit, linked-by-signed-
  tag, mutable-ref, and path-based source resolution.
- [ ] Issue: add unit and golden tests for bundle precedence, cycle rejection,
  depth rejection, glob changes, and path-escape failures.
- [ ] Issue: add golden fixtures for embedded, linked, path, and monorepo
  resolution outputs so later CLI and RuneCode parity tests share one baseline.

### Exit Criteria

- Embedded, linked, and path-based projects all resolve under one consistent
  model.
- Signed-tag verification is supported as an advanced MVP path.
- Bundle inheritance behaves deterministically across override and diamond cases.
- Resolution fails closed for cycles, escapes, and integrity mismatches.
- Shared fixtures exist for every supported source mode and precedence rule.

### RuneCode Companion-Track Checkpoints

- RuneCode can run shared embedded/linked/path source fixtures against its own
  future resolver.
- RuneCode can validate signed-tag verification parity and fail-closed behavior.
- RuneCode can confirm local and remote resolution produce the same selected
  file set from the same inputs.

## `v0.1.0-alpha.3` - Change Workflow, Standards, And Traceability

Primary outcome: make RuneContext usable as a change-oriented workflow system
with stable IDs, lightweight shaping, and reviewable standards linkage.

### Epic 1: Change identity and lifecycle

- [ ] Issue: implement year-scoped change ID allocation with monotonic counter
  plus collision-resistant suffix.
- [ ] Issue: implement minimum-mode change scaffolding with `status.yaml`,
  `proposal.md`, and `standards.md`.
- [ ] Issue: implement full-mode materialization for `design.md`, `tasks.md`,
  `references.md`, and `verification.md`.
- [ ] Issue: implement lifecycle state validation for `proposed`, `planned`,
  `implemented`, `verified`, `closed`, and `superseded`.
- [ ] Issue: implement bidirectional supersession consistency checks.
- [ ] Issue: implement merge-time change-ID collision detection, reallocation,
  and atomic local-reference rewriting for the rare case where branches still
  collide.

### Epic 2: Progressive disclosure and intake heuristics

- [ ] Issue: define the branching rules for `project`, `feature`, `bug`,
  `standard`, and `chore` work so minimum mode versus full mode is chosen
  consistently.
- [ ] Issue: define size and risk escalation rules so `small`, `medium`, and
  `large` work items shape correctly.
- [ ] Issue: define the deeper intake checklist for new-project work, including
  mission, target users, stack/runtime constraints, deployment/security
  constraints, success criteria, and non-goals.
- [ ] Issue: define the bug-workflow escalation rules for unclear root causes,
  security impact, schema impact, API impact, and behavior ambiguity.
- [ ] Issue: define the "ask more vs less" heuristics for when RuneContext must
  probe further versus infer defaults from repository conventions.
- [ ] Issue: ensure inferred assumptions are recorded in `proposal.md` when
  non-trivial decisions are made without prompting.

### Epic 3: Intent artifacts and standards linkage

- [ ] Issue: generate and validate `proposal.md` using the required heading
  order and explicit `N/A` rules.
- [ ] Issue: generate and validate `status.yaml` fields, including type, size,
  verification status, and promotion assessment placeholders.
- [ ] Issue: populate and refresh `standards.md` during change creation and
  shaping.
- [ ] Issue: enforce reviewable diffs for any automatic `standards.md` refresh.

### Epic 4: Standards authoring model

- [ ] Issue: validate standard frontmatter, including stable `id` path matching.
- [ ] Issue: implement `draft`, `active`, and `deprecated` standard-state
  handling.
- [ ] Issue: implement `replaced_by` and `aliases` support for migration and
  rename workflows.
- [ ] Issue: enforce the rule that standards are referenced by path instead of
  copied into change/spec bodies.
- [ ] Issue: ensure `suggested_context_bundles` remains advisory metadata only
  and never becomes authoritative bundle membership.

### Epic 5: History preservation and traceability

- [ ] Issue: implement close behavior that updates state without moving change
  folders into an archive tree.
- [ ] Issue: implement traceability fields connecting changes, specs, and
  decisions.
- [ ] Issue: validate that `depends_on`, `informed_by`, `related_changes`,
  `related_specs`, and `related_decisions` resolve to real artifacts or report
  clear diagnostics.
- [ ] Issue: ensure closed changes remain directly readable at stable paths.
- [ ] Issue: define the minimum traceability needed for future lineage/index
  tooling without building that lineage view yet.

### Epic 6: Workflow tests and fixtures

- [ ] Issue: add unit tests for change ID allocation, lifecycle transitions,
  supersession consistency, and collision reallocation behavior.
- [ ] Issue: add parser/validator tests for `proposal.md` and `standards.md`
  contracts.
- [ ] Issue: add golden fixtures for minimum-mode, full-mode, closed, and
  superseded change folders.
- [ ] Issue: add tests for dangling cross-artifact references and standards-
  maintenance review-diff behavior.

### Exit Criteria

- Every substantive work item can start in minimum mode and deepen only when
  needed.
- `proposal.md` is the canonical reviewable intent artifact.
- `standards.md` is always present and reviewably maintained.
- Change history stays accessible at stable paths.
- Minimum/full-mode branching and prompting heuristics are both tested.

### RuneCode Companion-Track Checkpoints

- RuneCode can bind active change IDs plus proposal sections into audit-history
  fixtures.
- RuneCode can generate reviewable `standards.md` updates rather than silent
  mutations.
- RuneCode can consume change close outputs as inputs to future promotion flows.

## `v0.1.0-alpha.4` - Deterministic Context Packs, Promotion, And Indexes

Primary outcome: generate deterministic resolved outputs and supporting indexes
that make RuneContext consumable by power users, automation, and future
RuneCode integration.

### Epic 1: Context-pack generation

- [ ] Issue: implement selected and excluded file inventories with per-file
  hashes.
- [ ] Issue: implement compact deterministic provenance showing which rules
  selected or excluded each file.
- [ ] Issue: implement top-level pack hashing over the canonicalized resolved
  pack.
- [ ] Issue: implement source metadata capture inside the context pack,
  including resolved commit and signed-tag verification posture.
- [ ] Issue: implement stable ordering rules for all generated pack content.
- [ ] Issue: implement compact deterministic provenance in the context pack
  while preserving a clean extension path for fuller provenance receipts in
  Verified mode.

### Epic 2: Pack explanation and limits

- [ ] Issue: implement human-readable and machine-readable pack output modes.
- [ ] Issue: implement `--explain`-style provenance output for include/exclude
  decisions.
- [ ] Issue: implement advisory warnings using the design defaults of `256`
  selected files, `1 MiB` referenced content bytes, and `256 KiB` provenance
  metadata.
- [ ] Issue: implement fail/rebuild behavior when files change between
  enumeration, hashing, and delivery preparation.

### Epic 3: Generated indexes and manifests

- [ ] Issue: implement overall `manifest.yaml` generation.
- [ ] Issue: implement generated change indexes grouped by lifecycle state.
- [ ] Issue: implement generated bundle inventory views showing parents and
  referenced patterns.
- [ ] Issue: ensure generated indexes use stable ordering and merge-friendly
  formatting.

### Epic 4: Promotion assessment

- [ ] Issue: implement structured promotion assessment records in `status.yaml`.
- [ ] Issue: implement the full promotion-assessment status lifecycle:
  `pending`, `none`, `suggested`, `accepted`, and `completed`.
- [ ] Issue: implement reviewable suggested promotion targets for `specs/`,
  `standards/`, and `decisions/`.
- [ ] Issue: implement explicit "no promotion needed" recording on close.

### Epic 5: Determinism and pack-quality tests

- [ ] Issue: add golden fixtures for resolved context packs, selected/excluded
  provenance, and top-level pack hashes.
- [ ] Issue: add regression tests for advisory-size and provenance-threshold
  warnings using the documented default values.
- [ ] Issue: add tests for changed-file fail-closed behavior between
  enumeration, hashing, and delivery preparation.
- [ ] Issue: add fixtures for generated manifest and change-index stability.

### Exit Criteria

- Any bundle selection can be flattened into a deterministic context pack.
- The context pack contains the top-level canonical hash required for future
  audit binding.
- Promotion assessment is structured and reviewable.
- Generated indexes aid browsing without becoming the source of truth.
- Deterministic outputs are protected by golden tests rather than manual spot
  checking.

### RuneCode Companion-Track Checkpoints

- RuneCode can test direct-resolver versus CLI parity using shared context-pack
  fixtures and expected pack hashes.
- RuneCode can draft typed isolate-delivery descriptor fixtures from the pack's
  resolved metadata and hashes.
- RuneCode can verify that over-limit context packs fail loudly rather than
  being silently truncated.

## `v0.1.0-alpha.5` - Assurance Tiers And Verifiable Tracing

Primary outcome: support both low-friction standalone use and stronger
verifiable tracing, while keeping assurance progressive rather than mandatory.

### Epic 1: Assurance-tier model

- [ ] Issue: implement persisted `plain` versus `verified` tier behavior.
- [ ] Issue: implement generated baseline artifact shape and baseline creation.
- [ ] Issue: implement receipt schemas and file conventions for context packs,
  changes, promotions, and verifications.
- [ ] Issue: implement receipt hashing, receipt IDs, and collision-resistant
  filenames.

### Epic 2: Verified enablement flow

- [ ] Issue: implement the Verified enablement flow from adoption commit through
  baseline generation.
- [ ] Issue: record initial resolved source posture and adoption metadata.
- [ ] Issue: implement receipt generation triggers for future verified
  operations.
- [ ] Issue: implement commit-policy guidance for what is committed, ignored, or
  treated as ephemeral in Verified mode.

### Epic 3: Backfill and historical provenance

- [ ] Issue: implement imported provenance class support for historical work.
- [ ] Issue: implement backfill traversal over git history and existing
  RuneContext artifacts.
- [ ] Issue: attach imported/backfilled evidence to the adoption baseline.
- [ ] Issue: ensure imported evidence remains visibly distinct from native
  verified evidence.

### Epic 4: Merge and concurrency behavior

- [ ] Issue: ensure assurance receipt generation does not depend on a shared
  mutable index.
- [ ] Issue: ensure receipts merge by file union where possible across branches.
- [ ] Issue: ensure concurrent verified work does not produce hidden
  correctness-critical state outside the repository.

### Epic 5: Assurance tests and fixtures

- [ ] Issue: add golden fixtures for baselines and each receipt family.
- [ ] Issue: add tests distinguishing `captured_verified` from
  `imported_git_history` provenance.
- [ ] Issue: add tests for Verified enablement, backfill flow, and merge-safe
  receipt generation.
- [ ] Issue: add fixtures RuneCode can reuse to test audited-workflow gating and
  provenance ingestion.

### Exit Criteria

- Plain mode remains lightweight and usable without extra receipt generation.
- Verified mode can generate baseline and receipt artifacts for future audit
  consumption.
- Historical backfill can strengthen trust without pretending to be native
  verified capture.
- Assurance behavior is covered by deterministic fixtures rather than narrative
  examples only.

### RuneCode Companion-Track Checkpoints

- RuneCode can gate audited workflows on `assurance_tier: verified`.
- RuneCode can ingest baseline and receipt fixtures into its audit/provenance
  model.
- RuneCode can reject `type: path` packs as verified provenance in audited
  flows.

## `v0.1.0-alpha.6` - Minimal CLI And Machine-Facing Operations

Primary outcome: expose the small universal command surface needed for
automation, CI, debugging, and non-agent workflows.

### Epic 1: Primary commands

- [ ] Issue: implement `runectx init`.
- [ ] Issue: implement `runectx status`.
- [ ] Issue: implement `runectx change new`.
- [ ] Issue: implement `runectx change shape`.
- [ ] Issue: implement `runectx bundle resolve`.
- [ ] Issue: implement `runectx change close`.

### Epic 2: Secondary and admin commands

- [ ] Issue: implement `runectx validate`.
- [ ] Issue: implement `runectx doctor`.
- [ ] Issue: implement `runectx standard discover`.
- [ ] Issue: implement `runectx promote`.
- [ ] Issue: implement `runectx assurance enable verified`.
- [ ] Issue: implement `runectx assurance backfill`.
- [ ] Note: `runectx update` is intentionally deferred to `v0.1.0-alpha.8`
  alongside release/install hardening.

### Epic 3: Universal machine-facing flags

- [ ] Issue: implement `--json` output contracts across machine-facing commands.
- [ ] Issue: implement `--non-interactive` behavior with clear inference or
  failure rules.
- [ ] Issue: implement `--dry-run` behavior for write operations.
- [ ] Issue: implement `--explain` output for resolution, standards selection,
  and promotion suggestions.

### Epic 4: Parity and automation readiness

- [ ] Issue: define stable exit codes and failure classes for automation.
- [ ] Issue: build CLI versus library parity fixtures.
- [ ] Issue: ensure all write commands surface reviewable diffs or proposed
  mutations rather than silent commits.

### Epic 5: CLI test coverage

- [ ] Issue: add integration tests for every primary command.
- [ ] Issue: add integration tests for every secondary/admin command.
- [ ] Issue: add snapshot or golden tests for `--json` outputs.
- [ ] Issue: add behavior tests for `--non-interactive`, `--dry-run`, and
  `--explain`.
- [ ] Issue: add tests for failure classes, diagnostics, and exit-code
  stability.

### Exit Criteria

- Power users can manage RuneContext entirely through the CLI.
- Automation and CI can consume structured command outputs.
- CLI semantics stay aligned with the canonical file model rather than becoming
  a competing source of truth.
- CLI behavior is protected by integration tests and machine-readable golden
  outputs.

### RuneCode Companion-Track Checkpoints

- RuneCode can run parity suites between its future direct integration and the
  CLI.
- RuneCode can consume JSON status, resolve, and close outputs in integration
  tests.
- RuneCode can validate non-interactive behavior for remote/server workflows.

## `v0.1.0-alpha.7` - Adapters And Adapter-Pack UX

Primary outcome: make RuneContext comfortable to use inside multiple coding
tools while preserving one core model.

### Epic 1: Canonical operations reference

- [ ] Issue: author the canonical in-project operations reference under
  `runecontext/operations/`.
- [ ] Issue: define adapter-to-core operation mapping rules.
- [ ] Issue: define how adapters consume or derive from the canonical
  operations reference without redefining semantics.

### Epic 2: Generic adapter

- [ ] Issue: author the `generic` adapter pack with plain markdown workflow
  docs.
- [ ] Issue: provide example flows for manual, CLI-assisted, and non-agent use.

### Epic 3: Tool-specific adapters

- [ ] Issue: author the `claude-code` adapter pack.
- [ ] Issue: author the `opencode` adapter pack.
- [ ] Issue: author the `codex` adapter pack.
- [ ] Issue: define compatibility-mode guidance for hosts with weaker
  interaction capabilities.

### Epic 4: Adapter packaging and sync

- [ ] Issue: implement adapter packaging for release artifacts.
- [ ] Issue: implement the `runectx adapter sync <tool>` command as the
  adapter-management CLI surface.
- [ ] Issue: define merge-aware adapter sync/update behavior.
- [ ] Issue: ensure adapters never introduce tool-specific source-of-truth
  files.

### Epic 5: Adapter tests and parity

- [ ] Issue: add smoke tests for the `generic`, `claude-code`, `opencode`, and
  `codex` adapters.
- [ ] Issue: add parity checks showing adapter flows map back to the same core
  operations and expected file mutations.
- [ ] Issue: add tests ensuring adapters do not introduce hidden state or
  adapter-only correctness requirements.

### Exit Criteria

- At least one tool-specific adapter is usable end to end.
- All adapters map back to the same underlying operations.
- Users can still work directly with repo files and CLI without any adapter.
- Adapter behavior is covered by parity and smoke tests rather than manual
  walkthroughs only.

### RuneCode Companion-Track Checkpoints

- RuneCode can use adapter docs as workflow fixtures for change, standards, and
  promotion review flows.
- RuneCode can verify that adapter UX does not smuggle in RuneCode-only hidden
  state or permissions.

## `v0.1.0-alpha.8` - Release, Install, Update, And End-To-End Hardening

Primary outcome: harden RuneContext as a distributable product with tested
install/update paths and end-to-end reference fixtures.

### Epic 1: Release packaging

- [ ] Issue: establish CI/CD platform parity with RuneCode:
  - Primary: Linux (x86_64 and arm64) and macOS (x86_64 and arm64) via Nix.
  - Portability: Windows via non-Nix smoke testing.
- [ ] Issue: package the schema bundle for releases across supported platforms.
- [ ] Issue: package adapter packs for releases.
- [ ] Issue: package optional `runectx` binaries for primary supported platforms:
  `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64`.
- [ ] Issue: emit release checksums, release manifest, signatures, and release
  notes.
- [ ] Issue: publish a RuneCode `<->` RuneContext compatibility matrix.

### Epic 2: Install and update flows

- [ ] Issue: document and test the canonical manual repo-install flow.
- [ ] Issue: implement `runectx update` as a diff-first, reviewable update
  flow.
- [ ] Issue: harden adapter sync/update to be namespaced and merge-aware.
- [ ] Issue: ensure `doctor` reports unsupported version combinations and
  integrity posture issues.

### Epic 3: Reference projects and fixtures

- [ ] Issue: create an embedded-mode reference project fixture.
- [ ] Issue: create a linked-by-commit reference project fixture.
- [ ] Issue: create a linked-by-signed-tag reference project fixture.
- [ ] Issue: create a Verified-mode reference project fixture.
- [ ] Issue: create a monorepo reference fixture with nested RuneContext roots.

### Epic 4: MVP readiness review

- [ ] Issue: run the full MVP acceptance matrix against reference fixtures.
- [ ] Issue: freeze v1 naming, schema, and lifecycle semantics.
- [ ] Issue: verify all generated artifacts remain derived rather than
  authoritative.
- [ ] Issue: confirm signed-tag verification is tested in both standalone and
  RuneCode companion-track readiness flows.

### Epic 5: Release and workflow test hardening

- [ ] Issue: add tests covering release artifact contents, checksums, manifests,
  adapter packs, and optional binaries.
- [ ] Issue: add end-to-end tests for manual repo install, CLI-managed install,
  and diff-first update flows.
- [ ] Issue: add regression tests asserting forbidden install/update patterns do
  not appear: required global installs, bash-only installers, overwriting
  existing tool config files, hidden runtime-manager dependencies, and silent
  auto-updates.
- [ ] Issue: add end-to-end tests over reference projects for embedded,
  linked-by-commit, linked-by-signed-tag, Verified-mode, and monorepo cases.

### Exit Criteria

- RuneContext can be installed manually, managed by CLI, and updated reviewably.
- Release artifacts are canonical, inspectable, and compatible with the repo-
  first distribution model.
- Signed-tag verification is included in MVP validation, not deferred.
- Install, update, and release guarantees are backed by automated end-to-end
  tests.

### RuneCode Companion-Track Checkpoints

- RuneCode can consume the compatibility matrix and enforce supported ranges.
- RuneCode can test local versus remote parity against released RuneContext
  artifacts.
- RuneCode can validate linked signed-tag sources in audited integration paths.

## `v0.1.0` - MVP Stabilization And Release

Primary outcome: ship the full RuneContext MVP with all v1 repository-side
features described in `docs/project_idea.md`.

### Final Release Work

- [ ] Issue: execute the MVP acceptance checklist in
  `docs/implementation-plan/mvp-acceptance.md`.
- [ ] Issue: finalize user-facing docs, compatibility docs, and release notes.
- [ ] Issue: run final schema, CLI, adapter, and fixture parity validation.
- [ ] Issue: tag and publish `v0.1.0`.

### RuneCode Companion-Track Checkpoints

- RuneCode can begin full integration work against a stable RuneContext MVP
  contract.
- RuneCode has parity fixtures for source resolution, context packs, Verified
  mode gating, and change-intent binding.
- RuneCode integration work can proceed without asking RuneContext to reopen its
  core semantics.
