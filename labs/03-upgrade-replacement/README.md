# Lab 3: catch an upgrade replacement

Goal: rehearse a deterministic baseline/candidate engine transition where the candidate schema and plan require resource replacement.

Windows:

```powershell
.\bin\planparity.exe rehearse upgrade `
  --from-engine terraform `
  --from-bin .\labs\.bin\baseline\terraform.exe `
  --to-engine terraform `
  --to-bin .\labs\.bin\candidate\terraform.exe `
  --config .\labs\03-upgrade-replacement `
  --state .\labs\03-upgrade-replacement\state.show.json `
  --allow-registry-access=false
```

Linux:

```bash
./bin/planparity rehearse upgrade \
  --from-engine terraform \
  --from-bin ./labs/.bin/baseline/terraform \
  --to-engine terraform \
  --to-bin ./labs/.bin/candidate/terraform \
  --config ./labs/03-upgrade-replacement \
  --state ./labs/03-upgrade-replacement/state.show.json \
  --allow-registry-access=false
```

Expected result:

- Candidate plan summary reports one replacement.
- Provider schema differences include `terraform_data`.
- A high-severity replacement finding appears.
- Original lab files remain byte-identical and temporary plan data is removed.
