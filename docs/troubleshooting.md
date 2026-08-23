# Troubleshooting

## An engine is not found

Run `planparity doctor`. Install Terraform or OpenTofu yourself, place it on `PATH`, or provide the explicit command flag. PlanParity intentionally does not download engines. An explicit binary must be a regular file named `terraform`, `terraform.exe`, `tofu`, or `tofu.exe`.

## Provider initialization fails

`init -backend=false` may still need provider or module registry access. Confirm network policy, use `--inherit-proxy` only when intended, or populate a safe provider cache before running offline. Rehearsal reports state that provider behavior is not an operating-system sandbox.

## The UI says “Session missing”

Open the complete URL printed by `planparity ui`; it includes a one-time fragment token. Do not bookmark or share it. Restart the UI if the token was lost. If a fixed port is occupied, run `planparity ui --port 0`.

## The frontend fallback page appears

Build the frontend before the Go binary:

```text
cd web
npm ci
npm run build
cd ..
go build ./cmd/planparity
```

The repository build scripts perform this order automatically.

## A state cannot be decoded

Confirm the selected engine can read that state format and that the file is a local raw state file. Use the engine that wrote the file first. Never post a real state file in a public issue; create a synthetic reproduction with values removed.

## Temporary directory access fails

PlanParity normally uses the operating-system temporary directory. In restricted development environments, set `PLANPARITY_TEMP_DIR` to a private writable directory. This variable affects only disposable working copies; it does not change the source configuration or state.

