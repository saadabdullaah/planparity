# Safety model

PlanParity is a local compatibility laboratory with a deliberately narrow authority boundary.

## Invariants

PlanParity has no command construction path for:

- `apply`
- `destroy`
- `state push`
- `state rm`
- `import`
- `taint` or `untaint`
- `force-unlock`

Allowed engine operations are a closed typed set. Arguments are passed directly to an executable without a shell.

## Input preservation

Inspection hashes a regular state file before and after engine decoding. Rehearsal hashes the original configuration tree and state, creates a private temporary workspace, copies only allowed regular files without following symbolic links/junctions, and verifies original hashes again before cleanup.

Copied configuration excludes state, plan, VCS, dependency-cache, and PlanParity-managed directories. Individual input files are bounded at 256 MiB and configuration copies are bounded at 100,000 files.

## Engine process boundary

Rehearsal provides a temporary home, temporary engine data directory, an allowlisted environment, per-command timeouts, bounded output, and process-tree termination. On Windows, child processes are attached to a kill-on-close Job Object. On Unix, engines run in a separate process group.

Known secret-like variables and cloud credential prefixes are removed. Proxy variables are inherited only when explicitly requested.

This process boundary is not a security sandbox. Provider plugins can execute local code. Use an operating-system sandbox or disposable machine when stronger isolation is required.

## Data minimization

Raw engine state, plan, and schema JSON is held only long enough to normalize it, then its byte buffer is cleared. Public models contain identity, structure, type names, sensitivity/unknown markers, fingerprints, actions, and sanitized diagnostics, but not ordinary values.

Terminal text strips control and bidirectional formatting characters. Browser rendering uses React text nodes with a strict content security policy and no raw HTML injection path.

## Local server

- Listener host must be a numeric loopback address.
- Host and Origin are validated to reduce DNS rebinding and cross-origin abuse.
- Non-public API requests require a cryptographically random process-local token.
- The token is delivered in the URL fragment, moved into memory, and not sent as part of the initial HTTP request.
- Results are held in bounded process memory and use `Cache-Control: no-store`.
- Requests are size-bounded and unknown JSON fields are rejected.
- The UI uses no browser persistence, analytics, CDN, or remote database.

The token grants authority to state and configuration files inside the UI path root printed at startup. The API normalizes each requested path and rejects traversal or links that resolve outside the root. Explicit engine binaries may live elsewhere under the operating-system account. Do not share the token. Restart PlanParity if it may have escaped.

## Interpretation boundary

PlanParity distinguishes `VERIFIED`, `INFERRED`, `UNVERIFIED`, `UNSUPPORTED`, and `FAILED`. Unknown behavior is never promoted to safe. A clean result is evidence about the tested local inputs and engines, not a production guarantee.

See [SECURITY.md](../../SECURITY.md) for vulnerability reporting.
