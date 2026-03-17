# Source Resolution Fixtures

These fixtures cover the alpha.2 source-resolution and discovery slice.

- `embedded-project/`: embedded-mode project with a local `runecontext/` tree.
- `path-project/`: local `type: path` project with a sibling RuneContext tree.
- `monorepo/`: nearest-ancestor discovery fixture with both root and nested `runecontext.yaml` files.
- `templates/minimal-runecontext/`: reusable RuneContext tree copied into dynamic git-source test repos.
- `golden/*.yaml`: expected structured source-resolution metadata snapshots.

The golden files focus on the shared source-resolution result shape used by tests:

- selected config path
- project root
- source root
- source mode
- source ref
- resolved commit when applicable
- verification posture
- diagnostics

Tests replace `${PROJECT_ROOT}` and `${COMMIT}` placeholders at runtime.
