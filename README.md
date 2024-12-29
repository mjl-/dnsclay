# dnsclay

DNS UPDATE/AXFR/NOTIFY to custom DNS API gateway.

Dnsclay implements a DNS server that translates DNS UPDATE (RFC 2136) and DNS
AXFR (RFC 5936, zone transfers) requests to the many custom cloud DNS operator
APIs for managing DNS records/zones. Dnsclay keeps a local copy of the records,
periodically synchronizes its copy with authoritative data at the cloud DNS
operator, and sends DNS NOTIFY (RFC 1996) messages to configured listeners
when any records changed. Dnsclay also has a web interface for managing the
configured zones, and for viewing and editing records.

Most cloud DNS operators implement their own custom APIs for changing DNS
records. Application developers are tempted to add support for long lists of
those custom APIs to their applications so they can make automated DNS changes
(even just for handling ACME verification through DNS). This is time-consuming
and error-prone. Developers can instead settle on the standard DNS interfaces
with UPDATE/AXFR/NOTIFY, talking either directly to DNS servers that implement
them (like BIND, Knot), or talking to dnsclay which does the translating.

For more information, see the documentation:

https://pkg.go.dev/github.com/mjl-/dnsclay


# Installing

Get the latest binary:

https://beta.gobuilds.org/github.com/mjl-/dnsclay@latest/linux-amd64-latest-stripped/

Or compile it locally (requires a recent Go toolchain):

	GOBIN=$PWD CGO_ENABLED=0 go install github.com/mjl-/dnsclay@latest

To start:

	./dnsclay serve

Running this for the first time creates an admin password for the web interface,
and a TLS private key for the DNS server. Use flags to the serve subcommand for
setting the IPs and ports to listen on.


# Providers

Support for all the cloud APIs is coming from the various community-maintained
providers at https://github.com/libdns. If your DNS operator of choice is
missing in dnsclay, check if someone has implemented a provider, or consider
implementing it yourself. See https://github.com/libdns/libdns.

All providers available at https://github.com/libdns at the time of writing have
been added, except:

- acmedns, only creates ACME TXT records
- acmeproxy, only creates ACME TXT records
- dinahosting, only creates ACME TXT records
- dode, only creates ACME TXT records
- vercel, cannot set TTL
- nicrudns, does not compile against latest libdns
- regfish, does not compile against latest libdns
- transip, requires a key in a file on disk
- openstack-designate, sherpadoc doesn't handle it being imported as openstack-designate and named openstackdesignate

## Adding a new provider

Adding a provider should be a matter of adding it to providers.txt (keep it
sorted!) and running "make build". It regenerates providers.go and syncs the Go
module dependencies. The config fields in the package's Provider should be
automatically processed, into both backend and frontend.


# About

Dnsclay is MIT-licensed, written by Mechiel Lukkien. Create an "issue" for bugs
or questions. Consider working on one of the open issues.  Please send
feedback/insights on automating DNS changes to mechiel@ueber.net.
