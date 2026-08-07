# Timeweb DNS for [libdns](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/Riskful/timeweb-libdns)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Timeweb DNS API](https://timeweb.cloud/api-docs#tag/Domeny/operation/getDomainDNSRecords), allowing you to manage DNS records.

## Authorize

To authorize you need to use Timeweb [Authorization](https://timeweb.cloud/my/login).

## Example

Minimal working example of getting DNS zone records.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/libdns/libdns/timeweb-libdns"
)

func main() {

	provider = timeweb.Provider{
		ApiURL:   os.Getenv("TIMEWEB_URL"),
		ApiToken: os.Getenv("TIMEWEB_API_TOKEN"),
	}
	zone = os.Getenv("TIMEWEB_ZONE")
	ctx = context.Background()

	records, err := provider.GetRecords(ctx, zone)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	fmt.Println(records)
}

```

Always yours [@Riskful](https://github.com/Riskful)