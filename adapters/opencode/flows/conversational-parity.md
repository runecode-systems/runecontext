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
