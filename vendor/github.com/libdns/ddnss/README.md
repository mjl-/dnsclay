**DEVELOPER INSTRUCTIONS:**

---

ddnss for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/ddnss)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for ddnss, allowing you to manage DNS records.

You can pass three parameters:

- `api_token` - api token from the ddnss.de user interface used for almost all actions
- `username` used in combination with:
- `password` - Used in the GetRecords() function, to pull the domains from the web-interface
