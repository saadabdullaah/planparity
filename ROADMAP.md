# PlanParity roadmap

The roadmap is organized around independently useful, testable releases. Dates are intentionally omitted; release quality is controlled by exit gates.

## Phase 0 — Scope lock

- Product charter and explicit non-goals.
- Apache-2.0 licensing.
- Initial backlog and safety policy.

## Phase 1 — Repository foundation

- Go CLI and React/TypeScript workspace.
- Windows and Linux build/test scripts.
- Formatting, linting, unit testing, version command, embedded UI proof, and architecture decisions.

## Phase 2 — Models and fixtures

- Normalized state, plan, finding, comparison, and engine models.
- Versioned report schema.
- Terraform/OpenTofu, corrupted-state, and secret-canary fixtures.

## Phase 3 — Engines and doctor (`v0.1.0-alpha.1`)

- Terraform and OpenTofu discovery, explicit paths, versions, capabilities, timeouts, and normalized errors.
- `planparity doctor` on Windows and Linux.

## Phase 4 — State inspection (`v0.1.0`)

- `show -json`, sanitized normalization, inventories, dependencies, terminal and JSON reporting.
- Byte-for-byte preservation of supplied state and secret-canary coverage.

## Phase 5 — State comparison (`v0.2.0`)

- Exact matching, structural/provider/dependency changes, probable move inference, confidence evidence, and deterministic output.

## Phase 6 — Restricted rehearsal (`v0.3.0`)

- Copied workspace, temporary home, reduced environment, command allowlist, `init -backend=false`, `plan -refresh=false`, normalized plans, output limits, and cleanup verification.

## Phase 7 — Upgrade and parity (`v0.4.0`)

- Old/new engine and provider comparison, state readability, plan differences, Terraform/OpenTofu parity, and preserved diagnostics.

## Phase 8 — Local API and UI (`v0.5.0-alpha.1`)

- Localhost-only HTTP server, embedded UI, session token, security middleware, jobs, server-sent events, engine/health endpoints, inspection/comparison endpoints, navigation, and sanitized state explorer.

## Later milestones

- Phase 9: complete visual comparison and rehearsal UI (`v0.5.0`).
- Phase 10: Windows/Linux and security hardening (`v0.9.0-beta.1`).
- Phase 11: documentation, four local labs, and stable release (`v1.0.0`).

## Version-one non-goals

No apply/remediation, hosted service, remote-state integration, accounts, external database, Terraform/OpenTofu downloading, CI product integration, policy language, plugins, Terragrunt, other IaC engines, AI analysis, or macOS support.

