# Tasks

- Expand `ChangeStatusEntry` with relationship, verification, and recency fields needed by human status rendering.
- Extend `BuildProjectStatusSummary` to populate those fields from validated change records.
- Preserve the existing active, closed, and superseded grouping while adding deterministic recency data for later sorting.
- Add tests proving the richer summary data does not break the current flat `status --json` contract.
