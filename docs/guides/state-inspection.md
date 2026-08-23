# Inspect Terraform or OpenTofu state

Inspection asks the selected local engine to decode a raw state file, then immediately converts the engine JSON into PlanParity's value-free model.

Terraform:

```powershell
planparity inspect --engine terraform --state .\terraform.tfstate
```

OpenTofu:

```powershell
planparity inspect --engine tofu --state .\terraform.tfstate
```

Use `--engine-bin` when the executable is not on `PATH`, `--timeout` for a bounded engine deadline, and `--format json --output report.json` for the versioned contract.

The report includes:

- Resource identity, mode, provider, schema version, status, and dependency addresses.
- Module containment.
- Provider usage and schema versions.
- Output names, sensitivity, and type fingerprints.
- Structural key names, sensitive paths, unknown paths, and fingerprints.

It excludes ordinary attribute and output values. PlanParity hashes the supplied file before and after inspection and fails if it changes.

If an engine cannot decode the state, retry with the engine family/version that wrote it. Do not convert or repair the original state in place.
