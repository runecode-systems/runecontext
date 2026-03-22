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

## `v0.1.0-alpha.1` - Core Model And Contracts - COMPLETED

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
  artifacts (closed schemas, deferred implementation to alpha.6).
- [x] Issue: codify the restricted machine-readable YAML profile with no anchors,
  aliases, duplicate keys, or custom tags; UTF-8 only; normalized formatting.
- [x] Issue: define the canonical JSON data model for hashing derived from YAML.
- [x] Issue: codify canonical hashing serialization rules; SHA256 for the hash
  algorithm, with later alpha.4 refinement for the context-pack-specific
  `runecontext-canonical-json-v1` token.
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
- [x] Issue: define traceability metadata conventions for `specs/` and
  `decisions/`.

### Epic 4: Canonical data rules

Completed as part of Epic 2 (consolidated with schema contracts for better audit coverage).

### Epic 5: Testing foundation

- [x] Issue: define repository-wide test layers and coverage expectations for
  unit tests, golden fixtures, CLI integration tests, adapter smoke tests, and
  reference-project tests.
- [x] Issue: define fixture taxonomy and storage conventions for schemas,
  markdown contracts, source resolution, context packs, assurance artifacts,
  and CLI JSON output.
- [x] Issue: establish the rule that new semantics cannot land without tests or
  fixtures in the same milestone.
- [x] Issue: add a narrow executable validation entrypoint so alpha.1 contracts
  can be enforced without waiting for the broader alpha.5 CLI surface.
- [x] Issue: define a stable line-oriented machine output contract for the
  narrow alpha.1 validation entrypoint without pre-empting the broader alpha.5
  `--json` work.
- [x] Issue: close review-identified fail-open gaps in the alpha.1 validation
  foundation, including project-level markdown enforcement, bundle validation,
  structured error handling, and restricted-YAML tag rejection.
- [x] Issue: close PR-review gaps in alpha.1 validation hardening, including
  content-root-aware project validation, full restricted-YAML style checks, and
  segment-safe spec/decision path matching.
- [x] Issue: close re-review correctness gaps in alpha.1 foundation work,
  including exact frontmatter delimiter parsing, valid Go module/toolchain
  directives, release metadata version alignment, and avoiding redundant
  project-validation rereads.

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
- Markdown contracts for `proposal.md`, `standards.md`, `specs/*.md`, and
  `decisions/*.md` are frozen with section ordering, frontmatter, path-matched IDs,
  and regeneration rules where applicable.
- The `allow_extensions: true` opt-in flag is defined in `runecontext.yaml` with visible
  warning behavior.
- The core file model is frozen enough for fixture generation.
- The policy-neutrality rule is explicit and testable.
- The LLM-input trust rule is explicit and testable.
- The test strategy and fixture taxonomy are defined early enough to shape every
  later alpha.
- An executable Go validation foundation exists for schema, markdown, YAML-profile,
  and alpha.1 traceability rules.
- A narrow `runectx validate [path]` entrypoint exists for whole-project contract
  enforcement without pulling alpha.5 command breadth into alpha.1.
- The alpha.1 validation entrypoint emits stable one-line `key=value` fields for
  success, invalid, and usage-error outcomes so CI and scripts can consume it
  before broader machine-facing flags land.

- Whole-project validation now covers required change markdown files,
  `runecontext/bundles/*.yaml`, extension opt-in enforcement, and restricted YAML
  tag rejection without panic-based failure paths.
- Whole-project validation follows `runecontext.yaml` source-root settings and
  rejects the remaining forbidden YAML styles (flow collections and multiline scalars).
- Frontmatter parsing only accepts exact `---` delimiter lines, release metadata
  matches the documented `v0.1.0-alpha.1` train, and Go module metadata is valid
  for standard tooling.
- Future alphas can build without reopening naming or ownership decisions.

### Historical Implementation Notes

- Alpha.1 delivered the schema/file-contract baseline through four core
  schemas, the contract/profile references in `schemas/`, and the initial
  shipped fixture taxonomy for standalone schema validation, project-level
  extension checks, and restricted YAML-profile rejection cases.
- The security-first baseline from that work remains the foundation for the MVP:
  closed schemas by default, explicit opt-in extensions for authored artifacts
  only, policy-neutral semantics, and deterministic generated artifacts.
- Later hardening passes folded back into the same baseline included embedded-
  root and whole-project symlink containment, explicit git transport guards,
  bundle traversal bounds, defensive-copy bundle results, synchronized schema
  compilation caching, safer bundle-file reads, and clearer resolved-path
  diagnostics.
- Additional follow-up fixes also refined markdown duplicate-heading fragments,
  thin CLI required-flag parsing, bundle traversal fail-closed behavior, and
  `status.yaml` rewrite error propagation.
- Alpha.4 then refined the original context-pack hashing contract further by
  moving to the explicit `runecontext-canonical-json-v1` token, requiring
  whole-second caller-supplied `generated_at`, normalizing UTF-8 text line
  endings before file hashing, and tightening portable `source_ref` rules for
  path-mode packs.

### RuneCode Companion-Track Checkpoints

- RuneCode can start version-gating against the root `runecontext.yaml`
  contract.
- RuneCode can parse `proposal.md` into a compact intent summary without
  granting it runtime authority.
- RuneCode can begin fixture-based validation of policy-neutrality assumptions.

## `v0.1.0-alpha.2` - Source Resolution And Bundle Engine - COMPLETED

Primary outcome: make storage modes and bundle semantics deterministic,
auditable, and safe for future local/remote parity.

### Implementation Notes

- Bundle rules, diagnostics, and generated file inventories should use
  RuneContext-root-relative paths consistently (for example,
  `project/mission.md` and `standards/security/**`). The aspect key still
  constrains the allowed subtree, and mismatched aspect/path combinations must
  fail closed.
- Source resolution should return structured metadata that later alpha.4,
  alpha.5, alpha.6, and RuneCode audit flows can reuse without semantic
  translation.
  That metadata should include the selected config path, project root,
  RuneContext source root, source mode, source ref, resolved commit when
  applicable, verification posture, and warnings/diagnostics.
- Embedded source paths and git `subdir` values must resolve inside the
  selected project root or fetched repository root respectively; absolute or
  escaping values fail closed. `type: path` remains allowed to point outside the
  project repo for developer-local workflows, but any resolved files and
  symlink targets must remain inside the declared local source tree.
- Git resolution must validate user-supplied URL/ref/commit values before
  invoking git, reject option-like values, run with an explicit minimal
  subprocess environment, and disable interactive prompting so correctness does
  not depend on hidden host credentials or config.
- Git URL validation should reject remote-helper forms and constrain subprocess
  protocol use to an explicit allowlist. Any user-visible git errors should
  redact embedded URLs or credentials rather than echoing raw transport details.
- Mutable git refs should be validated more strictly than a broad character
  whitelist so obviously invalid refs fail before any subprocess execution.
- RuneContext should not expose environment-variable configuration or use
  environment variables as semantic inputs. Correctness-critical behavior must
  come from repository state, explicit config files, or caller-supplied options.
  A minimal inherited process environment is allowed only for non-semantic OS
  plumbing such as executable lookup and temp-directory access.
- Git network/process steps must run with explicit timeouts and cancellation so
  validation and future CI flows cannot hang indefinitely on fetch operations.
- Pinned-commit resolution must work against ordinary advertised refs rather than
  relying on direct fetch-by-SHA support from the remote.
- Signed-tag verification must rely on explicitly supplied trusted-signer
  inputs on the trusted side. Alpha.2 should not depend on hidden machine-
  global Git, GPG, or home-directory trust configuration for correctness.
- `type: path` invalidity for remote/CI usage should be driven by an explicit
  caller-supplied execution mode rather than environment sniffing.
- Symlinks may be followed only when their fully resolved targets remain inside
  both the RuneContext root and the selected aspect root; otherwise resolution
  must fail closed.
- Whole-project artifact discovery and reads should apply the same resolved-path
  containment model so specs, decisions, and other validated project files
  cannot escape their selected subtree through symlinked files.
- Symlinked artifact-root directories that still resolve in-bounds (for example,
  a symlinked `specs/` or `standards/` directory) should remain valid; the
  guardrail is on the fully resolved target, not on whether the root directory
  entry itself is a symlink.
- Local path snapshots should exclude obvious repository-control directories like
  `.git/` and fail closed when practical depth, file-count, or byte-size bounds
  are exceeded, so alpha.2 snapshotting remains usable without silently copying
  arbitrarily large trees.
- Bundle traversal should also enforce practical depth and file-count bounds so
  pathological trees fail closed rather than consuming unbounded work.
- Once bundle traversal begins, missing or broken symlink targets discovered
  during exact or glob evaluation should also fail closed consistently rather
  than being silently treated as empty matches.
- Bundle exact and glob evaluation should canonicalize aspect roots before
  containment checks so in-bounds symlinked aspect directories are accepted
  consistently.
- Validation entrypoints that materialize temporary source trees must close and
  clean up those trees on success as well as failure.
- Alpha.2 should capture concrete per-glob match sets and structured
  diagnostics so later CLI and context-pack flows can compare changed match
  sets without inventing hidden persistent state in this milestone.

### Epic 1: Source modes and discovery

- [x] Issue: implement embedded-mode RuneContext resolution.
- [x] Issue: implement linked git source resolution by pinned commit SHA.
- [x] Issue: implement linked git source resolution by signed tag, including
  trusted-signer verification using explicit caller-supplied trust inputs,
  resolved signer identity capture, `expect_commit` validation, and fail-closed
  mismatch behavior.
- [x] Issue: implement linked git source resolution by mutable ref with
  required `allow_mutable_ref` opt-in and visible warnings.
- [x] Issue: implement local path source resolution with `unverified_local_source`
  posture, bounded symlink handling, and snapshot-oriented source-tree capture
  for later hashing/integrity flows.
- [x] Issue: implement monorepo nearest-ancestor discovery and selected-config
  reporting as structured resolution metadata.

### Epic 2: Context bundle semantics

- [x] Issue: implement bundle loading, `id` uniqueness checks, and unknown
  parent rejection.
- [x] Issue: implement depth-first, left-to-right parent linearization with
  duplicate ancestor collapse.
- [x] Issue: implement inheritance cycle rejection and maximum depth `8`
  enforcement.
- [x] Issue: implement ordered include/exclude rule evaluation with
  last-matching-rule-wins semantics per aspect family over RuneContext-root-
  relative bundle paths.
- [x] Issue: implement exact-path, glob, and authoring-time diagnostics for
  missing paths and changed match sets, including concrete per-rule matched file
  inventories for later comparison.

### Epic 3: Path and integrity guardrails

- [x] Issue: reject path traversal segments, absolute paths, and drive-qualified
  paths in bundle rules.
- [x] Issue: reject files that escape the RuneContext root or selected aspect
  roots through traversal or symlink resolution.
- [x] Issue: record resolved source metadata, including source mode, resolved
  commit, verification posture, selected config path, source root, warning set,
  and signed-tag signer details when present.
- [x] Issue: define remote/CI invalidity rules for `type: path` sources through
  explicit caller-supplied execution mode rather than environment inference.

### Epic 4: Resolution tests and fixtures

- [x] Issue: add unit tests for embedded, linked-by-commit, linked-by-signed-
  tag, mutable-ref, and path-based source resolution.
- [x] Issue: add unit and golden tests for bundle precedence, cycle rejection,
  depth rejection, glob changes, and path-escape failures.
- [x] Issue: add golden fixtures for embedded, linked, path, and monorepo
  resolution outputs so later CLI and RuneCode parity tests share one baseline.

### Exit Criteria

- Embedded, linked, and path-based projects all resolve under one consistent
  model.
- Signed-tag verification is supported as an advanced MVP path.
- Signed-tag verification uses explicit trusted-signer inputs rather than hidden
  machine-global trust state.
- Narrow alpha.2 validation entrypoints can accept explicit signed-tag trust
  material and surface structured failure reasons/diagnostics without inventing
  hidden trust discovery.
- Signed-tag verification and surfaced git diagnostics preserve actionable
  execution failures and avoid over-redacting normal git reflog syntax.
- Signed-tag timeout failures retain structured verification diagnostics, and
  explicit trust-input parsing rejects blank values before filesystem access.
- Signed-tag validation rejects empty `expect_commit` values with a clear
  caller-facing error rather than a confusing placeholder-derived format
  failure.
- Monorepo discovery reports the selected config path and source metadata in a
  structured form that later CLI and audit flows can reuse.
- Bundle inheritance behaves deterministically across override and diamond cases.
- Bundle rules and generated inventories use one consistent RuneContext-root-
  relative path model with aspect-boundary enforcement.
- Resolution fails closed for cycles, escapes, and integrity mismatches.
- `type: path` resolution is explicitly invalid in remote/CI mode unless a
  higher-level trusted wrapper deliberately downgrades the run.
- Shared fixtures exist for every supported source mode and precedence rule.

### RuneCode Companion-Track Checkpoints

- RuneCode can run shared embedded/linked/path source fixtures against its own
  future resolver.
- RuneCode can validate signed-tag verification parity and fail-closed behavior.
- RuneCode can confirm local and remote resolution produce the same selected
  file set from the same inputs.

## `v0.1.0-alpha.3` - Change Workflow, Standards, And Traceability - COMPLETED

Primary outcome: make RuneContext usable as a change-oriented workflow system
with stable IDs, lightweight shaping, and reviewable standards linkage.

### Implementation Notes

- Multiple non-closed changes may exist at the same time. RuneContext should not
  require one global active-change slot for the whole repository; instead,
  tooling should surface open versus closed work clearly and let specific runs
  select the relevant active change or changes.
- Lifecycle state and change shape are separate axes. `planned` should not imply
  full mode automatically, and `change shape` should be additive/idempotent
  rather than a destructive regeneration step. Terminal changes should reject
  later shaping so historical artifacts stay immutable.
- Very large or high-risk work should usually start as one minimum-mode change
  and then move into full mode early. `change new` should be able to recommend
  or directly trigger shaping when the requested work appears too large,
  ambiguous, or risky for minimum mode.
- Shaped changes should stay lean by default: `design.md` and
  `verification.md` are the baseline shaped artifacts, while `tasks.md` and
  `references.md` are supplemental shaped docs created only when they add real
  value.
- When a large feature is split into an umbrella change plus smaller sub-
  changes, RuneContext should model that as a linked graph of changes rather
  than a hidden alternate hierarchy. `related_changes` keeps the graph
  navigable; directional `depends_on` links capture prerequisite ordering.
- Tooling should help create and preserve those links when work is split, and
  validation should fail clearly when prerequisite or related-change references
  are missing, inconsistent, or no longer resolve.
- `superseded` should be treated as a terminal state distinct from `closed`:
  the work was replaced by successor change(s), must carry `superseded_by`,
  must remain bidirectionally consistent with successor `supersedes` links, and
  should still preserve stable-path readability, and must not leave
  `verification_status` at `pending`. If repairing a missing
  reciprocal supersession link would require mutating a terminal successor,
  tooling should fail closed instead.
- Structured traceability should remain artifact-first in machine-readable files
  while markdown docs may use machine-validated deep refs via stable
  `path#heading-fragment` syntax. Do not use line numbers as durable refs.
- Change ID slugs and automatically derived heading fragments should stay
  ASCII-safe so authored non-ASCII titles or headings never generate invalid
  machine-readable identifiers.
- Markdown deep-ref validation and tool-assisted rewrite behavior should ignore
  fenced code blocks, require RuneContext-root-relative paths instead of `./`,
  `../`, or absolute `/...` forms, reject line-number fragments such as `#L10`
  or `#42`, and use documented first-match rewrite semantics for scoped update
  rules.
- Markdown deep-ref tokenization must also stay UTF-8-safe so surrounding
  non-ASCII punctuation in prose does not get absorbed into a local markdown ref
  or produce false missing-artifact validation failures, while the
  machine-readable fragment token itself remains ASCII-bounded.
- Markdown deep-ref detection should ignore external URLs even when they contain
  `.md#fragment` suffixes, and the machine-addressable heading subset for
  alpha.3 is ATX `#` headings rather than Setext underlined headings.
- Fenced-code exclusion in alpha.3 includes ordinary fenced blocks and
  blockquote-prefixed fenced examples so quoted examples do not become live
  machine refs.
- Stable deep-ref targets for alpha.3 are the machine-indexed markdown areas:
  `specs/`, `decisions/`, `standards/`, and the top-level markdown files inside
  each `changes/<id>/` directory.
- Artifact traceability in alpha.3 is intentionally minimum viable and
  artifact-level: `related_specs` and `related_decisions` must mirror a real
  change reference on the target artifact, but they do not yet encode a stricter
  machine distinction between "originating" versus "revision" linkage.
  Validation errors should still clearly point reviewers back to the change
  `status.yaml` that needs the reciprocal artifact reference.
- Lifecycle helpers remain forward-only in alpha.3: tooling validates monotonic
  progress and terminal immutability rather than offering an explicit reopen or
  downgrade workflow.
- Automatically derived heading fragments must remain unique within a file even
  when headings naturally collide with suffixed forms such as `foo-2`; tooling
  should preserve deterministic, machine-validated fragments rather than
  silently overwriting earlier headings.
- Duplicate heading fragments should also follow deterministic markdown-anchor
  numbering (`foo`, `foo-1`, `foo-2`, ...) instead of skipping the first
  duplicate suffix, while still advancing past already occupied suffixed forms.
- Thin `runectx status`, `runectx change new`, `runectx change shape`, and
  `runectx change close` entrypoints may land here as narrow wrappers over the
  same core operations, with the broader CLI contract deferred to `alpha.5`.
  Explicit path arguments should remain explicit roots even when the caller
  passes `.`.
- Thin CLI parsing should fail with a direct `requires a value` usage error when
  a required string flag is followed by another long flag token, rather than
  consuming that next flag as data and producing a downstream parse error.
- Mixed standalone RuneContext and RuneCode teams must remain portable: the
  repository carries all correctness-critical state, while RuneCode may add
  richer audit evidence on top without becoming required for collaboration.

### Recommended Branch Cut 1: Change identity, lifecycle, stable-path history, and traceability core

- [x] Issue: implement year-scoped change ID allocation with monotonic counter
  plus collision-resistant suffix.
- [x] Issue: implement lifecycle state validation for `proposed`, `planned`,
  `implemented`, `verified`, `closed`, and `superseded`.
- [x] Issue: define terminal-state invariants so `superseded` is distinct from
  `closed`, requires successor references, and still records terminal metadata
  such as `closed_at` consistently.
- [x] Issue: implement bidirectional supersession consistency checks.
- [x] Issue: implement close behavior that updates state without moving change
  folders into an archive tree.
- [x] Issue: define concurrent-open-change behavior so multiple non-closed
  changes can coexist without requiring one global active change for the whole
  repository.
- [x] Issue: implement artifact-level traceability fields connecting changes,
  specs, and decisions.
- [x] Issue: validate that `depends_on`, `informed_by`, `related_changes`,
  `related_specs`, and `related_decisions` resolve to real artifacts or report
  clear diagnostics.
- [x] Issue: define asymmetric graph semantics so `depends_on` and
  `informed_by` remain directional while `related_changes` stays reciprocal for
  navigation.
- [x] Issue: implement machine-validated deep-link refs in markdown using
  stable `path#heading-fragment` syntax rather than line-number references.
- [x] Issue: implement heading-fragment rename/move rewriting for tool-assisted
  file renames and scoped markdown reference updates.
- [x] Issue: implement split-change helpers and validation rules so umbrella
  changes and sub-changes wire `related_changes` and directional `depends_on`
  links consistently when prerequisite ordering matters.
- [x] Issue: allow split-change `depends_on` edges to reference external
  prerequisite changes while still rejecting self-dependencies and intra-split
  dependency cycles.
- [x] Issue: ensure closed and superseded changes remain directly readable at
  stable paths.
- [x] Issue: define the minimum traceability needed for future lineage/index
  tooling without building that lineage view yet.

### Recommended Branch Cut 2: Standards authoring and migration semantics

- [x] Issue: validate standard frontmatter, including stable `id` path matching.
- [x] Issue: implement `draft`, `active`, and `deprecated` standard-state
  handling.
- [x] Issue: implement `replaced_by` and `aliases` support for migration and
  rename workflows.
- [x] Issue: tighten `replaced_by` to one canonical standards-path reference
  form rather than path-or-id ambiguity.
- [x] Issue: enforce the rule that standards are referenced by path instead of
  copied into change/spec bodies.
- [x] Issue: ensure `suggested_context_bundles` remains advisory metadata only
  and never becomes authoritative bundle membership.

Post-review clarifications:

- Deprecated standards remain directly selectable in applicable change sections
  and bundle selections, but validation/tooling must emit warnings and surface
  `replaced_by` guidance when available.
- Deprecated standards without a successor remain valid in `alpha.3`, but
  validation should emit a warning on the standard so missing migration guidance
  is visible during authoring and review.
- Draft standards fail closed only when directly selected as applicable/added
  standards or bundle members; draft and deprecated standards may still appear
  in `Standards Considered But Excluded` for reviewable migration notes.
- `aliases` are validated as migration metadata in `alpha.3`, but automated
  alias-based rewrite/resolution flows remain deferred to later tooling work;
  authored references must still use canonical standard paths, and no runtime
  alias lookup is performed in this branch cut.
- Path-based standard-reference enforcement in `alpha.3` covers `standards.md`,
  `proposal.md`, and `specs/*.md`; copied standard body text in those authored
  bodies is rejected to keep standards reviewable and non-duplicated, while
  fenced and quoted-fenced examples remain exempt from copied-body detection.
- `standards.md` bullet validation counts canonical standard path spans only, so
  a bullet may still contain other backticked code snippets in its descriptive
  text as long as it names exactly one standard path; any additional
  `standards/...` backticked reference, including non-canonical fragment forms,
  is rejected.
- CLI validation output should preserve enough structured diagnostic context
  (bundle/aspect/rule/pattern/matches/path) to make standards-migration and
  bundle-selection warnings actionable in automation, using RuneContext-root-
  relative paths rather than machine-specific absolute paths.
- Comparable-snippet precomputation for copied-content detection remains a
  deliberate post-`alpha.3` optimization rather than a correctness requirement
  for Branch Cut 2.

### Recommended Branch Cut 3: Progressive disclosure, intent artifacts, standards linkage, and thin change/status commands

- [x] Issue: define the branching rules for `project`, `feature`, `bug`,
  `standard`, and `chore` work so minimum mode versus full mode is chosen
  consistently.
- [x] Issue: define size and risk escalation rules so `small`, `medium`, and
  `large` work items shape correctly.
- [x] Issue: implement `change new` heuristics that recommend or immediately
  trigger full-mode shaping when a requested change appears too large,
  ambiguous, or high-risk for minimum mode.
- [x] Issue: define the deeper intake checklist for new-project work, including
  mission, target users, stack/runtime constraints, deployment/security
  constraints, success criteria, and non-goals.
- [x] Issue: define the bug-workflow escalation rules for unclear root causes,
  security impact, schema impact, API impact, and behavior ambiguity.
- [x] Issue: define the "ask more vs less" heuristics for when RuneContext must
  probe further versus infer defaults from repository conventions.
- [x] Issue: ensure inferred assumptions are recorded in `proposal.md` when
  non-trivial decisions are made without prompting.
- [x] Issue: implement minimum-mode change scaffolding with `status.yaml`,
  `proposal.md`, and `standards.md`.
- [x] Issue: implement shaped change materialization for `design.md` and
  `verification.md` by default, with `tasks.md` and `references.md` created
  only when needed and non-empty.
- [x] Issue: generate and validate `proposal.md` using the required heading
  order and explicit `N/A` rules.
- [x] Issue: generate and validate `status.yaml` fields, including type, size,
  verification status, traceability fields, and promotion assessment state
  scaffolding.
  - Optional string fields that are absent stay omitted on rewrite rather than
    being serialized as placeholder values such as `<nil>`.
  - Empty promotion-assessment maps preserve the default `pending` status on
    rewrite rather than serializing placeholder values or invalid enum entries.
  - YAML rewrite helpers propagate encoder close failures instead of ignoring
    them after a successful encode step.
- [x] Issue: populate and refresh `standards.md` during change creation and
  shaping.
- [x] Issue: enforce reviewable diffs for any automatic `standards.md` refresh.
- [x] Issue: implement thin `runectx status` reporting for active, closed, and
  superseded changes only, using a documented narrow machine-readable contract.
- [x] Issue: implement thin `runectx change new` as a narrow wrapper over the
  core change-creation operation.
- [x] Issue: implement thin `runectx change shape` as a narrow wrapper over the
  core shaping operation.
- [x] Issue: implement thin `runectx change close` as a narrow wrapper over the
  core close operation.

### Recommended Branch Cut 4: Rewrite-heavy edge cases and late alpha.3 polish

- [x] Issue: implement merge-time change-ID collision detection, reallocation,
  and atomic local-reference rewriting for the rare case where branches still
  collide.
  - Post-review hardening: reject terminal or externally referenced changes,
    stage outside the live `changes/` tree, reject symlinked change artifacts,
    keep reallocation rewrites scoped to local change-path references, preserve
    unchanged markdown bytes, preserve original line endings on successful
    rewrites, keep rewrite token boundaries UTF-8-safe, make close/create
    failure paths roll back or clean up instead of leaving partial state behind,
    preserve file permissions across transactional rewrites, reject symlinked
    reallocate rename roots before directory swaps, use a Windows-safe fallback
    when atomic file replacement targets already exist, and surface
    backup-cleanup as a warning rather than an ambiguous command failure.

### Cross-Cutting Workflow Tests and Fixtures

- [x] Issue: add unit tests for change ID allocation, lifecycle transitions,
  supersession consistency, and collision reallocation behavior.
- [x] Issue: add tests for terminal-state invariants, reciprocal
  `related_changes`, directional `depends_on`/`informed_by` semantics, and
  heading-fragment ref validation.
- [x] Issue: add parser/validator tests for `proposal.md`, `standards.md`, and
  deep-link markdown reference contracts.
- [x] Issue: add golden fixtures for minimum-mode, shaped, supplemental-doc,
  closed, and superseded change folders.
- [x] Issue: add tests for dangling cross-artifact references, heading-fragment
  rewrite behavior, and standards-maintenance review-diff behavior.
- [x] Issue: add tests ensuring fenced code examples do not validate or rewrite
  as live deep refs, and that absolute or traversal-style markdown deep-ref
  paths fail closed.

### Exit Criteria

- Every substantive work item can start in minimum mode and deepen only when
  needed.
- Shaped changes default to `design.md` and `verification.md`, while
  `tasks.md` and `references.md` appear only when they carry real content.
- Large or high-risk work is pushed toward full mode early enough that minimum
  mode does not become a hiding place for under-shaped work.
- `proposal.md` is the canonical reviewable intent artifact.
- `standards.md` is always present and reviewably maintained.
- Change history stays accessible at stable paths.
- `superseded` is validated as a terminal successor state distinct from
  `closed`.
- Multiple non-closed changes can coexist, and status/reporting flows do not
  depend on a single repository-wide active-change slot.
- Large features can be represented as an umbrella change plus linked
  sub-changes with consistent `related_changes` and prerequisite `depends_on`
  edges.
- Machine-readable graph links stay artifact-level, while markdown deep links
  can target stable heading fragments without relying on line numbers.
- Thin `status`, `change new`, `change shape`, and `change close` commands exist
  as wrappers over the same core operations later CLI work will broaden.
- Minimum/full-mode branching and prompting heuristics are both tested.

### RuneCode Companion-Track Checkpoints

- RuneCode can bind active change IDs plus proposal sections into audit-history
  fixtures.
- RuneCode can consume the same artifact-level and heading-fragment references
  that standalone RuneContext validates.
- RuneCode can generate reviewable `standards.md` updates rather than silent
  mutations.
- RuneCode can consume change close outputs as inputs to future promotion flows.

## `v0.1.0-alpha.4` - Deterministic Context Packs, Promotion, And Indexes

Primary outcome: generate deterministic resolved outputs and supporting indexes
that make RuneContext consumable by power users, automation, and future
RuneCode integration.

### Implementation Notes

- Persisted context-pack fields should use portable path and identity forms;
  host-specific absolute paths are acceptable for local diagnostics but must not
  become part of the canonical generated pack representation.
- `generated_at` should remain a required emitted context-pack field for human
  auditability, but it must stay outside the canonical `pack_hash` input so
  regenerating the same resolved content at a different time does not change the
  hash.
- Core context-pack builders should require an explicit `generated_at` input
  rather than silently defaulting to wall-clock time; if a CLI wants a default,
  that policy should live at the command boundary instead of the canonical pack
  engine.
- Core context-pack builders should also reject sub-second `generated_at`
  precision rather than silently truncating it so the timestamp contract stays
  explicit and reviewable.
- Alpha.4 should refine the persisted context-pack provenance shape so selected
  and excluded entries retain enough selector detail for explanation and future
  Verified receipts: `bundle`, `aspect`, `rule`, `pattern`, and `kind`.
- Context-pack request identity should use a hybrid model: the normal authored
  workflow still prefers one top-level bundle or an authored composite bundle,
  while the generated pack contract should preserve ordered
  `requested_bundle_ids` separately from resolved bundle linearization so
  RuneCode and other future runtimes do not need a schema refactor for ordered
  multi-bundle requests.
- Context-pack semantics must not embed evidence-service endpoints, locator
  metadata, tenancy/auth details, or other deployment-specific runtime routing.
  Those concerns belong to RuneCode-owned integration metadata rather than
  RuneContext core format meaning.
- Context packs are generated portable artifacts and should usually remain
  on-demand or ephemeral; future runtime systems may bind to `pack_hash`
  without requiring context packs themselves or high-frequency runtime evidence
  dumps to live in git.
- Alpha.4 context packs should advertise an explicit restricted canonicalization
  token rather than claiming full RFC 8785 JCS interoperability; the emitted
  pack profile should stay narrow, deterministic, and well-tested for the actual
  value shapes RuneContext writes.
- That restricted canonicalization profile should still carry dedicated tests for
  key ordering, control-character escaping, Unicode preservation, and
  HTML-sensitive characters such as `<`, `>`, and `&`.
- The same profile should also fail closed on invalid UTF-8 string content
  instead of silently normalizing it during canonicalization.
- Selected-file hashing should normalize text line endings before hashing so LF
  and CRLF checkouts of the same logical content still yield the same
  deterministic pack output across clean machines and operating systems.
- Portable path-source `source_ref` values should reject absolute, UNC,
  drive-qualified, and traversal-like path forms so persisted packs keep a clear
  cross-machine contract.
- Generated context-pack bundle identifiers should fail closed against the same
  lowercase-alphanumeric-plus-hyphen ID grammar used by authored bundle
  contracts, rather than accepting arbitrary non-whitespace strings.
- Generated indexes should standardize on `runecontext/manifest.yaml`,
  `runecontext/indexes/changes-by-status.yaml`, and
  `runecontext/indexes/bundles.yaml`, each using a closed schema, stable
  ordering, and merge-friendly formatting while remaining optional and
  regenerable.
- Alpha.4 close-time promotion assessment should deterministically record only
  `none` or `suggested`; `accepted` and `completed` remain explicit later
  workflow transitions rather than implied close outcomes.

### Recommended Branch Cut 1: Context-pack engine and determinism fixtures

- [x] Issue: align the machine-readable context-pack contract artifacts first,
  including `schemas/context-pack.schema.json`, related profile docs, and
  shipped fixtures, so the alpha.4 implementation starts from a schema/fixture
  contract that already reflects `requested_bundle_ids`, persisted selector
  `pattern`/`kind`, and the `generated_at` versus canonical-hash rule.
- [x] Issue: implement selected and excluded file inventories with per-file
  hashes.
- [x] Issue: implement compact deterministic provenance for selected and
  excluded files, persisting `bundle`, `aspect`, `rule`, `pattern`, and `kind`
  while preserving a clean extension path for fuller provenance receipts in
  Verified mode.
- [x] Issue: implement hybrid pack-request identity so generated packs can carry
  ordered `requested_bundle_ids` separately from resolved bundle linearization,
  while normal authored workflows still prefer one top-level bundle or an
  authored composite bundle.
- [x] Issue: implement source metadata capture inside the context pack,
  including resolved commit and signed-tag verification posture.
- [x] Issue: implement required `generated_at` emission together with top-level
  pack hashing over the canonicalized resolved pack, excluding both
  `pack_hash` and `generated_at` from the hash input.
- [x] Issue: harden canonical pack hashing with RFC 8785-compatible string and
  key-order behavior for the emitted pack shapes, or else lock the pack schema
  to an explicit RuneContext-owned canonicalization token with matching tests
  and documentation.
- [x] Issue: make the restricted canonicalization profile reject invalid UTF-8
  string content explicitly so pack hashing never silently rewrites malformed
  machine-readable values during canonicalization.
- [x] Issue: normalize text line endings before per-file hashing so deterministic
  pack output survives LF/CRLF checkout differences.
- [x] Issue: reject sub-second `generated_at` inputs and non-portable local
  `source_ref` traversal forms at the core pack-builder boundary.
- [x] Issue: align generated context-pack bundle-ID validation with the authored
  bundle-ID grammar, and keep reject fixtures specific enough that unknown-field
  failures are not masked by unrelated missing required fields.
- [x] Issue: implement stable ordering rules for all generated pack content.
- [x] Issue: add golden fixtures for resolved context packs, selected/excluded
  provenance, and top-level pack hashes.
- [x] Issue: add clean-machine parity tests showing that CLI and library pack
  generation stay deterministic without relying on host caches, home-directory
  state, or other hidden local metadata.
- [x] Issue: add negative tests for invalid bundle requests, non-portable local
  source references, missing selector provenance, and hashing failures so pack
  errors stay explicit and reviewable.

### Recommended Branch Cut 2: Pack explanation, thresholds, and fail-closed rebuild behavior

- [x] Issue: implement human-readable and machine-readable pack output modes.
- [x] Issue: give machine-readable pack reports an explicit schema version and
  standalone schema contract so RuneCode and other automation consumers can
  validate them independently of the embedded pack schema.
- [x] Issue: document that the report schema validates the envelope while the
  embedded `pack` payload still requires separate validation against
  `context-pack.schema.json` when consumers need full contract enforcement.
- [x] Issue: keep report warning counters (`value`, `threshold`) non-negative at
  schema level so machine-output contracts fail closed on impossible advisory
  payloads.
- [x] Issue: implement `--explain`-style provenance output for include/exclude
  decisions using the persisted selector detail from Branch Cut 1.
- [x] Issue: compare rebuild stability using canonicalized explanation content
  rather than brittle in-memory deep-struct comparison.
- [x] Issue: implement advisory warnings using the design defaults of `256`
  selected files, `1 MiB` referenced content bytes, and `256 KiB` provenance
  metadata.
- [x] Issue: document and test the threshold API semantics so a fully zero-valued
  threshold struct means "use defaults", explicit zero fields remain valid once
  any field is set, and negative values opt back into per-field defaults.
- [x] Issue: expose advisory-threshold defaults through copy-returning APIs or
  equivalent immutable contracts rather than mutable exported global structs.
- [x] Issue: keep pack-only construction and enriched report construction as
  separable flows even when they share the same rebuild/fail-closed engine.
- [x] Issue: implement fail/rebuild behavior when files change between
  enumeration, hashing, and delivery preparation.
- [x] Issue: propagate non-transient digest/read failures during rebuild
  stability checks instead of collapsing them into a generic "inputs changed"
  retry outcome.
- [x] Issue: document that rebuild stability is evaluated against the loaded
  project snapshot and selected-file content; bundle-definition reloads are not
  performed mid-build unless a later milestone explicitly broadens that scope.
- [x] Issue: add regression tests for advisory-size and provenance-threshold
  warnings using the documented default values.
- [x] Issue: add regression tests for unsupported canonical scalar types and
  machine-report schema validation so the restricted hashing/output contracts
  fail closed when future type drift appears.
- [x] Issue: add tests for changed-file fail-closed behavior between
  enumeration, hashing, and delivery preparation.
- [x] Issue: make test read-hook helpers nil-safe so test-only hook misuse does
  not panic the pack/report build path.

### Recommended Branch Cut 3: Promotion assessment on close

- [x] Issue: build promotion suggestion behavior on top of the alpha.3
  `promotion_assessment` structure already present in `status.yaml`.
- [x] Issue: on `change close`, deterministically record
  `promotion_assessment.status` as either `none` or `suggested`; keep
  `accepted` and `completed` available for later explicit promotion workflows.
- [x] Issue: implement reviewable suggested promotion targets for `specs/`,
  `standards/`, and `decisions/`.
- [x] Issue: implement explicit "no promotion needed" recording on close.
- [x] Issue: preserve explicitly advanced promotion lifecycle states
  (`accepted`, `completed`) if they already exist, so close-time reassessment
  does not regress later explicit promotion workflow outcomes.
- [x] Issue: lock in deterministic promotion-assessment behavior for both
  `closed` and `superseded` terminal lifecycle outcomes.
- [x] Issue: ensure suggested promotion target paths are emitted from normalized
  traceability references so close-time output stays slash-canonical and
  deterministic across platforms.
- [x] Issue: add tests for close-time promotion assessment determinism,
  explicit `none` outcomes, and stable suggested-target formatting.

### Recommended Branch Cut 4: Generated indexes and manifests

- [x] Issue: implement overall `manifest.yaml` generation at
  `runecontext/manifest.yaml`.
- [x] Issue: implement generated change indexes grouped by lifecycle state at
  `runecontext/indexes/changes-by-status.yaml`.
- [x] Issue: implement generated bundle inventory views showing parents and
  referenced patterns at `runecontext/indexes/bundles.yaml`.
- [x] Issue: define closed schemas for `manifest.yaml` and the generated index
  artifacts so standalone tooling and RuneCode can validate them without
  treating them as source of truth.
- [x] Issue: ensure generated indexes use stable ordering and merge-friendly
  formatting.
- [x] Issue: add fixtures for generated manifest, change-index, and bundle-index
  stability.
- [x] Issue: harden generated index builders to fail closed when artifact paths
  escape the RuneContext content root or when a change carries an unsupported
  lifecycle status.
- [x] Issue: tighten manifest and bundle-index path patterns to reject
  traversal, hidden, and empty path segments so external tooling can
  validate generated artifacts against a fail-closed path contract.
- [x] Issue: keep the generated `changes-by-status` schema aligned with
  `change-status.schema.json` so the lifecycle-typed `x-` prefix convention
  stays deterministic across schema and generator contracts.
- [x] Issue: persist the generation-order guarantee where `changes-by-status`
  and `bundles` indexes land before `manifest.yaml`, ensuring partial failures
  cannot leave the manifest pointing at missing indexes.

### Exit Criteria

- Any bundle selection can be flattened into a deterministic context pack.
- The context pack contains the top-level canonical hash required for future
  audit binding.
- The top-level `pack_hash` remains stable across regenerations of identical
  resolved content even when `generated_at` differs.
- Persisted context-pack artifacts remain portable and do not require host-
  specific absolute paths or deployment-specific evidence-service metadata to
  interpret correctly.
- Persisted context-pack provenance remains specific enough to explain
  include/exclude decisions without needing full Verified receipts.
- Promotion assessment is structured and reviewable.
- Generated indexes aid browsing without becoming the source of truth, and they
  remain optional/regenerable at the documented standard paths.
- Deterministic outputs are protected by golden tests rather than manual spot
  checking.

### RuneCode Companion-Track Checkpoints

- RuneCode can test direct-resolver versus CLI parity using shared context-pack
  fixtures and expected pack hashes.
- RuneCode can verify that the same resolved content produces the same
  `pack_hash` even when emitted `generated_at` timestamps differ.
- RuneCode can distinguish caller-requested bundle order from resolved bundle
  linearization when ingesting context packs.
- RuneCode can draft typed isolate-delivery descriptor fixtures from the pack's
  resolved metadata and hashes.
- RuneCode can ingest optional manifest and generated index artifacts from fixed
  portable paths when they are present.
- RuneCode can verify that over-limit context packs fail loudly rather than
  being silently truncated.

## `v0.1.0-alpha.5` - Minimal CLI And Machine-Facing Operations

Primary outcome: expose the small universal command surface needed for
automation, CI, debugging, non-agent workflows, and broader post-alpha.4
dogfooding across this repository and other repositories.

### Implementation Notes

- `alpha.3` may already expose thin `status` and change write wrappers. This
  milestone broadens those commands into the stable universal CLI contract
  rather than redefining their semantics.
- Alpha.5 should lock clear command boundaries so later work does not need to
  refactor the CLI surface: `status` is workflow summary, `validate` is
  authoritative contract enforcement, and `doctor` is environment/install/
  source-posture diagnosis.
- Alpha.5 should standardize one shared machine-facing JSON envelope and
  failure taxonomy across commands. Earlier line-oriented key/value output can
  remain as a documented historical thin-contract phase, but broader commands
  should converge on the same structured contract instead of inventing per-
  command payload shapes.
- Write-command `--dry-run` behavior should simulate the planned mutations and
  validate the resulting would-be project state rather than emitting prose-only
  intent.
- `runectx init` should land here as the repo-local, local-first scaffolding
  and command-UX front door for embedded and linked workflows, while alpha.8
  keeps the release/install hardening, network-policy enforcement, and end-to-
  end reference-fixture coverage for network-enabled install/update behavior.
- `runectx promote` should be the only durable-mutation surface for promotion
  state. Close-time assessment still settles to `none` or `suggested`; explicit
  promote workflows own transitions to `accepted` and `completed`.
- `runectx standard discover` should remain advisory-only. Interactive runs may
  offer a confirmed handoff into `runectx promote`, but `--non-interactive`
  must emit reusable candidate data and exit without mutation. That handoff
  should use explicit candidate data rather than hidden session state.
- Verified-mode enablement and backfill command surfaces move with the
  underlying assurance implementation in `alpha.6`; alpha.5 should not block
  aggressive Plain-mode dogfooding on those later assurance artifacts.

### Epic 1: Recommended Branch Cut 1 / Best Combined Branch

- [x] Issue: define stable exit codes, failure classes, and the shared
  machine-facing JSON envelope for automation.
- [x] Issue: implement `--json` output contracts across machine-facing
  commands.
- [x] Issue: implement `--non-interactive` behavior with clear prompt,
  inference, and failure rules.
- [x] Issue: implement `--dry-run` behavior for write operations by simulating
  planned mutations and validating the resulting would-be project state.
- [x] Issue: implement `--explain` output for resolution, standards selection,
  and promotion suggestions.
- [x] Issue: broaden `runectx status` from its alpha.3 narrow status-reporting
  contract into the stable CLI surface.
- [x] Issue: broaden `runectx change new` from its alpha.3 thin wrapper into the
  stable CLI surface.
- [x] Issue: broaden `runectx change shape` from its alpha.3 thin wrapper into
  the stable CLI surface.
- [x] Issue: broaden `runectx change close` from its alpha.3 thin wrapper into
  the stable CLI surface.
- [x] Issue: broaden `runectx validate` from the earlier narrow contract into
  the stable CLI surface.
- [x] Issue: build CLI-versus-library parity fixtures for the broadened command
  set.
- [x] Issue: ensure all write commands surface reviewable diffs or proposed
  mutations rather than silent commits.
- [x] Issue: add integration coverage for the broadened thin commands.
- [x] Issue: add snapshot or golden tests for shared `--json` outputs.
- [x] Issue: add behavior tests for `--non-interactive`, `--dry-run`, and
  `--explain`.
- [x] Issue: add tests for failure classes, diagnostics, and exit-code
  stability.

Implementation note: `--explain` is currently accepted and machine-visible for
`status`, `validate`, and `change*` commands, but those commands emit an
explicit `explain_warning` field while richer explanation payloads remain
pending for later alpha.5 work.

Implementation note: alpha.5 `--dry-run` now clones from the resolved
project root (not only the invocation directory), enforces clone safety/size
limits, and fails closed on absolute symlinks or relative symlinks that resolve
outside the selected project root.

### Epic 2: Recommended Branch Cut 2 / Read-Only Admin And Resolution Commands

- [x] Issue: implement `runectx bundle resolve` on top of the existing
  resolution/reporting core.
- [x] Issue: implement `runectx doctor` with a clearly separate environment,
  install, and source-posture diagnostic contract.
- [x] Issue: add integration tests for `bundle resolve` and `doctor`, including
  `--json` and `--explain` behavior where applicable.

### Epic 3: Recommended Branch Cut 3 / Local Init Workflow

- [x] Issue: implement repo-local, local-first `runectx init` scaffolding for
    embedded and linked workflows.
- [x] Issue: ensure alpha.5 `runectx init` does not depend on implicit network
    fetches; network-enabled install/upgrade hardening remains in
    `v0.1.0-alpha.8`.
- [x] Issue: add integration tests covering embedded and linked local init flows
    plus `--dry-run`, `--json`, and `--non-interactive` behavior.
- Note: init tests now cover embedded and linked scaffolding, machine-facing
  flags (`--dry-run`, `--json`, `--non-interactive`), plan reporting, and seed
  bundle validation while keeping the workflow local-first and network-free.
- [ ] Note: `runectx upgrade` is intentionally deferred to `v0.1.0-alpha.8`
  alongside release/install hardening.

### Epic 4: Recommended Branch Cut 4 / Explicit Promotion Workflow

- [ ] Issue: implement `runectx promote` as the only CLI surface that writes
  durable promotion mutations.
- [ ] Issue: define explicit `runectx promote` state transitions from
  `suggested` to `accepted` and `completed`.
- [ ] Issue: ensure promotion mutations remain reviewable and machine-readable
  rather than hidden behind implicit workflow state.
- [ ] Issue: add integration tests for explicit promotion flows, promotion
  failure classes, and reviewable output contracts.

### Epic 5: Recommended Branch Cut 5 / Advisory Standards Discovery

- [ ] Issue: implement advisory-only `runectx standard discover` candidate
  output.
- [ ] Issue: allow interactive `runectx standard discover` runs to hand off to
  `runectx promote` only after explicit user confirmation.
- [ ] Issue: ensure `runectx standard discover --non-interactive` emits reusable
  candidate data and exits without mutation.
- [ ] Issue: add integration tests for advisory discovery, interactive handoff,
  and non-interactive no-mutation behavior.

### Exit Criteria

- Power users can manage RuneContext entirely through the CLI.
- Automation and CI can consume structured command outputs.
- CLI semantics stay aligned with the canonical file model rather than becoming
  a competing source of truth.
- The repo can be initialized, resolved, promoted, and maintained through the
  same stable CLI surface used for day-to-day dogfooding in Plain mode.
- CLI behavior is protected by integration tests and machine-readable golden
  outputs.

### RuneCode Companion-Track Checkpoints

- RuneCode can run parity suites between its future direct integration and the
  CLI.
- RuneCode can consume JSON status, resolve, and close outputs in integration
  tests.
- RuneCode can validate non-interactive behavior for remote/server workflows.

## `v0.1.0-alpha.6` - Assurance Tiers And Verifiable Tracing

Primary outcome: support both low-friction standalone use and stronger
verifiable tracing, while keeping assurance progressive rather than mandatory.

### Implementation Notes

- `plain` and `verified` should share one authored workflow and one repository
  source model. Verified mode adds portable evidence requirements rather than an
  alternate authoring path.
- Standalone `runectx` must be able to generate the same minimal portable
  Verified receipts that mixed-team collaboration depends on.
- RuneCode may add richer parallel audit evidence, but that evidence must remain
  additive and must not become hidden required state for correctness.
- Verified commit policy should preserve RuneContext's low-noise posture:
  baselines and minimal portable receipts may be committed when needed, but
  high-frequency runtime evidence must stay outside the portable RuneContext
  tree.
- Deployment-specific evidence discovery and service-routing metadata must stay
  outside RuneContext core semantics even when RuneCode later consumes portable
  baselines and receipts.
- The broadened alpha.5 CLI remains the day-to-day surface for these workflows;
  alpha.6 should add assurance evidence and enablement on top of that same
  command surface rather than introducing an alternate source of truth or a
  parallel authoring model.
- Assurance receipts for durable knowledge promotion should attach to the
  explicit `runectx promote` workflow introduced in alpha.5 rather than
  inventing an assurance-only promotion mutation surface.

### Epic 1: Assurance-tier model

- [ ] Issue: implement persisted `plain` versus `verified` tier behavior.
- [ ] Issue: implement generated baseline artifact shape and baseline creation.
- [ ] Issue: implement receipt schemas and file conventions for context packs,
  changes, promotions, and verifications.
- [ ] Issue: implement receipt hashing, receipt IDs, and collision-resistant
  filenames.

### Epic 2: Verified enablement flow

- [ ] Issue: implement `runectx assurance enable verified`.
- [ ] Issue: implement the Verified enablement flow from adoption commit through
  baseline generation.
- [ ] Issue: record initial resolved source posture and adoption metadata.
- [ ] Issue: implement receipt generation triggers for future verified
  operations.
- [ ] Issue: ensure standalone `runectx` can generate the minimal portable
  receipt set required by a Verified repository without depending on RuneCode.
- [ ] Issue: implement commit-policy guidance for what is committed, ignored, or
  treated as ephemeral in Verified mode.

### Epic 3: Backfill and historical provenance

- [ ] Issue: implement `runectx assurance backfill`.
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
- [ ] Issue: add clean-machine and no-hidden-state tests showing that portable
  assurance artifacts do not depend on host caches, service availability, or
  deployment-specific local metadata for correctness.
- [ ] Issue: add fixtures RuneCode can reuse to test audited-workflow gating and
  provenance ingestion.

### Exit Criteria

- Plain mode remains lightweight and usable without extra receipt generation.
- Verified mode can generate baseline and receipt artifacts for future audit
  consumption.
- Verified repositories remain fully usable by mixed standalone RuneContext and
  RuneCode teams through the same portable receipt model.
- Verified repositories preserve a low-noise commit policy and do not require
  deployment-specific evidence-service metadata or external service availability
  for correctness.
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

## `v0.1.0-alpha.7` - Adapters And Adapter-Pack UX

Primary outcome: make RuneContext comfortable to use inside multiple coding
tools while preserving one core model.

### Implementation Notes

- Adapters should preserve the alpha.5 split between advisory `standard
  discover` output and explicit confirmed `promote` mutations rather than
  inventing hidden tool-specific auto-promotion behavior.
- Rich completion and suggestion UX should derive from the stable alpha.5 CLI
  contract so completions are never a second command-definition source of truth.
- Alpha.7 should target shell completion for Bash, Zsh, and Fish first.
  PowerShell and Windows command-prompt completion are deferred until after the
  MVP.
- Repo-aware suggestions must stay read-only, honor nearest-root discovery and
  explicit `--path`, and degrade gracefully outside a RuneContext project.
- Adapter-native suggestion UX should reuse the same underlying completion
  metadata/providers as shell completion rather than inventing adapter-only
  command semantics.

### Epic 1: Canonical operations reference

- [ ] Issue: author the canonical in-project operations reference under
  `runecontext/operations/`.
- [ ] Issue: define adapter-to-core operation mapping rules.
- [ ] Issue: define how adapters consume or derive from the canonical
  operations reference without redefining semantics.
- [ ] Issue: define a canonical completion metadata model derived from the
  stable CLI command, flag, and value contracts.
- [ ] Issue: define the adapter-pack rule that edits to authoritative
  RuneContext files must automatically trigger `runectx validate` before the
  tool considers the workflow step complete.

### Epic 2: Generic adapter

- [ ] Issue: author the `generic` adapter pack with plain markdown workflow
  docs.
- [ ] Issue: provide example flows for manual, CLI-assisted, and non-agent use.
- [ ] Issue: document completion and suggestion affordances for generic shell-
  based workflows.

### Epic 3: Tool-specific adapters

- [ ] Issue: author the `claude-code` adapter pack.
- [ ] Issue: author the `opencode` adapter pack.
- [ ] Issue: author the `codex` adapter pack.
- [ ] Issue: define compatibility-mode guidance for hosts with weaker
  interaction capabilities.
- [ ] Issue: add tool-native suggestion/autocomplete integrations that reuse the
  canonical completion metadata for hosts that support richer UX.
- [ ] Issue: add tool-native automation/skills that run `runectx validate`
  after edits to authoritative RuneContext files and surface failures
  immediately.

### Epic 4: Completion And Suggestion UX

- [ ] Issue: implement `runectx completion <bash|zsh|fish>` generation.
- [ ] Issue: support static command, subcommand, and flag completion from the
  canonical CLI contract.
- [ ] Issue: support enum/value completion for stable machine-facing and
  workflow flags.
- [ ] Issue: implement repo-aware dynamic suggestions for change IDs, bundle IDs,
  promotion target paths, and adapter names where applicable.
- [ ] Issue: ensure completion and suggestion flows never mutate project state
  and fail soft outside RuneContext repositories.

### Epic 5: Adapter packaging and sync

- [ ] Issue: implement adapter packaging for release artifacts as packs bundled
  with the selected RuneContext release.
- [ ] Issue: implement the `runectx adapter sync <tool>` command as the
  adapter-management CLI surface for local adapter materialization from the
  installed or pinned RuneContext release.
- [ ] Issue: define merge-aware adapter sync/update behavior, including managed
  file boundaries, reviewable diffs, and explicit local config updates where
  required.
- [ ] Issue: ensure adapter sync never fetches from GitHub or any other network
  source; network access is reserved for explicit `runectx init` and
  `runectx upgrade` flows.
- [ ] Issue: ensure adapters never introduce tool-specific source-of-truth
  files.

### Epic 6: Adapter tests and parity

- [ ] Issue: add smoke tests for the `generic`, `claude-code`, `opencode`, and
  `codex` adapters.
- [ ] Issue: add parity checks showing adapter flows map back to the same core
  operations and expected file mutations.
- [ ] Issue: add golden tests for generated Bash, Zsh, and Fish completion
  scripts.
- [ ] Issue: add parity tests proving completion metadata stays aligned with the
  actual command and flag surface.
- [ ] Issue: add fixture tests for repo-aware suggestions across embedded,
  linked, and monorepo projects.
- [ ] Issue: add tests ensuring adapters do not introduce hidden state or
  adapter-only correctness requirements.
- [ ] Issue: add tests ensuring adapter-driven edits to authoritative
  RuneContext files automatically invoke `runectx validate` while unrelated code
  edits do not trigger unnecessary validation.

### Exit Criteria

- At least one tool-specific adapter is usable end to end.
- All adapters map back to the same underlying operations.
- Users can still work directly with repo files and CLI without any adapter.
- Bash, Zsh, and Fish users can install shell completion for the stable CLI
  surface.
- Repo-aware suggestions help users discover valid change IDs, bundles,
  promotion targets, and adapter names without mutating project state.
- Adapter sync materializes the selected tool pack from the installed release
  without requiring network access.
- Adapter behavior is covered by parity and smoke tests rather than manual
  walkthroughs only.

### RuneCode Companion-Track Checkpoints

- RuneCode can use adapter docs as workflow fixtures for change, standards, and
  promotion review flows.
- RuneCode can verify that adapter UX does not smuggle in RuneCode-only hidden
  state or permissions.
- RuneCode can consume the same completion metadata or equivalent providers for
  richer in-tool suggestion UX without redefining command semantics.

## `v0.1.0-alpha.8` - Release, Install, Upgrade, And End-To-End Hardening

Primary outcome: harden RuneContext as a distributable product with tested
install/upgrade paths and end-to-end reference fixtures.

### Implementation Notes

- Several release-foundation pieces landed before the rest of alpha.8 is
  complete: the canonical Nix release-artifact builder, signed and attested
  repo-bundle plus binary publication workflow, release-manifest/checksum/SBOM
  generation, and maintainer/user verification docs. The remaining alpha.8 work
  is primarily upgrade flow, compatibility-matrix, adapter-pack, and
  reference-fixture hardening.
- Use the command name `upgrade` rather than `update`. `runectx upgrade`
  previews a reviewable upgrade plan, and `runectx upgrade apply` is the
  explicit user-authorized mutation step.
- `schema_version` remains the fail-closed parser contract for individual
  machine-readable files, while `runecontext_version` identifies the installed
  RuneContext release for compatibility checks and upgrade planning.
- Upgrade-triggering contract changes must be machine-detectable. If authored
  file structure changes, the relevant file schema or an explicit migration
  marker must change so mixed-version trees fail closed instead of being
  silently reinterpreted.
- `runectx upgrade` must be transactional and reviewable: stage work in
  tool-owned temporary space, run migrators there, validate the staged tree,
  then replace only the targeted live files; any in-flight failure rolls back
  automatically with detailed diagnostics.
- Successful upgrades are rolled back through normal project VCS history rather
  than a hidden RuneContext rollback store or any other second source of truth.
- Source-mode upgrade rules stay narrow: embedded upgrades may rewrite managed
  repo-local files, git upgrades update only the pinned source reference fields
  in `runecontext.yaml`, and `type: path` sources are externally managed and
  must never be mutated by `runectx`.
- Read-only commands (`status`, `validate`, `doctor`, `bundle resolve`, and
  similar surfaces) must never perform hidden upgrades or migrations. `validate`
  and `doctor` should instead detect unsupported version combinations or stale
  pre-upgrade files after merge/rebase and direct users to rerun
  `runectx upgrade`.

### Epic 1: Release packaging

- [ ] Issue: establish CI/CD platform parity with RuneCode and mirror its
  tag-driven release workflow shape:
  - Primary: Linux (x86_64 and arm64) and macOS (x86_64 and arm64) via Nix.
  - Portability: Windows via non-Nix smoke testing.
- [x] Issue: keep `nix build .#release-artifacts` as the canonical unsigned
  release builder; workflow steps may verify, sign, attest, and publish assets
  but must not redefine release contents outside the Nix build graph.
- [ ] Issue: package the schema bundle for releases across supported platforms.
- [ ] Issue: package adapter packs for releases.
- [x] Issue: package optional `runectx` binaries for primary supported platforms:
  `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64`.
- [x] Issue: keep repo bundles as the canonical install and audit path even after
  `runectx` binary archives are added as convenience assets.
- [x] Issue: verify pushed release tags against release metadata and fail closed
  on mismatches before packaging or publication.
- [x] Issue: emit release checksums, release manifest, signatures,
  attestations, SBOM, and release notes.
- [ ] Issue: publish a RuneCode `<->` RuneContext compatibility matrix.
- [x] Issue: publish through a protected `release` environment after unsigned
  assets are built and uploaded by the initial build job.

### Epic 2: Install and upgrade flows

- [ ] Issue: document and test the canonical manual repo-install flow around
  pinned GitHub release assets emitted by the Nix release builder.
- [ ] Issue: harden `runectx init` and implement `runectx upgrade` as the only
  CLI flows allowed to make network calls.
- [ ] Issue: implement preview-only `runectx upgrade` and explicit
  `runectx upgrade apply` mutation as the repo-upgrade command surface.
- [ ] Issue: implement the upgrade planner/migrator registry keyed by
  `runecontext_version`, file-level `schema_version`, and explicit migration
  markers where needed.
- [ ] Issue: implement transactional upgrade staging in tool-owned temporary
  space with validate-before-replace and automatic rollback on in-flight
  failure.
- [ ] Issue: ensure embedded upgrades detect locally modified managed files and
  stop with reviewable conflict guidance rather than overwriting them.
- [ ] Issue: ensure git upgrades mutate only pinned source reference fields in
  `runecontext.yaml` and never rewrite a linked source tree.
- [ ] Issue: ensure `type: path` sources are reported as externally managed and
  never mutated; the CLI should direct users to navigate to the owning source
  path and run the upgrade there.
- [ ] Issue: harden adapter sync/update to be namespaced and merge-aware, with
  normal adapter sync remaining local-only against installed release content.
- [ ] Issue: ensure `validate` and `doctor` report unsupported version
  combinations, stale mixed-version trees after merge/rebase, and integrity
  posture issues.

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
  adapter packs, optional binaries, attestations, and SBOM outputs.
- [ ] Issue: add end-to-end tests for manual repo install, CLI-managed install,
  and preview-first upgrade flows.
- [ ] Issue: add regression tests asserting forbidden install/upgrade patterns
  do
  not appear: required global installs, bash-only installers, overwriting
  existing tool config files, hidden runtime-manager dependencies, and silent
  auto-upgrades.
- [ ] Issue: add regression tests asserting adapter sync remains local-only and
  never performs implicit network fetches.
- [ ] Issue: add regression tests for upgrade transaction rollback,
  merge/rebase stale-file detection, and idempotent reruns of
  `runectx upgrade`.
- [ ] Issue: add end-to-end tests over reference projects for embedded,
  linked-by-commit, linked-by-signed-tag, Verified-mode, and monorepo cases.

### Exit Criteria

- RuneContext can be installed manually, managed by CLI, and upgraded
  reviewably.
- The release workflow mirrors RuneCode's tag-driven build/publish split while
  keeping RuneContext's unsigned asset set canonical in Nix.
- Release artifacts are canonical, inspectable, and compatible with the repo-
  first distribution model.
- Normal adapter sync is local-only; any network access is confined to explicit
  `init` and `upgrade` operations.
- Mixed-version trees fail closed and are repairable through explicit reruns of
  `runectx upgrade` rather than hidden background migration.
- Signed-tag verification is included in MVP validation, not deferred.
- Install, upgrade, and release guarantees are backed by automated end-to-end
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
