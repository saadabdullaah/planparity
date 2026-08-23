#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

unformatted="$(find . -name '*.go' -not -path './vendor/*' -not -path './web/node_modules/*' -not -path './.planparity-local/*' -not -path './.git/*' -print0 | xargs -0 gofmt -l)"
if [[ -n "$unformatted" ]]; then
  printf 'Go files are not formatted:\n%s\n' "$unformatted" >&2
  exit 1
fi

go vet ./...
go test -race -coverprofile coverage.coverprofile ./...

(
  cd web
  npm ci
  npm run lint
  npm run test:run
  npm run build
)
