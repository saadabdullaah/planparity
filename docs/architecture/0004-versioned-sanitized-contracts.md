# ADR 0004: Versioned sanitized contracts

- Status: Accepted
- Date: 2026-08-23

## Context

Raw state and plan JSON contain values that can be credentials or other sensitive information. CLI, API, reports, and UI need stable shared data without coupling to engine-specific formats.

## Decision

Normalization creates value-free domain models. Public JSON uses explicit schema version `1`. Attribute structure may contain type fingerprints, sorted key names, sensitivity paths, and unknown paths, but not ordinary values. All slices used in reports are deterministically ordered.

## Consequences

- Raw JSON cannot cross the application boundary.
- Contract changes require deliberate versioning.
- Secret canaries can validate every external output surface.

