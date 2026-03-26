# MVP Acceptance

`v0.1.0` is complete only when every item in this checklist is true for the
`runecontext` repository.

## 1. Core Model And Contracts

- [ ] The canonical on-disk model is documented and stable.
- [ ] The three-layer boundary between core, adapters, and RuneCode integration
  is explicit.
- [ ] `runecontext.yaml`, bundle files, change-status files, context packs,
  spec frontmatter, and decision frontmatter all have versioned schemas.
- [ ] The restricted YAML profile and canonical JSON hashing rules are frozen.
- [ ] Policy neutrality is explicit and tested: RuneContext text does not grant
  capabilities, approvals, or runtime authority.
- [ ] The LLM input trust boundary is explicit and tested: RuneContext text is
  treated as untrusted model input.

## 2. Storage Modes And Resolution

- [ ] Embedded mode works.
- [ ] Linked git mode by pinned commit works.
- [ ] Linked git mode by signed tag works, including trusted-signer validation,
  expected-commit checking, fail-closed mismatch behavior, and explicit trusted-
  signer inputs rather than hidden machine-global trust state.
- [ ] Mutable refs require explicit opt-in and warnings.
- [ ] Local path sources work for developer-local usage and are marked
  unverified/non-auditable.
- [ ] Local path sources are invalid for remote/CI mode unless a higher-level
  trusted wrapper explicitly downgrades the run.
- [ ] Embedded source paths and git subdirectories fail closed if they are
  absolute or escape the selected project/repository root.
- [ ] Embedded source roots are checked after symlink resolution against the
  selected project root so symlinked escapes fail closed.
- [ ] Git source resolution validates URL/ref/commit inputs, rejects option-like
  values, disables interactive prompting, and does not depend on hidden host
  credentials or global Git config for correctness.
- [ ] Git source resolution rejects remote-helper URL forms, constrains allowed
  transport protocols explicitly, and redacts transport secrets from surfaced
  subprocess errors.
- [ ] Git source verification surfaces executable/process failures as actionable
  diagnostics and avoids over-redacting normal git ref/reflog syntax.
- [ ] Signed-tag verification timeouts and explicit trust-input parse failures
  also fail closed with structured, machine-readable diagnostics.
- [ ] Signed-tag verification rejects empty `expect_commit` values with a clear
  validation error before commit-format checks run.
- [ ] Mutable git refs reject obviously invalid ref syntax before subprocess
  execution rather than relying on fetch failures alone.
- [ ] RuneContext does not use environment variables as user-facing
  configuration or semantic inputs; correctness-critical behavior comes from
  repo state, explicit config, or caller-supplied options only.
- [ ] Git source resolution uses explicit process/network timeouts so local and
  CI validation cannot hang indefinitely during fetch/checkout steps.
- [ ] Pinned-commit git resolution works without requiring the remote to support
  direct fetch-by-SHA behavior.
- [ ] Monorepo nearest-root discovery works and reports the selected config path
  as structured metadata.
- [ ] Local path snapshots exclude `.git/` and fail closed when practical
  snapshot depth, file-count, or byte-size limits are exceeded.
- [ ] Bundle traversal fails closed when practical depth or file-count limits are
  exceeded.
- [ ] Validation entrypoints clean up any temporary source materializations after
  successful validation as well as on error.
- [ ] Bundle resolution is deterministic, cycle-safe, depth-limited, and path-
  boundary-safe.
- [ ] Bundle rules, diagnostics, and generated inventories use one consistent
  RuneContext-root-relative path model with aspect-boundary enforcement.
- [ ] Whole-project validation fails closed when specs, decisions, bundle files,
  or other validated artifacts escape their selected subtree through symlinks.
- [ ] Whole-project validation and bundle traversal still accept symlinked root
  directories when the fully resolved target remains in-bounds.
- [ ] Bundle aspect-root containment checks canonicalize the aspect root before
  comparing against resolved bundle matches.

## 3. Change Workflow And Standards

- [x] Every substantive work item gets a stable change ID.
- [x] Stable change IDs remain ASCII-safe even when authored titles contain
  non-ASCII characters.
- [x] Minimum mode works with `status.yaml`, `proposal.md`, and `standards.md`.
- [x] Shaped work defaults to `design.md` and `verification.md`, while
  `tasks.md` and `references.md` remain supplemental files created only when
  they are needed and contain real content.
- [x] `change shape` is additive/idempotent and does not behave like a
  destructive regeneration pass over authored files.
- [x] Large or high-risk work is prompted or inferred toward full mode early,
  and non-interactive shaping rationale remains reviewable.
- [x] Work-type and size branching rules exist for project, feature, bug,
  standard, and chore changes.
- [x] Ask-more versus infer-more heuristics exist and inferred assumptions are
  captured in `proposal.md`.
- [x] `proposal.md` uses the required section order and validation rules.
- [x] `standards.md` is always present and reviewably maintained.
- [x] Standards are referenced by path rather than copied into change/spec
  bodies.
- [x] Standards frontmatter validation, deprecation, and rename/migration rules
  work.
- [x] Standards migration uses one canonical `replaced_by` path-reference form
  rather than mixed path/id ambiguity.
- [x] Deprecated standards may still be referenced directly for compatibility,
  but validation emits warnings and suggests `replaced_by` targets when present.
- [x] Deprecated standards without `replaced_by` remain valid in `alpha.3`, but
  validation emits a warning so missing migration guidance is reviewable.
- [x] Draft or deprecated standards may still appear in `Standards Considered
  But Excluded`, while draft standards fail closed for applicable selections and
  bundle membership.
- [x] `aliases` are validated as migration metadata and collision-checked even
  though automatic alias-driven rewrites and runtime alias lookup remain
  deferred.
- [x] Path-based standards references inside `proposal.md` and `specs/*.md`
  validate both deep-ref and plain backticked `standards/...md` forms.
- [x] Copied-standard-content enforcement ignores fenced and quoted-fenced code
  examples so reviewable excerpts do not trigger false positives.
- [x] `standards.md` bullets may include non-standard backticked code in their
  descriptions, but exactly one canonical standard path is required per bullet,
  and any extra `standards/...` reference is rejected.
- [x] Cross-artifact references in change metadata validate cleanly or produce
  clear diagnostics.
- [x] Standards-related validation and warning diagnostics use RuneContext-root-
  relative paths so CLI output is deterministic across machines.
- [x] Machine-readable traceability stays artifact-level, and human-readable
  markdown can use machine-validated `path#heading-fragment` deep refs without
  relying on brittle line numbers.
- [x] Automatically derived heading fragments remain ASCII-safe and valid for
  machine-readable deep refs even when headings contain non-ASCII text.
- [x] Markdown deep-ref validation and rewrite flows ignore fenced code blocks,
  reject absolute and traversal-style paths, and reject line-number fragments
  such as `#L10`, `#L10-L20`, and `#42`.
- [x] UTF-8-safe deep-ref tokenization does not absorb adjacent non-ASCII prose
  into machine-readable fragment tokens, which remain ASCII-bounded in
  `alpha.3`.
- [x] Quoted fenced-code examples such as blockquote-prefixed fences are also
  ignored by markdown deep-ref validation and rewrite flows.
- [x] External URLs containing `.md#fragment` are not misclassified as local
  RuneContext deep refs.
- [x] Alpha.3 deep refs target machine-indexed markdown under `specs/`,
  `decisions/`, `standards/`, and the top-level markdown files inside each
  `changes/<id>/` directory.
- [x] Alpha.3 machine-addressable markdown headings use ATX `#` headings; Setext
  underlined headings are not part of the guaranteed deep-ref contract yet.
- [x] Multiple non-closed changes can coexist without requiring one global
  active-change slot for the repository.
- [x] Large features can be represented as an umbrella change plus linked
  sub-changes using navigable `related_changes` links and directional
  `depends_on` prerequisites.
- [x] Change-splitting flows and validation preserve consistent `depends_on` /
  `related_changes` wiring when one sub-change must land before others.
- [x] Split-change helpers reject self-dependencies and intra-split dependency
  cycles while still allowing external prerequisite change IDs in `depends_on`.
- [x] `superseded` works as a terminal successor state distinct from `closed`
  and preserves reciprocal supersession links.
- [x] Terminal lifecycle states that represent completed work in alpha.3 do not
  leave `verification_status` at `pending`; both `closed` and `superseded`
  require a completed verification outcome.
- [x] Lifecycle helpers enforce forward-only progression and do not provide a
  built-in reopen/downgrade path in alpha.3.
- [x] Closed changes remain directly accessible at stable paths.
- [x] Rare change-ID reallocation stays fail-closed in alpha.3: terminal or
  externally referenced changes are rejected, only local change-path references
  inside the change are rewritten, unchanged markdown preserves its original
  bytes, successful rewrites keep the original newline style, rewrite token
  boundaries stay UTF-8-safe, and backup cleanup degrades to an explicit
  warning instead of a misleading hard failure after success.
- [x] Failed alpha.3 lifecycle mutations do not leave partial on-disk state:
  rejected closes restore prior status files, failed creates clean up transient
  change directories, mutation paths reject symlinked targets, reallocate also
  rejects symlinked rename roots, successful transactional rewrites preserve the
  original file permissions, and atomic file replacement works even when the
  destination already exists on Windows.
- [x] Missing spec/decision reciprocity diagnostics clearly identify that the
  reciprocal `related_specs` or `related_decisions` entry belongs on the
  referenced change `status.yaml`.

## 4. Context Packs, Promotion, And Indexes

- [ ] Context packs contain deterministic selected/excluded inventories.
- [ ] Context packs record per-file hashes and a top-level canonical pack hash.
- [ ] Context packs keep required `generated_at` metadata while excluding it
  from the canonical `pack_hash` input so identical resolved content hashes the
  same across regenerations.
- [ ] Core context-pack generation requires an explicit `generated_at` input and
  does not silently inject wall-clock timestamps inside the canonical pack
  builder.
- [ ] Core context-pack generation rejects sub-second `generated_at` inputs
  rather than silently truncating them.
- [ ] Context packs record the resolved source revision and verification posture.
- [ ] Context packs retain compact deterministic provenance, including enough
  selector detail to explain include/exclude outcomes, and leave room for fuller
  provenance receipts in Verified mode.
- [ ] Persisted context-pack fields use portable stable identifiers and path
  forms rather than host-specific absolute paths.
- [ ] Path-source `resolved_from.source_ref` values remain portable forward-slash
  relative paths without drive-qualified, UNC, or traversal-style segments.
- [ ] Deterministic context-pack hashes remain stable across LF/CRLF checkout
  differences for the same logical text content.
- [ ] Context packs use an explicit RuneContext-owned canonicalization token for
  their restricted emitted-shape serializer rather than overstating full RFC
  8785 interoperability.
- [ ] Context-pack canonicalization fails closed on invalid UTF-8 strings rather
  than silently replacing malformed bytes during hash preparation.
- [ ] Context packs can preserve ordered multi-bundle requests separately from
  resolved bundle linearization without forcing authored workflows away from one
  top-level bundle or authored composite bundles.
- [ ] Generated context-pack `id` and `requested_bundle_ids` values enforce the
  same fail-closed bundle-ID grammar as authored bundle contracts.
- [ ] Context-pack semantics do not embed deployment-specific evidence-service
  locator, endpoint, auth, or tenancy metadata.
- [ ] Size and provenance threshold warnings exist with the default advisory
  thresholds from the design document.
- [ ] Machine-readable context-pack reports carry an explicit schema version and
  validate against a dedicated report schema.
- [ ] The dedicated report schema clearly documents that the embedded `pack`
  still requires separate context-pack schema validation when consumers need
  full contract enforcement.
- [ ] Report advisory warning fields are constrained as non-negative counters in
  schema contracts (`value`, `threshold`) so impossible warning payloads fail
  validation.
- [ ] Pack-only generation remains available without forcing callers to pay for
  enriched report serialization logic when they only need the pack artifact.
- [ ] Advisory-threshold API semantics are documented and tested for default,
  explicit-zero, and negative-fallback cases.
- [ ] Advisory-threshold defaults are exposed without mutable process-wide
  global state so callers cannot silently rewrite default warning behavior.
- [ ] Fail-closed rebuild checks surface non-transient digest/read errors
  directly instead of collapsing them into a generic changed-input retry.
- [ ] Test-only context-pack read-hook helpers are nil-safe and fall back to the
  real file reader rather than panicking when hooks are unset.
- [ ] Fail-closed rebuild semantics are documented as operating against the
  loaded project snapshot and selected-file content, not hot-reloading bundle
  definitions from disk mid-attempt.
- [ ] Change close deterministically records promotion assessment as `none` or
  `suggested`, while later workflows may advance reviewable promotions to
  `accepted` and `completed`.
- [x] Generated indexes/manifests exist at standard optional paths and remain
  regenerable rather than becoming the source of truth.
- [x] Generated index builders fail closed when emitted artifact paths would
  escape the RuneContext content root or when a change lifecycle status falls
  outside the supported generated-index grouping contract.
- [x] Manifest and bundle-index path patterns reject traversal, hidden, and
  empty segments so external validation fails closed on unsafe paths.

## 5. CLI And Machine Interface

- [ ] The primary CLI commands exist: `init`, `status`, `change new`,
  `change shape`, `bundle resolve`, and `change close`.
- [x] `runectx status` can at minimum report active, closed, and superseded
  changes without requiring a single repository-wide active-change slot.
- [ ] The secondary/admin commands exist: `validate`, `doctor`,
  `standard discover`, `promote`, `assurance enable verified`, and
  `assurance backfill`.
- [ ] `status`, `validate`, and `doctor` have distinct stable responsibilities:
  workflow summary, authoritative contract enforcement, and
  environment/install/source-posture diagnosis.
- [x] Before alpha.5 is complete, any earlier validation entrypoints remain narrow
  wrappers around the same core contracts rather than alternate semantics.
- [x] Before alpha.5 is complete, any earlier `status`, `change new`,
  `change shape`, and `change close` entrypoints remain narrow wrappers around
  the same core operations rather than alternate semantics.
- [x] Before alpha.5 is complete, any earlier validation entrypoints use a
  documented and tested machine-readable output contract.
- [ ] Before alpha.5 is complete, any earlier validation entrypoints that expose
  signed-tag verification accept explicit trust inputs from the caller and
  surface structured signed-tag failure reasons/diagnostics.
- [ ] Early validation entrypoints fail closed with structured diagnostics rather
  than panics when schemas, YAML, markdown contracts, or project references are invalid.
- [ ] Early validation entrypoints honor declared project content roots and the
  full restricted-YAML profile rather than relying on default-path assumptions.
- [ ] Alpha-stage release metadata, module metadata, and parser behavior stay
  consistent with the documented release series and fail-closed contracts.
- [x] The adapter-management command exists: `runectx adapter sync <tool>`.
- [x] `runectx adapter sync <tool>` uses the installed or pinned RuneContext
  release contents rather than implicitly fetching adapter packs from the
  network.
- [x] `runectx adapter sync <tool>` preserves explicit boundaries between
  tool-managed files and user-owned config and does not silently rewrite
  arbitrary host-tool configuration.
- [x] Machine-facing flags exist and behave consistently: `--json`,
  `--non-interactive`, `--dry-run`, and `--explain`.
- [x] `runectx completion <bash|zsh|fish>` exists and stays aligned with the
  stable CLI command/flag/value surface.
- [x] One typed command/operation registry drives operations docs, completion,
  machine-readable completion metadata, and adapter-native suggestion surfaces.
- [x] Machine-facing JSON output uses one documented envelope and failure
  taxonomy across commands rather than drifting command by command.
- [x] Write-command `--dry-run` behavior simulates planned mutations and
  validates the resulting would-be project state without persisting files.
- Note: alpha.5 dry-run is fail-closed for unsafe symlink topology (absolute
  links and relative links resolving outside project root) and enforces
  clone-time resource limits to avoid unbounded local snapshot growth.
- Note: richer `--explain` payloads for resolution/standards/promotion remain
  incremental; current alpha.5 commands surface explicit `explain_warning`
  metadata when `--explain` is accepted but detailed explain output is pending.
- [x] `runectx init` covers repo-local embedded/linked scaffolding in alpha.5,
  while broader release/install hardening and the explicit network-enabled
  `runectx upgrade` flow remain deferred to alpha.8.
- [ ] `runectx standard discover` is advisory-only; interactive runs may chain
  into `promote` only after explicit confirmation, while `--non-interactive`
  discovery emits reusable candidate data and exits without mutation.
- [ ] `runectx promote` is the only CLI surface that advances promotion state to
  `accepted` and `completed` and writes durable target updates.
- [ ] CLI behavior stays aligned with the canonical file model.

## 6. Assurance

- [x] Plain mode works without extra assurance artifacts.
- [x] Verified mode can be enabled and persisted in `runecontext.yaml`.
- [ ] Plain and Verified use the same authored workflow and repository source
  model; Verified adds portable evidence requirements rather than alternate
  source-of-truth files.
- [x] Baseline and receipt families share one stable portable assurance envelope and expose the same flat schema fields for envelope metadata and receipt-specific identifiers
  with explicit artifact kind, stable subject identity, deterministic hashing
  metadata where applicable, and explicit provenance-class distinctions; golden fixtures live under `fixtures/assurance/golden/` to prove the layout.
- [x] Verified mode generates a baseline artifact.
- [x] Verified mode generates receipt families for context packs, changes,
  promotions, and verifications (with verification-family capture scoped to
  verification workflows that emit durable verification events).
- [x] Standalone `runectx` can generate the same minimal portable receipt set
  that a Verified repository requires for mixed-team collaboration.
- [x] `runectx bundle resolve` remains read-only in both tiers; Verified
  context-pack receipts come from an explicit capture flow that emits the pack
  and receipt from the same validated snapshot.
- [ ] RuneCode may attach richer parallel audit evidence without becoming the
  only place correctness-critical assurance state lives.
- [ ] Verified commit policy preserves a low-noise portable tree: baselines and
  minimal portable receipts may be committed, while high-frequency runtime
  evidence stays outside RuneContext's core repository model.
- [x] Verified mutation/capture flows fail closed if required portable receipt
  emission fails.
- [ ] Portable assurance artifacts do not depend on home-directory caches,
  external service availability, or deployment-specific evidence locator
  metadata for correctness.
- [x] `runectx validate` can check assurance artifact schemas plus repo-local
  integrity/linkage semantics without external services, hidden caches, or
  replayed historical commands.
- [x] Receipt filenames are collision-resistant and do not require a shared
  mutable index.
- [x] Backfill can generate imported historical provenance distinct from native
  verified capture, and does so additively without rewriting native
  post-adoption evidence.

## 7. Adapters

- [x] The canonical operations reference exists under `runecontext/operations/`.
- [x] The `generic` adapter exists.
- [x] The `claude-code`, `opencode`, and `codex` adapters exist.
- [x] Adapters differ in UX only, not in core semantics or source-of-truth
  files.
- [x] The `generic` adapter remains thin and documentation-first rather than
  becoming a second source of dynamic runtime behavior.
- [x] Adapter-layer features are implemented as thin UX over explicit core
  operations and stable candidate data rather than adapter-only hidden
  semantics.
- [x] At least one tool-specific adapter supports conversational flows for
  `change new`, `change shape`, `standard discover`, and `promote` while still
  producing reviewable outputs and preserving the same underlying semantics as
  the CLI/core operations.
- [x] When adapters expose discovery scoping or focus inputs, those inputs map
  to explicit underlying operation/CLI contract fields rather than living only
  in prompt text or hidden tool state.
- [x] `runectx standard discover` supports those explicit scope/focus inputs as
  stable underlying operation/CLI-contract fields for adapter-driven targeted
  discovery.
- [x] Adapters and CLI shell completion reuse one canonical completion metadata
  model or equivalent provider surface rather than defining separate command
  semantics.
- [ ] Adapter-driven edits to authoritative RuneContext files automatically run
  `runectx validate` and surface failures before the workflow step is treated as
  complete.
- [ ] Authoritative-file validation triggers are limited to authored RuneContext
  files rather than generated artifacts, adapter-managed files, or unrelated
  repository code.
- [x] Repo-local host-native adapter artifacts are synced as additive outputs for
  supported tools: OpenCode (`.opencode/skills/` and `.opencode/commands/`),
  Claude Code (`.claude/skills/` plus optional `.claude/commands/` shim), and
  Codex (`.agents/skills/` only).
- [x] Synced host-native artifacts use RuneContext-owned naming and ownership
  markers so conflicts with unrelated user-owned files fail closed and future
  uninstall/upgrade flows can target only RuneContext-managed artifacts.
- [x] Host-native ownership remains predictable without a `.runecontext/adapters`
  tracking tree by using stable naming plus strict ownership headers in synced
  host-native artifacts.
- [x] Supported hosts can keep synced prompt bodies minimal and machine-oriented
  through explicit `runectx adapter render-host-native` shell-output injection,
  without introducing adapter-only operation semantics.
- [x] Repo-aware completion/suggestion UX can surface valid change IDs, bundle
  IDs, promotion targets, and adapter names without mutating project state.
- [x] Compatibility mode is explicit and capability-based so weaker hosts lose
  convenience rather than changing RuneContext semantics.
- [ ] Adapter packs are packaged for release.

## 8. Release, Install, And Upgrade

- [ ] GitHub release artifacts exist for the repo-first distribution model.
- [ ] The GitHub release workflow mirrors RuneCode's tag-driven build/publish
  structure, including a protected publish step after unsigned assets are built.
- [ ] `nix build .#release-artifacts` is the canonical unsigned release builder;
  release publication uses those outputs rather than reassembling assets ad hoc
  in workflow YAML.
- [ ] Releases include schema bundle, adapter packs, checksums, release notes,
  compatibility information, signatures, attestations, and an SBOM.
- [ ] Optional `runectx` binaries are packaged.
- [ ] Linux and macOS `runectx` binary archives are published as signed and
  attested convenience assets without replacing the canonical repo bundles.
- [ ] Manual repo install is documented and tested.
- [ ] `runectx init` is local-only and scaffolds from already-installed release
  contents rather than fetching project files over the network.
- [ ] `runectx upgrade` is preview-first, diff-first, and reviewable.
- [ ] `runectx upgrade apply` is the only durable mutation surface for source
  upgrades and migrations.
- [ ] Project upgrade planning classifies state explicitly as `current`,
  `upgradeable`, `unsupported_project_version`, `mixed_or_stale_tree`, or
  `conflicted` before apply is allowed.
- [ ] The upgrade planner/migrator registry is driven by project-level
  `runecontext_version` transitions, with file-level `schema_version` checks and
  explicit migration markers acting as subordinate transform gates.
- [ ] Source upgrades stage work in tool-owned temporary space, validate the
  staged result before replacing live files, and auto-rollback on in-flight
  failure.
- [ ] Successful rollback guidance relies on normal VCS history rather than a
  hidden RuneContext rollback store.
- [ ] Embedded upgrades detect locally modified managed files and stop with
  a reviewable conflict set and fail closed rather than silently overwriting or
  auto-merging user changes.
- [ ] Git upgrades update only pinned source reference fields in
  `runecontext.yaml` and do not rewrite linked source trees.
- [ ] `type: path` sources are treated as externally managed and are never
  mutated in place; the CLI directs users to the owning source path.
- [ ] Adapter sync and re-sync behavior is namespaced and merge-aware.
- [ ] Adapter sync materializes local tool files and config updates from the
  installed release content rather than acting as a remote installer.
- [ ] Read-only commands such as `status`, `validate`, and `doctor` never
  perform hidden upgrades or migrations.
- [ ] `validate` and `doctor` detect unsupported version combinations and stale
  mixed-version trees after merge/rebase and direct users to rerun
  `runectx upgrade`.
- [ ] `doctor` also provides explicit upgrade-readiness diagnostics.
- [ ] `runectx` makes no network calls outside explicit `runectx upgrade`
  flows.
- [ ] Windows MVP support covers portability validation and repo-bundle install
  usability without requiring binary convenience parity.
- [ ] The following anti-patterns are absent: required global installs,
  bash-only installers, overwriting existing `.claude`/`.github` files,
  hidden runtime-manager dependencies, template-only primary distribution,
  implicit adapter-pack fetches during sync, and silent auto-upgrades.

## 9. RuneCode Readiness (Companion Track)

These are not shipped by this repository, but the MVP is not truly ready until
RuneContext makes them possible and testable.

- [ ] RuneCode can version-gate on `runecontext_version`.
- [ ] RuneCode can detect mixed-version/stale-file trees and require an explicit
  RuneContext upgrade before audited workflows continue.
- [ ] RuneCode can resolve the same bundles and packs from the same inputs with
  local/remote parity.
- [ ] RuneCode can bind pack hashes and active change intent into audit history.
- [ ] RuneCode can require Verified mode for normal audited workflows.
- [ ] RuneCode can validate linked signed-tag sources in audited flows.
- [ ] RuneCode can consume receipt and baseline fixtures without asking
  RuneContext to change core semantics.
- [ ] RuneCode can consume the same portable Verified receipts emitted by
  standalone `runectx` while also attaching richer parallel audit evidence.

## 10. Explicitly Not Required For `v0.1.0`

- [ ] `Anchored` assurance tier remains deferred.
- [ ] Rich lineage/index views remain deferred.
- [ ] Package-manager distribution channels remain deferred.
- [ ] Optional stricter pinned-glob mode may remain deferred if the base bundle
  semantics and warnings are complete.
- [ ] Prompt-hygiene or content-safety heuristics remain supplementary and may
  be deferred.

## 11. Test Coverage

- [ ] Unit tests cover schema validation, markdown contracts, source
  resolution, bundle semantics, lifecycle invariants, promotion-state handling,
  and assurance behavior.
- [ ] Unit tests cover git transport hardening, bundle-resolution defensive-copy
  behavior, and symlink escape rejection for both bundle rules and whole-project
  artifact validation.
- [ ] Golden fixtures cover deterministic outputs such as context packs,
  manifests, baselines, receipts, and machine-readable CLI output.
- [x] CLI integration tests cover all primary and secondary commands plus
  `--json`, `--non-interactive`, `--dry-run`, and `--explain` behavior.
- [ ] Adapter smoke tests and parity checks exist for `generic`, `claude-code`,
  `opencode`, and `codex`.
- [ ] Release/install/upgrade flows are covered by end-to-end tests over
  reference projects.
- [ ] Automated tests cover upgrade transaction rollback, stale-file detection
  after merge/rebase, and `type: path` no-mutation behavior.
- [ ] Signed-tag verification, Verified-mode gating, and RuneCode parity
  fixtures are all covered by automated tests.
