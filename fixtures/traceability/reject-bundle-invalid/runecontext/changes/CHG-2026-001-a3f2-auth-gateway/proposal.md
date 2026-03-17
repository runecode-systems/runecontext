## Summary
Exercise invalid bundle schema rejection.

## Problem
The project bundle is missing required fields.

## Proposed Change
Keep the rest of the project valid so validation fails on the bundle schema.

## Why Now
Whole-project validation should include bundles.

## Assumptions
The change itself remains valid.

## Out of Scope
Fixing the broken bundle.

## Impact
This fixture proves bundle validation is part of `runectx validate`.
