# Local API v1

The UI and CLI share the same Go application services. The API is a private process-local transport for the embedded browser interface, not a remotely deployable service.

## Bootstrap and authentication

`planparity ui` generates a random 256-bit token and prints or opens a URL shaped like:

```text
http://127.0.0.1:8765/#token=<ephemeral-token>
```

The fragment is not part of HTTP requests. The frontend reads it once, keeps it only in JavaScript memory, clears it from the address bar, and sends it through `X-PlanParity-Token` on protected API calls. Restarting the server invalidates every prior token and result.

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | Liveness and contract version |
| `GET` | `/api/v1/system` | Build, OS, execution mode, persistence policy |
| `GET` | `/api/v1/engines` | Local Terraform/OpenTofu availability |
| `POST` | `/api/v1/inspect` | Start state inspection |
| `POST` | `/api/v1/compare` | Start deterministic state comparison |
| `POST` | `/api/v1/rehearsals/plan` | Start restricted planning |
| `POST` | `/api/v1/rehearsals/upgrade` | Compare two engine versions |
| `POST` | `/api/v1/rehearsals/migrate` | Compare Terraform with OpenTofu |
| `GET` | `/api/v1/jobs/{id}` | Read job status |
| `GET` | `/api/v1/jobs/{id}/events` | Replay and stream SSE progress |
| `DELETE` | `/api/v1/jobs/{id}` | Request cancellation |
| `GET` | `/api/v1/results/{id}` | Read a completed sanitized result |
| `GET` | `/api/v1/results/{id}/export` | Download sanitized JSON |

All endpoints except the three read-only health/system endpoints require the token. Requests reject unknown JSON fields and bodies over 1 MiB by default. See `schemas/api-v1.json` for request contracts.

## Job lifecycle

Start endpoints return `202 Accepted` with a random job ID and status, events, and result URLs. Statuses are `QUEUED`, `RUNNING`, `COMPLETED`, `FAILED`, and `CANCELLED`. SSE clients can reconnect with `Last-Event-ID`; stored events are replayed before live events.

Jobs and results are memory-only, capped per process, and discarded on shutdown. A failed job returns a generic public failure rather than captured engine output.

## Browser and network boundary

- Only numeric loopback hosts can bind or pass host validation.
- Origins must match the active server or an explicit development origin.
- State-changing requests require the session token, providing CSRF protection.
- Analysis responses disable caching.
- CSP, frame denial, MIME sniffing protection, referrer isolation, and cross-origin isolation headers are set globally.
- There is no filesystem-browsing endpoint and no endpoint accepts arbitrary executable names.

