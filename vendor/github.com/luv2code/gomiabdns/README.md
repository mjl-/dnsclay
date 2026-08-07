# Go-MIABDNS
Mail-In-A-Box custom DNS API client for go.

# CLI tool

```sh
go install github.com/luv2code/go-miabdns/cmd/miabdns@latest

# get a list of all domains defined:
miabdns -email $MIAB_USER -password $MIAB_PASS -url "https://your-box/admin/dns/custom" -command list

# update CNAME with the IP of current machine (will add if it doesn't exist):
miabdns \
    -email $MIAB_USER \
    -password $MIAB_PASS \
    -url "https://your-box/admin/dns/custom" 
    -command update \
    -rname "dyndns.your-box" \
    -rtype "A" \
    -rvalue "$(wget -qO- ipinfo.io/ip)" # also with curl: $(curl -s ipinfo.io/ip)

# add a new record to CNAME to the dyndns record set above
miabdns \
    -email $MIAB_USER \
    -password $MIAB_PASS \
    -url "https://your-box/admin/dns/custom" 
    -command add \
    -rname "some-other-name.your-box" \
    -rtype "CNAME" \
    -rvalue "dyndns.your-box"

# delete a record.
miabdns \
    -email $MIAB_USER \
    -password $MIAB_PASS \
    -url "https://your-box/admin/dns/custom" 
    -command delete \
    -rname "some-other-name.your-box" \
    -rtype "CNAME"
```

# Using as a Library

This project was created for use in [github.com/libdns](https://github.com/libdns/libdns) in order to
create a dns provider for [caddy server](https://caddyserver.com).

You can find the libdns project [here](https://github.com/libdns/miab),
and the caddy dns provider [here](https://github.com/caddy-dns/miab)
