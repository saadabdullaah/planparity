# ADR 0002: External engine boundary

- Status: Accepted
- Date: 2026-08-23

## Context

Terraform and OpenTofu are external executables with overlapping but evolving capabilities and output formats. Calling them throughout application code would make version behavior, security controls, and testing inconsistent.

## Decision

All engine behavior is exposed through one capability-aware interface. Terraform and OpenTofu have separate adapters. Raw output is contained within the adapter/normalization boundary and discarded after conversion to value-free normalized models.

Unsupported capabilities return typed results. They are not converted into generic failures or silently omitted.

## Consequences

- Engine differences remain explicit and testable.
- Version-aware decoding can evolve independently.
- Fake engines can drive deterministic application, API, and UI tests.

