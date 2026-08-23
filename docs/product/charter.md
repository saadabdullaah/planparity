# Product charter

## Mission

PlanParity helps Terraform and OpenTofu users understand whether local state, configuration, providers, and engine versions are structurally compatible before production state is exposed to a transition.

## Audience

- Infrastructure engineers preparing an engine or provider upgrade.
- Teams migrating between Terraform and OpenTofu.
- Module maintainers validating identity-preserving refactors.
- Reviewers who need evidence beyond a conventional plan summary.

## Version-one promise

Given local configuration, local state, and user-provided engine binaries, PlanParity produces sanitized, deterministic evidence about state readability, resource identity, structural changes, provider schemas, and plan behavior. It labels every conclusion by verification state.

## Constraints

- Local-only processing and localhost-only UI.
- No state mutation or infrastructure application.
- No cloud account, hosted database, or paid API requirement.
- Windows and Linux are first-class.
- The CLI and local web UI share the same Go application services.

## Success criteria

- A new user completes the inspection, refactor, provider-upgrade, and cross-engine labs without a cloud account.
- Seeded secret values never escape backend normalization.
- Original inputs remain byte-identical after every workflow.
- Unsupported or externally dependent behavior is never reported as verified.
- A 10,000-resource normalized snapshot can be inspected and compared within published performance targets.

