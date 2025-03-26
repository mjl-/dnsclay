# Bunny.net DNS for [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/bunny)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Bunny.net](https://docs.bunny.net/reference/bunnynet-api-overview), allowing you to manage DNS records.

## Authenticating

To authenticate you need to supply a Bunny.net [API Key](https://dash.bunny.net/account/settings).

## Example

Here's a minimal example of how to get all DNS records for zone. See also: [provider_test.go](https://github.com/libdns/bunny/blob/master/provider_test.go)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/libdns/bunny"
)

func main() {
	apiKey := os.Getenv("BUNNY_API_KEY")
	if apiKey == "" {
		fmt.Printf("BUNNY_API_KEY not set\n")
		return
	}

	zone := os.Getenv("BUNNY_ZONE")
	if token == "" {
		fmt.Printf("BUNNY_ZONE not set\n")
		return
	}

	provider := &bunny.Provider{
		AccessKey: apiKey,
	}

	records, err := provider.GetRecords(context.WithTimeout(context.Background(), time.Duration(15*time.Second)), zone)
	if err != nil {
        fmt.Printf("Error: %s", err.Error())
        return
	}

	fmt.Println(records)
}

```

## Debugging

You can enable logging by configuring a custom logger or by setting `Debug` to true.

```go
	...

	// Logging is always enabled when using a custom logger
	provider := &bunny.Provider{
		AccessKey: apiKey,
		Logger: func(msg string, records []libdns.Record) {
			fmt.Printf("[bunny]: %s\n", msg)
		}
	}

	// Enable the default logger
	provider := &bunny.Provider{
		AccessKey: apiKey,
		Debug: true
	}
```
Example output using the default logger:

```shell
[bunny] fetching all records in zone example.com
[bunny] fetching zone ID for example.com
[bunny] done found zone ID 82940 for example.com
[bunny] done fetching 3 record(s) in zone example.com
[bunny]   TXT: ID=7648777, TTL=2m0s, Priority=0, Name=test1, Value=test1
[bunny]   TXT: ID=7648778, TTL=2m0s, Priority=0, Name=test2, Value=test2
[bunny]   TXT: ID=7648779, TTL=2m0s, Priority=0, Name=test3, Value=test3
[bunny] deleting TXT record in zone example.com
```
