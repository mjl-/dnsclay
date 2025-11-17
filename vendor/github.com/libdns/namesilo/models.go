package namesilo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

type intString int

func (i *intString) UnmarshalJSON(data []byte) error {
	number, err := strconv.Atoi(strings.Trim(string(data), "\""))
	*i = intString(number)
	return err
}

type record struct {
	ID       string    `json:"record_id"`
	Type     string    `json:"type"`
	Host     string    `json:"host"`
	Value    string    `json:"value"`
	TTL      intString `json:"ttl"`
	Distance intString `json:"distance,omitempty"`
}

func (n record) toLibDNS(zone string) (libdns.Record, error) {
	// format MX record
	if n.Type == "MX" {
		n.Value = fmt.Sprintf("%d %s", n.Distance, n.Value)
	}

	return libdns.RR{
		Type: n.Type,
		Name: libdns.RelativeName(n.Host, zone),
		Data: n.Value,
		TTL:  time.Duration(n.TTL) * time.Second,
	}.Parse()
}

func namesiloRecord(zone string, r libdns.Record) (record, error) {
	rr := r.RR()

	value := rr.Data
	distance := 0
	if rr.Type == "MX" {
		fields := strings.Fields(rr.Data)
		if expectedFields := 2; len(fields) != expectedFields {
			return record{}, fmt.Errorf("expected data to contain %d fields, but had %d", expectedFields, len(fields))
		}
		value = fields[1]
		convertedDistance, err := strconv.Atoi(fields[0])
		if err != nil {
			return record{}, nil
		}
		distance = convertedDistance
	}

	host := rr.Name
	if host == "@" {
		host = ""
	}

	return record{
		Type:     rr.Type,
		Host:     host,
		TTL:      intString(rr.TTL.Seconds()),
		Value:    value,
		Distance: intString(distance),
	}, nil
}

type namesiloResponse struct {
	Reply struct {
		Code           int             `json:"code"`
		Detail         string          `json:"detail"`
		ResourceRecord json.RawMessage `json:"resource_record,omitempty"`
	}
}
