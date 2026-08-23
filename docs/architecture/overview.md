# Architecture overview

PlanParity is a local modular monolith. The CLI and embedded React UI call the same Go application services.

```text
CLI ───────────────┐
                   ├─ Application service ─ Normalize / Compare / Risk / Report
React UI -> API/SSE┘                         |
                                             v
                                Terraform/OpenTofu adapters
                                             |
                                             v
                                Restricted process runner
```

## Packages

- `internal/command` owns Cobra commands and user-facing flag validation.
- `internal/server` owns the loopback listener, security middleware, API routes, SSE, and embedded assets.
- `internal/app` coordinates jobs and workflows. It is the shared boundary for CLI and UI.
- `internal/engine` discovers and adapts Terraform/OpenTofu executables.
- `internal/runner` constructs the allowlisted argument vectors and copied execution environment.
- `internal/normalize` converts raw engine JSON into value-free deterministic models.
- `internal/compare` performs exact matching, plan/schema comparison, and bounded move inference.
- `internal/risk` converts observed evidence into findings.
- `internal/report` writes terminal and JSON representations.
- `internal/model` defines the versioned public contract.

## Request sequence

1. The interface validates a typed request.
2. The application resolves a supported engine adapter.
3. Inspection reads a regular local state. Rehearsal creates independent copied workspaces.
4. The runner executes only an allowlisted command shape with bounded resources.
5. Raw JSON is normalized immediately and discarded.
6. Comparison/risk logic creates deterministic evidence and findings.
7. CLI reports or authenticated API responses expose only the sanitized model.

## Design records

- [ADR 0001: local-first modular monolith](0001-local-first-modular-monolith.md)
- [ADR 0002: external engine boundary](0002-external-engine-boundary.md)
- [ADR 0003: restricted runner](0003-restricted-runner.md)
- [ADR 0004: versioned sanitized contracts](0004-versioned-sanitized-contracts.md)

See [the safety model](../security/safety-model.md) for trust boundaries and [the local API guide](../local-api.md) for endpoints.
