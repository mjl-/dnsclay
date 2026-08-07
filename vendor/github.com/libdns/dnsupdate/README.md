[DNS UPDATE] for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/dnsupdate)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for the [DNS UPDATE] and [DNS AXFR] protocols, allowing you to manage DNS records.

DNS AXFR is used to list DNS records. DNS UPDATE is used to append, set and delete records.

## Usage

The DNS server needs to accept DNS transfer and update requests from the host where libdns is used.

### Example [Knot] configuration

This example configuration allows libdns usage from localhost.

```yaml
acl:
  - id: local
    address: [127.0.0.1, ::1]
    action: [transfer, update]
zone:
  - domain: example.com
    acl: [local]
```

### Example [bind] configuration

This example configuration allows libdns usage from localhost.

```
allow-transfer { 127.0.0.1; };
allow-update { 127.0.0.1; };
```

## Caveats

DNS doesn't have a concept of unique identifier for each record. The DNS UPDATE protocol doesn't support duplicate records (two records with the exact same header and value).

[DNS UPDATE]: https://www.rfc-editor.org/rfc/rfc2136
[DNS AXFR]: https://datatracker.ietf.org/doc/html/rfc5936
[Knot]: https://www.knot-dns.cz/
[bind]: https://www.isc.org/bind/
