$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$repoRoot = Split-Path -Parent $PSScriptRoot
$labBinary = Join-Path $repoRoot 'labs\.bin\planparity.exe'

Push-Location $repoRoot
try {
    & (Join-Path $repoRoot 'labs\build-fixture-engines.ps1')
    go build -trimpath -o $labBinary .\cmd\planparity

    $inspect = (& $labBinary inspect --engine terraform --engine-bin .\labs\.bin\terraform.exe --state .\labs\01-basic-inspection\state.show.json --format json | ConvertFrom-Json)
    if ($inspect.snapshot.resources.Count -ne 1) { throw 'Inspection lab did not report one resource.' }

    $comparison = (& $labBinary compare --engine terraform --engine-bin .\labs\.bin\terraform.exe --before .\labs\02-module-refactor\before.show.json --after .\labs\02-module-refactor\after.show.json --format json | ConvertFrom-Json)
    if ($comparison.summary.probable_moves -ne 1) { throw 'Module-refactor lab did not report one probable move.' }

    $upgrade = (& $labBinary rehearse upgrade --from-engine terraform --from-bin .\labs\.bin\baseline\terraform.exe --to-engine terraform --to-bin .\labs\.bin\candidate\terraform.exe --config .\labs\03-upgrade-replacement --state .\labs\03-upgrade-replacement\state.show.json --allow-registry-access=false --format json | ConvertFrom-Json)
    if ($upgrade.candidate.summary.replacements -ne 1) { throw 'Upgrade lab did not report one replacement.' }

    $migration = (& $labBinary rehearse migrate --terraform-bin .\labs\.bin\terraform.exe --tofu-bin .\labs\.bin\tofu.exe --config .\labs\04-terraform-to-tofu --state .\labs\04-terraform-to-tofu\state.show.json --allow-registry-access=false --format json | ConvertFrom-Json)
    if ($migration.comparison.equivalent) { throw 'Migration lab did not report a plan mismatch.' }

    Write-Output 'All four offline labs passed.'
}
finally {
    Pop-Location
}
