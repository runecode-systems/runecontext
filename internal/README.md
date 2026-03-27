# Internal

This directory holds shared Go packages for RuneContext implementation code.

## Current Packages

- `internal/cli/`
  - CLI parsing, metadata registries, machine-facing output contracts, and
    command behavior for validation, diagnostics, assurance, adapters,
    release/install UX, and upgrade flows
- `internal/contracts/`
  - shared validation, source-resolution, bundle, standards, change-workflow,
    context-pack, generated-artifact, and MVP-readiness implementation that the
    CLI and tests build on
