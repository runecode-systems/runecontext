# Change Workflow Fixtures

These fixtures capture the expected alpha.3 change-authoring outputs for the thin
change and status commands.

- `template-project/` is the embedded project used by the change-creation tests.
- `golden/minimum-change/` captures the minimum change shape.
- `golden/shaped-change/` captures automatic full shaping for project work.
- `golden/supplemental-change/` captures shaped work with optional `tasks.md`
  and `references.md` plus a standards refresh.
- `golden/closed-change/` captures a closed change at a stable path.
- `golden/superseded-change/` captures a superseded change folder that remains
  readable in place.
