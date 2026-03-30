# Design

## Overview
Add an explicit recursive option for umbrella lifecycle mutations, for example --recursive on change update and change close. Recursive propagation must be opt-in, never the default. It should apply only when the selected change is a project umbrella and should target only associated feature sub-changes that belong to that umbrella set, not every related_changes entry. The target set should be derived from a well-defined umbrella-sub-change relationship rule, and the command must fail closed if any targeted sub-change cannot take the requested transition. Recursive execution should be transactional across all affected status files so partial umbrella or sub-change mutation cannot persist.

## Target Selection Rules
- Recursive propagation should be available only when the selected root change is `type: project`.
- Eligible recursive targets should be limited to associated `type: feature` sub-changes for that umbrella.
- Generic `related_changes` entries must not be treated as recursive targets unless they satisfy the explicit umbrella/sub-change rule.

## Mutation Rules
- Non-recursive behavior remains the default.
- Recursive requests must surface which change IDs will be mutated before or during execution output.
- If any targeted sub-change cannot take the requested transition, the full recursive mutation must fail without partial writes.

## Command Scope
- `change update --recursive` should propagate non-terminal lifecycle mutations only to eligible feature sub-changes.
- `change close --recursive` should propagate terminal close behavior only to eligible feature sub-changes and require valid terminal metadata across the full target set.
