#!/usr/bin/env bash
set -uo pipefail

if grep '	' index.txt  | grep -E -v '^[^	 ]+	[^	]+$$'; then
	echo 'bad lines in rfc/index.txt'
	exit 1
fi

set -e

for number in $(sed -n 's/^\([0-9][0-9]*\)[ 	].*$/\1/p' index.txt); do
	if ! test -f "$number"; then
		curl https://www.rfc-editor.org/rfc/rfc$number.txt >$number || rm $number
	fi
done
