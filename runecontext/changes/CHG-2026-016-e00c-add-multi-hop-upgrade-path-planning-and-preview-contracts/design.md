# Design

## Overview
Replace the current single-edge target check with deterministic multi-hop path planning. The preview command must stay read-only and compute a concrete ordered sequence of upgrade hops from the project runecontext_version to the chosen target version. Planning should fail closed if no path exists, preserve the distinction between supported project versions and explicit migration edges, and report hop_count, ordered hop transitions, and per-hop plan actions so users can review exactly what apply would do before any mutation occurs.

## Planning Rules
- Migration edges should be modeled explicitly and searched as a graph rather than treated as exact target membership checks only.
- Path selection must be deterministic so preview output and apply behavior are stable across runs.
- Preview should support direct-edge and multi-hop paths using the same planner.
- Missing paths fail closed with clear guidance instead of silently choosing a version bump.

## Preview Contract
- `runectx upgrade` remains read-only and should act as the reviewable dry-run surface for upgrade planning.
- Preview output should include at least the current version, target version, state, hop count, ordered hop transitions, and per-hop plan actions.
- Preview should keep existing flat structured output conventions so scripts and adapters can consume the new fields without a separate rendering path.

## Migration Registry Direction
- The registry should be able to describe ordered transitions, not only whether a direct edge exists.
- Registry entries should be reusable by apply-time migration execution without duplicating transition knowledge in multiple places.
