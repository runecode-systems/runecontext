# Design

## Overview
Adopt a shared generated workflow contract for all adapter-native commands and skills. Replace the flat shared flows catalog with per-flow structured definitions that capture required outcome, guardrails, inputs to gather, decision rules, workflow steps, stop condition, recommended next commands, and examples. Extend tools/syncadapters to generate both richer flow markdown and a machine-readable workflow.json for every adapter pack. Update adapter render-host-native to load generated workflow.json from the installed/generated adapter root so shell-injection hosts render the same rich structure at runtime, while non-shell-injection hosts embed the same rendered content at sync time. Keep adapter sync as the only user-facing setup command and treat adapter render-host-native as an internal rendering primitive. Require every generated command or skill to stop after its immediate RuneContext outcome and only recommend the next RuneContext command instead of auto-chaining. Preserve the no-question-tool rule across all adapters.

## Shape Rationale
- Full mode was requested explicitly to deepen the change.
- Minimum mode is sufficient for the current size and risk signal.
