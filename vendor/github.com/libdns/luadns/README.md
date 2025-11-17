# LuaDNS for [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/luadns)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [LuaDNS](https://www.luadns.com/api.html), allowing you to manage DNS records.

Usage:

```go
// Init Provider struct.
provider := luadns.Provider{
	Email:  email,
	APIKey: key,
}

// List account zones.
zones, err := provider.ListZones(ctx)


// List zone records.
records, err := provider.GetRecords(ctx, zone)


// Create new records.
records, err := provider.AppendRecords(ctx, zone, []libdns.Record{
	libdns.RR{
		Name: "foo",
		Type: "TXT",
		Data: "foo",
		TTL:  3600 * time.Second,
	},
})


// Set zone records for input (name, type) pairs with supplied records.
records, err := provider.SetRecords(ctx, zone, []libdns.Record{
	libdns.RR{
		Name: "foo",
		Type: "TXT",
		Data: "bar",
		TTL:  3600 * time.Second,
	},
})


// Delete records when matching supplied name, type, data and TTL.
records, err := provider.DeleteRecords(ctx, zone, []libdns.Record{
	libdns.RR{
		Name: "foo",
		Type: "TXT",
	},
})
```

For a complete example see [_examples/main.go](_examples/main.go).
