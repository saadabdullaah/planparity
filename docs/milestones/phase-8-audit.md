# Phase 8 completion audit

This audit records the evidence for the `v0.5.0-alpha.1` local API and UI foundation. It covers the master plan through Phase 8; later visualization, platform-hardening, labs, packaging, and 1.0 documentation remain on the public roadmap.

## Phase gates

| Phase | Exit evidence |
| --- | --- |
| 0 — Scope | Public name is PlanParity; charter, non-goals, Apache-2.0 license, roadmap, and frozen v1 scope are present. |
| 1 — Foundation | Go CLI and embedded React workspace build on Windows; Linux scripts and a two-OS CI matrix use the same commands. `version` and `ui` are wired into the executable. |
| 2 — Models | Terraform and OpenTofu fixtures normalize into the same versioned, value-free contracts. Malformed fixtures fail safely and canary values are absent. |
| 3 — Engines | Real Terraform 1.12.1 and OpenTofu 1.12.1 binaries report versions and all required capabilities through `doctor`. Missing/invalid paths return typed public failures. |
| 4 — Inspect | Real-engine inspection leaves the state hash unchanged. A populated raw-state canary fixture returns its resource address without returning the seeded value. |
| 5 — Compare | Identical states have zero differences; move, ambiguity, structural, provider, and dependency behavior is deterministic. A 10,000-resource comparison benchmark remains well below the target on the development machine. |
| 6 — Runner | Only version, validate, show, provider schema, backend-disabled init, and refresh-disabled plan operations are constructible. Copied inputs remain byte-identical and temporary workspaces are removed. |
| 7 — Parity | Real plan, same-engine upgrade, and Terraform/OpenTofu migration rehearsals complete. Tests cover plan mismatch, new deletion/replacement, provider-schema differences, decode failure, and unverified provider isolation. |
| 8 — API/UI | A compiled server binds to loopback, authenticates with a fragment-bootstrapped memory-only token, completes a job over the real API, and returns a sanitized one-row resource list. Raw canary and session token are absent from browser content. |

## Automated verification

- `go test ./...`
- `go vet ./...`
- `npm run lint`
- `npm run test:run`
- `npm run build`
- Windows `scripts/test.ps1` and `scripts/build.ps1`
- GitHub Actions on Windows and Linux
- CodeQL for Go and JavaScript/TypeScript
- Dependency review and weekly Dependabot updates

## Safety assertions

- Original state modified: **NO**
- Original configuration modified: **NO**
- Apply executed by PlanParity or its audit: **NO**
- Destroy executed: **NO**
- State push executed: **NO**
- Raw state stored in browser persistence: **NO**
- Seeded canary observed in public result or UI content: **NO**

The restricted runner is still not an operating-system sandbox. Providers and data sources can attempt network access or local execution even when refresh is disabled, so every rehearsal reports that boundary as unverified.

