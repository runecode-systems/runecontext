# Tasks

- Refactor upgrade planning to distinguish compatible older projects, migration-required hops, and project-newer-than-cli failures.
- Allow zero-hop version-bump-only project upgrades when the project is compatible with the installed CLI but no migration edge exists.
- Reserve explicit migration edges for transitions that require real migration logic rather than plain version bumps.
- Update validate and doctor diagnostics so older compatible projects recommend running project upgrade, while newer-than-cli projects explicitly require upgrading the CLI binary.
- Add tests covering no-migration version bump apply, migration-required apply, older-compatible preview, and project-newer-than-cli failure messaging.
