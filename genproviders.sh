#!/usr/bin/env bash
set -euo pipefail
(
cat <<EOF
package main

// WARNING: Automatically generated, do not edit manually. Add new providers to providers.txt, keeping it sorted, and run "make build".

import (
EOF

sed 's/^/	/' providers.txt

cat <<EOF

)

// KnownProviders ensures all providers types are included in sherpadoc API documentation.
type KnownProviders struct {
EOF

for i in $(cat providers.txt | cut -f1 -d' '); do
	echo "	X$i $i.Provider"
done

cat <<EOF
}

// providers is used for instantiating a provider by name.
var providers = map[string]any{
EOF

for i in $(cat providers.txt | cut -f1 -d' '); do
	echo "	\"$i\": $i.Provider{},"
done

cat <<EOF
}

// providerURLs map provider names to repository URLs for help.
var providerURLs = map[string]string{
EOF

sed 's/^\(.*\) \(.*\)$/"\1": \2,/' providers.txt

cat <<EOF
}
EOF

)>providers.go
gofmt -w -s providers.go
