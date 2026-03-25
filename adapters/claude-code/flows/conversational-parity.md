# Claude Code Conversational Parity

This document defines conversational-flow parity for the Claude Code adapter.

## Mapping Rule

Each conversational flow must map to explicit `runectx` operations and stable
CLI inputs. No adapter-only lifecycle or mutation semantics are allowed.

## Flow Mappings

- `change new` conversation -> `runectx change new --title --type [--size] [--shape] [--bundle] [--description] [--path]`
- `change shape` conversation -> `runectx change shape CHANGE_ID [--design] [--verification] [--task] [--reference] [--path]`
- `standard discover` conversation -> `runectx standard discover [--path] [--change CHANGE_ID] [--confirm-handoff] [--target TYPE:PATH]`
- `promote` conversation -> `runectx promote CHANGE_ID [--accept|--complete] [--target TYPE:PATH] [--path]`

## Candidate Data Rule

- Standards-discovery candidate targets come from `runectx standard discover`
  output fields such as `candidate_promotion_target_*`.
- Promotion candidate/target values are explicit user-visible values and remain
  reusable across turns.
- The adapter must not rely on hidden session-only target state.

## Reviewability

- Conversations should end with reviewable command proposals or emitted output.
- Durable mutations remain `runectx` writes so standard diffs are visible in git.

## Host Capabilities

- Hosts must declare the interaction surface they expose for this adapter by listing the capabilities documented under "Recommended Branch Cut 4: Remaining tool-specific adapters, compatibility mode, and parity hardening" in `docs/implementation-plan/adapter-host-capabilities.md`. Explicit capability declarations keep adapter behavior aligned with each host's constraints.
- Prompts: a prompt-less host receives the same command proposals as a shell user would; the adapter falls back to static guidance plus the matching `runectx` CLI commands for each step.
- Shell access: when the host cannot run shell helpers, the adapter never executes `runectx` commands directly and instead guides reviewers to invoke the explicit CLI calls it would otherwise run.
- Hooks: hosts without hook support defer validation to the next explicit `runectx validate` call, and the adapter surfaces the same diagnostics it would produce with a hook-enabled host.
- Dynamic suggestions: if suggestion APIs are unavailable, the adapter uses the canonical completion metadata and enumerates CLI-friendly numbered choices inside the conversation instead of relying on inline auto-complete.
- Structured output: hosts that cannot consume structured JSON instead receive a human-readable summary accompanied by the specific CLI flags/values necessary to reproduce the same effect, ensuring the same semantics survive manual automation.
