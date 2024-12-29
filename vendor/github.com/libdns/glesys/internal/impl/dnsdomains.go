// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2017 GleSYS Internet Services AB

package impl

import "context"

// DNSDomainService provides functions to interact with dns domains
type DNSDomainService struct {
	client clientInterface
}

// DNSDomainRecord - data in the domain
type DNSDomainRecord struct {
	DomainName string `json:"domainname"`
	Data       string `json:"data"`
	Host       string `json:"host"`
	RecordID   int    `json:"recordid"`
	TTL        int    `json:"ttl"`
	Type       string `json:"type"`
}

// AddRecordParams - parameters for updating domain records
type AddRecordParams struct {
	DomainName string `json:"domainname"`
	Data       string `json:"data"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	TTL        int    `json:"ttl,omitempty"`
}

// UpdateRecordParams - parameters for updating domain records
type UpdateRecordParams struct {
	RecordID int    `json:"recordid"`
	Data     string `json:"data,omitempty"`
	Host     string `json:"host,omitempty"`
	Type     string `json:"type,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

// ListRecords - return a list of all records for domain
func (s *DNSDomainService) ListRecords(context context.Context, domainname string) (*[]DNSDomainRecord, error) {
	data := struct {
		Response struct {
			Records []DNSDomainRecord
		}
	}{}
	err := s.client.post(context, "domain/listrecords", &data, struct {
		Name string `json:"domainname"`
	}{domainname})
	return &data.Response.Records, err
}

// AddRecord - add a domain record
func (s *DNSDomainService) AddRecord(context context.Context, params AddRecordParams) (*DNSDomainRecord, error) {
	data := struct {
		Response struct {
			Record DNSDomainRecord
		}
	}{}
	err := s.client.post(context, "domain/addrecord", &data, params)
	return &data.Response.Record, err
}

// UpdateRecord - update a domain record
func (s *DNSDomainService) UpdateRecord(context context.Context, params UpdateRecordParams) (*DNSDomainRecord, error) {
	data := struct {
		Response struct {
			Record DNSDomainRecord
		}
	}{}
	err := s.client.post(context, "domain/updaterecord", &data, params)
	return &data.Response.Record, err
}

// DeleteRecord deletes a record
func (s *DNSDomainService) DeleteRecord(context context.Context, recordID int) error {
	return s.client.post(context, "domain/deleterecord", nil, struct {
		RecordID int `json:"recordid"`
	}{recordID})
}
