# Contract Validation Foundation

This package provides the executable alpha.1 validation foundation for RuneContext.

## Current Coverage

- JSON Schema validation for machine-readable YAML contracts
- restricted YAML profile checks for duplicate keys and anchors/aliases
- strict markdown parsing for `proposal.md` and `standards.md`
- strict YAML-frontmatter validation for `specs/*.md` and `decisions/*.md`
- project-level traceability checks across changes, bundles, specs, and decisions
- content-root-aware project validation that follows `runecontext.yaml` source settings

## Intentional Scope

- This package enforces the frozen alpha.1 contracts only.
- It does not yet implement source resolution, CLI commands, standards discovery,
  promotion flows, or assurance workflows.
- Later alphas should build on this package rather than re-encoding contract
  rules ad hoc.
