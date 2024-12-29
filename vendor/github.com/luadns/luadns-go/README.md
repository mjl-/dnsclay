# LuaDNS Go Client

This is the Go client for [LuaDNS REST API](https://www.luadns.com/api.html).


[![Build Status](https://github.com/luadns/luadns-go/actions/workflows/ci.yml/badge.svg)](https://github.com/luadns/luadns-go/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/luadns/luadns-go?status.svg)](https://godoc.org/github.com/luadns/luadns-go)


Usage:

``` go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	api "github.com/luadns/luadns-go"
)

var email string
var key string

func main() {
	flag.StringVar(&email, "email", "joe@example.com", "your email address")
	flag.StringVar(&key, "key", "", "your API key")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	c := api.NewClient(email, key)
	user, err := c.Me(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("email:  ", user.Email)
	fmt.Println("name:   ", user.Name)
	fmt.Println("package:", user.Package)

	zones, err := c.ListZones(ctx, &api.ListParams{Query: "example.org"})
	if err != nil {
		log.Fatalln(err)
	}

	for _, z := range zones {
		fmt.Println("===> zone", z.Name)
		records, err := c.ListRecords(ctx, z, &api.ListParams{})
		if err != nil {
			log.Fatalln(err)
		}

		for _, r := range records {
			fmt.Println("    ", r.Name, r.Type, r.Content, r.TTL)
		}
	}
}
```

## Example
```
$ go run examples/main.go -email=YOUR_EMAIL -key=YOUR_API_KEY
email:   joe@example.com
name:    Example User
package: Free
===> zone example.org
     example.org. SOA ns1.luadns.net. hostmaster.luadns.net. 1693213328 1200 120 604800 3600 3600
     example.org. NS ns1.luadns.net. 86400
     example.org. NS ns2.luadns.net. 86400
     example.org. NS ns3.luadns.net. 86400
     example.org. NS ns4.luadns.net. 86400
     example.org. A 1.1.1.1 86400
     mail.example.org. CNAME ghs.google.com. 86400
     www.example.org. CNAME example.org. 86400
     example.org. MX 5 aspmx.l.google.com. 86400
     _sip._udp.example.org. SRV 0 0 5060 sip.example.com. 86400
     example.org. TXT v=spf1 a mx include:_spf.google.com ~all 86400
```
