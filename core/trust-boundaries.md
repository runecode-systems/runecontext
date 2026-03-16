# Trust Boundaries

This document codifies the policy-neutrality and LLM-input trust-boundary rules
for RuneContext.

## Policy Neutrality Rule

RuneContext content is project knowledge and workflow guidance. It is not a
runtime authority.

### Allowed Uses

RuneContext content may influence:

- context assembly
- presentation and summarization
- review suggestions
- change shaping and traceability
- advisory warnings and diagnostics

### Forbidden Uses

RuneContext content must not directly:

- influence policy-engine allow or deny results
- choose approval profiles
- widen capabilities
- bypass typed manifests, approvals, or runtime protocol objects
- act as an executable privileged workflow definition

### Trusted-Side Mapping Rule

If a trusted component maps RuneContext metadata into a stricter review posture,
that mapping must be:

- explicit
- allowlisted
- auditable
- unable to convert a denied action into an allowed action on its own

## LLM Input Trust Boundary

RuneContext content is also untrusted model input.

This includes standards, decisions, proposals, specs, and any linked RuneContext
content fetched from another repository or local path.

### Required Assumption

Implementations must assume that RuneContext text may contain prompt-injection
attempts, misleading instructions, or compromised content.

### Primary Defenses

Protection against RuneContext-based prompt injection must rely on:

- typed policy and manifest boundaries
- broker/runtime boundaries
- approval gates
- isolation constraints
- deterministic machine-readable contracts

### Supplementary Defenses

Prompt-hygiene scanning or content-safety heuristics may be added later, but
they are supplementary defenses rather than the primary trust boundary.

## Testable Invariants For Later Epics

Later schemas, fixtures, CLI behavior, adapters, and RuneCode integration should
be testable against the following invariants:

- Selecting a bundle or resolving a pack does not grant capabilities.
- Standards, changes, specs, and decisions may guide review, but they do not
  become policy authority.
- A compromised or linked RuneContext source is treated as untrusted content,
  not as trusted runtime control input.
- Protective behavior relies on typed boundaries and fail-closed behavior, not
  on trusting the text itself.

## Consequences For Surfaces

- Adapters may explain workflows and surface relevant files, but they must not
  imply that RuneContext text can approve or authorize runtime behavior.
- The CLI may validate, resolve, explain, and scaffold, but it must not create
  hidden policy side channels.
- RuneCode may bind RuneContext-derived artifacts into audit flows, but it must
  keep runtime trust decisions in RuneCode's own typed security model.
