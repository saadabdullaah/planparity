# Command reference

PlanParity analyzes only local files. Every state-taking command requires an explicit path; reports default to the terminal and can be emitted as versioned JSON.

## Check the toolchain

```text
planparity version
planparity doctor
planparity doctor --terraform-bin C:\tools\terraform.exe --tofu-bin C:\tools\tofu.exe
```

`doctor` tests binary discovery, machine-readable versions, validation, state/plan output, provider schemas, and refresh-disabled planning support.

## Inspect state

```text
planparity inspect --engine terraform --state terraform.tfstate
planparity inspect --engine tofu --engine-bin C:\tools\tofu.exe --state terraform.tfstate --format json --output inspection.json
```

The report includes identities, structure, providers, modules, outputs, dependencies, lifecycle status, and findings. Attribute values are excluded.

## Compare states

```text
planparity compare --engine terraform --before before.tfstate --after after.tfstate
planparity compare --engine tofu --before before.tfstate --after after.tfstate --probable-move-threshold 0.90 --format json
```

Exact identity differences are deterministic. Probable moves are labeled as inferred, include confidence evidence, and preserve ambiguity.

## Rehearse a plan

```text
planparity rehearse plan --engine terraform --config ./infra --state ./terraform.tfstate
```

PlanParity copies the configuration and state, creates a temporary home and engine data directory, runs `init -backend=false`, and runs `plan -refresh=false`. Registry access is allowed by the CLI default because provider installation may require it; use an already populated local provider cache when operating offline. Proxy variables are excluded unless `--inherit-proxy` is supplied.

## Rehearse an upgrade

```text
planparity rehearse upgrade \
  --from-engine terraform --from-bin C:\tools\terraform-1.13\terraform.exe \
  --to-engine terraform --to-bin C:\tools\terraform-new\terraform.exe \
  --config ./infra --state ./terraform.tfstate
```

Both executions receive independent copied workspaces. Plan and provider-schema results are normalized and compared.

## Compare Terraform and OpenTofu

```text
planparity rehearse migrate \
  --terraform-bin C:\tools\terraform.exe \
  --tofu-bin C:\tools\tofu.exe \
  --config ./infra --state ./terraform.tfstate
```

## Start the local UI

```text
planparity ui
planparity ui --port 0 --open-browser=false
```

The server binds to numeric loopback only. Port `0` selects an available port. The CLI prints a bootstrap URL containing an ephemeral token in its fragment; the UI moves the token into memory and removes the fragment from browser history.

## Output and exit behavior

- `--format terminal` is human-readable; `--format json` follows `schemas/report-v1.json`.
- `--output` writes the selected report instead of standard output.
- Inputs are never modified, regardless of success or failure.
- Failures avoid echoing engine output or state detail to the local API.

