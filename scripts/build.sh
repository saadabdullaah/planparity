#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$repo_root/bin"

(
  cd "$repo_root/web"
  npm ci
  npm run build
)

build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
version="${PLANPARITY_VERSION:-dev}"
commit="$(git -C "$repo_root" rev-parse --short HEAD 2>/dev/null || printf unknown)"
ldflags="-s -w -X github.com/saadabdullaah/planparity/internal/buildinfo.version=$version -X github.com/saadabdullaah/planparity/internal/buildinfo.commit=$commit -X github.com/saadabdullaah/planparity/internal/buildinfo.date=$build_date"
go build -trimpath -ldflags "$ldflags" -o "$repo_root/bin/planparity" ./cmd/planparity
