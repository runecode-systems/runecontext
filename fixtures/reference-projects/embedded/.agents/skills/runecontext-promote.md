---
name: runecontext-promote
description: Advance RuneContext promotion state
---
<!-- runecontext-managed-artifact: host-native-v1 -->
<!-- runecontext-tool: codex -->
<!-- runecontext-kind: flow_asset -->
<!-- runecontext-id: runecontext:promote -->
# RuneContext Skill: promote

- canonical_flow_source: `adapters/codex/flows/promote.md`
- adapter_role: `flow_asset`
- operation_identifier: `runecontext:promote`
- command_path: `promote`
- usage: `runectx promote [--json] [--non-interactive] [--dry-run] [--explain] CHANGE_ID [--accept | --complete] [--target TYPE:PATH (summary auto-filled per target type)] [--path PATH]`
- required_positionals: `CHANGE_ID`
