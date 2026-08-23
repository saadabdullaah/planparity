# Five-minute quick start

This path proves the complete local workflow without a cloud account or provider download.

## 1. Build PlanParity

Windows:

```powershell
.\scripts\build.ps1
```

Linux:

```bash
./scripts/build.sh
```

The result is `bin/planparity.exe` on Windows or `bin/planparity` on Linux.

## 2. Build the deterministic lab engines

Windows:

```powershell
.\labs\build-fixture-engines.ps1
```

Linux:

```bash
./labs/build-fixture-engines.sh
```

These fixtures implement only the machine-readable commands used by the labs. They do not contact registries or providers.

## 3. Inspect a snapshot

Windows:

```powershell
.\bin\planparity.exe inspect --engine terraform --engine-bin .\labs\.bin\terraform.exe --state .\labs\01-basic-inspection\state.show.json
```

Linux:

```bash
./bin/planparity inspect --engine terraform --engine-bin ./labs/.bin/terraform --state ./labs/01-basic-inspection/state.show.json
```

Confirm that one resource is reported, values are hidden, and the seeded lab value is absent.

## 4. Open the local UI

Windows:

```powershell
.\bin\planparity.exe ui --terraform-bin .\labs\.bin\terraform.exe
```

Linux:

```bash
./bin/planparity ui --terraform-bin ./labs/.bin/terraform
```

Open the complete URL printed in the terminal. The fragment token is moved into memory and removed from browser history. Choose **Inspect**, enter the absolute path to the lab snapshot, and start the analysis.

## 5. Continue through the labs

Follow [all four offline labs](../labs/README.md) to exercise probable moves, upgrade replacement detection, and Terraform/OpenTofu plan differences.

PlanParity does not execute `apply`, `destroy`, or state mutation. Read the [safety model](security/safety-model.md) before using real infrastructure files.
