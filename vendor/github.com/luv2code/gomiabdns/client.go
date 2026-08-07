// Package gomiabdns provides a interface for the Mail-In-A-Box dns API
package gomiabdns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RecordType is the type of DNS Record. For ex. CNAME.
type RecordType string

const (
	// A record type.
	A RecordType = "A"
	// AAAA record type.
	AAAA RecordType = "AAAA"
	// CAA record type.
	CAA RecordType = "CAA"
	// CNAME record type.
	CNAME RecordType = "CNAME"
	// MX record type.
	MX RecordType = "MX"
	// NS record type.
	NS RecordType = "NS"
	// TXT record type.
	TXT RecordType = "TXT"
	// SRV record type.
	SRV RecordType = "SRV"
	// SSHFP record type.
	SSHFP RecordType = "SSHFP"
)

// Client provides a target for methods interacting with the DNS API.
type Client struct {
	ApiUrl *url.URL
}

// New returns a new client ready to call the provided endpoint.
func New(apiUrl, email, password string) *Client {
	parsedUrl, err := url.Parse(apiUrl)
	parsedUrl.User = url.UserPassword(email, password)
	if err != nil {
		panic(err)
	}
	return &Client{
		ApiUrl: parsedUrl,
	}
}

// GetHosts returns all defined records if name and recordType are both empty string.
// If values are provided for both name and recordType, only the records that match both are returned.
// If one or the other of name and recordType are empty string, no records are returned.
func (c *Client) GetHosts(ctx context.Context, name string, recordType RecordType) ([]DNSRecord, error) {
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodGet, apiUrl.String(), "")
	if err != nil {
		return nil, err
	}
	return unmarshalRecords(apiResp)
}

// AddHost adds a record. name, recordType, and value are all required. If a record exists with the same value,
// no new record is created. Use this method for creating multple A records for dns loadbalancing. Or use it
// to create multiple different TXT records.
func (c *Client) AddHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" || recordType == "" || value == "" {
		return fmt.Errorf(
			"Missing parameters to AddHost. all are required. name: %s, recordType: %s, value: %s ",
			name,
			recordType,
			value,
		)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodPost, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

// UpdateHost will create or update a record that corresponds with the name and recordType.
// If multiple records with the same name and type exists, they will all be removed and replaced
// with a single one that matches the parameters passed to this method. name, recordType, and value
// are all required.
func (c *Client) UpdateHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" || recordType == "" || value == "" {
		return fmt.Errorf(
			"Missing parameters to UpdateHost. all are required. name: %s, recordType: %s, value: %s ",
			name,
			recordType,
			value,
		)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodPut, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

// DeleteHost will delete records that match the passed paramters.
func (c *Client) DeleteHost(ctx context.Context, name string, recordType RecordType, value string) error {
	if name == "" {
		return fmt.Errorf("Missing parameter to DeleteHost. Name is required. name: %s", name)
	}
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := doRequest(ctx, http.MethodDelete, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

func doRequest(ctx context.Context, method, requestURL, value string) ([]byte, error) {
	var r io.Reader
	if value != "" {
		r = strings.NewReader(value)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, r)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = resp.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

func getApiWithPath(apiUrl *url.URL, name string, rtype RecordType) *url.URL {
	if name != "" {
		if rtype != "" {
			return apiUrl.JoinPath(name, string(rtype))
		} else {
			return apiUrl.JoinPath(name)
		}
	}
	return apiUrl
}

func unmarshalRecords(data []byte) ([]DNSRecord, error) {
	var result []DNSRecord
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DNSRecord represents the host data returned from the API
type DNSRecord struct {
	QualifiedName string     `json:"qname"`
	RecordType    RecordType `json:"rtype"`
	SortOrder     struct {
		ByCreated int `json:"created"`
		ByName    int `json:"qname"`
	} `json:"sort-order"`
	Value string `json:"value"`
	Zone  string `json:"zone"`
}
