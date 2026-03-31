# Tasks

- Update assurance enable, capture, backfill, shared receipt writers, and related CLI output so alpha.13 writes only runecontext/assurance artifacts.
- Update assurance validation and indexing to treat runecontext/assurance as canonical post-alpha.13 layout.
- Implement the real alpha.12 to alpha.13 staged migration that moves legacy assurance artifacts into runecontext/assurance and rewrites baseline imported_evidence backfill paths.
- Add apply and validation tests covering successful migration, imported_evidence rewrites, staged rollback on failure, and canonical alpha.13 post-upgrade validation.
