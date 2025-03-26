// NOTE: generated by gendoc.sh

/*
Dnsclay implements a DNS server that translates DNS UPDATE (RFC 2136) and DNS
AXFR (RFC 5936, zone transfers) requests to the many custom cloud DNS operator
APIs for managing DNS records/zones. Dnsclay keeps a local copy of the records,
periodically synchronizes its copy with authoritative data at the cloud DNS
operator, and sends DNS NOTIFY (RFC 1996) messages to configured listeners when
any records changed. Dnsclay also has a web interface for managing the
configured zones, and for viewing and editing records.

Most cloud DNS operators implement their own custom APIs for changing DNS
records. Application developers are tempted to add support for long lists
of those custom APIs to their applications so they can make automated DNS
changes (even just for handling ACME verification through DNS). This is
time-consuming and error-prone. Developers can instead settle on the standard
DNS interfaces with UPDATE/AXFR/NOTIFY, talking either directly to DNS servers
that implement them (like BIND, Knot), or talking to dnsclay which does the
translating.

Dnsclay implements TLS with the option for client certificate authentication
(mutual TLS) based on public keys (ignoring certificate
name/expiration/constraints, keeping it simple). DNS TSIG (RFC 8945) is also
supported.

Dnsclay helps diagnosing errors by returning error responses with Extended DNS
Errors (RFC 8914) to requests with EDNS0.

Dnsclay does not answer regular DNS queries for records (recursive or
authoritative), with the exception of giving authoritative answers to SOA
queries. Clients can use this to check if the zone has been updated before
deciding to do an AXFR of the full zone.

One of the implemented backend providers, "rfc2136", connects to DNS servers
implementing the standard DNS UPDATE/AXFR protocols, making dnsclay a web-based
zone editor for standard DNS servers.

Like secondary DNS servers, dnsclay periodically fetches the SOA record from
authoritative name servers, and does an AXFR if the zone serial changes.
Dnsclay also periodically does a full sync regardless of SOA serial, since some
DNS operators don't change the serial when a zone changes. In such cases,
dnsclay will keep track of its own serial, so its clients can properly detect
zone changes. The "refresh interval" from the SOA record is not used, since it
is often configured to work only with the setup of the primary/secondary
servers of the DNS operator.  After a change to a zone, either because of DNS
UPDATE through dnsclay or by dnsclay detecting a record change at the DNS
operator, dnsclay will temporarily increase the interval with which it checks
again for a new update, speculating more changes are coming. Timely
notification of DNS record changes is useful during lock-step changes like key
rollovers. Cloud DNS operators typically don't have a mechanism to notify
applications of changes to records.

# Limitations

DNS UPDATE/AXFR/NOTIFY may look relatively complicated to application
developers interested in making automated DNS changes. They may be expecting a
HTTP/JSON API. If one is standardized, dnsclay could implement it.

Changes in a DNS UPDATE request must be applied atomically: Either all the
changes in a request must be applied, or none. Dnsclay cannot implement this
requirement for all requests. With the libdns API, records cannot be added and
removed atomically.

Cloud DNS operators may have unexpected limitations. If standard DNS resource
record types are not implemented, adding them may result in an error.

The dnsclay server does not process multiple messages on a single TCP connection
in parallel. It reads a request, process it, and writes a response, then starts
on the next request. Multiple connections, and UDP packets, are handled in
parallel.

# Usage for "dnsclay"

	usage: dnsclay serve [flags]
	       dnsclay genkey >privkey-ed25519.pkcs8.pem
	       dnsclay dns [flags] notify [flags] addr zone
	       dnsclay dns [flags] update [flags] addr zone [add|del name type ttl value] ...
	       dnsclay dns [flags] xfr [flags] addr zone
	       dnsclay version
	       dnsclay license

# Usage for "dnsclay serve"

	usage: dnsclay serve [flags]
	  -adminaddr string
	    	address to serve admin interface on (default "localhost:8053")
	  -adminpasswordpath string
	    	file with admin password for http basic auth; if absent, a random password is generated and written (default "adminpassword")
	  -dns-notify-tcpaddr string
	    	comma-separated tcp address to listen for dns notify messages on
	  -dns-notify-tlsaddr string
	    	comma-separated tls address to listen for dns notify messages on
	  -dns-udpaddr string
	    	comma-separated udp address to serve dns notify and authoritative soa requests on (default "localhost:1053")
	  -dns-upxfr-tcpaddr string
	    	comma-separated tcp address to serve dns update and axfr requests on (default "localhost:1053")
	  -dns-upxfr-tlsaddr string
	    	comma-separated tls address to serve dns update and axfr requests on (default "localhost:1853")
	  -loglevel value
	    	log level: error, warn, info, debug (default INFO)
	  -metricsaddr string
	    	address to serve prometheus metrics on; can be same as adminaddr, no authentication needed (default "localhost:8053")
	  -tlscertpem string
	    	path to pem file with one or more certificates; if empty, an ephemeral minimalistic certificate is generated for the private key
	  -tlskeypem string
	    	path to pem file with pkcs#8 private key file, for dns tls server; if empty an ephemeral tls key is generated at startup; if left at default, file is created if missing (default "server.privkey-ed25519.pkcs8.pem")
	  -trace string
	    	if non-empty, comma-separated formats to log dns request/response traces: text for textual format, json for json, jsonindent for multi-line indented json

# Usage for "dnsclay dns notify"

	usage: dnsclay dns [flags] notify [-trace] [-json] addr zone
	  -json
	    	print dns packets in json too
	  -trace
	    	print dns packets

# Usage for "dnsclay dns update"

	usage: dnsclay dns [flags] update [-trace] [-json] addr zone [add name ttl type value | delname name | deltype name type | delrecord name type value] ...
	  -json
	    	print dns packets in json too
	  -trace
	    	print dns packets

# Usage for "dnsclay dns xfr"

	usage: dnsclay dns [flags] xfr [-i serial] [-json] addr zone
	  -i uint
	    	serial number for IXFR request instead of default AXFR
	  -json
	    	print dns packets in json too
	  -query
	    	print dns query

# Providers

The following providers are implemented in dnsclay, with community-provided
implementations maintained at https://github.com/libdns:

  - alidns
  - autodns
  - azure
  - bunny
  - civo
  - cloudflare
  - cloudns
  - ddnss
  - desec
  - digitalocean
  - directadmin
  - dnsimple
  - dnsmadeeasy
  - dnspod
  - dnsupdate
  - domainnameshop
  - dreamhost
  - duckdns
  - dynu
  - dynv6
  - easydns
  - exoscale
  - gandi
  - gcore
  - glesys
  - godaddy
  - googleclouddns
  - he
  - hetzner
  - hexonet
  - hosttech
  - huaweicloud
  - infomaniak
  - inwx
  - ionos
  - katapult
  - leaseweb
  - linode
  - loopia
  - luadns
  - mailinabox
  - metaname
  - mythicbeasts
  - namecheap
  - namedotcom
  - namesilo
  - nanelo
  - netcup
  - netlify
  - nfsn
  - njalla
  - ovh
  - porkbun
  - powerdns
  - rfc2136
  - route53
  - scaleway
  - selectel
  - tencentcloud
  - timeweb
  - totaluptime
  - vultr
*/
package main
