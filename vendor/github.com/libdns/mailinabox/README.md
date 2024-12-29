[Mail-In-A-Box](https://mailinabox.email/) custom DNS API provider
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/mailinabox)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Mail-In-A-Box](https://mailinabox.email/) custom DNS API,
allowing you to manage DNS records.


```go
import (
	"context"
	"fmt"

	"github.com/libdns/mailinabox"
)

func GetSubDomains() []string {
	zone := "[your mailinabox root domain]." // <- note the trailing .
	provider := &mailinabox.Provider{
		APIURL:       "https://[your mailinabox box]/admin/dns/custom",
		EmailAddress: "[create a special account on your box for managing domains]",
		Password:     "[password of the special dns account]",
	}
	records, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		fmt.Printf("Error fetching records: %s", err)
		return nil
	}

	subDomains := make([]string, len(records))
	for i, record := range records {
		subDomains[i] = record.Name
	}
	return subDomains
}
```
