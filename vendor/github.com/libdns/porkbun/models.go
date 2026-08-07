package porkbun

import (
	"fmt"
	"github.com/libdns/libdns"
	"strconv"
	"time"
)

type pkbnRecord struct {
	Content string `json:"content"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Notes   string `json:"notes"`
	Prio    string `json:"prio"`
	TTL     string `json:"ttl"`
	Type    string `json:"type"`
}

type pkbnRecordsResponse struct {
	Records []pkbnRecord `json:"records"`
	Status  string       `json:"status"`
}

type ApiCredentials struct {
	Apikey       string `json:"apikey"`
	Secretapikey string `json:"secretapikey"`
}

type pkbnResponseStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
type pkbnPingResponse struct {
	pkbnResponseStatus
	YourIP string `json:"yourIp"`
}

type pkbnCreateResponse struct {
	pkbnResponseStatus
	// TODO contact support endpoint isn't returning the ID despite it being in their docs.
	// ID string `json:"id"`
}

func (record pkbnRecord) toLibdnsRecord(zone string) libdns.Record {
	ttl, _ := time.ParseDuration(record.TTL + "s")
	priority, _ := strconv.Atoi(record.Prio)
	return libdns.Record{
		ID:       record.ID,
		Name:     libdns.RelativeName(record.Name, LibdnsZoneToPorkbunDomain(zone)),
		Priority: uint(priority),
		TTL:      ttl,
		Type:     record.Type,
		Value:    record.Content,
	}
}

func (a pkbnResponseStatus) Error() string {
	return fmt.Sprintf("%s: %s", a.Status, a.Message)
}

type pkbnRecordPayload struct {
	*ApiCredentials
	Content string `json:"content"`
	Name    string `json:"name"`
	TTL     string `json:"ttl"`
	Type    string `json:"type"`
}
