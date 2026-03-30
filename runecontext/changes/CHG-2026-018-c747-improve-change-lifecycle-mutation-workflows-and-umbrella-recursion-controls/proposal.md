## Summary
Improve change lifecycle mutation workflows and umbrella recursion controls

## Problem
The current lifecycle command split has two practical gaps. First, `change update` can target `status=verified`, but it cannot record a completed `verification_status`, which makes a valid open-but-verified change awkward or impossible to express through the CLI. Second, umbrella project changes cannot explicitly cascade lifecycle mutations to their feature sub-changes, which leaves teams to repeat equivalent updates manually or invent ad hoc conventions outside the CLI.

## Proposed Change
Track the lifecycle-mutation improvement as an umbrella over two linked deliverables:

- `CHG-2026-019-c1af-allow-change-update-to-record-completed-verification-state` for non-terminal verification-state mutation support.
- `CHG-2026-020-75ba-add-explicit-recursive-umbrella-lifecycle-propagation-for-sub-changes` for opt-in recursive umbrella propagation limited to associated feature sub-changes.

## Why Now
Recent dogfooding made both gaps concrete: verified work often wants to remain open long enough for normal lifecycle bookkeeping, and umbrella changes often need coordinated lifecycle transitions across their sub-changes. Both problems affect ordinary repository maintenance and should be captured before more ad hoc local workarounds accumulate.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Recursive lifecycle propagation must be explicit and opt-in rather than hidden default behavior.
- Only feature sub-changes associated with a project umbrella are eligible recursive targets; arbitrary `related_changes` entries are not.

## Out of Scope
- Automatic recursive lifecycle propagation by default.
- Propagation to every `related_changes` entry regardless of change type or umbrella membership.
- Changing the meaning of terminal close or promotion assessment.

## Impact
The umbrella keeps non-terminal verification-state fixes and explicit recursive umbrella propagation aligned so lifecycle commands remain safe, explicit, and implementable without hidden hierarchy assumptions.
