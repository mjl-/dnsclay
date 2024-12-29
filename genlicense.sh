#!/bin/sh
rm -r web/license
set -e
mkdir -p web/license
cp LICENSE web/license
for p in $(cd vendor && find . -type f \( -iname '*license*' -or -iname '*notice*' -or -iname '*patent*' \) -not -name '*.go'); do
	(set +e; mkdir -p $(dirname web/license/$p))
	cp vendor/$p web/license/$p
done
