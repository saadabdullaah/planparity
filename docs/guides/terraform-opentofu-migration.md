# Compare Terraform and OpenTofu

Migration rehearsal runs Terraform as the baseline and OpenTofu as the candidate over independent copies.

```powershell
planparity rehearse migrate `
  --terraform-bin C:\Tools\Terraform\terraform.exe `
  --tofu-bin C:\Tools\OpenTofu\tofu.exe `
  --config .\infra `
  --state .\infra\terraform.tfstate
```

The result answers whether each engine can read the copied state, whether normalized provider schemas match, and whether normalized plan actions are equivalent.

A failed parity check does not necessarily mean one engine is wrong. It means the observed local behavior differs and needs review. Engine diagnostics are sanitized and preserved, while raw plan/state JSON remains inside the backend process.

Start with synthetic or independently backed-up state. Review [known limitations](../known-limitations.md) before interpreting provider-dependent behavior.
