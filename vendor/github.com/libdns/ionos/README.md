# IONOS DNS API for `libdns`

This package implements the libdns interfaces for the [IONOS DNS
API](https://developer.hosting.ionos.de/docs/dns)

## Authenticating

To authenticate you need to supply a IONOS API Key, as described on
https://developer.hosting.ionos.de/docs/getstarted

## Example

Here's a minimal example of how to get all DNS records for zone.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/libdns/ionos"
)

func main() {
	token := os.Getenv("LIBDNS_IONOS_TOKEN")
	if token == "" {
		panic("LIBDNS_IONOS_TOKEN not set")
	}

	zone := os.Getenv("LIBDNS_IONOS_ZONE")
	if zone == "" {
		panic("LIBDNS_IONOS_ZONE not set")
	}

	p := &ionos.Provider{
		AuthAPIToken: token,
	}

	records, err := p.GetRecords(context.TODO(), zone)
	if err != nil {
		panic(err)
	}

	out, _ := json.MarshalIndent(records, "", "  ")
	fmt.Println(string(out))
}
```

## Test

The file `provisioner_test.go` contains an end-to-end test suite, using the
original IONOS API service (i.e. no test doubles - be careful). To run the
tests:

```console
$ export LIBDNS_IONOS_TEST_ZONE=mydomain.org
$ export LIBDNS_IONOS_TEST_TOKEN=aaaaaaaaaaa.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
$ go  test -v
go test -v
=== RUN   Test_AppendRecords
=== RUN   Test_AppendRecords/testcase_0
=== RUN   Test_AppendRecords/testcase_1
=== RUN   Test_AppendRecords/testcase_2
--- PASS: Test_AppendRecords (6.71s)
    --- PASS: Test_AppendRecords/testcase_0 (2.51s)
    --- PASS: Test_AppendRecords/testcase_1 (2.15s)
    --- PASS: Test_AppendRecords/testcase_2 (2.05s)
=== RUN   Test_DeleteRecords
=== RUN   Test_DeleteRecords/clear_record.ID=true
=== RUN   Test_DeleteRecords/clear_record.ID=false
--- PASS: Test_DeleteRecords (9.62s)
    --- PASS: Test_DeleteRecords/clear_record.ID=true (4.81s)
    --- PASS: Test_DeleteRecords/clear_record.ID=false (4.80s)
=== RUN   Test_GetRecords
--- PASS: Test_GetRecords (4.41s)
=== RUN   Test_UpdateRecords
=== RUN   Test_UpdateRecords/clear_record.ID=true
=== RUN   Test_UpdateRecords/clear_record.ID=false
--- PASS: Test_UpdateRecords (10.14s)
    --- PASS: Test_UpdateRecords/clear_record.ID=true (5.84s)
    --- PASS: Test_UpdateRecords/clear_record.ID=false (4.30s)
PASS
ok  	github.com/libdns/ionos	30.884s
```

## Author

Original Work (C) Copyright 2020 by matthiasng (based on https://github.com/libdns/hetzner),
this version (C) Copyright 2021 by Jan Delgado (github.com/jandelgado).

License: MIT

