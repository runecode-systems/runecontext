# Portable Layout And Ownership

This document freezes the authoritative portable RuneContext layout for
`v0.1.0-alpha.1`.

## Scope

- `runecontext.yaml` lives at the project root.
- In embedded mode, the portable RuneContext tree lives under `./runecontext/`.
- In linked mode, the same portable tree is resolved from the selected source.
- This document defines reserved path names, ownership, and generation rules.
- This document freezes reserved path names and ownership rules for the MVP
  layout, even when validation or automation for some paths is implemented in a
  later milestone.

## Ownership Dimensions

Each path is described using the following dimensions:

- `authority`
  - `canonical source`: portable source of truth
  - `derived artifact`: generated output that must never become the sole source
    of truth
  - `generated evidence`: generated assurance/provenance artifact
- `authoring mode`
  - `hand-authored`: primarily written directly by humans, even if tooling may
    scaffold it
  - `reviewable canonical`: tooling may refresh it, but it remains canonical,
    reviewable, and manually editable
  - `mixed`: a subtree that intentionally contains both hand-authored and
    reviewable canonical files
  - `generated`: machine-generated and not hand-authored as the primary workflow
- `requiredness`
  - `required`: required for the relevant mode or artifact shape
  - `optional`: allowed but not required
  - `conditional`: required only when a feature, mode, or artifact shape is in
    use

## Root Configuration

| Path | Authority | Authoring mode | Requiredness | Notes |
| --- | --- | --- | --- | --- |
| `runecontext.yaml` | canonical source | hand-authored | required | Project-root configuration for version, assurance tier, and source selection. |

`runecontext.yaml` is part of the authoritative RuneContext model even though it
does not live inside the embedded `runecontext/` folder.

## Reserved Top-Level Portable Tree

| Path | Authority | Authoring mode | Requiredness | Notes |
| --- | --- | --- | --- | --- |
| `runecontext/project/` | canonical source | hand-authored | optional | Durable project-wide knowledge such as mission, roadmap, and stack. |
| `runecontext/standards/` | canonical source | hand-authored | optional | Reusable normative documents. |
| `runecontext/bundles/` | canonical source | hand-authored | optional | Reusable context selectors. |
| `runecontext/changes/` | canonical source | mixed | optional | Change folders remain at stable paths across their lifecycle. |
| `runecontext/specs/` | canonical source | hand-authored | optional | Stable current-state subsystem or feature specs. |
| `runecontext/decisions/` | canonical source | hand-authored | optional | Durable ADR-like decisions. |
| `runecontext/operations/` | canonical source | hand-authored | optional | Canonical in-project reference/source material for underlying RuneContext operations. |
| `runecontext/schemas/` | derived artifact | generated | optional | Versioned schema assets distributed with RuneContext releases. These are not a substitute for the authoritative schema sources in this repository. |
| `runecontext/assurance/` | generated evidence | generated | conditional | Exists only when Verified assurance is enabled. |
| `runecontext/manifest.yaml` | derived artifact | generated | optional | Regenerable inventory/index output; never the sole source of truth. |

## Change Folder Contract

Each change lives at a stable path under `runecontext/changes/<change-id>/`.

### Minimum Shape

| Path | Authority | Authoring mode | Requiredness | Notes |
| --- | --- | --- | --- | --- |
| `runecontext/changes/<change-id>/status.yaml` | canonical source | hand-authored | required | Machine-readable lifecycle and traceability state. Tooling may scaffold and update it, but it remains canonical project state. |
| `runecontext/changes/<change-id>/proposal.md` | canonical source | hand-authored | required | Canonical reviewable intent artifact. |
| `runecontext/changes/<change-id>/standards.md` | canonical source | reviewable canonical | required | Always present. Tooling may refresh it, but it remains canonical and user-reviewable. |

### Full-Mode Additions

| Path | Authority | Authoring mode | Requiredness | Notes |
| --- | --- | --- | --- | --- |
| `runecontext/changes/<change-id>/design.md` | canonical source | hand-authored | conditional | Materialized only when deeper shaping is needed. |
| `runecontext/changes/<change-id>/tasks.md` | canonical source | hand-authored | conditional | Detailed implementation/task breakdown. |
| `runecontext/changes/<change-id>/references.md` | canonical source | hand-authored | conditional | External references and related links. |
| `runecontext/changes/<change-id>/verification.md` | canonical source | hand-authored | conditional | Verification notes and evidence pointers. |

## Assurance Artifacts

| Path | Authority | Authoring mode | Requiredness | Notes |
| --- | --- | --- | --- | --- |
| `runecontext/assurance/baseline.yaml` | generated evidence | generated | conditional | Generated when Verified mode is enabled. Commit policy is mode-dependent, but the file remains machine-generated. |
| `runecontext/assurance/receipts/context-packs/` | generated evidence | generated | conditional | Usually ephemeral receipt family. |
| `runecontext/assurance/receipts/changes/` | generated evidence | generated | conditional | Change-event evidence receipts. |
| `runecontext/assurance/receipts/promotions/` | generated evidence | generated | conditional | Promotion evidence receipts. |
| `runecontext/assurance/receipts/verifications/` | generated evidence | generated | conditional | Verification evidence receipts. |

## Reserved-Name Rule

- The top-level names in this document are reserved for RuneContext's portable
  model.
- Tooling and adapters should not repurpose those paths for unrelated hidden
  state.
- Additional files may exist, but they must not change the meaning of the
  canonical paths defined here.

## Tool-Specific Files Boundary

- Tool-specific adapter-synced files, runtime files, and host-tool configuration
  should live outside the portable `runecontext/` tree.
- The portable `runecontext/` tree may contain canonical shared source material,
  including `runecontext/operations/`, but it should not become a dumping
  ground for host-specific runtime state.
- If an adapter needs tool-specific files such as `.claude/` content or other
  host-tool config, those files should live in the tool's own expected location
  rather than inside `runecontext/`.

## Review-Only Outputs

- Review-only proposed diffs, diagnostics, or advisory reports are valid
  workflow outputs.
- Those outputs are not part of the authoritative v1 portable on-disk layout.
- No review-only report may become a hidden correctness-critical source of
  truth.

## Layout Invariants

- Closed changes stay at stable paths under `runecontext/changes/`.
- Generated files remain derived from canonical source or generated evidence
  rules; they do not replace the canonical source files.
- `standards.md` is a reviewable canonical file, not a disposable generated
  cache.
- `manifest.yaml` is optional and regenerable.
- `assurance/` is conditional on Verified mode and must not be required for
  Plain-mode correctness.
