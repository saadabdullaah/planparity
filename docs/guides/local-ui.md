# Local UI guide

Start the UI with:

```powershell
planparity ui
```

PlanParity binds to numeric loopback only. The printed URL contains an ephemeral fragment token. Open the complete URL; the application keeps the token in memory and removes it from browser history.

## Home

The home screen reports detected Terraform/OpenTofu engines, links to the three workflow families, states the non-mutation contract, and keeps a process-memory-only recent list.

## Inspect

Choose an engine and absolute state path. Results provide an overview, module/provider inventory, searchable resource table, dependency graph, findings, and sanitized JSON export.

## Compare

Choose before/after state paths. Results provide summary counts, matched changes with before/after structure, probable moves with evidence, an impact graph, provider differences, and filtered findings.

## Rehearse

Choose Plan, Upgrade, or Parity. Advanced controls expose explicit executable paths, proxy inheritance, and registry authorization. Results show plan actions, compatibility checks, provider schema differences, limitations, findings, and side-by-side engine summaries.

## Browser data boundary

- Results and history stay in JavaScript memory only.
- No local storage, session storage, IndexedDB, service worker, analytics, CDN, or remote font is used.
- Raw state/plan values are never returned by the API.
- Export downloads only the sanitized server result.

Restart the process to invalidate the session and clear all jobs/results.
