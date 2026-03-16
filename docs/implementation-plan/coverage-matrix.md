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
| Research Findings / Design Implications | Planning principles, progressive disclosure, reviewable diffs | `README.md`, `alpha.3`, `alpha.6`, `alpha.7` | Yes |
| Goals | Acceptance criteria and MVP boundaries | `README.md`, `mvp-acceptance.md` | Yes |
| Non-Goals | Scope guardrails and post-MVP separation | `README.md`, `post-mvp.md` | Yes |
| Product Decomposition | Core/adapters/RuneCode repository boundary | `alpha.1` | Yes |
| Why The Three Layers Need To Exist | Boundary enforcement and ownership rules | `alpha.1`, `alpha.7` | Yes |
| Packaging, Repositories, And Releases | Repo structure and release model | `alpha.1`, `alpha.8` | Yes |
| Releases and Installation | Install lanes, update flow, compatibility matrix | `alpha.8` | Yes |
| Optional Assurance And Verifiable Tracing | Plain/Verified model, baseline, receipts, backfill | `alpha.5` | Yes |
| RuneCode Context And Integration Constraints | Companion-track test and contract checklist | `alpha.1`-`alpha.8`, `mvp-acceptance.md` | Yes |
| Usage Scenarios | Validation of local, remote, and non-RuneCode flows | `alpha.2`, `alpha.6`, `alpha.8` | Yes |
| Terminology | Normative glossary and naming rules | `alpha.1` | Indirect |
| Why `context bundle` Was Chosen | Naming and user-facing vocabulary | `alpha.1` | Indirect |
| Storage Modes | Embedded, linked, path, and monorepo behavior | `alpha.2` | Yes |
| Project Root Configuration | Root schema, versioning, assurance tier | `alpha.1`, `alpha.2` | Yes |
| Embedded Mode | Resolver implementation and reference fixture | `alpha.2`, `alpha.8` | Yes |
| Linked Mode | Commit, mutable ref, signed tag, and path handling | `alpha.2`, `alpha.8` | Yes |
| Monorepo Support | Discovery and reference fixtures | `alpha.2`, `alpha.8` | Yes |
| Core On-Disk Layout | Authored/generated ownership and scaffolding | `alpha.1`, `alpha.6` | Indirect |
| Machine-Readable Schema Versioning | Schema behavior, unknown-field handling, YAML profile | `alpha.1` | Yes |
| Generated Artifact Commit Policy | Assurance and release/install policy | `alpha.5`, `alpha.8` | Yes |
| Project Files | Project-context scaffolding and conventions | `alpha.1`, `alpha.6` | Indirect |
| Standards | Frontmatter rules, lifecycle, migration behavior | `alpha.3` | Yes |
| Context Bundles | Bundle schema and resolution engine | `alpha.1`, `alpha.2` | Yes |
| Changes | ID allocation and lifecycle workflow | `alpha.3` | Yes |
| Minimum And Full Change Shapes | Progressive disclosure and file materialization | `alpha.3` | Yes |
| Branching Logic By Work Type And Size | Intake depth, escalation, and minimum/full mode branching | `alpha.3` | Yes |
| When To Ask More Vs Less | Prompting heuristics and assumption capture | `alpha.3` | Yes |
| Proposal.md Structure | Generator/parser/validator rules | `alpha.1`, `alpha.3` | Yes |
| Standards.md Structure | Normalized structure and auto-maintenance | `alpha.1`, `alpha.3` | Yes |
| Automatic Standards Maintenance | Change creation/shaping refresh and review diffs | `alpha.3`, `alpha.6` | Yes |
| Stable Specs | Durable spec conventions and traceability metadata | `alpha.3` | Yes |
| Decisions | Durable decision conventions and traceability metadata | `alpha.3` | Yes |
| Traceability And Future Lineage | Minimum traceability now, richer lineage later | `alpha.3`, `alpha.4`, `post-mvp.md` | Yes |
| Context Bundle Semantics | Ordered rule application and validation rules | `alpha.2` | Yes |
| Deterministic Resolved Output | Context pack generation, provenance, pack hash | `alpha.4` | Yes |
| Standards Membership And Authoring Model | Manual versus assisted standards workflow | `alpha.3` | Yes |
| Change Lifecycle | State machine and close rules | `alpha.3` | Yes |
| Promotion / Promotion Assessment | Structured promotion suggestions and close flow | `alpha.4` | Yes |
| Archive/Promotion Rule | Stable-path close behavior | `alpha.3` | Yes |
| Historical Traceability Requirements | Future-safe linkage and readable history | `alpha.3`, `alpha.4` | Yes |
| Minimal Process And User Experience | Small mental model and progressive disclosure | `alpha.3`, `alpha.6`, `alpha.7` | Yes |
| Invocation Surfaces And Command Architecture | CLI surface, command semantics, flags | `alpha.6`, `alpha.7` | Yes |
| Adapters | Thin adapters and capability model | `alpha.7` | Yes |
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
- Decision: project root config carries `runecontext_version` and
  `assurance_tier`.
  - Planned capture: `alpha.1`, `alpha.2`
- Decision: change IDs use `CHG-YYYY-NNN-RAND-short-slug`.
  - Planned capture: `alpha.3`
- Decision: context packs must include a top-level canonical hash.
  - Planned capture: `alpha.4`
- Decision: RuneCode isolate delivery uses typed transport and hash-addressed
  artifacts.
  - Planned capture: RuneCode companion track from `alpha.4` onward
- Decision: RuneContext content must stay policy-neutral.
  - Planned capture: `alpha.1`, companion-track validation throughout
- Decision: RuneContext starts with `Plain` and `Verified`; `Anchored` is later.
  - Planned capture: `alpha.5`, `post-mvp.md`
- Decision: standards updates must remain reviewable and path-referenced.
  - Planned capture: `alpha.3`, `alpha.7`
- Decision: every substantive work item gets a minimum change shape first.
  - Planned capture: `alpha.3`
- Decision: `proposal.md` is the canonical reviewable intent artifact.
  - Planned capture: `alpha.1`, `alpha.3`
- Decision: `standards.md` is always present and tooling-maintained.
  - Planned capture: `alpha.1`, `alpha.3`
- Decision: closing a change must not move it into an archive tree in v1.
  - Planned capture: `alpha.3`
- Decision: traceability must be strong enough for a future lineage/index view.
  - Planned capture: `alpha.3`, `alpha.4`, `post-mvp.md`
- Decision: promotion is selective and reviewable, not silent auto-promotion.
  - Planned capture: `alpha.4`
- Decision: users must be able to use embedded or dedicated-repo storage.
  - Planned capture: `alpha.2`, `alpha.8`
- Decision: adapters are the primary UX; CLI is the power-user and automation
  surface.
  - Planned capture: `alpha.6`, `alpha.7`
- Decision: `adapter pack` is the packaged tool UX surface, and
  `runecontext/operations/` is the canonical in-project reference/source area
  for underlying RuneContext operations.
  - Planned capture: `alpha.1`, `alpha.7`

## Deferred But Captured

The following ideas are intentionally captured in planning but deferred beyond
the MVP:

- `Anchored` assurance tier
- richer lineage/index views
- package-manager distribution channels
- optional stricter pinned-glob mode
- optional prompt-hygiene/content-safety heuristics

Those are grouped in `docs/implementation-plan/post-mvp.md`.
