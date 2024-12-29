Huawei Cloud for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/huaweicloud)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Huawei Cloud DNS](https://www.huaweicloud.com/product/dns.html), allowing you to manage DNS records.

## Authenticating

To authenticate you need to supply your AccessKeyId and SecretAccessKey to the Provider.

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
	"context"
	"fmt"

	"github.com/libdns/huaweicloud"
)

func main() {
	provider := huaweicloud.Provider{
		AccessKeyId: "<AccessKeyId form your huaweicloud console>",
		SecretAccessKey: "<SecretAccessKey form your huaweicloud console>",
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
