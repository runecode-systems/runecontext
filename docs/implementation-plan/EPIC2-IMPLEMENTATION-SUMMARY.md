# Epic 2 Implementation Summary: Schema and File-Contract Baseline

**Status**: Complete for v0.1.0-alpha.1  
**Date Completed**: 2026-03-16  
**Version**: alpha.1  
**Branch**: `1.2_schema_and_file_contract_baseline`

---

## Executive Summary

Epic 2 establishes the authoritative schema and file-contract baseline for RuneContext v0.1.0-alpha.1. The implementation follows a **security-first, closed-schema approach** that prioritizes auditability, determinism, and policy neutrality while maintaining lightweight extensibility for advanced users.

All JSON schemas now target **JSON Schema Draft 2020-12** so conditional variants can stay closed safely via `unevaluatedProperties: false` where needed.

### Key Decision: Closed Schemas with Opt-In Extensions

Rather than allowing arbitrary unknown fields (which creates security/audit risks), all machine-readable artifacts use **closed JSON schemas by default**. A single optional `extensions` object is available when explicitly enabled, enforcing namespaced keys and non-authoritative semantics.

---

## Delivered Artifacts

### 1. JSON Schemas (schemas/)

#### runecontext.schema.json
- **Scope**: Project-root configuration file.
- **Key Fields**: `schema_version` (const 1), `runecontext_version`, `assurance_tier` (enum: plain|verified), `source` (object), `allow_extensions` (boolean, default false).
- **Source Types**: embedded, git (commit/signed_tag/ref), path (local developer-local).
- **Git Source Detail**: Supports pinned commits, signed tags with verification, mutable refs (opt-in only), local subdir resolution.
- **Dialect**: JSON Schema Draft 2020-12 with `unevaluatedProperties: false` on `source` so variant-specific fields remain closed without Draft-7 conditional pitfalls.
- **Closed**: Yes. Unknown fields rejected.
- **Extensions**: Yes, optional (when `allow_extensions: true`).
- **Extensions Keys**: Extension keys require explicit dot-separated ownership segments (`owner.name.more`) with lowercase alphanumerics plus `_` and `-` inside each non-empty segment.

#### bundle.schema.json
- **Scope**: Context bundle selectors at `bundles/*.yaml`.
- **Key Fields**: `schema_version` (const 1), `id` (unique), `extends` (parent IDs, max depth 8), `includes`/`excludes` (aspect-aware maps).
- **Aspect Families**: project, standards, specs, decisions.
- **Inheritance**: Depth-first, left-to-right linearization; last-matching-rule-wins semantics.
- **Pattern Grammar**: Include/exclude entries are RuneContext bundle patterns (exact path, `*`, and recursive `**`).
- **Closed**: Yes. Unknown fields rejected.
- **Extensions**: Yes, optional. Full acceptance still requires project-level validation to confirm the root `runecontext.yaml` has `allow_extensions: true`.

#### change-status.schema.json
- **Scope**: Change lifecycle metadata at `changes/<change-id>/status.yaml`.
- **Key Fields**: `schema_version`, `id` (CHG-YYYY-NNN-RAND-slug), `title`, `status` (lifecycle enum), `type` (base enum + x- prefix), `size`, `verification_status` (pending|passed|failed|skipped), traceability fields, `promotion_assessment`.
- **Type Enum**: project, feature, bug, standard, chore, or custom x-* values.
- **Promotion Assessment**: Structured status (pending|none|suggested|accepted|completed) with suggested targets (target_type, target_path, summary).
- **Supersession Validation**: When `status=superseded`, `superseded_by` is required in-schema; bidirectional consistency remains a project-level cross-file validation rule.
- **Closed**: Yes. Unknown fields rejected.
- **Extensions**: Yes, optional. Full acceptance still requires project-level validation to confirm the root `runecontext.yaml` has `allow_extensions: true`.

#### context-pack.schema.json
- **Scope**: Generated deterministic runtime artifact (output of bundle resolution).
- **Key Fields**: `schema_version` (const 1), `canonicalization` (const rfc8785-jcs), `pack_hash_alg` (const sha256), `pack_hash` (SHA256 hex), `id`, `resolved_from` (metadata), `selected`/`excluded` (file inventories).
- **Resolved Metadata**: source_mode, source_ref, source_verification (enum: pinned_commit|verified_signed_tag|unverified_mutable_ref|unverified_local_source|embedded), context_bundle_ids, and `source_commit` only when `source_mode` is `git`.
- **Source Consistency**: `embedded` sources must use `source_verification: embedded`; `path` sources must use `source_verification: unverified_local_source`; only `git` sources may carry `source_commit`.
- **Hashing**: `pack_hash` is SHA256 over the RFC 8785 JCS serialization of the complete pack minus the `pack_hash` field itself, ensuring no circular inputs.
- **File Inventory**: Path, SHA256 hash, selected_by (ordered rules showing last-matching-rule-wins).
- **Canonical Shape**: `selected` always contains `project`, `standards`, `specs`, and `decisions`; `excluded` uses the same four-key layout whenever present so equivalent packs hash identically.
- **Alpha Refinement**: This canonical shape rule is treated as an alpha.1 contract refinement before a stable v1 release, so generators and fixtures must align before downstream consumers rely on pack hashes.
- **Closed**: Yes, fully. **No extensions permitted in v1 generated artifacts.**

### 2. Contracts and Profiles (schemas/)

#### MACHINE-READABLE-PROFILE.md
Defines the restricted YAML subset for all machine-readable files:
- **No YAML Anchors/Aliases**: Each value written out in full.
- **No Duplicate Keys**: YAML/profile validation must reject duplicates during parsing or decode configuration before schema validation runs.
- **No Implicit Type Coercion**: yes/no not coerced to booleans; use true/false explicitly.
- **No Custom Tags**: YAML tags forbidden.
- **UTF-8 Only**: No other encodings permitted.
- **Normalized Formatting**: 2-space indentation, no trailing whitespace, Unix LF, single trailing newline in generated artifacts.
- **Schema Dialect**: JSON Schema Draft 2020-12 across all shipped schemas.

**Canonical Data Model for Hashing**:
- Parse YAML to nested structure (objects, arrays, strings, numbers, booleans, nulls).
- Sort object keys lexicographically.
- Serialize to compact JSON.
- Apply RFC 8785 JCS canonicalization (minimizes whitespace, normalizes numbers, sorts keys).
- Compute SHA256 hash of JCS output.

**Unknown-Field Behavior**:
- **Closed Schema Default**: Unknown top-level fields rejected.
- **Extensions Opt-In**: Optional `extensions` object allowed when `runecontext.yaml` sets `allow_extensions: true`; bundle/status enforcement is project-scoped because the opt-in flag lives in the root file.
- **Namespaced Keys**: Extension keys must follow the `owner.name.more` pattern—`[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?)+`—so dots only separate non-empty namespace segments.
- **Non-Authoritative**: Extension values are data, not semantics; cannot affect validation, bundle resolution, lifecycle, or policy.
- **Not in Generated Artifacts**: Context packs and assurance artifacts contain no extensions in v1.

#### MARKDOWN-CONTRACTS.md
Defines strict structural contracts for human-readable markdown files:

**proposal.md** (Canonical Reviewable Intent Artifact):
- Required sections (level-2 headings, exact order):
  1. Summary
  2. Problem
  3. Proposed Change
  4. Why Now
  5. Assumptions
  6. Out of Scope
  7. Impact
- Each required section must contain content or explicit `N/A`.
- Additional custom sections allowed after required section block.
- Tooling must parse and validate section ordering.

**standards.md** (Automatically Maintained):
- Recommended sections:
  - Applicable Standards (normalized list, regenerable by tooling)
  - Standards Added Since Last Refresh (optional)
  - Standards Considered But Excluded (optional)
  - Resolution Notes (optional)
- Auto-maintained by tooling on `change new`, `change shape`, and planning.
- Refreshes must be reviewable diffs; never silent rewrites.
- Always present; never disposable.

### 3. Test Fixtures (`fixtures/schema-contracts/`)

The shipped fixture set currently contains **15 YAML fixtures** plus a README that distinguishes standalone schema validation from project-level and YAML-profile validation:

#### Valid Cases
- `valid-runecontext-no-extensions.yaml`: Closed schema, no extensions.
- `valid-runecontext-with-extensions-optin.yaml`: Root opt-in plus a real namespaced `extensions` object.
- `valid-git-source-signed-tag.yaml`: Signed-tag git source contract.
- `valid-bundle-closed-schema.yaml`: Bundle with closed schema.
- `valid-bundle-with-extensions.yaml`: Bundle with namespaced extension keys; requires project-level opt-in.
- `valid-change-status.yaml`: Complete change status with all fields.
- `valid-custom-type.yaml`: Change with custom type `x-migration`.
- `valid-superseded-change.yaml`: Superseded change with required `superseded_by`.
- `valid-context-pack.yaml`: Generated context pack with consistent git provenance, canonical four-aspect shape, and valid 64-character hashes.

#### Reject Cases
- `reject-unknown-field-runecontext.yaml`: Unknown top-level field fails validation.
- `reject-unknown-schema-version.yaml`: Unknown schema version fails closed.
- `reject-bad-extension-key.yaml`: Invalid extension key fails namespacing rules.
- `reject-context-pack-unknown-field.yaml`: Unknown field in generated artifact fails validation.
- `reject-yaml-anchors-aliases.yaml`: YAML anchors/aliases violate profile.
- `reject-extensions-without-optin.yaml`: Project-level rejection case for bundle/status extensions without root opt-in.

### 4. Updated Documentation

#### docs/implementation-plan/milestone-breakdown.md
- Marked 14 Epic 2 issues as completed (checked).
- Consolidated Epic 4 into Epic 2 (canonical data rules integrated with schema contracts).
- Updated exit criteria to reflect closed schemas, extensions opt-in, YAML profile, JCS hashing, strict owner-style extension names, and canonical context-pack shape.
- Simplified markdown contract epics to reflect completed sections.

### 5. Post-Review Hardening Follow-Up

The alpha.2 bundle and source-resolution implementation now includes the
additional hardening agreed during code review triage:

- **Embedded root canonicalization**: embedded RuneContext roots are checked
  after symlink resolution against the selected project root so symlinked
  escapes fail closed.
- **Git transport hardening**: git URLs now reject remote-helper forms,
  subprocesses run with an explicit protocol allowlist, and surfaced git errors
  redact URLs or embedded credentials.
- **Whole-project symlink containment**: project walks and file reads for
  bundle/spec/decision/change validation now apply resolved-path containment so
  symlinked files cannot escape the selected subtree.
- **Bundle traversal bounds**: bundle glob enumeration now enforces practical
  depth and file-count caps in addition to the existing inheritance-depth guard.
- **Immutable bundle results**: cached bundle resolutions are returned as
  defensive copies so later callers cannot mutate hidden cached state.
- **Concurrent schema safety**: schema compilation cache access is synchronized
  for concurrent validator use.

### 6. Additional Tests

Follow-up tests were added for:

- rejection of git remote-helper URL forms
- redaction behavior for surfaced git transport errors
- explicit git subprocess environment guards
- embedded-root symlink escape rejection
- whole-project spec symlink escape rejection
- defensive-copy behavior for cached bundle resolutions

### 7. PR Review Follow-Up Fixes

The re-run PR review surfaced a small set of additional follow-up fixes, all of
which are now incorporated:

- **Safe bundle-file reads**: bundle catalog loading now reuses the same
  containment-aware file-read path as other project artifacts instead of raw
  `os.ReadFile` after discovery.
- **Canonical aspect roots**: exact and glob bundle evaluation now canonicalize
  the selected aspect root before containment checks so in-bounds symlinked
  aspect directories are accepted consistently.
- **Symlinked root directories**: whole-project discovery now accepts symlinked
  root directories like `specs/` or `bundles/` when they resolve inside the
  selected subtree.
- **Clearer diagnostics**: non-regular-file errors now report the resolved path
  rather than the unresolved logical path.

Additional regression tests now cover:

- walking a symlinked root directory
- bundle resolution with a symlinked in-bounds aspect root
- removal of stale unused setup in the embedded symlink-escape test

### 8. Final PR Cleanup Pass

The final cleanup pass for the re-run PR review made two small source-resolution
refinements:

- **Clearer path helper naming**: the helper that resolves a root and target pair
  is now named to reflect canonicalization only, rather than implying it also
  enforces containment.
- **Less redundant open-path validation**: local snapshot copying now validates
  the exact resolved path being opened against the declared local source root,
  preserving the hardening intent without re-resolving the same path under a
  misleading helper name.

### 9. Additional PR Review Follow-Up

One more review pass surfaced a short set of narrow correctness fixes that are
now incorporated without changing the broader alpha.1-alpha.3 scope:

- **Markdown duplicate-heading fragments**: automatically derived fragments now
  use deterministic duplicate numbering (`foo`, `foo-1`, `foo-2`, ...) while
  still skipping already occupied suffixed forms.
- **Bundle traversal consistency**: broken or disappearing paths discovered
  during bundle walking now fail closed consistently instead of being partially
  normalized into an empty-path sentinel before a later `Stat` failure.
- **Thin CLI flag parsing**: required string flags now reject another long flag
  token as a missing value so callers get an immediate usage error instead of a
  confusing downstream parse failure.
- **Status YAML rewrite robustness**: YAML encoding now propagates `Close`
  failures as well as `Encode` failures when rendering rewritten `status.yaml`
  content.

This follow-up intentionally does not narrow `type: git` to remote-only URLs.
The current alpha.2 contract still allows local repository paths for testing and
developer-local workflows, so changing that behavior would be a deliberate
product decision rather than a narrow review fix.

---

## Security and Audit Implications

### Policy Neutrality (core/trust-boundaries.md)
- ✅ RuneContext content (standards, bundles, changes, proposals) is never runtime authority.
- ✅ Extensions cannot affect policy, approvals, or capabilities.
- ✅ All extensions are non-authoritative data; they cannot grant permissions.

### LLM Input Trust Boundary (core/trust-boundaries.md)
- ✅ Closed schemas prevent hidden semantics from untrusted sources.
- ✅ Namespaced extension keys reduce prompt-injection surface.
- ✅ Unknown schema_version fails closed; no speculative permissiveness.
- ✅ Deterministic hashing (RFC 8785 JCS) prevents semantic drift across implementations.

### Auditability
- ✅ All authoritative files (bundles, status, proposals) are schema-versioned and machine-readable.
- ✅ Generated artifacts (context packs, baselines, receipts) are fully closed in v1; no hidden fields.
- ✅ Extensions require explicit opt-in with visible warnings; audit trail is clear.
- ✅ JCS canonical hashing ensures bit-for-bit reproducibility across implementations.

---

## Alignment with RuneCode Integration

### Version Gating (RuneCode Companion Track)
- RuneCode can read `runecontext_version` from `runecontext.yaml` and apply version checks before deeper resolution.
- Unknown versions fail closed; RuneCode refuses to proceed if RuneContext version is out of supported range.

### Context Pack Hashing and Binding
- Context packs carry top-level `pack_hash` (SHA256 over RFC 8785 JCS of the full pack minus `pack_hash`).
- RuneCode can bind and sign this hash in audit/provenance flows.
- Deterministic hashing ensures local/remote parity.
- Because canonical empty-aspect encoding is now explicit, any generator that previously omitted empty aspect keys must be updated before comparing or persisting alpha.1 pack hashes.

### Proposal.md Intent Binding
- RuneCode can parse `proposal.md` by its strict section ordering.
- Intent artifacts can be bound into audit history without granting runtime authority.

### Standards and Bundle Resolution
- RuneCode uses the same closed-schema validation; no ambiguity between local and remote.
- Extension opt-in is visible; RuneCode can audit and warn users about projects using extensions.

---

## Backwards Compatibility and Forward Extensibility

### v1 Semantics (Frozen)
- Schema version 1 is authoritative and frozen for v0.1.0-alpha.1 through v0.1.0 (MVP).
- Unknown schema_version values are always rejected.
- Unknown enum values in known schemas are always rejected.

### Future Extensibility (alpha.2+)
- If new schema fields are needed, increment `schema_version` to 2.
- Implementations may preserve unknown fields from v1 when round-tripping, but must fail on unknown schema_version.
- Extension opt-in policy remains stable and forward-compatible.
- Draft 2020-12 remains the target schema dialect unless future validator support forces a migration plan.
- Alpha contract refinements that land before a stable v1 release must still update fixtures, generators, and documentation in the same change so downstream implementations stay in sync.

### Generated Artifact Stability
- Context pack schema is fully closed in v1; no extensions.
- Assurance artifacts (baselines, receipts) are designed as fully closed.
- Future runecode binding or richer provenance receipts (alpha.5+) will use new artifact types or future schema versions, not mutate existing ones.

---

## Testing Strategy

### Unit Test Validation
1. **Schema Validation**: Each fixture (valid and reject) validated against its corresponding JSON schema.
2. **Profile Compliance**: YAML profile rules (no anchors, UTF-8, etc.) enforced in validation layer.
3. **Enum Enforcement**: Unknown enum values rejected; custom x-* types accepted for status.type.
4. **Extension Opt-In**: Root schema enforces its own opt-in directly; bundle/status files require project-level validation against the root config, with warnings when extensions are accepted.

### Parity Tests
1. **Local/Remote**: Same bundles and status files validated identically in Go and TypeScript implementations.
2. **Hashing**: Same context packs produce identical SHA256 hashes across implementations using RFC 8785 JCS.
3. **Markdown Parsing**: proposal.md and standards.md parsed identically by tools.

### Integration Tests
1. **End-to-End**: Create a minimal project with embedded RuneContext; validate all files against schemas.
2. **Extension Scenario**: Enable `allow_extensions: true`; validate that extensions are accepted and warnings issued.
3. **Source Verification**: Validate source_verification enum in context packs (pinned, signed-tag, mutable, local, embedded).
4. **CI Split**: Nix jobs run `just nix-ci` inside the dev shell, while Windows portability jobs run the portable `just ci` target with no Nix dependency.

---

## Known Deferred Items (Not In Alpha.1)

The following are explicitly deferred to later alphas:

1. **Alpha.2 - Actual Source Resolution Logic**: Schemas are frozen, but implementation of resolver, symlink handling, and signed-tag verification logic happens in alpha.2.
2. **Alpha.3 - Change Workflow Implementation**: Change ID allocation, lifecycle transitions, collision detection deferred to alpha.3.
3. **Alpha.3 - Automatic Standards Maintenance**: Tooling for refreshing standards.md during change creation deferred to alpha.3, though alpha.1 now includes executable structure validation and fixture coverage.
4. **Alpha.5 - Assurance Artifacts**: Baseline and receipt schema inventory is defined; implementation deferred to alpha.5.
5. **Alpha.7 - Adapter Packs**: Tool-specific adapter UX and operations reference deferred to alpha.7.

---

## Files Delivered

### Schemas (4 JSON files)
- `schemas/runecontext.schema.json`
- `schemas/bundle.schema.json`
- `schemas/change-status.schema.json`
- `schemas/context-pack.schema.json`

### Contracts (2 Markdown files)
- `schemas/MACHINE-READABLE-PROFILE.md`
- `schemas/MARKDOWN-CONTRACTS.md`

### Fixtures (15 YAML files + 1 README)
- `fixtures/schema-contracts/README.md`
- `fixtures/schema-contracts/valid-*.yaml` (9 files)
- `fixtures/schema-contracts/reject-*.yaml` (6 files)

### Documentation Updates
- `docs/implementation-plan/milestone-breakdown.md` (marked issues complete, updated epics and exit criteria)

### Git Commit
- Commit `b8c1d73`: "feat(alpha.1/epic2): Add closed-schema contracts and extension policy"

---

## Next Steps (Unblocked by This Epic)

1. **Alpha.2**: Implement source resolution logic using the frozen `runecontext.schema.json` and `source_verification` enum.
2. **Alpha.3**: Implement change ID allocation and lifecycle management using frozen `change-status.schema.json`.
3. **Alpha.4**: Implement context pack generation and RFC 8785 JCS hashing using frozen `context-pack.schema.json`.
4. **RuneCode Companion Track**: Begin fixture-based validation of policy-neutrality and version-gating using these schemas and fixtures.

---

## Questions and Clarifications

### Extension Namespacing
**Q**: Why ownership-style namespacing (e.g., `io.runecode.custom`) rather than simple `x-` prefix?  
**A**: Ownership namespaces prevent collisions, make typos obvious (misplaced dots fail validation), and audit trails show origin (io.runecode = RuneCode integration, dev.acme = company internal). The rule `[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?)+` enforces this while still allowing underscores and dashes inside each non-empty segment.

### Why No Extensions in Generated Artifacts?
**Q**: Why are context packs fully closed; no extensions allowed?  
**A**: Generated artifacts are outputs of deterministic algorithms; they must be bit-for-bit identical across implementations. Allowing extensions would create ambiguity about what is canonical. Future richer provenance (alpha.5+) will use new artifact types, not mutate existing ones.

### Why Closed Schemas by Default?
**Q**: Doesn't closed schema limit future extensibility?  
**A**: No. If new fields are needed, increment `schema_version` to 2. Implementations refuse to load unknown schema_version values. This ensures clarity: unknown versions are always rejected, not speculatively interpreted.

---

## Acceptance Checklist (v0.1.0-alpha.1 Epic 2 Complete)

- [x] JSON schemas authored for all core machine-readable files.
- [x] Restricted YAML profile frozen (no anchors, UTF-8, normalized formatting, etc.).
- [x] RFC 8785 JCS canonicalization and SHA256 hashing defined.
- [x] Unknown-field behavior frozen (closed by default, opt-in extensions with namespacing).
- [x] Markdown structure contracts frozen (proposal.md section ordering, standards.md auto-maintenance).
- [x] All source_verification enum values defined and documented.
- [x] all type enum values (base + x- prefix) documented.
- [x] Test fixture README distinguishes standalone schema, project-level, and YAML-profile validation cases for the shipped 15-fixture set.
- [x] Milestone-breakdown.md updated with completion status and new exit criteria.
- [x] Security/audit implications documented (policy neutrality, LLM trust, closed schemas).
- [x] RuneCode companion-track checkpoints identified (version-gating, parity testing, policy-neutrality validation).
- [x] Backwards compatibility and forward extensibility strategy documented.
