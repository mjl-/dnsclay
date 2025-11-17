package infomaniak

import (
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Default TTL that is applied if none is provided - infomaniak requires a TTL
const defaultTtlSecs = 300

// ToLibDnsRecord maps a infomaniak dns record to a libdns record
func (ikr *IkRecord) ToLibDnsRecord(zoneMapping *ZoneMapping) (libdns.Record, error) {
	switch ikr.Type {
	case "A", "AAAA":
		return ikr.toAddressRecord(zoneMapping)
	case "CAA":
		return ikr.toCaaRecord(zoneMapping), nil
	case "CNAME":
		return ikr.toCNameRecord(zoneMapping), nil
	case "MX":
		return ikr.toMxRecord(zoneMapping), nil
	case "NS":
		return ikr.toNsRecord(zoneMapping), nil
	case "SRV":
		return ikr.toServiceRecord(zoneMapping), nil
	case "TXT":
		return ikr.toTextRecord(zoneMapping), nil
	default:
		return libdns.RR{
			Name: zoneMapping.ToRelativeLibdnsName(ikr.Source),
			TTL:  ikr.getTtlAsTimeDuration(),
			Type: ikr.Type,
			Data: ikr.Target,
		}.Parse()
	}
}

// getTtlAsTimeDuration returns the TTL of a infomaniak DNS record as a time duration
func (ikr *IkRecord) getTtlAsTimeDuration() time.Duration {
	return time.Duration(ikr.TtlInSec * int(time.Second))
}

// toAddressRecord parses an infomaniak DNS record as a libdns Address record
func (ikr *IkRecord) toAddressRecord(zoneMapping *ZoneMapping) (libdns.Address, error) {
	addr, err := netip.ParseAddr(ikr.Target)
	if err != nil {
		return libdns.Address{}, err
	}

	return libdns.Address{
		Name: zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:  ikr.getTtlAsTimeDuration(),
		IP:   addr,
	}, nil
}

// toCaaRecord parses an infomaniak DNS record as a libdns CAA record
func (ikr *IkRecord) toCaaRecord(zoneMapping *ZoneMapping) libdns.CAA {
	return libdns.CAA{
		Name:  zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:   ikr.getTtlAsTimeDuration(),
		Flags: uint8(ikr.Description.Flags.Value),
		Tag:   ikr.Description.Tag.Value,
		Value: ikr.getLastTargetValue(),
	}
}

// toCNameRecord parses an infomaniak DNS record as a libdns CNAME record
func (ikr *IkRecord) toCNameRecord(zoneMapping *ZoneMapping) libdns.CNAME {
	return libdns.CNAME{
		Name:   zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:    ikr.getTtlAsTimeDuration(),
		Target: ikr.Target,
	}
}

// toMxRecord parses an infomaniak DNS record as a libdns MX record
func (ikr *IkRecord) toMxRecord(zoneMapping *ZoneMapping) libdns.MX {
	return libdns.MX{
		Name:       zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:        ikr.getTtlAsTimeDuration(),
		Preference: uint16(ikr.Description.Priority.Value),
		Target:     ikr.getLastTargetValue(),
	}
}

// toNsRecord parses an infomaniak DNS record as a libdns NS record
func (ikr *IkRecord) toNsRecord(zoneMapping *ZoneMapping) libdns.NS {
	return libdns.NS{
		Name:   zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:    ikr.getTtlAsTimeDuration(),
		Target: ikr.Target,
	}
}

// toServiceRecord parses an infomaniak DNS record as a libdns SRV record
func (ikr *IkRecord) toServiceRecord(zoneMapping *ZoneMapping) libdns.SRV {
	parts := strings.SplitN(ikr.Source, ".", 2)
	return libdns.SRV{
		Service:   strings.TrimPrefix(parts[0], "_"),
		Transport: strings.TrimPrefix(ikr.Description.Protocol.Value, "_"),
		Name:      zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:       ikr.getTtlAsTimeDuration(),
		Priority:  uint16(ikr.Description.Priority.Value),
		Weight:    uint16(ikr.Description.Weight.Value),
		Port:      uint16(ikr.Description.Port.Value),
		Target:    ikr.getLastTargetValue(),
	}
}

// toTextRecord parses an infomaniak DNS record as a libdns TXT record
func (ikr *IkRecord) toTextRecord(zoneMapping *ZoneMapping) libdns.TXT {
	return libdns.TXT{
		Name: zoneMapping.ToRelativeLibdnsName(ikr.Source),
		TTL:  ikr.getTtlAsTimeDuration(),
		Text: ikr.Target,
	}
}

// getLastTargetValue parses last value of the record's target
func (ikr *IkRecord) getLastTargetValue() string {
	parts := strings.Split(ikr.Target, " ")
	targetValue := parts[len(parts)-1]
	unquoted, err := strconv.Unquote(targetValue)
	if err == nil {
		targetValue = unquoted
	}
	return targetValue
}

// ToRelativeLibdnsName converts a relative name from the infomaniak managed zone
// to the input zone of the libdns caller
func (zoneMapping *ZoneMapping) ToRelativeLibdnsName(relativeName string) string {
	return zoneMapping.convertZone(relativeName, zoneMapping.InfomaniakManagedZone, zoneMapping.LibDnsZone)
}

// ToRelativeInfomaniakName converts a relative name from input zone of the libdns caller
// to the infomaniak managed zone
func (zoneMapping *ZoneMapping) ToRelativeInfomaniakName(relativeName string) string {
	return zoneMapping.convertZone(relativeName, zoneMapping.LibDnsZone, zoneMapping.InfomaniakManagedZone)
}

// convertZone converts a relative name from a source zone to a target zone
func (zoneMapping *ZoneMapping) convertZone(relativeName string, sourceZone string, targetZone string) string {
	return libdns.RelativeName(libdns.AbsoluteName(relativeName, sourceZone), targetZone)
}

// ToInfomaniakRecord maps a libdns record to a infomaniak dns record
func ToInfomaniakRecord(libdnsRec libdns.Record, zoneMapping *ZoneMapping) IkRecord {
	rr := libdnsRec.RR()

	rec := IkRecord{
		Source:   zoneMapping.ToRelativeInfomaniakName(rr.Name),
		Type:     rr.Type,
		TtlInSec: int(rr.TTL.Seconds()),
		Target:   rr.Data,
	}

	if rec.TtlInSec < 60 {
		rec.TtlInSec = defaultTtlSecs
	}

	return rec
}
