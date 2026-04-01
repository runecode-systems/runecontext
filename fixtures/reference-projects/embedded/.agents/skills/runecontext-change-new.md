---
name: runecontext-change-new
description: Create a new RuneContext change
---
<!-- runecontext-managed-artifact: host-native-v1 -->
<!-- runecontext-tool: codex -->
<!-- runecontext-kind: flow_asset -->
<!-- runecontext-id: runecontext:change-new -->
# RuneContext Skill: change new

- canonical_flow_source: `build/generated/adapters/codex/flows/change-new.md`
- adapter_role: `flow_asset`
- operation_identifier: `runecontext:change-new`
- command_path: `change new`
- usage: `runectx change new [--json] [--non-interactive] [--dry-run] [--explain] --title TITLE --type TYPE [--size SIZE] [--bundle ID] [--shape minimum|full] [--description TEXT] [--path PATH]`
- required_flags: `--title --type`
