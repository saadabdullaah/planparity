# PlanParity

PlanParity is a local compatibility laboratory for Terraform and OpenTofu. It inspects state, compares snapshots, and rehearses engine or provider transitions in isolated working copies before production state is exposed.

![PlanParity connects two infrastructure interpretations through a local evidence graph](web/public/assets/planparity-hero.png)

> [!IMPORTANT]
> PlanParity never runs `apply`, `destroy`, or `state push`. It is an analysis tool, not an operating-system sandbox or a guarantee that a production change is safe.

## Project status

The version-one implementation is complete for Windows and Linux. It includes the full CLI, responsive local UI, restricted rehearsal runner, security hardening, offline labs, documentation, and release automation described in the [roadmap](ROADMAP.md).

Verification evidence is recorded in the [v1.0.0 completion audit](docs/milestones/v1.0.0-audit.md).

Start with the [five-minute quick start](docs/quickstart.md), then read the [safety model](docs/security/safety-model.md) before using real infrastructure files.

## Workflows

```text
planparity version
planparity doctor
planparity inspect --engine terraform --state terraform.tfstate
planparity compare --engine terraform --before before.tfstate --after after.tfstate
planparity rehearse plan
planparity rehearse upgrade
planparity rehearse migrate
planparity ui
```

The local UI is served at `http://127.0.0.1:8765` by default. PlanParity rejects non-loopback bind addresses.

The interface provides searchable state/module/provider inventories, dependency and impact graphs, before/after comparison, probable-move evidence, plan and schema parity views, filtered findings, in-memory recent analyses, and sanitized JSON export.

## Safety model

- Supplied configuration and state are treated as read-only.
- Engine execution happens in a copied temporary workspace.
- Known cloud credentials and secret-like environment variables are removed.
- State values are normalized away before reports or API responses are created.
- Raw engine JSON is never sent to the browser.
- Results distinguish verified, inferred, unverified, unsupported, and failed behavior.

Provider plugins and data sources can still attempt network access even when refresh is disabled. See [SECURITY.md](SECURITY.md) for the complete boundary.

## Development prerequisites

- Go 1.26 or newer.
- Node.js 24 LTS and npm for frontend development.
- Terraform and/or OpenTofu for integration tests and real analyses.
- PowerShell 7 on Windows or a POSIX shell on Linux.

## Build from source

Windows:

```powershell
.\scripts\build.ps1
.\bin\planparity.exe version
```

Linux:

```bash
./scripts/build.sh
./bin/planparity version
```

Run all checks with `scripts/test.ps1` or `scripts/test.sh`.

No-cloud demo fixtures are available in [labs](labs/README.md).

## Repository layout

- `cmd/planparity`: executable entry point.
- `internal`: application, engine, normalization, comparison, runner, reporting, and server packages.
- `web`: React and TypeScript local interface.
- `schemas`: versioned JSON contracts.
- `testdata`: sanitized deterministic fixtures.
- `docs`: architecture, commands, security, tutorials, and troubleshooting.

See the [documentation index](docs/README.md), [command reference](docs/commands.md), [local API guide](docs/local-api.md), and [troubleshooting guide](docs/troubleshooting.md).

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a change. Security issues must follow [SECURITY.md](SECURITY.md) rather than the public issue tracker.

## License

Apache License 2.0. See [LICENSE](LICENSE).
