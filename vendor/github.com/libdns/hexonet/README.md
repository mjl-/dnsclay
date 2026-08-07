Hexonet provider for [`libdns`](https://github.com/libdns/libdns)
=======================


This package implements the [libdns interfaces](https://github.com/libdns/libdns) for
[Hexonet](https://hexonet.net/), allowing you to
manage DNS records.

To configure this, simply specify the username and the password.


    package main

    import (
        "context"

        "github.com/libdns/libdns"
        "github.com/libdns/hexonet"
    )

    func main() {
        p := &hexonet.Provider{
            Username: "abcde12312312", // required
            Password:  "@#$#12312312",        // required
        }

        _, err := p.AppendRecords(context.Background(), "example.org.", []libdns.Record{
            {
                Name:  "_acme_whatever",
                Type:  "TXT",
                Value: "123456",
            },
        })
        if err != nil {
            panic(err)
        }

    }
