#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

./labs/build-fixture-engines.sh
go build -trimpath -o ./labs/.bin/planparity ./cmd/planparity

inspect="$(./labs/.bin/planparity inspect --engine terraform --engine-bin ./labs/.bin/terraform --state ./labs/01-basic-inspection/state.show.json --format json)"
[[ "$inspect" != *"PP_LAB_VALUE_MUST_NOT_APPEAR"* ]]
[[ "$inspect" == *'"resource_count": 1'* ]]

comparison="$(./labs/.bin/planparity compare --engine terraform --engine-bin ./labs/.bin/terraform --before ./labs/02-module-refactor/before.show.json --after ./labs/02-module-refactor/after.show.json --format json)"
[[ "$comparison" == *'"probable_moves": 1'* ]]

upgrade="$(./labs/.bin/planparity rehearse upgrade --from-engine terraform --from-bin ./labs/.bin/baseline/terraform --to-engine terraform --to-bin ./labs/.bin/candidate/terraform --config ./labs/03-upgrade-replacement --state ./labs/03-upgrade-replacement/state.show.json --allow-registry-access=false --format json)"
[[ "$upgrade" == *'"replacements": 1'* ]]

migration="$(./labs/.bin/planparity rehearse migrate --terraform-bin ./labs/.bin/terraform --tofu-bin ./labs/.bin/tofu --config ./labs/04-terraform-to-tofu --state ./labs/04-terraform-to-tofu/state.show.json --allow-registry-access=false --format json)"
[[ "$migration" == *'"equivalent": false'* ]]

printf 'All four offline labs passed.\n'
