#!/bin/sh

build_method_call() {
	local method_name=$1
	shift
	echo -n "<?xml version=\"1.0\"?>"
	echo -n "<methodCall>"
	echo -n "<methodName>$method_name</methodName><params>"

	for param in "$@"; do
		echo -n "$param"
	done
	
	echo -n "</params></methodCall>"
}

build_param() {
	local value=$1
	local type=${2:-string}
	echo -n "<param><value><$type>"
		echo -n "$value"
	echo -n "</$type></value></param>"
}


request_zone_records(){
    build_method_call "getZoneRecords" \
        "$(build_param $LOOPIA_USER)" \
        "$(build_param $LOOPIA_PASSWORD)" \
        "$(build_param $ZONE)" \
        "$(build_param $1)"
}

request_domains() {
    build_method_call "getDomains" \
        "$(build_param $LOOPIA_USER)" \
        "$(build_param $LOOPIA_PASSWORD)"
}

request_subdomains() {
    build_method_call "getSubdomains" \
        "$(build_param $LOOPIA_USER)" \
        "$(build_param $LOOPIA_PASSWORD)" \
        "$(build_param $ZONE)"
}


getit(){
    echo "request $1"
    echo "\n--\n"
    curl -X POST https://api.loopia.se/RPCSERV \
        -H "Content-Type: application/xml" \
        -H "Accept: application/xml" \
        -d "$1"
    echo "\n--"
}

method=$1
shift 1
echo ""
case $method in
    getZoneRecords)
        getit "$(request_zone_records $1)"
        ;;
    getDomains)
        getit "$(request_domains)"
        ;;
    getSubdomains)
        getit "$(request_subdomains $1)"
        ;;
    *)
        echo -n "unknown command $method"
        ;;
esac
echo ""