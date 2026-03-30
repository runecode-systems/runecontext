# Tasks

- Introduce an explicit migration-registry model that can describe ordered upgrade hops instead of only exact from-to edge membership.
- Implement deterministic path search for upgrade preview and apply planning, including clear failure output when no path exists.
- Extend preview output contracts to expose hop_count, hop_N_from, hop_N_to, and readable per-hop action summaries while keeping upgrade read-only.
- Document that upgrade preview is the assessment path and that apply remains the only mutating command.
- Add tests covering direct-edge paths, multi-hop paths, alias targets, current-version no-op previews, and fail-closed no-path results.
