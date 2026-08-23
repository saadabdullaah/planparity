# Build from source

## Prerequisites

- Go 1.26 or newer.
- Node.js 24 and npm.
- PowerShell 7 on Windows, or Bash on Linux.
- Git.

Node.js is needed only to build and test the embedded frontend. Users of a release binary do not need Node.js.

## Windows

```powershell
git clone https://github.com/saadabdullaah/planparity.git
Set-Location planparity
.\scripts\test.ps1
.\scripts\build.ps1
.\bin\planparity.exe version
```

## Linux

```bash
git clone https://github.com/saadabdullaah/planparity.git
cd planparity
./scripts/test.sh
./scripts/build.sh
./bin/planparity version
```

The build scripts install exact frontend dependencies from `web/package-lock.json`, compile the React interface, embed it into the Go binary, and include commit/build metadata.

## Useful development commands

```text
go test ./...
go vet ./...
cd web && npm ci && npm run lint && npm run test:run && npm run build
```

The complete Windows and Linux command sequences live in `scripts/test.ps1` and `scripts/test.sh`.
