# Tasks

- Define the umbrella scope and safety model for real multi-hop upgrade planning with staged execution.
- Deliver path planning, preview output, and explicit migration registry semantics through CHG-2026-016-e00c-add-multi-hop-upgrade-path-planning-and-preview-contracts.
- Deliver staged per-hop execution, per-hop validation, and final atomic replacement through CHG-2026-017-d68f-add-staged-per-hop-upgrade-execution-and-validation.
- Deliver compatibility refinement and version-bump-only project upgrade behavior through CHG-2026-022-58d8-refine-project-upgrade-compatibility-and-version-bump-only-semantics.
- Deliver explicit CLI release-check and self-update flows through CHG-2026-023-2423-add-explicit-cli-self-update-and-release-check-flows.
- Keep the command contract explicit: preview-only project `runectx upgrade`, mutating-only project `runectx upgrade apply`, explicit CLI `runectx upgrade cli` self-update flows, and fail-closed behavior when a project requires a newer CLI release.
