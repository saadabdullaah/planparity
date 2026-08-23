# Lab 4: explain a Terraform/OpenTofu mismatch

Goal: run the same copied configuration and state through two deterministic fixture engines whose normalized plan actions differ.

Windows:

```powershell
.\bin\planparity.exe rehearse migrate `
  --terraform-bin .\labs\.bin\terraform.exe `
  --tofu-bin .\labs\.bin\tofu.exe `
  --config .\labs\04-terraform-to-tofu `
  --state .\labs\04-terraform-to-tofu\state.show.json `
  --allow-registry-access=false
```

Linux:

```bash
./bin/planparity rehearse migrate \
  --terraform-bin ./labs/.bin/terraform \
  --tofu-bin ./labs/.bin/tofu \
  --config ./labs/04-terraform-to-tofu \
  --state ./labs/04-terraform-to-tofu/state.show.json \
  --allow-registry-access=false
```

Expected result:

- Terraform reports `NO_OP`; OpenTofu reports `UPDATE`.
- Plan equivalence is false.
- The parity check and finding explain that the actions differ.
- Both engines operate only on independent copied workspaces.
