// Package nfsn implements a DNS record management client compatible with the libdns interfaces for
// nearlyfreespeech.net (NFSN)
package nfsn

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
)

const apiBase = "https://api.nearlyfreespeech.net"
const authHeader = "X-NFSN-Authentication"

// NFSN enforces a minimum TTL of 3 minutes
const minimumTTL = 180 * time.Second

// Constants used for API salt generation
const saltChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijjklmnopqrstuvwxyz0123456789"
const saltLen = 16

// Provider facilitates DNS record manipulation with nearlyfreespeech.net
type Provider struct {
	// NFSN Member Login.
	Login string `json:"login,omitempty"`

	// NFSN API Key. API Keys can be generated from the "Profile" tab in the NFSN member interface.
	APIKey string `json:"api_key,omitempty"`

	client    *http.Client
	clientMtx sync.Mutex
}

type nfsnRecord struct {
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Data  string `json:"data,omitempty"`
	TTL   int    `json:"ttl,omitempty"`
	Scope string `json:"scope,omitempty"`
	Aux   int    `json:"aux,omitempty"`
}

// The pieces necessary to make a request to create/update a record in NFSN. Differs slightly from
// the fields in libdns.Record
type nfsnRecordParameters struct {
	Name string
	Type string
	Data string
	TTL  int
}

func (nRecord nfsnRecord) Record() (libdns.Record, error) {
	record := libdns.Record{
		Type:  nRecord.Type,
		Name:  nRecord.Name,
		Value: nRecord.Data,
		TTL:   time.Second * time.Duration(nRecord.TTL),
	}

	switch nRecord.Type {
	case "HTTPS":
	case "MX":
		record.Priority = uint(nRecord.Aux)
	case "SRV":
	case "URI":
		// Priority is in the 'aux' field from NFSN
		record.Priority = uint(nRecord.Aux)

		// Data is "weight port target", libdns expects weight in the record
		parts := strings.SplitN(nRecord.Data, " ", 2)

		if len(parts) != 3 {
			return libdns.Record{}, fmt.Errorf("%s record %s has incorrect format", nRecord.Name, nRecord.Data)
		}

		weight, err := strconv.Atoi(parts[0])

		if err != nil {
			return libdns.Record{}, err
		}

		record.Weight = uint(weight)
		record.Value = parts[1]
	}

	return record, nil
}

func toNfsnRecordParameters(record libdns.Record) url.Values {
	var dataBuilder strings.Builder

	switch record.Type {
	case "HTTPS":
	case "MX":
		dataBuilder.WriteString(fmt.Sprintf("%d ", record.Priority))
	case "SRV":
	case "URI":
		dataBuilder.WriteString(fmt.Sprintf("%d %d ", record.Priority, record.Weight))
	}

	dataBuilder.WriteString(record.Value)

	parameters := url.Values{}
	parameters.Set("name", record.Name)
	parameters.Set("type", record.Type)
	parameters.Set("data", dataBuilder.String())

	ttl := record.TTL

 	if ttl < minimumTTL {
		ttl = minimumTTL
	}

	parameters.Set("ttl", fmt.Sprintf("%d", int(ttl.Seconds())))

	return parameters
}

// Constructs a value to pass into an X-NFSN-Authentication header.
//
// The header value has the format [LOGIN];[TIMESTAMP];[SALT];[HASH]
//
// * LOGIN is the member's login name
// * TIMESTAMP is a 32 bit unsigned unix timestamp
// * SALT is a random, 16 character, alphanumeric string
// * HASH is sha1("[LOGIN];[TIMESTAMP];[SALT];[API_KEY];[REQUEST_URI];[BODY_HASH]")
//   - LOGIN, TIMESTAMP, SALT are the same as above
//   - API_KEY is the member's private API key
//   - REQUEST_URI is the PATH portion of the request URI
//   - BODY_HASH is the SHA1 hash of the request body (or of the empty string, if no request body is
//     present)
//
// Takes `timestamp` and `salt` values for testing.
func (p *Provider) innerGetAuthValue(req *http.Request, timestamp time.Time, salt string) (string, error) {
	var bodyBytes []byte
	var err error

	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)

		if err != nil {
			return "", err
		}
	}

	// Restore the body so it can be read again later
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyHash := sha1.Sum(bodyBytes)

	// Build the text to hash
	hText := fmt.Sprintf("%s;%d;%s;%s;%s;%x", p.Login, timestamp.Unix(), salt, p.APIKey, req.URL.Path, bodyHash)
	hHash := sha1.Sum([]byte(hText))

	// Format the auth value to send on the wire
	authVal := fmt.Sprintf("%s;%d;%s;%x", p.Login, timestamp.Unix(), salt, hHash)
	return authVal, nil
}

// Generate a random salt usable for generating an X-NFSN-Authentication header value. See
// `innerGetAuthValue` for details.
func genSalt() (string, error) {
	bytes := make([]byte, saltLen)
	readLen, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	if readLen != saltLen {
		return "", fmt.Errorf("Failed to read enough random bytes")
	}

	var sb strings.Builder

	for b := range bytes {
		sb.WriteByte(saltChars[b%len(saltChars)])
	}

	return sb.String(), nil
}

func uriForZone(zone string, resource string) string {
	return fmt.Sprintf("%s/dns/%s/%s", apiBase, strings.TrimRight(zone, "."), resource)
}

// See `innerGetAuthValue` for details.
func (p *Provider) getAuthValue(req *http.Request) (string, error) {
	salt, err := genSalt()

	if err != nil {
		return "", err
	}

	return p.innerGetAuthValue(req, time.Now(), salt)
}

func (p *Provider) ensureClient() {
	if p.client == nil {
		p.clientMtx.Lock()
		defer p.clientMtx.Unlock()

		if p.client == nil {
			p.client = &http.Client{}
		}
	}
}

// Makes a request with the given parameters (see `http.NewRequestWithContext`), adding necessary
// auth information before executing it.
func (p *Provider) makeRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Response, error) {
	p.ensureClient()
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	authValue, err := p.getAuthValue(req)

	if err != nil {
		return nil, err
	}

	req.Header.Add(authHeader, authValue)

	resp, err := p.client.Do(req)

	if err != nil {
		return nil, err
	}

	var bodyBytes []byte

	if resp.Body != nil {
		bodyBytes, _ = io.ReadAll(resp.Body)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned non-success status code %s with response body %s. Original error: %w", resp.Status, string(bodyBytes), err)
	}

	return resp, err
}

// Execute the given `verb` for each record in `records`. Accumulate successfully process records
// and return them at the end. If only some records are processed, returns those that were
// successfull _and_ an error.
func (p *Provider) processRecords(ctx context.Context, zone string, verb string, records []libdns.Record) ([]libdns.Record, error) {
	uri := uriForZone(zone, verb)
	var successfulRecords []libdns.Record

	for _, record := range records {
		params := toNfsnRecordParameters(record)
		_, err := p.makeRequest(ctx, "POST", uri, strings.NewReader(params.Encode()))

		if err != nil {
			return successfulRecords, err
		}

		successfulRecords = append(successfulRecords, record)
	}

	return successfulRecords, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	resp, err := p.makeRequest(ctx, "POST", uriForZone(zone, "listRRs"), nil)

	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var nRecords []nfsnRecord
	err = json.Unmarshal(bodyBytes, &nRecords)

	if err != nil {
		return nil, err
	}

	records := make([]libdns.Record, 0, len(nRecords))

	for _, nRecord := range nRecords {
		record, err := nRecord.Record()

		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added. In the case where
// only some records succeed returns both the records that were added and an error.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.processRecords(ctx, zone, "addRR", records)
}

// SetRecords sets the records in the zone, either by updating existing records or creating new
// ones. It returns the updated records. In the case where only some records succeed returns both
// the records that were replaced and an error.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.processRecords(ctx, zone, "replaceRR", records)
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted. In the
// case where only some records succeed returns both the records that were deleted and an error.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	return p.processRecords(ctx, zone, "removeRR", records)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
