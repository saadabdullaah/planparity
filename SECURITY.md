# Security policy

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability or secret exposure. Use GitHub private vulnerability reporting for this repository. Include affected versions, reproduction steps, impact, and any suggested mitigation. Never attach real infrastructure state or credentials.

## Security boundary

PlanParity is a local analysis tool. It is intentionally incapable of requesting the following Terraform or OpenTofu operations:

- `apply`
- `destroy`
- `state push`
- `state rm`
- `import`
- `taint`
- `untaint`
- `force-unlock`

Allowed engine commands are defined centrally and tested. Supplied workspaces and state files are copied before rehearsal execution. The runner uses a temporary home, a reduced environment, timeouts, output limits, and best-effort cleanup.

PlanParity is not an operating-system sandbox. Providers and data sources execute third-party code and can attempt network access even when planning uses `-refresh=false`. Reports must identify these limits and must not describe unverified behavior as safe.

Every rehearsal reports:

```text
Execution mode: LOCAL-RESTRICTED

PlanParity copied the workspace, disabled refresh, removed known
credentials, and avoided apply operations. This is not an operating
system security sandbox and does not guarantee zero network activity.
```

## Sensitive data rules

- Never commit genuine state files, plans, credentials, tokens, or private configuration.
- Test fixtures contain synthetic values only.
- Fixture secrets use documented canaries that tests assert never appear in logs, reports, API responses, the DOM, or browser storage.
- Raw engine JSON is normalized in the backend and is not persisted or returned by the API.
- State attribute values are omitted; structural type information and sensitivity markers may be retained.

## Supported versions

Until the first stable release, users should build from the latest tagged pre-release and should not use PlanParity against irreplaceable state without an independent backup.

