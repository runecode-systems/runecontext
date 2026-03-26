# Fixtures

This directory holds the repository-wide fixture taxonomy for RuneContext contract and behavior tests.

## Fixture Taxonomy

- `schema-contracts/`
  - standalone and project-level YAML fixtures for JSON Schema validation and YAML-profile rejection
- `markdown-contracts/`
  - human-readable markdown fixtures for `proposal.md` and `standards.md` structure validation
- `traceability/`
  - multi-file project fixtures for spec/decision frontmatter, path-matched IDs, and cross-artifact traceability checks
- `bundle-resolution/`
  - multi-file project fixtures and goldens for bundle inheritance, precedence, diagnostics, and path-boundary guardrails
- `source-resolution/`
  - source-mode, discovery, and structured resolution metadata fixtures for alpha.2
- `change-workflow/`
  - alpha.3 project fixtures and golden outputs for change creation, shaping,
    closing, supersession, and rewrite-safe history preservation
- `context-packs/`
  - alpha.4 context-pack and report goldens for deterministic output coverage
- `generated-indexes/`
  - alpha.4 manifest and generated index golden outputs
- `reference-projects/`
  - alpha.8 reference fixtures for embedded, linked, verified, and monorepo scenarios

## Reserved Future Fixture Families

These directories are reserved by convention for later alphas so tests and tooling can grow without reshuffling the tree:

- `assurance/`
- `cli-json/`
- `adapters/`

## Storage Conventions

- Prefer `valid-*` and `reject-*` naming for single-file fixtures.
- Prefer one self-contained project tree per cross-artifact fixture case.
- Keep fixtures reviewable and hand-authored unless a fixture is intentionally generated as a golden output.
- Put helper explanations in README files, not in ad hoc sidecar notes.
- Reuse the same fixtures across Go and future TypeScript/RuneCode parity suites where the contract is shared.

## Coverage Expectations

- New machine-readable semantics land with schema fixtures in the same milestone.
- New markdown/document contracts land with parser fixtures in the same milestone.
- New cross-artifact rules land with project fixtures in the same milestone.
- New user-facing write or CLI flows land with integration or end-to-end fixtures in the same milestone.
- RuneCode-facing contract changes land with parity fixtures in the same milestone.

## Release Rule

RuneContext does not treat tests or fixtures as cleanup work. New semantics must not land without the corresponding tests or fixtures in the same milestone.
