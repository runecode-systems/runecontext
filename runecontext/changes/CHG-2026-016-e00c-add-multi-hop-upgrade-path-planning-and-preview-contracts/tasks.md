# Tasks

- Introduce an explicit migration-registry model that can describe ordered upgrade hops instead of only exact from-to edge membership.
- Implement deterministic planning for project upgrade preview, including zero-hop version-bump-only upgrades, migration-required multi-hop paths, and clear failure output when the project is newer than the installed CLI.
- Extend preview output contracts to expose hop_count, hop_N_from, hop_N_to, and readable per-hop action summaries while keeping upgrade read-only and clearly differentiating version-bump-only plans from migration-required plans.
- Document that upgrade preview is the assessment path and that apply remains the only mutating command.
- Add tests covering direct-edge paths, multi-hop paths, compatible older version-bump-only previews, alias targets, current-version no-op previews, and project-newer-than-cli fail-closed results.
