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
  expected-commit checking, and fail-closed mismatch behavior.
- [ ] Mutable refs require explicit opt-in and warnings.
- [ ] Local path sources work for developer-local usage and are marked
  unverified/non-auditable.
- [ ] Monorepo nearest-root discovery works.
- [ ] Bundle resolution is deterministic, cycle-safe, depth-limited, and path-
  boundary-safe.

## 3. Change Workflow And Standards

- [ ] Every substantive work item gets a stable change ID.
- [ ] Minimum mode works with `status.yaml`, `proposal.md`, and `standards.md`.
- [ ] Full mode works by materializing deeper files only when needed.
- [ ] Work-type and size branching rules exist for project, feature, bug,
  standard, and chore changes.
- [ ] Ask-more versus infer-more heuristics exist and inferred assumptions are
  captured in `proposal.md`.
- [ ] `proposal.md` uses the required section order and validation rules.
- [ ] `standards.md` is always present and reviewably maintained.
- [ ] Standards are referenced by path rather than copied into change/spec
  bodies.
- [ ] Standards frontmatter validation, deprecation, and rename/migration rules
  work.
- [ ] Cross-artifact references in change metadata validate cleanly or produce
  clear diagnostics.
- [ ] Closed changes remain directly accessible at stable paths.

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
- [ ] Verified mode generates a baseline artifact.
- [ ] Verified mode generates receipt families for context packs, changes,
  promotions, and verifications.
- [ ] Receipt filenames are collision-resistant and do not require a shared
  mutable index.
- [ ] Backfill can generate imported historical provenance distinct from native
  verified capture.

## 6. CLI And Machine Interface

- [ ] The primary CLI commands exist: `init`, `status`, `change new`,
  `change shape`, `bundle resolve`, and `change close`.
- [ ] The secondary/admin commands exist: `validate`, `doctor`,
  `standard discover`, `promote`, `assurance enable verified`, and
  `assurance backfill`.
- [ ] Before alpha.6 is complete, any earlier validation entrypoints remain narrow
  wrappers around the same core contracts rather than alternate semantics.
- [ ] Before alpha.6 is complete, any earlier validation entrypoints use a
  documented and tested machine-readable output contract.
- [ ] Early validation entrypoints fail closed with structured diagnostics rather
  than panics when schemas, YAML, markdown contracts, or project references are invalid.
- [ ] Early validation entrypoints honor declared project content roots and the
  full restricted-YAML profile rather than relying on default-path assumptions.
- [ ] Alpha-stage release metadata, module metadata, and parser behavior stay
  consistent with the documented release series and fail-closed contracts.
- [ ] The adapter-management command exists: `runectx adapter sync <tool>`.
- [ ] Machine-facing flags exist and behave consistently: `--json`,
  `--non-interactive`, `--dry-run`, and `--explain`.
- [ ] CLI behavior stays aligned with the canonical file model.

## 7. Adapters

- [ ] The canonical operations reference exists under `runecontext/operations/`.
- [ ] The `generic` adapter exists.
- [ ] The `claude-code`, `opencode`, and `codex` adapters exist.
- [ ] Adapters differ in UX only, not in core semantics or source-of-truth
  files.
- [ ] Adapters are packaged for release.

## 8. Release, Install, And Update

- [ ] GitHub release artifacts exist for the repo-first distribution model.
- [ ] Releases include schema bundle, adapter packs, checksums, release notes,
  and compatibility information.
- [ ] Optional `runectx` binaries are packaged.
- [ ] Manual repo install is documented and tested.
- [ ] `runectx update` is diff-first and reviewable.
- [ ] Adapter sync/update is namespaced and merge-aware.
- [ ] The following anti-patterns are absent: required global installs,
  bash-only installers, overwriting existing `.claude`/`.github` files,
  hidden runtime-manager dependencies, template-only primary distribution, and
  silent auto-updates.

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
