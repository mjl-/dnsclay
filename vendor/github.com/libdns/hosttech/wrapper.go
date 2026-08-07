package hosttech

import (
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
)

type HosttechZoneListResponseWrapper struct {
	Data []HosttechZone `json:"data"`
}

type HosttechListResponseWrapper struct {
	Data []HosttechRecordWrapper `json:"data"`
}

type HosttechSingleResponseWrapper struct {
	Data HosttechRecordWrapper `json:"data"`
}

type HosttechRecordWrapper struct {
	value HosttechRecord
}

func (h HosttechRecordWrapper) toLibdnsRecord(zone string) libdns.Record {
	return h.value.toLibdnsRecord(zone)
}

func (h HosttechRecordWrapper) fromLibdnsRecord(record libdns.Record) {
	h.value.fromLibdnsRecord(record)
}

func (h *HosttechRecordWrapper) UnmarshalJSON(b []byte) error {
	var base Base
	err := json.Unmarshal(b, &base)
	if err != nil {
		return err
	}
	switch base.Type {
	case "AAAA":
		record := AAAARecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "A":
		record := ARecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "NS":
		record := NSRecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "CNAME":
		record := CNAMERecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "MX":
		record := MXRecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "TXT":
		record := TXTRecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	case "TLSA":
		record := TLSARecord{}
		err = json.Unmarshal(b, &record)
		h.value = HosttechRecord(record)
	default:
		err = fmt.Errorf(`record type "%s" is not supported"`, base.Type)
	}
	if err != nil {
		return err
	}

	return nil
}

func LibdnsRecordToHosttechRecordWrapper(record libdns.Record) (HosttechRecord, error) {
	var hosttechRecord HosttechRecord

	switch record.Type {
	case "AAAA":
		hosttechRecord = AAAARecord{}
	case "A":
		hosttechRecord = ARecord{}
	case "NS":
		hosttechRecord = NSRecord{}
	case "CNAME":
		hosttechRecord = CNAMERecord{}
	case "MX":
		hosttechRecord = MXRecord{}
	case "TXT":
		hosttechRecord = TXTRecord{}
	case "TLSA":
		hosttechRecord = TLSARecord{}
	default:
		return nil, fmt.Errorf(`record type "%s" is not supported"`, record.Type)
	}

	return hosttechRecord.fromLibdnsRecord(record), nil
}
