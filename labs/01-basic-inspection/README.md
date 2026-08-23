# Lab 1: inspect a state snapshot

Goal: inspect one synthetic state and confirm that identity and structure appear while attribute values do not.

Windows:

```powershell
.\bin\planparity.exe inspect `
  --engine terraform `
  --engine-bin .\labs\.bin\terraform.exe `
  --state .\labs\01-basic-inspection\state.show.json
```

Linux:

```bash
./bin/planparity inspect \
  --engine terraform \
  --engine-bin ./labs/.bin/terraform \
  --state ./labs/01-basic-inspection/state.show.json
```

Expected result:

- One `terraform_data.example` resource.
- One provider and one output metadata entry.
- `Values: HIDDEN` in terminal output.
- The string `PP_LAB_VALUE_MUST_NOT_APPEAR` is absent.

Repeat with `--format json` and inspect the value-free contract.
