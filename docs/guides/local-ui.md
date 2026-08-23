# Local UI guide

Start the UI with:

```powershell
planparity ui
```

The current directory is the browser API's path root. Start PlanParity from the directory that contains the configurations and state snapshots you want to analyze, or set an explicit root:

```text
planparity ui --path-root D:\infra
```

The terminal prints the resolved root. State and configuration inputs outside that tree, including links that resolve outside it, are rejected before a job starts. Explicit Terraform and OpenTofu executable paths remain independent because those tools are commonly installed elsewhere.

PlanParity binds to numeric loopback only. The printed URL contains an ephemeral fragment token. Open the complete URL; the application keeps the token in memory and removes it from browser history.

## Home

The home screen reports detected Terraform/OpenTofu engines, links to the three workflow families, states the non-mutation contract, and keeps a process-memory-only recent list.

## Inspect

Choose an engine and a state path inside the printed UI path root. Results provide an overview, module/provider inventory, searchable resource table, dependency graph, findings, and sanitized JSON export.

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
