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
- [ ] Standards are referenced by path rather than copied into change/spec
  bodies.
- [ ] Standards frontmatter validation, deprecation, and rename/migration rules
  work.
- [ ] Standards migration uses one canonical `replaced_by` path-reference form
  rather than mixed path/id ambiguity.
- [ ] Deprecated standards may still be referenced directly for compatibility,
  but validation emits warnings and suggests `replaced_by` targets when present.
- [ ] Deprecated standards without `replaced_by` remain valid in `alpha.3`, but
  validation emits a warning so missing migration guidance is reviewable.
- [ ] Draft or deprecated standards may still appear in `Standards Considered
  But Excluded`, while draft standards fail closed for applicable selections and
  bundle membership.
- [ ] `aliases` are validated as migration metadata and collision-checked even
  though automatic alias-driven rewrites and runtime alias lookup remain
  deferred.
- [ ] Path-based standards references inside `proposal.md` and `specs/*.md`
  validate both deep-ref and plain backticked `standards/...md` forms.
- [ ] Copied-standard-content enforcement ignores fenced and quoted-fenced code
  examples so reviewable excerpts do not trigger false positives.
- [ ] `standards.md` bullets may include non-standard backticked code in their
  descriptions, but exactly one canonical standard path is required per bullet,
  and any extra `standards/...` reference is rejected.
- [ ] Cross-artifact references in change metadata validate cleanly or produce
  clear diagnostics.
- [ ] Standards-related validation and warning diagnostics use RuneContext-root-
  relative paths so CLI output is deterministic across machines.
- [ ] Machine-readable traceability stays artifact-level, and human-readable
  markdown can use machine-validated `path#heading-fragment` deep refs without
  relying on brittle line numbers.
- [ ] Automatically derived heading fragments remain ASCII-safe and valid for
  machine-readable deep refs even when headings contain non-ASCII text.
- [ ] Markdown deep-ref validation and rewrite flows ignore fenced code blocks,
  reject absolute and traversal-style paths, and reject line-number fragments
  such as `#L10`, `#L10-L20`, and `#42`.
- [ ] Quoted fenced-code examples such as blockquote-prefixed fences are also
  ignored by markdown deep-ref validation and rewrite flows.
- [ ] External URLs containing `.md#fragment` are not misclassified as local
  RuneContext deep refs.
- [ ] Alpha.3 deep refs target machine-indexed markdown under `changes/`,
  `specs/`, `decisions/`, and `standards/`.
- [ ] Alpha.3 machine-addressable markdown headings use ATX `#` headings; Setext
  underlined headings are not part of the guaranteed deep-ref contract yet.
- [x] Multiple non-closed changes can coexist without requiring one global
  active-change slot for the repository.
- [ ] Large features can be represented as an umbrella change plus linked
  sub-changes using navigable `related_changes` links and directional
  `depends_on` prerequisites.
- [ ] Change-splitting flows and validation preserve consistent `depends_on` /
  `related_changes` wiring when one sub-change must land before others.
- [ ] Split-change helpers reject self-dependencies and intra-split dependency
  cycles while still allowing external prerequisite change IDs in `depends_on`.
- [x] `superseded` works as a terminal successor state distinct from `closed`
  and preserves reciprocal supersession links.
- [x] Lifecycle helpers enforce forward-only progression and do not provide a
  built-in reopen/downgrade path in alpha.3.
- [x] Closed changes remain directly accessible at stable paths.

## 4. Context Packs, Promotion, And Indexes

- [ ] Context packs contain deterministic selected/excluded inventories.
- [ ] Context packs record per-file hashes and a top-level canonical pack hash.
- [ ] Context packs record the resolved source revision and verification posture.
- [ ] Context packs retain compact deterministic provenance and leave room for
  fuller provenance receipts in Verified mode.
- [ ] Size and provenance threshold warnings exist with the default advisory
  thresholds from the design document.
- [ ] Promotion assessment is structured and reviewable.
- [ ] Generated indexes/manifests exist but remain optional and regenerable.

## 5. Assurance

- [ ] Plain mode works without extra assurance artifacts.
- [ ] Verified mode can be enabled and persisted in `runecontext.yaml`.
- [ ] Plain and Verified use the same authored workflow and repository source
  model; Verified adds portable evidence requirements rather than alternate
  source-of-truth files.
- [ ] Verified mode generates a baseline artifact.
- [ ] Verified mode generates receipt families for context packs, changes,
  promotions, and verifications.
- [ ] Standalone `runectx` can generate the same minimal portable receipt set
  that a Verified repository requires for mixed-team collaboration.
- [ ] RuneCode may attach richer parallel audit evidence without becoming the
  only place correctness-critical assurance state lives.
- [ ] Receipt filenames are collision-resistant and do not require a shared
  mutable index.
- [ ] Backfill can generate imported historical provenance distinct from native
  verified capture.

## 6. CLI And Machine Interface

- [ ] The primary CLI commands exist: `init`, `status`, `change new`,
  `change shape`, `bundle resolve`, and `change close`.
- [x] `runectx status` can at minimum report active, closed, and superseded
  changes without requiring a single repository-wide active-change slot.
- [ ] The secondary/admin commands exist: `validate`, `doctor`,
  `standard discover`, `promote`, `assurance enable verified`, and
  `assurance backfill`.
- [ ] Before alpha.6 is complete, any earlier validation entrypoints remain narrow
  wrappers around the same core contracts rather than alternate semantics.
- [x] Before alpha.6 is complete, any earlier `status`, `change new`,
  `change shape`, and `change close` entrypoints remain narrow wrappers around
  the same core operations rather than alternate semantics.
- [ ] Before alpha.6 is complete, any earlier validation entrypoints use a
  documented and tested machine-readable output contract.
- [ ] Before alpha.6 is complete, any earlier validation entrypoints that expose
  signed-tag verification accept explicit trust inputs from the caller and
  surface structured signed-tag failure reasons/diagnostics.
- [ ] Early validation entrypoints fail closed with structured diagnostics rather
  than panics when schemas, YAML, markdown contracts, or project references are invalid.
- [ ] Early validation entrypoints honor declared project content roots and the
  full restricted-YAML profile rather than relying on default-path assumptions.
- [ ] Alpha-stage release metadata, module metadata, and parser behavior stay
  consistent with the documented release series and fail-closed contracts.
- [ ] The adapter-management command exists: `runectx adapter sync <tool>`.
- [ ] `runectx adapter sync <tool>` uses the installed or pinned RuneContext
  release contents rather than implicitly fetching adapter packs from the
  network.
- [ ] Machine-facing flags exist and behave consistently: `--json`,
  `--non-interactive`, `--dry-run`, and `--explain`.
- [ ] CLI behavior stays aligned with the canonical file model.

## 7. Adapters

- [ ] The canonical operations reference exists under `runecontext/operations/`.
- [ ] The `generic` adapter exists.
- [ ] The `claude-code`, `opencode`, and `codex` adapters exist.
- [ ] Adapters differ in UX only, not in core semantics or source-of-truth
  files.
- [ ] Adapter-driven edits to authoritative RuneContext files automatically run
  `runectx validate` and surface failures before the workflow step is treated as
  complete.
- [ ] Adapters are packaged for release.

## 8. Release, Install, And Update

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
- [ ] `runectx update` is diff-first and reviewable.
- [ ] Adapter sync/update is namespaced and merge-aware.
- [ ] Adapter sync materializes local tool files and config updates from the
  installed release content rather than acting as a remote installer.
- [ ] `runectx` makes no network calls outside explicit `init` and `update`
  flows.
- [ ] The following anti-patterns are absent: required global installs,
  bash-only installers, overwriting existing `.claude`/`.github` files,
  hidden runtime-manager dependencies, template-only primary distribution,
  implicit adapter-pack fetches during sync, and silent auto-updates.

## 9. RuneCode Readiness (Companion Track)

These are not shipped by this repository, but the MVP is not truly ready until
RuneContext makes them possible and testable.

- [ ] RuneCode can version-gate on `runecontext_version`.
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
- [ ] CLI integration tests cover all primary and secondary commands plus
  `--json`, `--non-interactive`, `--dry-run`, and `--explain` behavior.
- [ ] Adapter smoke tests and parity checks exist for `generic`, `claude-code`,
  `opencode`, and `codex`.
- [ ] Release/install/update flows are covered by end-to-end tests over
  reference projects.
- [ ] Signed-tag verification, Verified-mode gating, and RuneCode parity
  fixtures are all covered by automated tests.
