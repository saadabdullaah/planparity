$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot

$goFiles = Get-ChildItem -Path $repoRoot -Filter '*.go' -Recurse | Where-Object {
    $_.FullName -notmatch '[\\/](vendor|node_modules|\.planparity-local|\.git)[\\/]'
}
if ($goFiles.Count -gt 0) {
    $unformatted = gofmt -l @($goFiles.FullName)
    if ($unformatted) {
        throw "Go files are not formatted:`n$($unformatted -join "`n")"
    }
}

go vet ./...
if ($IsLinux) {
    go test -race -coverprofile (Join-Path $repoRoot 'coverage.coverprofile') ./...
}
else {
    go test -coverprofile (Join-Path $repoRoot 'coverage.coverprofile') ./...
}

Push-Location (Join-Path $repoRoot 'web')
try {
    npm ci
    npm run lint
    npm run test:run
    npm run build
}
finally {
    Pop-Location
}
