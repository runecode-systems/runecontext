# Tasks

- Define a typed per-hop migration interface with explicit apply and verify phases for version transitions that need migration logic beyond a plain runecontext_version rewrite.
- Execute upgrade apply against a staged project copy, validating the staged tree after each hop and failing closed on the first migration or verification error.
- Replace real project files only after the full hop chain, final validation, and managed-artifact checks succeed; otherwise roll back cleanly.
- Support default no-op hop logic for simple version transitions while allowing dedicated migration code for hops that rewrite files, layouts, or managed artifacts.
- Add tests covering multi-hop apply success, per-hop validation failure rollback, hop-specific verification failure rollback, and staged managed-artifact refresh behavior.
