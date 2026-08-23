# Compare state snapshots

Comparison independently inspects two local state files through the same engine, then compares their sanitized models.

```powershell
planparity compare `
  --engine terraform `
  --before .\state-before.tfstate `
  --after .\state-after.tfstate
```

Linux uses the same flags with shell line continuations.

The report separates:

- Added and removed identities.
- Matched resources whose provider, schema, structure, or dependencies changed.
- Provider usage differences.
- Probable moves inferred from type, provider, structure, name, mode, and dependency signals.

Probable moves are always `INFERRED` unless independent engine evidence verifies them. Adjust the bounded threshold only when you understand the false-positive tradeoff:

```powershell
planparity compare --engine terraform --before before.tfstate --after after.tfstate --probable-move-threshold 0.90
```

Use the JSON report or local UI to inspect move signals and before/after sanitized resource metadata.
