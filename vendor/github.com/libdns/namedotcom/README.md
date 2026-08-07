name.com for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/namedotcom)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for name.com, allowing you to manage DNS records.

## Authenticating
To initiate the provider you need to supply the following parameters:
```go
provider := namedotcom.Provider {
	Token : "NAMEDOTCOM_API_TOKEN",
	User :  "NAMEDOTCOM_USER_NAME",
	Server: "https://api.name.com", // full url scheme expected here..
}
```

## Example
Here's a basic example of how to list, update and delete records using this provider
```go
package main

import (
	"context"
	"github.com/libdns/libdns"
	"os"
	"github.com/libdns/namedotcom"
	"log"
)

func main() {
	ctx := context.Background()

	zone := "example.com."

	// configure the name.com DNS provider 
	provider := namedotcom.Provider{
		Token : os.GetEnv("NAMEDOTCOM_API_TOKEN"),
		User :     os.GetEnv("NAMEDOTCOM_USER_NAME"),
		Server:    os.GetEnv("NAMEDOTCOM_SERVER"),
	}

	// list and iterate through all records
	recs, err := provider.GetRecords(ctx, zone)
	if err != nil {
		log.Fatal(err)
	}

	for _, rec := range recs {
		log.Println(rec)
	}


	// attempts an upsert, PUT or POST for the given record
	newRecs, err = provider.SetRecords(ctx, zone, []libdns.Record{
		Type:  "A",
		Name:  "sub",
		Value: "1.2.3.4",
	})

	// delete records deletes the given record by ID.
	deletedRecs, err = provider.DeleteRecords(ctx, zone, []libdns.Record{
		Type:  "A",
		Name:  "sub",
		Value: "1.2.3.4",
	})

}
```
