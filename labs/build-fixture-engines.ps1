$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$repoRoot = Split-Path -Parent $PSScriptRoot
$outputRoot = Join-Path $PSScriptRoot '.bin'
$source = Join-Path $PSScriptRoot 'support\fake-engine'

$targets = @(
    (Join-Path $outputRoot 'terraform.exe'),
    (Join-Path $outputRoot 'tofu.exe'),
    (Join-Path $outputRoot 'baseline\terraform.exe'),
    (Join-Path $outputRoot 'candidate\terraform.exe')
)

foreach ($target in $targets) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $target) | Out-Null
    go build -trimpath -o $target $source
}

Write-Output "Built deterministic fixture engines in $outputRoot"
