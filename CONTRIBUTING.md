# Contributing to PlanParity

Correctness, determinism, and protection of infrastructure state take priority over convenience.

## Development workflow

1. Create a focused branch from `main`.
2. Keep commits small, descriptive, and independently understandable.
3. Add or update tests for every behavioral change.
4. Run backend, frontend, secret-canary, and formatting checks locally.
5. Open a pull request explaining behavior, safety impact, and verification.

Use Conventional Commit-style subjects, for example `feat(compare): detect probable resource moves` or `test(runner): reject apply commands`.

## Required checks

Windows:

```powershell
.\scripts\test.ps1
.\scripts\build.ps1
```

Linux:

```bash
./scripts/test.sh
./scripts/build.sh
```

## Engineering rules

- Never introduce an execution path for `apply`, `destroy`, or state mutation.
- Never log raw engine JSON or attribute values.
- Do not add a dependency without documenting why the standard library is insufficient.
- Keep Terraform and OpenTofu behind the shared engine interface.
- Preserve deterministic ordering in normalized data and reports.
- Treat unsupported behavior explicitly; never silently downgrade it to success.
- New API fields require schema and compatibility consideration.
- Windows and Linux remain first-class.

## Fixtures

Fixtures must be synthetic, deterministic, and safe to publish. Any fake secret must use a recognizable canary and be covered by an assertion that the canary cannot escape sanitized boundaries.

## Scope discipline

Version-one scope is frozen in [ROADMAP.md](ROADMAP.md). A new version-one feature must replace or defer another feature of similar cost. Hosted services, remote-state integrations, automated remediation, and additional IaC engines remain out of scope.

