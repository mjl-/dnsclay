package dnssdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	defaultBaseURL = "https://api.gcore.com/dns"
	tokenHeader    = "APIKey"
	defaultTimeOut = 10 * time.Second
)

// Client for DNS API.
type Client struct {
	HTTPClient *http.Client
	UserAgent  string
	BaseURL    *url.URL
	authHeader func() string
	Debug      bool
}

// ZonesFilter find zones
type ZonesFilter struct {
	Names []string
}

type authHeader string

// BearerAuth by header
func BearerAuth(token string) func() authHeader {
	return func() authHeader {
		return authHeader(fmt.Sprintf("Bearer %s", token))
	}
}

// PermanentAPIKeyAuth by header
func PermanentAPIKeyAuth(token string) func() authHeader {
	return func() authHeader {
		return authHeader(fmt.Sprintf("%s %s", tokenHeader, token))
	}
}

func (zf ZonesFilter) query() string {
	if len(zf.Names) == 0 {
		return ""
	}
	return url.Values{"name": zf.Names}.Encode()
}

// NewClient constructor of Client.
func NewClient(authorizer func() authHeader, opts ...func(*Client)) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	cl := &Client{
		authHeader: func() string { return string(authorizer()) },
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: defaultTimeOut},
	}
	for _, op := range opts {
		op(cl)
	}
	return cl
}

// CreateZone adds new zone.
// https://apidocs.gcore.com/dns#tag/zones/operation/CreateZone
func (c *Client) CreateZone(ctx context.Context, name string) (uint64, error) {
	res := CreateResponse{}
	params := AddZone{Name: name}
	err := c.do(ctx, http.MethodPost, "/v2/zones", params, &res)
	if err != nil {
		return 0, fmt.Errorf("request: %w", err)
	}
	if res.Error != "" {
		return 0, APIError{StatusCode: http.StatusOK, Message: res.Error}
	}

	return res.ID, nil
}

// Zones gets first 100 zones.
// https://apidocs.gcore.com/dns#tag/zones/operation/Zones
func (c *Client) Zones(ctx context.Context, filters ...func(zone *ZonesFilter)) ([]Zone, error) {
	res := ListZones{}
	filter := ZonesFilter{}
	for _, op := range filters {
		op(&filter)
	}
	err := c.do(ctx, http.MethodGet, "/v2/zones?limit=100&"+filter.query(), nil, &res)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	return res.Zones, nil
}

/*
ZonesParam parameter for ZonesWithParam method

offset
integer <uint64>
Amount of records to skip before beginning to write in response.

limit
integer <uint64>
Max number of records in response

order_by
string
Field name to sort by

order_direction
string
Enum: "asc" "desc"

Ascending or descending order
id
Array of integers <int64> [ items <int64 > ]

to pass several ids id=1&id=3&id=5...
client_id
Array of integers <int64> [ items <int64 > ]

to pass several client_ids client_id=1&client_id=3&client_id=5...
reseller_id
Array of integers <int64> [ items <int64 > ]

iam_reseller_id
Array of integers <int64> [ items <int64 > ]

name
Array of strings
to pass several names name=first&name=second...

case_sensitive
boolean

exact_match
boolean

enabled
boolean

status
string

dynamic
boolean
Zones with dynamic RRsets

healthcheck
boolean
Zones with RRsets that have healthchecks

updated_at_from
string <date-time>

updated_at_to
string <date-time>
*/
type ZonesParam struct {
	Offset         uint64
	Limit          uint64
	OrderBy        string
	OrderDirection string
	ID             []uint64
	ClientID       []uint64
	ResellerID     []uint64
	IAMResellerID  []uint64
	Name           []string
	CaseSensitive  bool
	ExactMatch     bool
	Enabled        bool
	Status         string
	Dynamic        bool
	Healthcheck    bool
	UpdatedAtFrom  time.Time
	UpdatedAtTo    time.Time
}

func (zp ZonesParam) query() string {
	form := url.Values{}
	if zp.Offset > 0 {
		form.Add("offset", fmt.Sprint(zp.Offset))
	}
	if zp.Limit > 0 {
		form.Add("limit", fmt.Sprint(zp.Limit))
	}
	if zp.OrderBy != "" {
		form.Add("order_by", zp.OrderBy)
	}
	if zp.OrderDirection != "" {
		form.Add("order_direction", zp.OrderDirection)
	}
	if len(zp.ID) > 0 {
		for _, id := range zp.ID {
			form.Add("id", fmt.Sprint(id))
		}
	}
	if len(zp.ClientID) > 0 {
		for _, id := range zp.ClientID {
			form.Add("client_id", fmt.Sprint(id))
		}
	}
	if len(zp.ResellerID) > 0 {
		for _, id := range zp.ResellerID {
			form.Add("reseller_id", fmt.Sprint(id))
		}
	}
	if len(zp.IAMResellerID) > 0 {
		for _, id := range zp.IAMResellerID {
			form.Add("iam_reseller_id", fmt.Sprint(id))
		}
	}
	if len(zp.Name) > 0 {
		for _, name := range zp.Name {
			form.Add("name", name)
		}
	}
	if zp.CaseSensitive {
		form.Add("case_sensitive", "true")
	}
	if zp.ExactMatch {
		form.Add("exact_match", "true")
	}
	if zp.Enabled {
		form.Add("enabled", "true")
	}
	if zp.Status != "" {
		form.Add("status", zp.Status)
	}
	if zp.Dynamic {
		form.Add("dynamic", "true")
	}
	if zp.Healthcheck {
		form.Add("healthcheck", "true")
	}
	if !zp.UpdatedAtFrom.IsZero() {
		form.Add("updated_at_from", zp.UpdatedAtFrom.Format(time.RFC3339))
	}
	if !zp.UpdatedAtTo.IsZero() {
		form.Add("updated_at_to", zp.UpdatedAtTo.Format(time.RFC3339))
	}
	return form.Encode()
}

// ZonesWithParam gets zones with params.
func (c *Client) ZonesWithParam(ctx context.Context, param ZonesParam) (res ListZones, err error) {
	err = c.do(ctx, http.MethodGet, "/v2/zones?"+param.query(), nil, &res)
	if err != nil {
		return res, fmt.Errorf("request: %w", err)
	}

	return res, nil
}

// AllZones get all zones per 1k
func (c *Client) AllZones(ctx context.Context, nameFilters []string) ([]Zone, error) {
	offset := 0
	const limit = 1000
	var zones []Zone
	for z := 0; z < 10; z++ {
		param := ZonesParam{
			Offset: uint64(offset),
			Limit:  uint64(limit),
			Name:   nameFilters,
		}
		zoneRes, err := c.ZonesWithParam(ctx, param)
		if err != nil {
			return zones, err
		}
		zones = append(zones, zoneRes.Zones...)
		if zoneRes.Error != `` {
			return zones, fmt.Errorf("request: %s", zoneRes.Error)
		}
		fetchedZones := len(zoneRes.Zones)
		if fetchedZones == 0 || fetchedZones < limit {
			break
		}
		offset += limit
	}
	return zones, nil
}

// ZonesWithRecords gets first 100 zones with records information.
func (c *Client) ZonesWithRecords(ctx context.Context, filters ...func(zone *ZonesFilter)) ([]Zone, error) {
	zones, err := c.Zones(ctx, filters...)
	if err != nil {
		return nil, fmt.Errorf("all zones: %w", err)
	}
	gr, _ := errgroup.WithContext(ctx)
	for i, z := range zones {
		z := z
		i := i
		gr.Go(func() error {
			zone, errGet := c.Zone(ctx, z.Name)
			if errGet != nil {
				return fmt.Errorf("%s: %w", z.Name, errGet)
			}
			zones[i] = zone
			return nil
		})
	}
	err = gr.Wait()
	if err != nil {
		return nil, fmt.Errorf("zone info: %w", err)
	}

	return zones, nil
}

// AllZonesWithRecords gets all zones with records information.
func (c *Client) AllZonesWithRecords(ctx context.Context, nameFilters []string) ([]Zone, error) {
	zones, err := c.AllZones(ctx, nameFilters)
	if err != nil {
		return nil, fmt.Errorf("all zones: %w", err)
	}
	gr, _ := errgroup.WithContext(ctx)
	for i, z := range zones {
		z := z
		i := i
		gr.Go(func() error {
			zone, errGet := c.Zone(ctx, z.Name)
			if errGet != nil {
				return fmt.Errorf("%s: %w", z.Name, errGet)
			}
			zones[i] = zone
			return nil
		})
	}
	err = gr.Wait()
	if err != nil {
		return nil, fmt.Errorf("zone info: %w", err)
	}

	return zones, nil
}

// DeleteZone gets zone information.
// https://apidocs.gcore.com/dns#tag/zones/operation/DeleteZone
func (c *Client) DeleteZone(ctx context.Context, name string) error {
	name = strings.Trim(name, ".")
	uri := path.Join("/v2/zones", name)

	err := c.do(ctx, http.MethodDelete, uri, nil, nil)
	if err != nil {
		return fmt.Errorf("request %s: %w", name, err)
	}

	return nil
}

// Zone gets zone information.
// https://apidocs.gcore.com/dns#tag/zones/operation/Zone
func (c *Client) Zone(ctx context.Context, name string) (Zone, error) {
	name = strings.Trim(name, ".")
	zone := Zone{}
	uri := path.Join("/v2/zones", name)

	err := c.do(ctx, http.MethodGet, uri, nil, &zone)
	if err != nil {
		return Zone{}, fmt.Errorf("get zone %s: %w", name, err)
	}

	return zone, nil
}

const nsRecordType = "NS"

// ZoneNameservers gets zone nameservers.
func (c *Client) ZoneNameservers(ctx context.Context, name string) ([]string, error) {
	name = strings.Trim(name, ".")
	uri := fmt.Sprintf("/v2/zones/%s/rrsets?all=true&type=%s", name, nsRecordType)

	var rrsets RRSets
	err := c.do(ctx, http.MethodGet, uri, nil, &rrsets)
	if err != nil {
		return nil, fmt.Errorf("get rrsets %s: %w", name, err)
	}

	resp := make([]string, 0)
	exists := make(map[string]struct{})

	for _, rrset := range rrsets.RRSets {
		for _, record := range rrset.Records {
			for _, content := range record.Content {
				contentStr := fmt.Sprint(content)
				if _, ok := exists[contentStr]; ok {
					continue
				}

				exists[contentStr] = struct{}{}
				resp = append(resp, contentStr)
			}
		}
	}

	return resp, nil
}

// RRSet gets RRSet item.
// https://apidocs.gcore.com/dns#tag/rrsets/operation/RRSet
func (c *Client) RRSet(ctx context.Context, zone, name, recordType string) (RRSet, error) {
	zone, name = strings.Trim(zone, "."), strings.Trim(name, ".")
	var result RRSet
	uri := path.Join("/v2/zones", zone, name, recordType)

	err := c.do(ctx, http.MethodGet, uri, nil, &result)
	if err != nil {
		return RRSet{}, fmt.Errorf("request %s -> %s: %w", zone, name, err)
	}

	return result, nil
}

// DeleteRRSet removes RRSet type records.
// https://apidocs.gcore.com/dns#tag/rrsets/operation/DeleteRRSet
func (c *Client) DeleteRRSet(ctx context.Context, zone, name, recordType string) error {
	zone, name = strings.Trim(zone, "."), strings.Trim(name, ".")
	uri := path.Join("/v2/zones", zone, name, recordType)

	err := c.do(ctx, http.MethodDelete, uri, nil, nil)
	if err != nil {
		// Support DELETE idempotence https://developer.mozilla.org/en-US/docs/Glossary/Idempotent
		statusErr := new(APIError)
		if errors.As(err, statusErr) && statusErr.StatusCode == http.StatusNotFound {
			return nil
		}

		return fmt.Errorf("delete record request: %w", err)
	}

	return nil
}

// DeleteRRSetRecord removes RRSet record.
func (c *Client) DeleteRRSetRecord(ctx context.Context, zone, name, recordType string, contents ...string) error {
	// get current records info
	rrSet, err := c.RRSet(ctx, zone, name, recordType)
	if err != nil {
		errAPI := new(APIError)
		if errors.As(err, errAPI) && errAPI.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("rrset: %w", err)
	}
	if len(rrSet.Records) == 0 {
		return nil
	}
	// setup new records
	newRecords := make([]ResourceRecord, 0, len(rrSet.Records))
LOOP:
	for _, record := range rrSet.Records {
		if len(record.Content) == 0 {
			continue
		}
		for _, toDelete := range contents {
			if toDelete == record.ContentToString() {
				continue LOOP
			}
		}
		newRecords = append(newRecords, record)
	}
	rrSet.Records = newRecords
	// delete on empty content
	if len(rrSet.Records) == 0 {
		err = c.DeleteRRSet(ctx, zone, name, recordType)
		if err != nil {
			err = fmt.Errorf("delete rrset: %w", err)
		}
		return err
	}
	// update with removing deleted content
	err = c.UpdateRRSet(ctx, zone, name, recordType, rrSet)
	if err != nil {
		err = fmt.Errorf("update rrset: %w", err)
	}
	return err
}

// AddZoneOpt setup RRSet
type AddZoneOpt func(*RRSet)

// WithFilters add filters to RRSet
func WithFilters(filters ...RecordFilter) AddZoneOpt {
	return func(set *RRSet) {
		set.AddFilter(filters...)
	}
}

// AddZoneRRSet create or extend resource record.
func (c *Client) AddZoneRRSet(ctx context.Context,
	zone, recordName, recordType string,
	values []ResourceRecord, ttl int, opts ...AddZoneOpt) error {

	record := RRSet{TTL: ttl, Records: values}
	for _, op := range opts {
		op(&record)
	}

	records, err := c.RRSet(ctx, zone, recordName, recordType)
	if err == nil && len(records.Records) > 0 {
		record.Records = append(record.Records, records.Records...)
		return c.UpdateRRSet(ctx, zone, recordName, recordType, record)
	}

	return c.CreateRRSet(ctx, zone, recordName, recordType, record)
}

// CreateRRSet https://apidocs.gcore.com/dns#tag/rrsets/operation/CreateRRSet
func (c *Client) CreateRRSet(ctx context.Context, zone, name, recordType string, record RRSet) error {
	zone, name = strings.Trim(zone, "."), strings.Trim(name, ".")
	uri := path.Join("/v2/zones", zone, name, recordType)

	return c.do(ctx, http.MethodPost, uri, record, nil)
}

// UpdateRRSet https://apidocs.gcore.com/dns#tag/rrsets/operation/UpdateRRSet
func (c *Client) UpdateRRSet(ctx context.Context, zone, name, recordType string, record RRSet) error {
	zone, name = strings.Trim(zone, "."), strings.Trim(name, ".")
	uri := path.Join("/v2/zones", zone, name, recordType)

	return c.do(ctx, http.MethodPut, uri, record, nil)
}

// DNSSecDS https://api.gcore.com/docs/dns#tag/DNSSEC/operation/GetDNSSECDS
func (c *Client) DNSSecDS(ctx context.Context, zone string) (DNSSecDS, error) {
	zone = strings.Trim(zone, ".")
	uri := path.Join("/v2/zones", zone, "dnssec")

	var dnsSecDS DNSSecDS
	err := c.do(ctx, http.MethodGet, uri, nil, &dnsSecDS)
	if err != nil {
		return DNSSecDS{}, fmt.Errorf("get dnssec: %w", err)
	}

	return dnsSecDS, nil
}

// ToggleDnssec https://api.gcore.com/docs/dns#tag/DNSSEC/operation/ToggleDNSSEC
func (c *Client) ToggleDnssec(ctx context.Context, zone string, enable bool) (DNSSecDS, error) {
	zone = strings.Trim(zone, ".")
	uri := path.Join("/v2/zones", zone, "dnssec")

	var dnssecDS DNSSecDS
	err := c.do(ctx, http.MethodPatch, uri, map[string]bool{"enabled": enable}, &dnssecDS)
	if err != nil {
		return DNSSecDS{}, fmt.Errorf("toggle dnssec: %w", err)
	}

	return dnssecDS, nil
}

func (c *Client) do(ctx context.Context, method, uri string, bodyParams interface{}, dest interface{}) error {
	var bs []byte
	if bodyParams != nil {
		var err error
		bs, err = json.Marshal(bodyParams)
		if err != nil {
			return fmt.Errorf("encode bodyParams: %w", err)
		}
	}

	endpoint, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, uri))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if c.Debug {
		log.Printf("[DEBUG] dns api request: %s %s %s \n", method, uri, bs)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(string(bs)))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader())
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusMultipleChoices {
		all, _ := ioutil.ReadAll(resp.Body)
		e := APIError{
			StatusCode: resp.StatusCode,
		}
		err := json.Unmarshal(all, &e)
		if err != nil {
			e.Message = string(all)
		}
		return e
	}

	// try read all so we can put breakpoint here
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if dest == nil {
		return nil
	}

	// nolint: wrapcheck
	return json.NewDecoder(bytes.NewReader(body)).Decode(dest)
}
