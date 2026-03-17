## Summary
Add a strict markdown validator foundation for RuneContext change intent artifacts.

## Problem
Alpha.1 froze the `proposal.md` contract in prose, but the repository did not yet have executable parsing and validation coverage for it.

## Proposed Change
Implement a Go validator that parses required sections in exact order and fails closed when a section is missing, duplicated, empty, or reordered.

## Why Now
Later CLI and adapter work will depend on this contract, so the repository needs a tested parser before more surfaces build on it.

## Assumptions
The validator can treat fenced code blocks as regular section content once the surrounding section boundary is known.

## Out of Scope
Authoring richer proposal semantics beyond the frozen alpha.1 section contract.

## Impact
This gives the repo deterministic parser fixtures and a foundation for future `runectx validate` behavior.

## Additional Context
Extra sections remain allowed after the required block.
