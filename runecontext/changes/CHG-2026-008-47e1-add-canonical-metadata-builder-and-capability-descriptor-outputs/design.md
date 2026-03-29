# Design

## Overview
Implement the core metadata work in one contracts-driven pipeline: add a closed versioned schema for the capability descriptor, define a typed Go builder that derives descriptor content from canonical CLI registry and contract/release metadata, expose that descriptor via a dedicated `runectx metadata` command, and embed the identical payload into the release manifest.

## Shape Rationale
- Large, ambiguous, or high-risk feature work should move to full mode early.

## Core Output Rules
- The descriptor is a derived artifact and must not become a second semantic authority.
- Unknown descriptor schema versions, unknown fields, incompatible layout/profile tokens, and unknown enum values fail closed.
- Compatibility reporting separates supported project versions from explicit upgrade edges.
- Runtime/layout reporting includes both repo-bundle and installed share-layout discovery facts.
