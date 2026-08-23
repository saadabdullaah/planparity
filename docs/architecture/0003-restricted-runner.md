# ADR 0003: Allowlisted restricted runner

- Status: Accepted
- Date: 2026-08-23

## Context

Engine planning loads third-party provider code and can inspect state containing secrets. Refresh suppression alone does not prevent every network attempt or side effect.

## Decision

PlanParity constructs engine invocations from typed operations mapped to a closed command allowlist. It never accepts arbitrary argument vectors from API clients. Each rehearsal uses a copied workspace, copied state, temporary home and data directory, reduced environment, bounded output, deadline, and verified cleanup.

The runner is explicitly not described as an OS sandbox. Reports identify real cloud behavior as unverified unless independently observed through an approved future mechanism.

## Consequences

- Command construction is auditable and fuzzable.
- Original paths remain outside engine working directories.
- Provider code can still consume CPU, memory, or network within OS permissions; documentation and result labels must preserve that limitation.

