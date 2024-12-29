#!/usr/bin/env bash
set -euo pipefail
(
cat <<EOF
package main

// WARNING: Automatically generated, do not edit manually. Add new providers to providers.txt, keeping it sorted, and run "make build".

import (
EOF

for i in $(cat providers.txt); do
	# can generalize on next occurrence of special import name
	prefix=''
	if test $i = 'openstack-designate'; then
		prefix='openstackdesignate '
	fi
	echo "	$prefix\"github.com/libdns/$i\""
done

cat <<EOF

)

// KnownProviders ensures all providers types are included in sherpadoc API documentation.
type KnownProviders struct {
EOF

for i in $(cat providers.txt); do
	if test $i = 'openstack-designate'; then
		i='openstackdesignate'
	fi

	echo "	X$i $i.Provider"
done

cat <<EOF
}

// providers is used for instantiating a provider by name.
var providers = map[string]any{
EOF

for i in $(cat providers.txt); do
	if test $i = 'openstack-designate'; then
		i='openstackdesignate'
	fi

	echo "	\"$i\": $i.Provider{},"
done

cat <<EOF
}
EOF

)>providers.go
gofmt -w -s providers.go
