# GCore for [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/gcore)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for GCore, allowing you to manage DNS records.

## Authenticating

To authenticate you need to supply a GCore [API Key](https://gcore.com/docs/edge-ai/inference-at-the-edge/create-and-manage-api-keys).

## Example

Here's a minimal example of how to get all DNS records for zone.

```go
// Package main provides a simple example of how to use the libdns-gcore package.
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    gcore "github.com/libdns/gcore"
)

func main() {
    apiKey := os.Getenv("GCORE_API_KEY")
    if apiKey == "" {
        fmt.Printf("GCORE_API_KEY not set\n")
        return
    }

    if len(os.Args) < 2 {
        fmt.Printf("Usage: %s <zone>\n", filepath.Base(os.Args[0]))
        os.Exit(1)
    }

    zone := os.Args[1]

    provider := &gcore.Provider{
        APIKey: apiKey,
    }

    records, err := provider.GetRecords(context.Background(), zone)
    if err != nil {
        fmt.Printf("Error: %s", err.Error())
        return
    }

    fmt.Println(records)
}
```
