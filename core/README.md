# Core Contracts

This directory holds the normative RuneContext core contracts for
`v0.1.0-alpha.1` Epic 1.

These documents freeze the naming, ownership, layout, and trust-boundary rules
that later schemas, fixtures, CLI behavior, adapters, and RuneCode integration
must share.

## Normative Documents

- `core/terminology.md`
  - canonical project terminology, naming policy, and customer-facing wording
- `core/boundaries.md`
  - ownership boundaries between RuneContext Core, adapters, CLI surfaces, and
    RuneCode integration
- `core/layout.md`
  - authoritative portable on-disk layout plus ownership and generation rules
- `core/trust-boundaries.md`
  - policy-neutrality and untrusted-LLM-input rules

## Usage Rules

- These documents are normative unless a later versioned contract explicitly
  replaces them.
- Customer-facing docs and adapters should prefer the simpler nouns `project`,
  `standards`, `bundles`, and `changes` for day-to-day UX.
- `project context` remains the architecture/spec umbrella term.
- Normative writing should avoid bare `context` when a more specific term is
  available.
