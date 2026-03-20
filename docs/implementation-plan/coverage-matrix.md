# Coverage Matrix

This document maps the contents of `docs/project_idea.md` into the milestone
plan so it is clear that the planning documents capture the full design.

## Section-To-Milestone Map

| Source section in `docs/project_idea.md` | Planned capture | Primary milestone(s) | RuneCode companion track |
| --- | --- | --- | --- |
| Executive Summary | Distributed across the full MVP plan | `alpha.1`-`alpha.8`, `v0.1.0` | Yes |
| Why RuneContext Exists | Product positioning and portability guardrails | `alpha.1`, `README.md` | Yes |
| Best Ideas To Keep From Agent OS | Markdown-first, low-ceremony, path-referenced standards | `alpha.1`, `alpha.3` | Yes |
| OpenSpec Ideas To Mix In | Change orientation, lifecycle state, traceability | `alpha.3`, `alpha.4` | Yes |
| Research Findings / Design Implications | Planning principles, progressive disclosure, reviewable diffs | `README.md`, `alpha.3`, `alpha.5`, `alpha.7` | Yes |
| Goals | Acceptance criteria and MVP boundaries | `README.md`, `mvp-acceptance.md` | Yes |
| Non-Goals | Scope guardrails and post-MVP separation | `README.md`, `post-mvp.md` | Yes |
| Product Decomposition | Core/adapters/RuneCode repository boundary | `alpha.1` | Yes |
| Why The Three Layers Need To Exist | Boundary enforcement and ownership rules | `alpha.1`, `alpha.7` | Yes |
| Packaging, Repositories, And Releases | Repo structure, release model, and Nix-built release workflow shape | `alpha.1`, `alpha.8` | Yes |
| Releases and Installation | Install lanes, update flow, compatibility matrix, and local-only adapter sync | `alpha.8` | Yes |
| Optional Assurance And Verifiable Tracing | Plain/Verified model, shared authored workflow, portable receipts, and backfill | `alpha.6` | Yes |
| RuneCode Context And Integration Constraints | Companion-track test and contract checklist | `alpha.1`-`alpha.8`, `mvp-acceptance.md` | Yes |
| Usage Scenarios | Validation of local, remote, and non-RuneCode flows | `alpha.2`, `alpha.5`, `alpha.6`, `alpha.8` | Yes |
| Terminology | Normative glossary and naming rules | `alpha.1` | Indirect |
| Why `context bundle` Was Chosen | Naming and user-facing vocabulary | `alpha.1` | Indirect |
| Storage Modes | Embedded, linked, path, and monorepo behavior | `alpha.2` | Yes |
| Project Root Configuration | Root schema, versioning, assurance tier | `alpha.1`, `alpha.2` | Yes |
| Embedded Mode | Resolver implementation and reference fixture | `alpha.2`, `alpha.8` | Yes |
| Linked Mode | Commit, mutable ref, signed tag, and path handling | `alpha.2`, `alpha.8` | Yes |
| Monorepo Support | Discovery and reference fixtures | `alpha.2`, `alpha.8` | Yes |
| Core On-Disk Layout | Authored/generated ownership and lean shaped-change scaffolding | `alpha.1`, `alpha.3`, `alpha.5` | Indirect |
| Machine-Readable Schema Versioning | Schema behavior, unknown-field handling, YAML profile | `alpha.1` | Yes |
| Generated Artifact Commit Policy | Assurance and release/install policy | `alpha.6`, `alpha.8` | Yes |
| Project Files | Project-context scaffolding and conventions | `alpha.1`, `alpha.5` | Indirect |
| Standards | Frontmatter rules, lifecycle, migration behavior | `alpha.3` | Yes |
| Context Bundles | Bundle schema and resolution engine | `alpha.1`, `alpha.2` | Yes |
| Changes | ID allocation and lifecycle workflow | `alpha.3` | Yes |
| Minimum And Full Change Shapes | Progressive disclosure and lean shaped-file materialization | `alpha.3` | Yes |
| Branching Logic By Work Type And Size | Intake depth, escalation, and minimum/full mode branching | `alpha.3` | Yes |
| When To Ask More Vs Less | Prompting heuristics and assumption capture | `alpha.3` | Yes |
| Proposal.md Structure | Generator/parser/validator rules | `alpha.1`, `alpha.3` | Yes |
| Standards.md Structure | Normalized structure and auto-maintenance | `alpha.1`, `alpha.3` | Yes |
| Automatic Standards Maintenance | Change creation/shaping refresh and review diffs | `alpha.3`, `alpha.5` | Yes |
| Stable Specs | Durable spec conventions and traceability metadata | `alpha.1`, `alpha.3` | Yes |
| Decisions | Durable decision conventions and traceability metadata | `alpha.1`, `alpha.3` | Yes |
| Traceability And Future Lineage | Artifact-level traceability, heading-fragment deep refs, richer lineage later | `alpha.1`, `alpha.3`, `alpha.4`, `post-mvp.md` | Yes |
| Context Bundle Semantics | Ordered rule application and validation rules | `alpha.2` | Yes |
| Deterministic Resolved Output | Context pack generation, provenance, pack hash | `alpha.4` | Yes |
| Standards Membership And Authoring Model | Manual versus assisted standards workflow | `alpha.3` | Yes |
| Change Lifecycle | State machine and close rules | `alpha.3` | Yes |
| Promotion / Promotion Assessment | Structured promotion suggestions and close flow | `alpha.4` | Yes |
| Archive/Promotion Rule | Stable-path close behavior | `alpha.3` | Yes |
| Historical Traceability Requirements | Future-safe linkage and readable history | `alpha.3`, `alpha.4` | Yes |
| Minimal Process And User Experience | Small mental model and progressive disclosure | `alpha.3`, `alpha.5`, `alpha.7` | Yes |
| Invocation Surfaces And Command Architecture | Thin early change/status wrappers, broader CLI surface, and machine-facing flags | `alpha.3`, `alpha.5`, `alpha.7` | Yes |
| Adapters | Thin adapters, capability model, and auto-validation workflow hooks | `alpha.7` | Yes |
| RuneCode Integration Details / Required Capabilities | Companion-track fixtures and acceptance checkpoints | `alpha.2`-`alpha.8`, `mvp-acceptance.md` | Yes |
| Context Pack Delivery Into Isolates | Pack-hash, artifact, and typed-delivery readiness for companion integration | `alpha.4`-`alpha.8`, `mvp-acceptance.md` | Yes |
| Reviewable Intent In RuneCode History | Change/proposal binding and audit-history readiness | `alpha.3`-`alpha.8`, `mvp-acceptance.md` | Yes |
| Policy Neutrality Rule | Contract that RuneContext never grants capabilities or approvals | `alpha.1`, `mvp-acceptance.md` | Yes |
| LLM Input Trust Boundary | Contract that RuneContext text is untrusted model input; heuristics remain supplementary | `alpha.1`, `post-mvp.md` | Yes |
| Generated Indexes And Manifests | Manifest and generated inventories | `alpha.4` | Yes |
| Standards Referencing Rule | Path-reference rule for standards | `alpha.3` | Yes |
| Recommended Implementation Plan | Recast into alpha-based milestone plan | `README.md`, `milestone-breakdown.md` | Yes |
| Explicit Design Decisions From The Discussion | Cross-checked below and reflected throughout the plan | All | Yes |
| Final Recommendation | Captured as the MVP goal statement | `README.md`, `mvp-acceptance.md` | Yes |

## Explicit Decision Extraction Checklist

- Decision: the reusable selector is called a `context bundle` and the folder is
  `bundles/`.
  - Planned capture: `alpha.1`
- Decision: bundle files require `schema_version`, `id`, and `includes`, with
  optional `extends` and `excludes`.
  - Planned capture: `alpha.1`, `alpha.2`
- Decision: bundle inheritance uses depth-first, left-to-right linearization
  and last-matching-rule-wins evaluation.
  - Planned capture: `alpha.2`
- Decision: linked sources prefer immutable commits or verified signed tags.
  - Planned capture: `alpha.2`, `alpha.8`
- Decision: signed-tag verification must use explicitly supplied trusted-signer
  inputs rather than hidden machine-global trust configuration.
  - Planned capture: `alpha.2`
- Decision: embedded source paths and git subdirectories must remain contained
  within the selected project/repository root, while `type: path` may point to
  developer-local trees outside the repo but must still remain contained within
  its declared local tree after symlink resolution.
  - Planned capture: `alpha.2`
- Decision: git source resolution must reject option-like URL/ref values, avoid
  hidden host credentials/config, and run with explicit subprocess timeouts.
  - Planned capture: `alpha.2`
- Decision: mutable git refs should fail fast on obviously invalid ref syntax
  before any git subprocess is invoked.
  - Planned capture: `alpha.2`
- Decision: RuneContext should not rely on environment variables for
  user-facing configuration or correctness-critical semantics; only minimal
  non-semantic process environment plumbing is allowed.
  - Planned capture: `alpha.2`, `alpha.5`
- Decision: pinned-commit git resolution must not assume remote support for
  direct fetch-by-SHA.
  - Planned capture: `alpha.2`
- Decision: local path snapshotting should exclude obvious repo-control
  directories and apply practical file/depth/byte limits.
  - Planned capture: `alpha.2`, `alpha.4`
- Decision: validation helpers that materialize temporary source trees must own
  cleanup on both success and failure paths.
  - Planned capture: `alpha.2`
- Decision: project root config carries `runecontext_version` and
  `assurance_tier`.
  - Planned capture: `alpha.1`, `alpha.2`
- Decision: change IDs use `CHG-YYYY-NNN-RAND-short-slug`.
  - Planned capture: `alpha.3`
- Decision: context packs must include a top-level canonical hash.
  - Planned capture: `alpha.4`
- Decision: context packs keep required `generated_at` metadata, but canonical
  `pack_hash` inputs exclude regeneration-only timestamps so identical resolved
  content hashes the same across regenerations.
  - Planned capture: `alpha.4`
- Decision: core context-pack builders require explicit `generated_at` input;
  command surfaces may supply defaults later, but the canonical engine should
  not hide wall-clock time injection.
  - Planned capture: `alpha.4`
- Decision: core context-pack builders reject sub-second `generated_at` values
  so timestamp precision changes are explicit instead of silently truncated.
  - Planned capture: `alpha.4`
- Decision: selected-file hashing should normalize text line endings so LF and
  CRLF checkouts of the same logical content still produce portable deterministic
  pack hashes.
  - Planned capture: `alpha.4`
- Decision: path-source `source_ref` values persisted into context packs must be
  portable forward-slash relative paths without drive-qualified, UNC, or
  traversal segments.
  - Planned capture: `alpha.4`
- Decision: alpha.4 context packs should use an explicit RuneContext-owned
  canonicalization token for their restricted emitted-shape serializer rather
  than claiming full RFC 8785 JCS interoperability prematurely.
  - Planned capture: `alpha.4`
- Decision: the restricted context-pack canonicalization profile must reject
  invalid UTF-8 string content rather than silently normalizing malformed bytes
  during hashing.
  - Planned capture: `alpha.4`
- Decision: persisted context-pack provenance keeps `bundle`, `aspect`, `rule`,
  `pattern`, and `kind` so explanation and later receipts do not need a format
  refactor.
  - Planned capture: `alpha.4`
- Decision: context-pack request identity uses a hybrid model: authored
  workflows still prefer one top-level bundle or authored composite bundles,
  while generated packs may record ordered `requested_bundle_ids` separately
  from resolved bundle linearization.
  - Planned capture: `alpha.4`
- Decision: generated context-pack bundle identifiers should use the same
  fail-closed ID grammar as authored bundle contracts.
  - Planned capture: `alpha.4`
- Decision: machine-readable context-pack reports should carry an explicit
  schema version and standalone schema so report consumers can validate that
  envelope independently of the embedded pack schema.
  - Planned capture: `alpha.4`
- Decision: report-envelope validation and embedded-pack validation remain
  distinct contracts; consumers that need full guarantees must validate both.
  - Planned capture: `alpha.4`
- Decision: report advisory warning counters should be non-negative in schema
  contracts (`value`, `threshold`) to reject impossible machine payloads.
  - Planned capture: `alpha.4`
- Decision: fail-closed rebuild checks should retry only on genuinely transient
  file-instability signals; non-transient digest/read failures must surface as
  direct errors.
  - Planned capture: `alpha.4`
- Decision: advisory-threshold configuration treats a fully zero-valued struct as
  "use defaults", while explicit field zeros remain meaningful once any field is
  set and negative values opt back into per-field defaults.
  - Planned capture: `alpha.4`
- Decision: advisory-threshold defaults should be exposed as immutable or copy-
  returning values rather than mutable exported globals.
  - Planned capture: `alpha.4`
- Decision: rebuild stability checks operate against the loaded project snapshot
  and selected-file content rather than reloading bundle-definition files from
  disk mid-attempt.
  - Planned capture: `alpha.4`
- Decision: test-only context-pack read-hook helpers should fail safe by falling
  back to the real file reader when unset instead of panicking.
  - Planned capture: `alpha.4`
- Decision: RuneCode isolate delivery uses typed transport and hash-addressed
  artifacts.
  - Planned capture: RuneCode companion track from `alpha.4` onward
- Decision: RuneContext content must stay policy-neutral.
  - Planned capture: `alpha.1`, companion-track validation throughout
- Decision: RuneContext starts with `Plain` and `Verified`; `Anchored` is later.
  - Planned capture: `alpha.6`, `post-mvp.md`
- Decision: standards updates must remain reviewable and path-referenced.
  - Planned capture: `alpha.3`, `alpha.7`
- Decision: every substantive work item gets a minimum change shape first.
  - Planned capture: `alpha.3`
- Decision: shaped changes should default to `design.md` and `verification.md`,
  while `tasks.md` and `references.md` remain supplemental and are only created
  when needed.
  - Planned capture: `alpha.3`
- Decision: multiple non-closed changes may exist concurrently; RuneContext
  should not require one repository-wide active-change slot.
  - Planned capture: `alpha.3`, `alpha.5`
- Decision: large or high-risk work should usually escalate from minimum mode to
  full mode early, and tooling should be able to recommend that escalation when
  a new change is obviously too large, ambiguous, or risky.
  - Planned capture: `alpha.3`, `alpha.5`
- Decision: lifecycle state and change shape are separate axes; `planned` should
  not imply full mode automatically, and shaping should be additive/idempotent.
  - Planned capture: `alpha.3`
- Decision: very large features may be modeled as one umbrella change plus
  linked sub-changes, with `related_changes` preserving navigation and
  directional `depends_on` links preserving prerequisite ordering.
  - Planned capture: `alpha.3`, `alpha.5`
- Decision: `proposal.md` is the canonical reviewable intent artifact.
  - Planned capture: `alpha.1`, `alpha.3`
- Decision: `standards.md` is always present and tooling-maintained.
  - Planned capture: `alpha.1`, `alpha.3`
- Decision: closing a change must not move it into an archive tree in v1.
  - Planned capture: `alpha.3`
- Decision: `superseded` is a terminal state distinct from `closed` and must use
  reciprocal successor/predecessor links.
  - Planned capture: `alpha.3`
- Decision: traceability must be strong enough for a future lineage/index view.
  - Planned capture: `alpha.3`, `alpha.4`, `post-mvp.md`
- Decision: structured machine validation should stay artifact-level, while
  markdown may use machine-validated `path#heading-fragment` deep refs instead
  of brittle line-number links.
  - Planned capture: `alpha.3`, `post-mvp.md`
- Decision: markdown deep-ref parsing and rewrite flows must ignore fenced code
  blocks, reject non-root-relative path forms, and use documented deterministic
  rewrite semantics.
  - Planned capture: `alpha.3`
- Decision: external URLs containing `.md#fragment` are not local deep refs,
  alpha.3 addressable headings use ATX `#` syntax, and machine-indexed markdown
  targets remain the indexed change/spec/decision/standards content areas.
  - Planned capture: `alpha.3`
- Decision: change ID slugs and derived heading fragments must remain ASCII-safe
  so non-ASCII authored text never produces invalid machine identifiers.
  - Planned capture: `alpha.3`
- Decision: alpha.3 traceability remains artifact-level and intentionally does
  not yet enforce a stricter originating-vs-revision semantic mirror contract.
  - Planned capture: `alpha.3`, `alpha.4`
- Decision: alpha.3 lifecycle helpers are forward-only and do not define a
  dedicated reopen/downgrade workflow.
  - Planned capture: `alpha.3`
- Decision: promotion is selective and reviewable, not silent auto-promotion.
  - Planned capture: `alpha.4`
- Decision: alpha.4 close-time promotion assessment records only `none` or
  `suggested`; explicit later workflows own `accepted` and `completed`.
  - Planned capture: `alpha.4`, `alpha.6`
- Decision: close-time promotion reassessment must preserve already-advanced
  promotion states (`accepted`, `completed`) and remain deterministic across
  both `closed` and `superseded` terminal change outcomes.
  - Planned capture: `alpha.4`
- Decision: close-time suggested promotion target paths are sourced from
  normalized traceability records so `target_path` values remain canonical and
  platform-independent.
  - Planned capture: `alpha.4`
- Decision: users must be able to use embedded or dedicated-repo storage.
  - Planned capture: `alpha.2`, `alpha.8`
- Decision: bundle rules and generated inventories use consistent
  RuneContext-root-relative paths, while the aspect key constrains the allowed
  subtree and mismatches fail closed.
  - Planned capture: `alpha.2`, `alpha.4`
- Decision: generated inventories should live at standard optional paths
  (`runecontext/manifest.yaml`, `runecontext/indexes/changes-by-status.yaml`,
  `runecontext/indexes/bundles.yaml`) and use closed schemas without becoming
  the source of truth.
  - Planned capture: `alpha.4`
- Decision: `type: path` remote/CI invalidity should be controlled by explicit
  caller mode rather than environment inference.
  - Planned capture: `alpha.2`
- Decision: adapters are the primary UX; CLI is the power-user and automation
  surface.
  - Planned capture: `alpha.3`, `alpha.5`, `alpha.7`
- Decision: `adapter pack` is the packaged tool UX surface, and
  `runecontext/operations/` is the canonical in-project reference/source area
  for underlying RuneContext operations.
  - Planned capture: `alpha.1`, `alpha.7`
- Decision: the release workflow should mirror RuneCode's tag-driven
  build/publish structure, with Nix defining the canonical unsigned release
  asset set.
  - Planned capture: `alpha.8`
- Decision: Linux/macOS `runectx` binary archives should be shipped as signed and
  attested convenience assets without replacing the canonical repo bundles.
  - Planned capture: `alpha.8`
- Decision: adapter packs ship with the selected RuneContext release, and
  `runectx adapter sync <tool>` materializes them locally rather than fetching
  them implicitly from GitHub.
  - Planned capture: `alpha.7`, `alpha.8`
- Decision: adapter packs should automatically run `runectx validate` after
  edits to authoritative RuneContext files and surface failures immediately.
  - Planned capture: `alpha.7`
- Decision: `runectx` must not make network calls outside explicit `init` and
  `update` flows.
  - Planned capture: `alpha.7`, `alpha.8`
- Decision: `plain` and `verified` should share one authored workflow, and
  standalone `runectx` must be able to emit the same portable minimal receipts a
  Verified repo requires while RuneCode adds richer parallel evidence.
  - Planned capture: `alpha.6`

## Deferred But Captured

The following ideas are intentionally captured in planning but deferred beyond
the MVP:

- `Anchored` assurance tier
- richer lineage/index views
- package-manager distribution channels
- optional stricter pinned-glob mode
- optional prompt-hygiene/content-safety heuristics

Those are grouped in `docs/implementation-plan/post-mvp.md`.
