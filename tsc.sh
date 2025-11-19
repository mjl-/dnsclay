#!/usr/bin/env bash
set -euo pipefail

./node_modules/.bin/tsc | sed -E 's/^([^\(]+)\(([0-9]+),([0-9]+)\):/\1:\2:\3: /'
unexpand -t4 <web/dnsclay.js >web/dnsclay.js.tmp && mv web/dnsclay.js.tmp web/dnsclay.js
