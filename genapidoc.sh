#!/usr/bin/env bash
set -euo pipefail

CGO_ENABLED=0 go run vendor/github.com/mjl-/sherpadoc/cmd/sherpadoc/*.go \
	-adjust-function-names none \
	-rename "$(cat providers.txt | sed s/-// | awk '{ printf("%s Provider Provider_%s,", $$1, $$1) }')"'sherpadoc Arg sherpadocArg,sherpadoc Function sherpadocFunction,sherpadoc Ints sherpadocInts,sherpadoc Section sherpadocSection,sherpadoc Strings sherpadocStrings,sherpadoc Struct sherpadocStruct,sherpadoc Field sherpadocField' \
	-dropfields 'digitalocean.Provider.Client,dnspod.Provider.Client,dynu.Provider.Once,dynu.Provider.Client,scaleway.Provider.Client,selectel.Provider.ZonesCache' \
	API >web/api.json

