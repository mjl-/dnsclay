`nfsn` for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/nfsn)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for
[nearlyfreespeech.net](https://www.nearlyfreespeech.net), allowing you to manage DNS records hosted
there.

## Configuration

The `Provider` requires two pieces of information to operate:

1. A `Login` - this is your NFSN login, i.e. the "Login Name" you provide when authenticating
   against https://members.nearlyfreespeech.net
2. An `APIKey` - obtained by logging into NFSN, navigating to the 'profile' tab, and selecting
   'Manage API Key'. More details on obtaining and managing API keys are available in the [NFSN
   FAQs](https://members.nearlyfreespeech.net/faq).

## Caveats

The API that backs `SetRecords` only supports `A` and `AAAA` records. All other record types need to
be deleted and re-created in separate steps.

## CLI

`cli/cli.go` contains a (bare bones) CLI driver for the package. To use it, put an NFSN API key in a
file on disk. By default the tool will look for a file called `api_key.txt` in the current working
directory. The `-f` argument allows specifying an alternative location and/or filename. The `-z`
zone argument and the `-l` login argument are required for all commands. The CLI tool supports the
following operations in the `-o` argument:

* `GetRecords` retrieves the set of DNS records for the specified zone and prints them to stdout.
* `AddRecord` adds a new record. Takes the `-t` type, `-n` name, and `-d` data arguments.
* `DeleteRecord` deltes a record. Takes the `-t` type, `-n` name, and `-d` data arguments.
* `SetRecord` replaces an existing A or AAAA record transactionally (the API does not support other
  types of records). Takes the `-t` type, `-n` name, and `-d` data arguments.

## Reference

_Note: these require an NFSN account to access._

* [NFSN API Introduction](https://members.nearlyfreespeech.net/wiki/API/Introduction)
* [NFSN API Reference](https://members.nearlyfreespeech.net/wiki/API/Reference)
