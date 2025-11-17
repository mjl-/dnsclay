package tencentcloud

import (
	"errors"
	"strconv"
	"time"

	"github.com/libdns/libdns"
)

var ErrRecordNotFound = errors.New("record not found")
var ErrNotValid = errors.New("returned value is not valid")

type Provider struct {
	SecretId     string
	SecretKey    string
	SessionToken string
	Region       string
}

type CreateModifyRecordRequest struct {
	Domain     string `json:"Domain"`
	SubDomain  string `json:"SubDomain,omitempty"`
	RecordType string `json:"RecordType,omitempty"`
	RecordLine string `json:"RecordLine,omitempty"`
	Value      string `json:"Value,omitempty"`
	TTL        int64  `json:"TTL,omitempty"`
	RecordId   uint64 `json:"RecordId,omitempty"`
}

type FindRecordRequest struct {
	Domain     string `json:"Domain"`
	RecordType string `json:"RecordType,omitempty"`
	RecordLine string `json:"RecordLine,omitempty"`
	Subdomain  string `json:"Subdomain,omitempty"`
	Limit      int64  `json:"Limit,omitempty"`
}

type DeleteRecordRequest struct {
	Domain   string `json:"Domain"`
	RecordId uint64 `json:"RecordId"`
}

type Response struct {
	Response ResponseData `json:"Response"`
}

type ResponseData struct {
	RecordList []RecordInfo `json:"RecordList,omitempty"`
	RecordId   uint64       `json:"RecordId,omitempty"`
	Error      *ErrorInfo   `json:"Error,omitempty"`
}

type RecordInfo struct {
	RecordId int64  `json:"RecordId"`
	Type     string `json:"Type"`
	Name     string `json:"Name"`
	Value    string `json:"Value"`
	TTL      int64  `json:"TTL"`
	MX       int    `json:"MX,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type record struct {
	Type  string
	Name  string
	Value string
	TTL   time.Duration
	MX    int
}

func (r record) libdnsRecord() (libdns.Record, error) {
	if r.Type == "MX" {
		r.Value = strconv.Itoa(r.MX) + " " + r.Value
	}
	return libdns.RR{
		Type: r.Type,
		Name: r.Name,
		Data: r.Value,
		TTL:  r.TTL,
	}.Parse()
}

func fromLibdnsRecord(r libdns.Record) record {
	rr := r.RR()

	host := rr.Name
	if host == "@" {
		host = ""
	}

	if rr.TTL == 0 {
		rr.TTL = 600
	}

	return record{
		Type:  rr.Type,
		Name:  host,
		Value: rr.Data,
		TTL:   rr.TTL,
	}
}
