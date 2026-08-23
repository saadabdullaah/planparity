# PlanParity offline labs

These four labs exercise the public CLI without cloud credentials, provider downloads, or network access. They use tiny deterministic fixture engines that implement only the Terraform/OpenTofu JSON command shapes PlanParity needs. They are test harnesses, not replacement infrastructure engines.

Build PlanParity first, then build the fixture engines:

```powershell
.\scripts\build.ps1
.\labs\build-fixture-engines.ps1
```

```bash
./scripts/build.sh
./labs/build-fixture-engines.sh
```

Complete the labs in order:

1. [Basic inspection](01-basic-inspection/README.md)
2. [Module refactor](02-module-refactor/README.md)
3. [Upgrade replacement](03-upgrade-replacement/README.md)
4. [Terraform/OpenTofu mismatch](04-terraform-to-tofu/README.md)

Every fixture is synthetic. The generated `.bin` directory is ignored by Git.
