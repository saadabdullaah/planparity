$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$repoRoot = Split-Path -Parent $PSScriptRoot
$webRoot = Join-Path $repoRoot 'web'
$binRoot = Join-Path $repoRoot 'bin'

Push-Location $webRoot
try {
    npm ci
    npm run build
}
finally {
    Pop-Location
}

New-Item -ItemType Directory -Force -Path $binRoot | Out-Null
$buildDate = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
$version = if ($env:PLANPARITY_VERSION) { $env:PLANPARITY_VERSION } else { 'dev' }
$commit = 'unknown'
if (Get-Command git -ErrorAction SilentlyContinue) {
    $candidate = git rev-parse --short HEAD 2>$null
    if ($LASTEXITCODE -eq 0) { $commit = $candidate }
}
$ldflags = "-s -w -X github.com/saadabdullaah/planparity/internal/buildinfo.version=$version -X github.com/saadabdullaah/planparity/internal/buildinfo.commit=$commit -X github.com/saadabdullaah/planparity/internal/buildinfo.date=$buildDate"
go build -trimpath -ldflags $ldflags -o (Join-Path $binRoot 'planparity.exe') ./cmd/planparity

Write-Output "Built $binRoot\planparity.exe"
