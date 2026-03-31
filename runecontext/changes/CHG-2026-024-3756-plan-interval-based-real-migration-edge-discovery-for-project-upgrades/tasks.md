# Tasks

- Refine upgrade planning to collect ordered explicit migration edges inside the current-to-target interval without inventing synthetic hops.
- Preserve zero-hop preview and staged version-bump-only apply when no explicit migration edge lies inside the requested interval.
- Ensure fresh projects initialized at the installed version write only the current version\'s canonical layout and never replay historical migration edges.
- Add planner and apply tests covering interval-based migration discovery, skipped historical migrations for fresh projects, and direct upgrade preview/apply behavior when only later edges are real.
