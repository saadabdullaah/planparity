# Install on Linux

PlanParity supports modern x86-64 Linux distributions. Terraform and OpenTofu are separate tools and are not downloaded by PlanParity.

## Release binary

Download the Linux archive and matching checksums, then verify and install it:

```bash
sha256sum --check checksums.txt --ignore-missing
tar -xzf planparity_VERSION_linux_amd64.tar.gz
install -m 0755 planparity "$HOME/.local/bin/planparity"
```

Ensure `$HOME/.local/bin` is on `PATH`, then run:

```bash
planparity version
planparity doctor
```

Use explicit executable paths when needed:

```bash
planparity doctor --terraform-bin /opt/terraform/terraform --tofu-bin /opt/opentofu/tofu
```

PlanParity rejects non-executable engine files and resolves symbolic links before validating the expected executable name. Rehearsal configuration links are rejected because copied workspaces never follow them.

For source builds, see [Build from source](build-from-source.md).
