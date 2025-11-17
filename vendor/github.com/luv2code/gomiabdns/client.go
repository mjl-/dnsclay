// Package gomiabdns provides a interface for the Mail-In-A-Box dns API
package gomiabdns

import (
	"context"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	
	"github.com/tidwall/gjson"
	"github.com/pquerna/otp/totp"
)

var apikey string

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
	email string
	password string
	totp_secret string
}

// New returns a new client ready to call the provided endpoint.
func New(apiUrl, email, password string, totp_secret string) *Client {
	parsedUrl, err := url.Parse(apiUrl)
	
	if err != nil {
		panic(err)
	}
	return &Client{
		ApiUrl: parsedUrl,
		email: email,
		password: password,
		totp_secret: totp_secret,
	}
}

// GetHosts returns all defined records if name and recordType are both empty string.
// If values are provided for both name and recordType, only the records that match both are returned.
// If one or the other of name and recordType are empty string, no records are returned.
func (c *Client) GetHosts(ctx context.Context, name string, recordType RecordType) ([]DNSRecord, error) {
	apiUrl := getApiWithPath(c.ApiUrl, name, recordType)
	apiResp, err := c.doRequest(ctx, http.MethodGet, apiUrl.String(), "")
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
	apiResp, err := c.doRequest(ctx, http.MethodPost, apiUrl.String(), value)
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
	apiResp, err := c.doRequest(ctx, http.MethodPut, apiUrl.String(), value)
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
	apiResp, err := c.doRequest(ctx, http.MethodDelete, apiUrl.String(), value)
	if err != nil {
		return err
	}
	fmt.Println(string(apiResp))
	return nil
}

// GetZones returns all zones that the MiaB box is responsible for.
func (c *Client) GetZones(ctx context.Context) ([]DNSZone, error) {
	apiUrl := c.ApiUrl.JoinPath("dns", "zones")
	
        //fmt.Println("apiUrl: " + apiUrl.String())
        apiResp, err := c.doRequest(ctx, http.MethodGet, apiUrl.String(), "")
        if err != nil {
                return nil, err
        }
        return unmarshalZones(apiResp)
}

// GetZonefile returns the zonefile for the indicate zone
func (c *Client) GetZonefile(ctx context.Context, zone string) (string, error) {
	apiUrl := c.ApiUrl.JoinPath("dns", "zonefile", zone)

        apiResp, err := c.doRequest(ctx, http.MethodGet, apiUrl.String(), "")
        if err != nil {
                return "", err
        }
        return string(apiResp), nil
}

func (c *Client) doLogin(ctx context.Context) (error) {
	requestURL := c.ApiUrl.JoinPath("login").String()
	var r io.Reader
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "JSON")
	if apikey != "" {
		// already logged in
		return nil
	}
	
	req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(c.email+":"+c.password)))

	// If totp secret is configured, use it to generate a totp token
	if c.totp_secret != "" {
		token, err := totp.GenerateCode(c.totp_secret, time.Now())
		if err != nil {
			err := fmt.Errorf("Error generating TOTP token: " + err.Error())
			return err
		}
		req.Header.Add("x-auth-token", token)
	}
	
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	if err != nil {
		return err
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	bodystr := string(body)
	status := gjson.Get(bodystr, "status").String()
	
	if status == "ok" {
		privileges := gjson.Get(bodystr, "privileges").String()
		if !strings.Contains(privileges, "admin") {
			err = fmt.Errorf("Account does not have admin priveleges")
			return err
		}
		apikey = gjson.Get(bodystr, "api_key").String()
	} else if status == "invalid" {
		apikey = ""
		reason := gjson.Get(bodystr, "reason").String()
		err = fmt.Errorf("Invalid response: " + reason)
		return err
	} else {
		apikey = ""
		err = fmt.Errorf("Unforeseen return value: " + status)
		return err
	}
	
	return nil
}
        
func (c *Client) doRequest(ctx context.Context, method, requestURL, value string) ([]byte, error) {
	var r io.Reader
	if value != "" {
		r = strings.NewReader(value)
	}

	err := c.doLogin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not login: " + err.Error())
	}
	
	if apikey == "" {
		return nil, fmt.Errorf("Could not login")
	}
	
	req, err := http.NewRequestWithContext(ctx, method, requestURL, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "json")
	req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(c.email + ":" + apikey)))
	
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	
	if err != nil {
		return nil, err
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return body, nil
}

func getApiWithPath(apiUrl *url.URL, name string, rtype RecordType) *url.URL {
	apiUrl = apiUrl.JoinPath("dns", "custom")
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
		var errorResult APIStatus
		if err2 := json.Unmarshal(data, &errorResult); err2 != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Error while decoding json: " + errorResult.Reason)
	}
	return result, nil
}

func unmarshalZones(data []byte) ([]DNSZone, error) {
        var result []DNSZone
        //fmt.Println(string(data))
        if err := json.Unmarshal(data, &result); err != nil {
        	var errorResult APIStatus
                if err2 := json.Unmarshal(data, &errorResult); err2 != nil {
                        return nil, err
                }
                return nil, fmt.Errorf("Error while decoding json: " + errorResult.Reason)
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

// DNSZone represents the zone data returned from the API
type DNSZone string

// Represents status returned from the API
type APIStatus struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
