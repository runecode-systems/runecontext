## Summary
Add explicit recursive umbrella lifecycle propagation for sub-changes

## Problem
Teams using umbrella project changes currently have to repeat lifecycle mutations across each feature sub-change manually. At the same time, automatically mutating every `related_changes` entry would be unsafe because `related_changes` is a navigability field, not an implicit recursive target set.

## Proposed Change
Add an explicit recursive option, such as `--recursive`, for `change update` and `change close` so a project umbrella can optionally cascade lifecycle mutations to its associated feature sub-changes only.

## Why Now
The need shows up naturally once umbrella/sub-change workflows are used heavily: people want coordinated lifecycle bookkeeping, but they also need the CLI to preserve explicit relationship semantics and avoid broad unsafe cascades.

## Assumptions
- Used all non-draft standards as a conservative fallback because no standards were selected through context bundles.
- Recursive propagation must be opt-in and transactional.
- Eligible targets are only feature sub-changes associated with the selected project umbrella.

## Out of Scope
- Default recursive behavior.
- Propagation to non-feature related changes, dependencies, or unrelated graph neighbors.
- Weakening lifecycle validation to allow partial cascades.

## Impact
The change gives umbrella owners an explicit, reviewable way to keep project and feature lifecycle state aligned without turning `related_changes` into a hidden mutation hierarchy.
