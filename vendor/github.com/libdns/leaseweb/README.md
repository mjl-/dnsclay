Leaseweb provider for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/leaseweb)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Leaseweb](https://leaseweb.com/), allowing you to manage DNS records.

## Usage

Generate an API Key via the [Leaseweb customer portal](https://secure.leaseweb.com/); under Administration -> API Key.

Place API Key in the configuration as `APIKey`.

## Gotcha's

- Leaseweb expects full domain (sub.example.com); where libdns does not (sub as name and example.com as zone).
- libdns might providea TTL of 0; Leaseweb validates on their supported values (defauling to 60 for now).
- Leaseweb does not expect a trailing dot in the zone; libdns provides one, so we remove it.

## Compiling

### Docker

Run:

```
docker run --rm -it -v "$PWD":/go/src/leaseweb -w /go/src/leaseweb golang:1.16
```

which will drop you in an interactive bash prompt where `go` and friends are available.

For example you can build the code with `go build`.

## Example

```go
package main

import (
	"context"
	"fmt"
	"github.com/libdns/leaseweb"
)

func main() {
	provider := leaseweb.Provider{APIKey: "<LEASEWEB API KEY>"}

	records, err  := provider.GetRecords(context.TODO(), "example.com")
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, record := range records {
		fmt.Printf("%s %v %s %s\n", record.Name, record.TTL.Seconds(), record.Type, record.Value)
	}
}
```

## Contributing / developing

Easiest way to get your hands dirty is to use [Caddy](https://github.com/caddyserver/caddy).

Check contributing instructions at [caddy-dns/leaseweb](https://github.com/caddy-dns/leaseweb).