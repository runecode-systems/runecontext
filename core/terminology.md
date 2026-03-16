# Terminology And Naming

This document freezes the core RuneContext terminology for `v0.1.0-alpha.1`.

## Naming Policy

- Normative docs, schemas, and machine-readable contracts should use the
  canonical terms in this document.
- Customer-facing docs and adapters should prefer the smaller working set of
  `project`, `standards`, `bundles`, and `changes` for day-to-day UX.
- `project context` remains the umbrella architecture term for the broader
  durable knowledge layer.
- Avoid bare `context` in normative writing when a more specific term is
  available.
- `context bundle` may be shortened to `bundle` after first mention when the
  meaning is unambiguous.
- `context pack` may be shortened to `pack` after first mention when the
  meaning is unambiguous.

## Canonical Core Terms

| Term | Definition | Preferred customer-facing wording |
| --- | --- | --- |
| `standard` | A reusable normative document describing a rule, practice, or convention. | `standard` |
| `project context` | The broader durable knowledge layer that includes project files, standards, context bundles, changes, specs, and decisions. | Usually avoid in day-to-day UX; prefer `project`, `standards`, `bundles`, or `changes` as appropriate. |
| `context bundle` | A named reusable selector of project-context inputs across one or more aspect families that may inherit from other bundles. | `bundle` |
| `context pack` | A resolved deterministic artifact generated from one or more context bundles plus optional change or project inputs for runtime use. | Usually `context pack`; use `pack` only after first mention. |
| `change` | A proposed or in-flight body of work with lifecycle state and stable identity. | `change` |
| `spec` | A stable current document describing a feature or subsystem after durable knowledge is promoted out of changes. | `spec` |
| `decision` | An ADR-like durable architectural or policy decision. | `decision` |

## Product And Surface Names

| Term | Meaning |
| --- | --- |
| `RuneContext` | The product, repository, and portable format family. |
| `runectx` | The CLI surface. The CLI is important, but it is not the canonical definition of RuneContext semantics. |
| `adapter` | A packaged tool-specific integration layer that maps a host tool's UX onto RuneContext operations without redefining the format. |
| `adapter pack` | The packaged prompts, skills, commands, workflow docs, or similar host-tool materials that a user invokes inside a tool. An adapter pack is part of the adapter UX, not a competing source of truth. |
| `operations reference` | The canonical in-project reference/source material for underlying RuneContext operations, stored under `runecontext/operations/`. |

## Storage And Assurance Terms

| Term | Meaning | Machine value |
| --- | --- | --- |
| `embedded mode` | RuneContext lives inside the project repository. | `embedded` |
| `linked mode` | The project points to an external RuneContext source. | `git` or `path` in `source.type` |
| `Plain` | Default lightweight assurance tier for standalone/manual/tool-assisted use. | `plain` |
| `Verified` | Assurance tier with generated baseline and receipt artifacts. | `verified` |
| `Anchored` | Future assurance tier reserved for post-MVP work. | `anchored` |

Customer-facing prose may use `Plain mode` and `Verified mode`, but persisted
config values stay lowercase.

## Disambiguation Rules

- Use `project context` for the overall durable knowledge layer.
- Use `context bundle` for reusable selectors.
- Use `context pack` for resolved runtime input.
- Use `model context window` when referring to LLM token context.
- Use `lifecycle status` when referring to a change's `status` field in prose.

## Terms Not Used As Primary Terms

- Do not use `profile` as the primary term for a reusable selector. RuneCode
  already uses approval-profile terminology, and the concepts must stay
  distinct.
- Do not use `bundle` and `pack` interchangeably.
- `command pack` may appear in older design materials, including the read-only
  `docs/project_idea.md`, but the normative term is `adapter pack`.
- Do not use `task`, `ticket`, or `issue` as the primary RuneContext lifecycle
  noun when the portable artifact is a `change`.
- Do not use `archive` as the primary model for closed changes in v1. Closing a
  change is a lifecycle transition, not a relocation into a separate archive
  tree.

## User-Facing Language Guidance

- Adapters and customer-facing docs should usually say `project`, `standards`,
  `bundles`, and `changes`.
- Architecture docs, schemas, and integration docs may use `project context`
  when they need the broader umbrella term.
- `context pack` is mainly an advanced/runtime/integration term. Use it in user
  docs only when discussing deterministic resolution, machine-readable output,
  or RuneCode integration.
