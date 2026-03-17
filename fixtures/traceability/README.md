# Traceability Fixtures

These fixtures exercise the strict alpha.1 traceability foundation for `changes/`, `specs/`, and `decisions/`.

- `valid-project/` is a self-contained embedded-mode project fixture with consistent cross-artifact references.
- `reject-*` directories each model one fail-closed traceability violation.
- `specs/*.md` and `decisions/*.md` use YAML frontmatter with IDs that must match the path-relative stem.
- Project fixtures may also exercise root-config rules that affect hand-authored artifacts, such as extension opt-in behavior.
