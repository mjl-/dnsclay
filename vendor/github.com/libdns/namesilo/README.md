NameSilo DNS for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/github.com/libdns/namesilo.svg)](https://pkg.go.dev/github.com/libdns/namesilo)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for the [NameSilo DNS API](https://www.namesilo.com/api-reference), allowing you to manage DNS records.

## Authentication

In order to use NameSilo DNS for libdns, the NameSilo API key is required as the token. One can obtain it on the API Manager page within one's account.

## Example

The following example shows how to retrieve the DNS records.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/libdns/namesilo"
)

func main() {
	token := os.Getenv("LIBDNS_NAMESILO_TOKEN")
	if token == "" {
		fmt.Println("LIBDNS_NAMESILO_TOKEN not set")
		return
	}

	zone := os.Getenv("LIBDNS_NAMESILO_ZONE")
	if token == "" {
		fmt.Println("LIBDNS_NAMESILO_ZONE not set")
		return
	}

	p := &namesilo.Provider{
		AuthAPIToken: token,
	}

	ctx := context.Background()
	records, err := p.GetRecords(ctx, zone)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	fmt.Println(records)
}

```
