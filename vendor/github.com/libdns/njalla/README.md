Njalla for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/github.com/libdns/njalla?status.svg)](https://pkg.go.dev/github.com/libdns/njalla)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Njalla](https://njal.la/), allowing you to manage DNS records.

## Configuration

To use this provider, you'll need to obtain an API token from Njalla. You can generate an API token from your Njalla account settings.

## Example

```go
package main

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/libdns/libdns"
	"github.com/libdns/njalla"
)

func main() {
	// Create the provider
	provider := njalla.Provider{
		APIToken: "your-njalla-api-token",
	}

	// Define your zone
	zone := "example.com"

	// Add a record
	records, err := provider.AppendRecords(context.TODO(), zone, []libdns.Record{
		libdns.Address{
			Name: "test",
			IP:   netip.MustParseAddr("192.0.2.1"),
			TTL:  time.Hour,
		},
	})
	if err != nil {
		fmt.Printf("Error adding record: %v\n", err)
		return
	}
	fmt.Printf("Records added: %v\n", records)

	// Get all records
	allRecords, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		fmt.Printf("Error getting records: %v\n", err)
		return
	}
	fmt.Printf("All records: %v\n", allRecords)

	// Delete records
	deletedRecords, err := provider.DeleteRecords(context.TODO(), zone, records)
	if err != nil {
		fmt.Printf("Error deleting records: %v\n", err)
		return
	}
	fmt.Printf("Deleted records: %v\n", deletedRecords)
}
```

## Caveats

- This provider is compatible with libdns v1.1.0 and follows the updated interfaces that use the new Record type system.
- The following record types are fully supported: A, AAAA, CNAME, TXT, MX, SRV
- Other record types are supported using the generic RR structure.
