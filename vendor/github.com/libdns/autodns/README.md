
\<autodns\> for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/autodns)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for \<autodns\>, allowing you to manage DNS records.

Example:

```
package main

import (
	"context"
	"os"
	"log"

	"github.com/libdns/autodns"
)

func main() {
	provider := autodns.Provider{
		Username: os.Getenv("AUTODNS_USERNAME"),
		Password: os.Getenv("AUTODNS_PASSWORD"),
	}

	records, err := provider.GetRecords(context.TODO(), "zone.example.org")
	if err != nil {
		log.Fatalf("unexpected error: %s", err)
	}

	fmt.Printf("%#v", records)
}
```

As an alternative, configure the provider struct with the following:

| Field      | Description (default)      | Required |
|------------|----------------------------|----------|
| Username   | username, empty            | yes      |
| Password   | password, empty            | yes      |
| Endpoint   | https://api.autodns.com/v1 | no       |
| Context    | 4                          | no       |
| HttpClient | `&http.Client{}`           | no       |
