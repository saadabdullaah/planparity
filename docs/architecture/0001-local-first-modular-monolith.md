# ADR 0001: Local-first modular monolith

- Status: Accepted
- Date: 2026-08-23

## Context

The CLI and browser UI require identical engine execution, normalization, comparison, risk, and reporting behavior. A distributed architecture would add deployment, authentication, persistence, and secret-handling surfaces that conflict with the local-only product promise.

## Decision

PlanParity is a single Go executable organized as internal modules. CLI commands call application services directly. The web UI calls the same services through a localhost-only HTTP API. Compiled frontend assets are embedded in the executable.

Jobs and results are memory-resident. The browser receives sanitized domain models only. No hosted service or external database is introduced.

## Consequences

- One release artifact contains the backend and UI.
- Service boundaries remain testable without network calls.
- Process lifetime defines result lifetime.
- Horizontal scaling and multi-user concerns are intentionally irrelevant.

