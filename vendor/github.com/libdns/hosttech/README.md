hosttech for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/hosttech)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [hosttech.ch](https://hosttech.ch), allowing you to manage DNS records.

## Example Use
See for an example [here](./provider_example.go).

## Constraints
Some constraints.
### Supported record types
Because the Hosttech API does not provide a way to manipulate a generic "Type,Name,Value"-Record, not every type of record can be set. Currently supported are:
- AAAA
- A
- NS
- CNAME
- MX
- TXT
- TLSA

Any unsupported record types returns an error.

### Minimal TTL
The Time-to-Life has to be at least 600 seconds. If you try to set a lower value, the client will
automatically set it to 600 seconds. Smaller values would be rejected by the Hosttech API.

## Further documentation
Any further documentation that could be helpful:
 - [Hosttech DNS API documentation](https://api.ns1.hosttech.eu/api/documentation)


