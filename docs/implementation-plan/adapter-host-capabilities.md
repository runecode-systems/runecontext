# Adapter Host Capabilities

This document captures the normative adapter compatibility-mode guidance referenced
by the alpha.7 `Recommended Branch Cut 4: Remaining tool-specific adapters,
compatibility mode, and parity hardening` workstream.

Adapters must describe the interaction surfaces their host tools can expose before
attempting to exercise those capabilities. At minimum, each host should make an
explicit declaration covering the following capability classes:

1. `prompts`: whether the host can surface and capture guided prompts.
2. `shell_access`: whether the host can launch shell helpers or run CLI commands
   on the user's behalf.
3. `hooks`: whether the host can register and run scripts such as validation or
   mutation hooks.
4. `dynamic_suggestions`: whether the host exposes inline suggestion/autocomplete
   APIs derived from the canonical completion metadata.
5. `structured_output`: whether the host can consume JSON/diagnostic payloads
   for downstream integrations.

If a host cannot offer a capability, the adapter must fall back to the documented
downgrade behavior instead of inventing hidden semantics:

- **Prompt fallback**: prompt-driven flows degrade into static guidance plus the
  equivalent explicit `runectx` commands, so conversational skips still read like
  reviewable CLI proposals.
- **Shell fallback**: hosts without shell access show the command steps and
  candidate data instead of running helpers, leaving execution to the user via
  the documented CLI contract.
- **Hook fallback**: when hooks are unavailable, validation behavior is deferred
  to the next explicit `runectx validate` run, and the adapter surfaces the same
  diagnostics it would have produced had a hook run.
- **Dynamic suggestion fallback**: hosts lacking suggestion APIs present the
  canonical completion metadata as static text or numbered choices derived from
  the shared registry rather than guessing inline completions.
- **Structured output fallback**: hosts that accept only text receive a
  human-readable summary plus the underlying CLI flag/value set that encodes the
  same semantic payload, so automation consumers still have a reviewable
  alternative.

Treat this document as the stable compatibility-mode contract for all adapters,
and update it whenever the capability expectations evolve.
