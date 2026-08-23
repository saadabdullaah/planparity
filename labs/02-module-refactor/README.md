# Lab 2: identify a probable module move

Goal: compare two snapshots where the same structural resource moved into `module.platform`.

Windows:

```powershell
.\bin\planparity.exe compare `
  --engine terraform `
  --engine-bin .\labs\.bin\terraform.exe `
  --before .\labs\02-module-refactor\before.show.json `
  --after .\labs\02-module-refactor\after.show.json
```

Linux:

```bash
./bin/planparity compare \
  --engine terraform \
  --engine-bin ./labs/.bin/terraform \
  --before ./labs/02-module-refactor/before.show.json \
  --after ./labs/02-module-refactor/after.show.json
```

Expected result:

- No unmatched added or removed resources after identity correlation.
- One probable move from `terraform_data.service` to `module.platform.terraform_data.service`.
- The move remains `INFERRED`, with confidence evidence, rather than being called verified.
