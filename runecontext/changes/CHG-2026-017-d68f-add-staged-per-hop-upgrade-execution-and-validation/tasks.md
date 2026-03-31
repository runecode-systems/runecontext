# Tasks

- Define a typed per-hop migration interface with explicit apply and verify phases for version transitions that need migration logic beyond a plain runecontext_version rewrite.
- Execute upgrade apply against a staged project copy, supporting both zero-hop version-bump-only upgrades and migration-required hop chains while failing closed on the first migration or verification error.
- Replace real project files only after the full staged chain, final pinned-version rewrite, final validation, and managed-artifact checks succeed; otherwise roll back cleanly.
- Support default final version-bump behavior for simple compatible transitions while allowing dedicated migration code for hops that rewrite files, layouts, or managed artifacts.
- Add tests covering zero-hop version-bump apply success, multi-hop apply success, per-hop validation failure rollback, hop-specific verification failure rollback, and staged managed-artifact refresh behavior.
