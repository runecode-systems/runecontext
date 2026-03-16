# Layer Boundaries And Ownership

This document defines the ownership boundary between RuneContext Core, adapters,
CLI surfaces, and RuneCode integration.

## Boundary Rule

RuneContext has one canonical portable core model. Everything else is a surface
over or consumer of that model.

## Layer Model

| Layer | Owns | Does not own |
| --- | --- | --- |
| RuneContext Core | On-disk layout, markdown/yaml/json conventions, schemas, bundle semantics, change lifecycle semantics, generated artifact conventions, naming contracts, and trust-boundary rules. | Tool-specific UX, runtime permissions, approvals, provider integrations, or RuneCode-only audit behavior. |
| RuneContext adapters | Tool-specific adapter packs, prompts, skills, bootstrap helpers, and UX translation into core operations. | Redefining file formats, changing lifecycle semantics, inventing adapter-only source-of-truth files, or making security decisions. |
| `runectx` CLI and future libraries | Implementing and exposing core operations for automation, debugging, and parity testing. | Becoming the source of truth for semantics or storing correctness-critical hidden state outside the repository. |
| RuneCode integration | Deterministic resolution into runtime-ready packs, audit/provenance binding, typed delivery into isolates, and audited workflow integration. | Mutating the portable RuneContext model into a RuneCode-only format or treating markdown as runtime authority. |

## Core Responsibilities

RuneContext Core is the authoritative home for:

- naming and disambiguation rules
- portable file and folder contracts
- schema and validation contracts
- bundle resolution and pack-generation semantics
- lifecycle and traceability rules
- policy-neutrality and LLM trust-boundary rules

## Adapter Responsibilities

Adapters may:

- present the workflow in a host tool's native style
- simplify wording for everyday use
- guide the user toward the same underlying core operations

Adapters may not:

- define alternate lifecycle states
- redefine what a bundle or pack means
- add hidden adapter-only correctness requirements
- grant capabilities, approvals, or runtime authority

## CLI And Library Placement Rule

- The CLI is an implementation surface, not a separate semantic authority.
- Future Go packages may implement the core rules, but the normative contract
  still lives in the portable docs and versioned schemas.
- Direct-library use and CLI use must converge on the same results.

## RuneCode Integration Rule

- RuneCode should be the best runtime for RuneContext, not the only runtime.
- RuneCode may consume bundles, changes, standards, specs, and decisions as
  project knowledge inputs.
- RuneCode must not treat RuneContext text as a substitute for signed manifests,
  typed policy inputs, approvals, or runtime protocol objects.

## Cross-Layer Invariants

- Core defines semantics once.
- Adapters vary UX only.
- CLI and direct-library flows must stay in parity with Core.
- RuneCode integration may derive runtime artifacts from RuneContext, but it may
  not redefine RuneContext's portable source model.
- Any trusted-side mapping from RuneContext metadata to stricter review posture
  must be explicit, allowlisted, and auditable.
