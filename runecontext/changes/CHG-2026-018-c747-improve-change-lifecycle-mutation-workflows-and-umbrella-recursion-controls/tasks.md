# Tasks

- Define the umbrella scope and safety rules for lifecycle mutation updates and explicit recursive umbrella propagation.
- Deliver non-terminal verification-status mutation support through CHG-2026-019-c1af-allow-change-update-to-record-completed-verification-state.
- Deliver explicit recursive umbrella propagation through CHG-2026-020-75ba-add-explicit-recursive-umbrella-lifecycle-propagation-for-sub-changes.
- Preserve explicit command semantics: no hidden recursive behavior by default, no propagation to all related_changes, and fail-closed lifecycle validation for all affected records.
