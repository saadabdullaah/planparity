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

## Phase 9 - Visualization and rehearsal UI (`v0.5.0`)

- Complete state, comparison, probable-move, dependency/impact graph, rehearsal, parity, findings, export, responsive, and accessible interface.

## Phase 10 - Windows/Linux and security hardening (`v0.9.0-beta.1`)

- Cross-platform path, executable, process-tree, temporary workspace, link, permission, output-limit, environment, injection, and secret-boundary coverage.

## Phase 11 - Documentation, labs, and stable release (`v1.0.0`)

- Complete installation, build, workflow, UI, safety, limitations, troubleshooting, architecture, and contribution documentation.
- Four deterministic offline labs covering inspection, module moves, upgrade replacement, and Terraform/OpenTofu mismatch.
- Reproducible Windows/Linux release archives, checksums, and automated GitHub releases.

All version-one phase deliverables are implemented. Future work must remain outside the frozen version-one scope unless it replaces an existing commitment.

## Version-one non-goals

No apply/remediation, hosted service, remote-state integration, accounts, external database, Terraform/OpenTofu downloading, CI product integration, policy language, plugins, Terragrunt, other IaC engines, AI analysis, or macOS support.
