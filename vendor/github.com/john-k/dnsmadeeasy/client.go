package dnsmadeeasy

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	DNSManagedPath string = "/dns/managed/"
	DNSRecordsPath string = "{domainId}/records"
	DNSRecordPath  string = "{domainId}/records/{recordId}"
)

type BaseURL string

const (
	Sandbox BaseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0/"
	Prod    BaseURL = "https://api.dnsmadeeasy.com/V2.0/"
)

type Client struct {
	APIKey      string
	SecretKey   string
	BaseURL     BaseURL
	resty       *resty.Client
	zoneIdCache map[string]int
}

// Construct a client using the supplied values
func GetClient(APIKey string, SecretKey string, url BaseURL) *Client {
	r := resty.New().SetBaseURL(string(url))
	return &Client{APIKey, SecretKey, url, r, nil}
}

// Convenience function to determine the error status of a response
// from DNS Made Easy
func checkRespForError(resp *resty.Response, err error) (*resty.Response, error) {
	// first bubble up any error passed to us
	if err != nil {
		return resp, err
	}

	var data map[string]interface{}

	// next check for json-formatted errors in the response body
	err = json.Unmarshal(resp.Body(), &data)
	// no error indicates that we were able to de-serialize some json
	if err == nil {
		if data["error"] != nil {
			// translate the array of strings that is DME's error json element
			// ie { "error": [ "", "" ] }
			resp_errors := data["error"].([]interface{})
			if len(resp_errors) > 0 {
				var error string
				if len(resp_errors) == 1 {
					error = resp_errors[0].(string)
				} else {
					for idx, err := range resp_errors {
						error += fmt.Sprintf("%d: %s\n", idx, err.(string))
					}
				}
				return resp, errors.New(error)
			}
		}
	}

	// lastly, check for an HTTP error code
	status := resp.StatusCode()
	if status < 200 || status >= 300 {
		return resp, fmt.Errorf("request returned http error code %d", status)
	}

	// if we got here, there are no errors
	return resp, nil
}

// Convenience function to calculate the authentication headers
// expected by DNS Made Easy
func (c *Client) addAuthHeaders(req *resty.Request) {
	requestDate := time.Now().UTC().Format(http.TimeFormat)

	// Calculate the hexadecimal HMAC SHA1 of requestDate using APIKey
	key := []byte(c.SecretKey)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(requestDate))
	hmacString := hex.EncodeToString(h.Sum(nil))

	req.Header.Add("X-Dnsme-Apikey", c.APIKey)
	req.Header.Add("X-Dnsme-Requestdate", requestDate)
	req.Header.Add("X-Dnsme-Hmac", hmacString)
}

// Convenience function to construct a request with common headers
func (c *Client) newRequest() *resty.Request {
	req := c.resty.R().ExpectContentType("application/json").
		SetHeader("Content-Type", "application/json")
	c.addAuthHeaders(req)
	return req
}

type Domain struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	CreatedAt          int      `json:"created"`
	UpdatedAt          int      `json:"updated"`
	FolderID           int      `json:"folderId"`
	ProcessMulti       bool     `json:"processMulti"`
	ActiveThirdParties []string `json:"activeThirdParties"`
	GtdEnabled         bool     `json:"gtdEnabled"`

	// Identifies which action is currently pending
	// Values:
	//  0 - nothing pending
	//  1 - pending creation (usable but not fully created throughout the system)
	//  3 - pending deletion
	PendingActionID int `json:"pendingActionId"`
}

type DomainsResp struct {
	TotalRecords int      `json:"totalRecords"`
	TotalPages   int      `json:"totalPages"`
	Domains      []Domain `json:"data"`
	CurrentPage  int      `json:"page"`
}

// Creates a new domain
func (c *Client) CreateDomain(domainName string) (Domain, error) {
	var newDomain Domain

	createDomainBody := fmt.Sprintf(`{"name":"%s"}`, domainName)
	req := c.newRequest().
		SetResult(&newDomain).
		SetBody(createDomainBody)

	_, err := checkRespForError(req.Post(DNSManagedPath))
	if err != nil {
		return Domain{}, err
	}

	return newDomain, nil
}

// Removes a domain and all associated records
func (c *Client) DeleteDomain(domainID int) error {
	_, err := checkRespForError(c.newRequest().
		Delete(fmt.Sprint(DNSManagedPath, domainID)))
	return err
}

// Returns the domain record for a given domain ID
func (c *Client) GetDomain(domainID int) (Domain, error) {
	var domain Domain
	_, err := checkRespForError(c.newRequest().
		SetResult(&domain).
		Get(DNSManagedPath + fmt.Sprint(domainID)))
	if err != nil {
		return Domain{}, err
	}
	return domain, nil
}

// Returns a map of Name:ID for all domains managed by the
// given account
func (c *Client) EnumerateDomains() (map[string]int, error) {
	domains := map[string]int{}

	var respDomains DomainsResp
	_, err := checkRespForError(c.newRequest().
		SetResult(&respDomains).
		Get(DNSManagedPath))
	if err != nil {
		return nil, err
	}

	for _, domain := range respDomains.Domains {
		domains[domain.Name] = domain.ID
	}

	return domains, nil
}

// Finds the numerical ID for a given domain name
func (c *Client) IdForDomain(domain string) (int, error) {
	justPopulated := false
	if c.zoneIdCache == nil {
		domainMap, err := c.EnumerateDomains()
		if err != nil {
			return 0, err
		}
		c.zoneIdCache = domainMap
		justPopulated = true
	}

	zoneId, ok := c.zoneIdCache[domain]
	if ok {
		return zoneId, nil
	} else {
		// if we didn't just populate the cache, refresh it in case
		// our domain exists now
		if !justPopulated {
			domainMap, err := c.EnumerateDomains()
			if err != nil {
				return 0, err
			}
			c.zoneIdCache = domainMap
			justPopulated = true
		}
		zoneId, ok := c.zoneIdCache[domain]
		if ok {
			return zoneId, nil
		}
	}

	return 0, errors.New("Domain not found")
}

type Record struct {
	// A unique name per record Type
	Name string `json:"name"`

	// A unique identifier for this record
	ID int `json:"id,omitempty"`

	// Can be one of: A, AAAA, ANAME, CNAME, HTTPRED, MX
	//                NS, PTR, SRV, TXT, SPF, or SOA
	Type string `json:"type"`

	// Differs per record type
	Value string `json:"value"`

	// 1 if the record is the record is domain specific
	// 0 if the record is part of a template
	Source int `json:"source,omitempty"`

	// The time to live of the record
	Ttl int `json:"ttl"`

	// Global Traffic Director location.
	// Values: DEFAULT, US_EAST, US_WEST, EUROPE,
	//         ASIA_PAC, OCREANIA, SOUTH_AMERICA
	GtdLocation string `json:"gtdLocation"`

	// The domain ID of this record
	SourceId int `json:"sourceId,omitempty"`

	// Indicates if DNS Failover is enabled for an A record
	Failover bool `json:"failover,omitempty"`

	// Indicates if System Monitoring is enabled for an A record
	Monitor bool `json:"monitor,omitempty"`

	// For HTTP Redirection Records
	HardLink bool `json:"hardLink,omitempty"`

	// Indicates if the record has dynamic DNS enabled
	DynamicDns bool `json:"dynamicDns,omitempty"`

	// Indicates if an A record is in failed status
	Failed bool `json:"failed,omitempty"`

	// The priority for an MX record
	MxLevel int `json:"mxLevel,omitempty"`

	// The priority for an SRV record
	Priority int `json:"priority,omitempty"`

	// The weight for an SRV record
	Weight int `json:"weight,omitempty"`

	// The port for an SRV record
	Port int `json:"port,omitempty"`
}

type RecordsResp struct {
	TotalRecords int      `json:"totalRecords"`
	TotalPages   int      `json:"totalPages"`
	Records      []Record `json:"data"`
	CurrentPage  int      `json:"page"`
}

func (c *Client) EnumerateRecords(domainId int) ([]Record, error) {
	var respRecords RecordsResp
	req := c.newRequest().
		SetResult(&respRecords).
		SetPathParam("domainId", fmt.Sprint(domainId))

	_, err := checkRespForError(req.Get(DNSManagedPath + DNSRecordsPath))
	if err != nil {
		return nil, err
	}

	return respRecords.Records, nil
}

// Deletes all records for the supplied domain
func (c *Client) DeleteAllRecords(domainID int) error {
	_, err := checkRespForError(c.newRequest().
		SetPathParam("domainId", fmt.Sprint(domainID)).
		Delete(DNSManagedPath + DNSRecordsPath))
	return err
}

// Deletes records with numerical IDs for the supplied domain
//
// NOTE: will silently continue if a recordId that doesn't belong to the
// given domainId is passed
func (c *Client) DeleteRecords(domainId int, recordIds []int) ([]int, error) {
	var queryString string

	// build query string of ids=X&ids=Y&ids=Z
	// we can't use other convenience methods since they use
	// map[string] and only the last id would made it
	for idx, id := range recordIds {
		if idx > 0 {
			queryString += "&"
		}
		queryString += fmt.Sprintf("ids=%d", id)
	}

	req := c.newRequest().
		SetPathParam("domainId", fmt.Sprint(domainId)).
		SetPathParam("recordId", "").
		SetQueryString(queryString)

	_, err := checkRespForError(req.Delete(DNSManagedPath + DNSRecordPath))
	if err != nil {
		return nil, err
	}
	return recordIds, nil
}

// Creates a single record in the supplied domain
func (c *Client) CreateRecord(domainId int, record Record) (Record, error) {
	var newRecord Record

	req := c.newRequest().
		SetResult(&newRecord).
		SetBody(&record).
		SetPathParam("domainId", fmt.Sprint(domainId))

	_, err := checkRespForError(req.Post(DNSManagedPath + DNSRecordsPath))
	if err != nil {
		return Record{}, err
	}

	return newRecord, nil
}

// Create many records at once in the supplied domain
//
// NOTE: is transactional; an error in creating any record causes none to be created
func (c *Client) CreateRecords(domainId int, records []Record) ([]Record, error) {
	var newRecords []Record

	req := c.newRequest().
		SetResult(&newRecords).
		SetBody(&records).
		SetPathParam("domainId", fmt.Sprint(domainId))

	_, err := checkRespForError(req.Post(DNSManagedPath + DNSRecordsPath + "/createMulti"))
	if err != nil {
		return []Record{}, err
	}

	return newRecords, nil
}

func (c *Client) UpdateRecords(domainId int, records []Record) ([]Record, error) {
	var updatedRecords []Record

	req := c.newRequest().
		SetResult(&updatedRecords).
		SetBody(&records).
		SetPathParam("domainId", fmt.Sprint(domainId))

	_, err := checkRespForError(req.Post(DNSManagedPath + DNSRecordsPath + "/updateMulti"))
	if err != nil {
		return []Record{}, err
	}

	return updatedRecords, nil
}
