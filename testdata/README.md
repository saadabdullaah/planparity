# Test fixtures

All fixtures are synthetic and safe to publish. The string in `security/canary.txt` is deliberately embedded in raw state and plan inputs. Tests must prove it never appears in normalized models, terminal output, reports, API responses, or frontend rendering.

Fixture categories:

- `terraform` and `opentofu`: representative `show -json` state output.
- `comparisons`: added, removed, and probable-move cases.
- `plans`: normalized action and replacement cases.
- `corrupted`: malformed or unsupported input.
- `security`: public fake-secret canary definition.
- `integration/populated`: valid raw state with the canary in ordinary attributes, used to prove a real engine and the local API emit only a value-free resource inventory.
