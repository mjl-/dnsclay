# ClouDNS for [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/cloudns)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for ClouDNS, allowing you to manage
DNS records.

## Installation

To install this package, use `go get`:

```sh
go get github.com/libdns/cloudns
```

## Usage

Here is an example of how to use this package to manage DNS records:

```go
package main

import (
	"context"
	"fmt"
	"github.com/libdns/cloudns"
	"github.com/libdns/libdns"
	"time"
)

func main() {
	provider := &cloudns.Provider{
		AuthId:       "your_auth_id",
		SubAuthId:    "your_sub_auth_id",
		AuthPassword: "your_auth_password",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get records
	records, err := provider.GetRecords(ctx, "example.com")
	if err != nil {
		fmt.Printf("Failed to get records: %s\n", err)
		return
	}
	fmt.Printf("Records: %+v\n", records)

	// Append a record
	newRecord := libdns.Record{
		Type:  "TXT",
		Name:  "test",
		Value: "test-value",
		TTL:   300 * time.Second,
	}
	addedRecords, err := provider.AppendRecords(ctx, "example.com", []libdns.Record{newRecord})
	if err != nil {
		fmt.Printf("Failed to append record: %s\n", err)
		return
	}
	fmt.Printf("Added Records: %+v\n", addedRecords)
}
```

## Configuration

The `Provider` struct has the following fields:

- `AuthId` (string, optional): Your ClouDNS authentication ID.
- `SubAuthId` (string, optional): Your ClouDNS sub-authentication ID.
- `AuthPassword` (string): Your ClouDNS authentication password.

## Testing

To run the tests, you need to set up your ClouDNS credentials and zone in the test file `provider_test.go`. The tests
require a live ClouDNS account.

```go
var (
TAuthId = "your_auth_id"
TSubAuthId = "your_sub_auth_id"
TAuthPassword = "your_auth_password"
TZone = "example.com"
)
```

Run the tests using the following command:

```sh
go test ./...
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
