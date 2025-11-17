<!--
SPDX-FileCopyrightText: 2024 Peter Magnusson <me@kmpm.se>

SPDX-License-Identifier: MIT
-->

Glesys for [`libdns`](https://github.com/libdns/libdns)
==============================================================

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [Glesys](https://glesys.se), allowing you to manage DNS records.
It utilizes [glesys-go](https://github.com/glesys/glesys-go) for API communication.

## Usage
```golang
include (
    glesys "github.com/libdns/glesys"
)
p := &glesys.Provider{
    Project: "your project/username usually clXXXXXX",
    ApiKey: "api-key",
}

zone := "example.org"
records, err := p.GetRecords(ctx, zone)
```
For more examples check the `_examples` folder in the source.

## Noteworthy
To do everything this library can do the Glesys API user needs permissions to the following...

- Domain.addrecord
- Domain.deleterecord
- Domain.listrecords
- Domain.updaterecord

## Development
### Testing
```shell
make test
```

If you have a domain available at glesys and want to do integration testing
then set 3 environment variables first.
```shell
export GLESYS_PROJECT="<your glesys project id>"
export GLESYS_KEY="<your glesys api-key>"
export GLESYS_ZONE="<yourdomain.touse>"
```
This will leave a `TXT` record called `_libdns-test` with the text of the current date and time 
in your DNS settings.

There is a "secret" way of enabling some debug output from libdns.
If you set the environment key `LIBDNS_GLESYS_DEBUG` to `true` (or something
parsable to a boolean true) then you will see som classic debug prints.
