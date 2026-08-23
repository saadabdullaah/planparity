#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_root="$repo_root/labs/.bin"
source_path="$repo_root/labs/support/fake-engine"

mkdir -p "$output_root/baseline" "$output_root/candidate"
go build -trimpath -o "$output_root/terraform" "$source_path"
go build -trimpath -o "$output_root/tofu" "$source_path"
go build -trimpath -o "$output_root/baseline/terraform" "$source_path"
go build -trimpath -o "$output_root/candidate/terraform" "$source_path"

printf 'Built deterministic fixture engines in %s\n' "$output_root"
