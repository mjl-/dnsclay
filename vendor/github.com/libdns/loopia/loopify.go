package loopia

import (
	"fmt"
	"strings"

	"github.com/libdns/libdns"
)

// loopia does not have support for propper subdomains so
// we need so that zone only contains <domain>.<tld>
func loopify(name, zone string) (string, string) {
	components := strings.Split(zone, ".")
	split := 2
	l := len(components)
	if components[l-1] == "" {
		split = 3
	}
	if l > split {
		name = fmt.Sprintf("%s.%s", name, strings.Join(components[:l-split], "."))
		zone = strings.Join(components[len(components)-split:], ".")
	}
	return name, zone
}

// modifies records in place
func loopifyRecords(zone string, records []libdns.Record) (hostSuffix string, domain string) {
	hostSuffix, domain = loopify("", zone)
	if hostSuffix != "" && len(records) > 0 {
		for i, r := range records {
			records[i].Name = r.Name + hostSuffix
		}
	}
	return hostSuffix, domain
}

// unLoopify modifies name and zone so that name should only contain hostname and
// everything else should end up in zone.
func unLoopify(name, zone string) (string, string) {
	components := strings.Split(name, ".")
	l := len(components)
	if l > 1 {
		name = components[0]
		zone = fmt.Sprintf("%s.%s", strings.Join(components[1:], "."), zone)
	}
	return name, zone
}

func unLoopifyName(hostSuffix string, record *libdns.Record) {
	if hostSuffix != "" {
		record.Name = strings.TrimSuffix(record.Name, hostSuffix)
	}
}

func unLoopifyRecords(hostSuffix string, records []libdns.Record) {
	if len(records) > 0 {
		for i, _ := range records {
			unLoopifyName(hostSuffix, &records[i])
		}
	}
}
