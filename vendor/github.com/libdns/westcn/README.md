West.cn for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/westcn)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [west.cn](https://west.cn), allowing you to manage DNS records.

This package references and uses the implementation of [lego](https://github.com/go-acme/lego).

## Authenticating

To authenticate you need to supply your Username and APIPassword to the Provider.

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
	"context"
	"fmt"

	"github.com/libdns/westcn"
)

func main() {
	provider := westcn.Provider{
		Username: "<Username form your west.cn account>",
		APIPassword: "<APIPassword form your west.cn account>",
	}

	records, err  := provider.GetRecords(context.TODO(), "example.com.")
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, record := range records {
		fmt.Printf("%s %v %s %s\n", record.Name, record.TTL.Seconds(), record.Type, record.Value)
	}
}
```

For complete demo check [_example/main.go](_example/main.go)
