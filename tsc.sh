#!/usr/bin/env bash
set -euo pipefail

./node_modules/.bin/tsc | sed -E 's/^([^\(]+)\(([0-9]+),([0-9]+)\):/\1:\2:\3: /'
./node_modules/.bin/prettier --write web/dnsclay.js
