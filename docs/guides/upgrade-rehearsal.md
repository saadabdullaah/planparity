# Rehearse an engine or provider upgrade

An upgrade rehearsal runs a baseline and candidate executable against independent copies of the same configuration and state.

```powershell
planparity rehearse upgrade `
  --from-engine terraform `
  --from-bin C:\Tools\Terraform-Old\terraform.exe `
  --to-engine terraform `
  --to-bin C:\Tools\Terraform-New\terraform.exe `
  --config .\infra `
  --state .\infra\terraform.tfstate
```

Each run receives:

- A copied configuration directory and copied state.
- A temporary home and engine data directory.
- An allowlisted environment with known credentials removed.
- `init -backend=false` and `plan -refresh=false` only.
- A deadline, bounded stdout/stderr, process-tree termination, and cleanup.

PlanParity compares normalized plan actions and provider schema fingerprints. Review candidate-only deletes/replacements, changed actions, schema differences, failed checks, and every `UNVERIFIED` limitation.

`--allow-registry-access=false` records that registry access was not authorized. It is a declared operating mode, not a network sandbox. Providers remain third-party local code.
