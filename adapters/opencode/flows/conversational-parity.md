# OpenCode Conversational Parity

This document defines conversational-flow parity for the OpenCode adapter.

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

- Hosts must declare the interaction surface they expose for this adapter using the capability guidance gathered under "Recommended Branch Cut 4: Remaining tool-specific adapters, compatibility mode, and parity hardening" in `docs/implementation-plan/adapter-host-capabilities.md`. Keeping those declarations explicit ensures adapters remain compatible with every supported host.
- Prompts: if the host does not support prompts, the adapter falls back to presenting CLI commands and asking for reviews, mirroring the same command-proposal cadence a shell user would follow.
- Shell access: without shell execution privileges, the adapter never runs `runectx` commands itself and instead guides reviewers to run the corresponding CLI call manually.
- Hooks: hosts that cannot register hooks are handled by deferring validation to the next `runectx validate` run and surfacing the same diagnostics as a hook-equipped host.
- Dynamic suggestions: when suggestion APIs are unavailable, the adapter uses the stable completion metadata produced by RuneContext and renders it as static text or numbered choices written into the conversation history.
- Structured output: hosts that only accept plain text do not receive structured JSON/diagnostic payloads; the adapter instead emits the same human-readable summary and reverts to the CLI contract for any required downstream automation.
