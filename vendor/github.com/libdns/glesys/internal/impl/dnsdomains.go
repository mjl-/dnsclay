// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2017 GleSYS Internet Services AB
package impl

import (
	"context"
)

// DNSDomainService provides functions to interact with dns domains
type DNSDomainService struct {
	client clientInterface
}

// DNSDomain represents a domain
type DNSDomain struct {
	Name                  string           `json:"domainname"`
	Available             bool             `json:"available,omitempty"`
	CreateTime            string           `json:"createtime,omitempty"`
	DisplayName           string           `json:"displayname,omitempty"`
	Expire                int              `json:"expire,omitempty"`
	Minimum               int              `json:"minimum,omitempty"`
	Prices                []DNSDomainPrice `json:"prices,omitempty"`
	PrimaryNameServer     string           `json:"primarynameserver,omitempty"`
	RecordCount           int              `json:"recordcount,omitempty"`
	Refresh               int              `json:"refresh,omitempty"`
	RegistrarInfo         RegistrarInfo    `json:"registrarinfo,omitempty"`
	ResponsiblePerson     string           `json:"responsibleperson,omitempty"`
	Retry                 int              `json:"retry,omitempty"`
	TTL                   int              `json:"ttl,omitempty"`
	UsingGlesysNameserver string           `json:"usingglesysnameserver,omitempty"`
}

// DNSDomainPrice represents the price for a single domain
type DNSDomainPrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Years    int     `json:"years"`
}

// AddDNSDomainParams - used for adding existing domains to GleSYS
// use CreateRecords = false to not create additional records.
type AddDNSDomainParams struct {
	Name              string `json:"domainname"`
	CreateRecords     string `json:"createrecords,omitempty"`
	Expire            int    `json:"expire,omitempty"`
	Minimum           int    `json:"minimum,omitempty"`
	PrimaryNameServer string `json:"primarynameserver,omitempty"`
	Refresh           int    `json:"refresh,omitempty"`
	ResponsiblePerson string `json:"responsibleperson,omitempty"`
	Retry             int    `json:"retry,omitempty"`
	TTL               int    `json:"ttl,omitempty"`
}

// EditDNSDomainParams - used when editing domain parameters
type EditDNSDomainParams struct {
	Name              string `json:"domainname"`
	Expire            int    `json:"expire,omitempty"`
	Minimum           int    `json:"minimum,omitempty"`
	PrimaryNameServer string `json:"primarynameserver,omitempty"`
	Refresh           int    `json:"refresh,omitempty"`
	ResponsiblePerson string `json:"responsibleperson,omitempty"`
	Retry             int    `json:"retry,omitempty"`
	TTL               int    `json:"ttl,omitempty"`
}

// RegistrarInfo contains information about the registrar for the domain
type RegistrarInfo struct {
	AutoRenew        string `json:"autorenew"`
	State            string `json:"state"`
	StateDescription string `json:"statedescription,omitempty"`
	Expire           string `json:"expire,omitempty"`
	TLD              string `json:"tld,omitempty"`
	InvoiceNumber    string `json:"invoicenumber,omitempty"`
}

// RegisterDNSDomainParams - parameters used when registering a domain
type RegisterDNSDomainParams struct {
	Name         string `json:"domainname"`
	Address      string `json:"address"`
	City         string `json:"city"`
	Country      string `json:"country"`
	Email        string `json:"email"`
	Firstname    string `json:"firstname"`
	Lastname     string `json:"lastname"`
	NationalID   int    `json:"nationalid"`
	Organization string `json:"organization"`
	PhoneNumber  string `json:"phonenumber"`
	ZipCode      string `json:"zipcode"`

	FaxNumber string `json:"fax,omitempty"`
	NumYears  int    `json:"numyears,omitempty"`
}

// DeleteDNSDomainParams - parameters for deleting a domain from the dns system.
// Set ForceDeleteEmail to true, to delete a domain AND email accounts for the domain.
type DeleteDNSDomainParams struct {
	Name             string `json:"domainname"`
	ForceDeleteEmail string `json:"forcedeleteemail,omitempty"`
}

// RenewDNSDomainParams - parameters to send when renewing a domain.
type RenewDNSDomainParams struct {
	Name     string `json:"domainname"`
	NumYears int    `json:"numyears"`
}

// SetAutoRenewParams - parameters to send for renewing a domain automatically.
type SetAutoRenewParams struct {
	Name         string `json:"domainname"`
	SetAutoRenew string `json:"setautorenew"`
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

// ChangeNameserverParams - parameters for updating the nameservers for domain
type ChangeNameserverParams struct {
	DomainName string `json:"domainname"`
	NS1        string `json:"NS1"`
	NS2        string `json:"NS2"`
	NS3        string `json:"NS3,omitempty"`
	NS4        string `json:"NS4,omitempty"`
}

// Available - checks if the domain is available
func (s *DNSDomainService) Available(context context.Context, search string) (*[]DNSDomain, error) {
	data := struct {
		Response struct {
			Domain []DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/available", &data, search)
	return &data.Response.Domain, err
}

// AddDNSDomain - add an existing domain to your GleSYS account
func (s *DNSDomainService) AddDNSDomain(context context.Context, params AddDNSDomainParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/add", &data, params)
	return &data.Response.Domain, err
}

// Delete - deletes a domain from the dns system
func (s *DNSDomainService) Delete(context context.Context, params DeleteDNSDomainParams) error {
	return s.client.post(context, "domain/delete", nil, params)
}

// Details - return details about the domain
func (s *DNSDomainService) Details(context context.Context, domainname string) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/details", &data, struct {
		Name string `json:"domainname"`
	}{domainname})
	return &data.Response.Domain, err
}

// Edit - edit domain parameters
func (s *DNSDomainService) Edit(context context.Context, params EditDNSDomainParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/edit", &data, params)
	return &data.Response.Domain, err
}

// Export - return the zonefile for the domain
func (s *DNSDomainService) Export(context context.Context, domainname string) (string, error) {
	data := struct {
		Response struct {
			Zonefile string
		}
	}{}
	err := s.client.post(context, "domain/export", &data, struct {
		Name string `json:"domainname"`
	}{domainname})
	return data.Response.Zonefile, err
}

// List - return a list of all domains in your account
func (s *DNSDomainService) List(context context.Context) (*[]DNSDomain, error) {
	data := struct {
		Response struct {
			Domains []DNSDomain
		}
	}{}
	err := s.client.get(context, "domain/list", &data)
	return &data.Response.Domains, err
}

// GenerateAuthCode - return a authcode for the domain
func (s *DNSDomainService) GenerateAuthCode(context context.Context, domainname string) (string, error) {
	data := struct {
		Response struct {
			Authcode string
		}
	}{}
	err := s.client.post(context, "domain/generateauthcode", &data, struct {
		Name string `json:"domainname"`
	}{domainname})
	return data.Response.Authcode, err
}

// Register - Register a domain
func (s *DNSDomainService) Register(context context.Context, params RegisterDNSDomainParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/register", &data, params)
	return &data.Response.Domain, err
}

// Renew - Renew a domain
func (s *DNSDomainService) Renew(context context.Context, params RenewDNSDomainParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/renew", &data, params)
	return &data.Response.Domain, err
}

// SetAutoRenew - Set a domain to renew automatically
func (s *DNSDomainService) SetAutoRenew(context context.Context, params SetAutoRenewParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/setautorenew", &data, params)
	return &data.Response.Domain, err
}

// Transfer - Transfer a domain
func (s *DNSDomainService) Transfer(context context.Context, params RegisterDNSDomainParams) (*DNSDomain, error) {
	data := struct {
		Response struct {
			Domain DNSDomain
		}
	}{}
	err := s.client.post(context, "domain/transfer", &data, params)
	return &data.Response.Domain, err
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

// ChangeNameservers - change the nameservers for domain
func (s *DNSDomainService) ChangeNameservers(context context.Context, params ChangeNameserverParams) error {
	return s.client.post(context, "domain/changenameservers", nil, params)
}
