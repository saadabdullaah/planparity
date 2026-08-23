# Install on Windows

PlanParity supports Windows 10 and 11 on x86-64. Terraform and OpenTofu are separate tools and are not downloaded by PlanParity.

## Release binary

1. Download the Windows archive and `checksums.txt` from the matching GitHub release.
2. Verify the archive:

```powershell
(Get-FileHash .\planparity_VERSION_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
Get-Content .\checksums.txt
```

3. Extract the archive to a user-owned tools directory.
4. If Windows marked the file as downloaded, run:

```powershell
Unblock-File C:\Tools\PlanParity\planparity.exe
```

5. Add that directory to your user `PATH`, then open a new terminal.

## Verify the toolchain

```powershell
planparity version
planparity doctor
```

Use explicit paths when an engine is not on `PATH`:

```powershell
planparity doctor --terraform-bin C:\Tools\Terraform\terraform.exe --tofu-bin C:\Tools\OpenTofu\tofu.exe
```

Paths with spaces and drive-letter paths are supported. PlanParity validates that explicit binaries are regular files named `terraform.exe` or `tofu.exe`.

For source builds, see [Build from source](build-from-source.md).
