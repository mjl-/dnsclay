// Package desec implements a DNS record management client compatible with the libdns interfaces for
// [deSEC].
//
// # Updates are not atomic
//
// The deSEC API doesn't map 1:1 to the libdns API. The main issue is that libdns works on the
// record level while the deSEC API works on the RRset level. This API impedence mismatch makes
// it impossible to update individual records atomically using the libdns API. The implementation
// here goes to great lengths to avoid interference of multiple concurrent requests, but that
// only works within a single process.
//
// If multiple processes are modifying a deSEC zone concurrently, care must be taken that the
// different processes operate on different [resource record sets]. Otherwise multiple concurrent
// operations will override one another. The easiest way to protect against that is to use different
// names within the zone for different processes.
//
// If multiple processes operate on the same resource record set, it's possible for two concurrently
// running writes to result in inconsistent records.
//
// # TTL attribute
//
// For a the same reason as above, the  TTL attribute cannot be set on the per record level. If
// multiple different TTLs are specified for different records of the same name and type, one of
// them wins. It's not defined which on that is.
//
// # Large zones (> 500 resource record sets)
//
// deSEC requires the use of pagination for zones with more than 500 RRSets. This is a reasonable
// limit for a general purpose library like libdns and no effort is made to handle zones with more
// than 500 RRSets. Methods that can fail with more than 500 RRSets have a godoc comment explaining
// this.
//
// # Rate Limiting
//
// deSEC applies [rate limiting], this implementation will retry when running into a rate limit
// while observing context cancellation. In practice this means that calls to methods of this
// provider can take multiple seconds and longer. It's therefore very important to set a deadline in
// the context.
//
// [deSEC]: http://desec.io
// [resource record sets]: https://desec.readthedocs.io/en/latest/dns/rrsets.html
// [rate limiting]: https://desec.readthedocs.io/en/latest/rate-limits.html
package desec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/libdns/libdns"
)

// writeToken is used to synchronize all writes to deSEC to make sure the API here adheres to the
// libdns contract. This is necessary because libdns operates in units of records (zone, name, type,
// value) while deSEC operates in units of record sets (zone, name, type). This makes it necessary
// to perform read-modify cycles that can't be done atomically due to limitations in the deSEC API.
//
// This also can't be a mutex, because the time it takes to perform a write is theoretically
// unbounded due to rate limiting (https://desec.readthedocs.io/en/latest/rate-limits.html) and
// using a mutex here would make it impossible to adhere to context cancellation.
//
// Rate limiting on the deSEC side also means that this is unlikely to become a bottleneck, though
// use cases may exist that cause this single synchronization point to become one.
var writeToken = make(chan struct{}, 1)

func acquireWriteToken(ctx context.Context) error {
	select {
	case <-writeToken:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseWriteToken() {
	writeToken <- struct{}{}
}

func init() {
	writeToken <- struct{}{}
}

// Provider facilitates DNS record manipulation with deSEC.
type Provider struct {
	// Token is a token created on https://desec.io/tokens. A basic token without the permission
	// to manage tokens is sufficient.
	Token string `json:"token,omitempty"`
}

// GetRecords lists all the records in the zone.
//
// Caveat: This method will fail if there are more than 500 RRsets in the zone. See package
// documentation for more detail.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	// https://desec.readthedocs.io/en/latest/dns/rrsets.html#retrieving-all-rrsets-in-a-zone
	rrsets, err := p.listRRSets(ctx, zone)
	if err != nil {
		return nil, err
	}
	var records []libdns.Record
	for _, rrset := range rrsets {
		records0, err := libdnsRecords(rrset)
		if err != nil {
			return nil, fmt.Errorf("parsing RRSet: %v", err)
		}
		records = append(records, records0...)
	}
	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	rrs := make([]libdns.RR, 0, len(records))
	for _, r := range records {
		rrs = append(rrs, r.RR())
	}

	err := acquireWriteToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for inflight requests to finish: %v", err)
	}
	defer releaseWriteToken()

	rrsets := make(map[rrKey]*rrSet)

	// Fetch or create base rrsets to append to
	for _, rr := range rrs {
		key := rrKey{rrSetSubname(rr), rr.Type}
		if _, ok := rrsets[key]; ok {
			continue
		}

		rrset, err := p.getRRSet(ctx, zone, key)
		switch {
		case errors.Is(err, errNotFound):
			// no RRSet exists, create one
			rrset = rrSet{
				Subname: key.Subname,
				Type:    key.Type,
				Records: nil,
				TTL:     rrSetTTL(rr),
			}
		case err != nil:
			return nil, fmt.Errorf("retrieving RRSet: %v", err)
		}
		rrsets[key] = &rrset
	}

	// Merge records into base
	dirty := make(map[rrKey]struct{})
	var ret []libdns.Record
	for _, rr := range rrs {
		key := rrKey{rrSetSubname(rr), rr.Type}
		rrset := rrsets[key]

		v := rrSetRecord(rr)
		if slices.Contains(rrset.Records, v) {
			// Don't modify existing records, if all records in a record set already exist, the
			// record set will not be marked dirty and excluded from the update request.
			continue
		}

		rrset.Records = append(rrset.Records, v)

		r, err := libdns.RR{
			Name: rr.Name,
			Type: rr.Type,
			TTL:  libdnsTTL(*rrset),
			Data: rr.Data,
		}.Parse()
		if err != nil {
			return nil, fmt.Errorf("parsing RR: %v", err)
		}
		ret = append(ret, r)

		// Mark this key as dirty, only dirty keys will result in an update.
		dirty[key] = struct{}{}
	}

	update := make([]rrSet, 0, len(dirty))
	for key := range dirty {
		update = append(update, *rrsets[key])
	}
	if err := p.putRRSets(ctx, zone, update); err != nil {
		return nil, fmt.Errorf("writing RRSets: %v", err)
	}
	return ret, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := acquireWriteToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for inflight requests to finish: %v", err)
	}
	defer releaseWriteToken()

	// Group records by rrKey (name, type)
	rrsetMap := make(map[rrKey]*rrSet)
	for _, r := range records {
		rr := r.RR()
		key := rrKey{rrSetSubname(rr), rr.Type}
		rrset := rrsetMap[key]
		if rrset == nil {
			rrset = &rrSet{
				Subname: key.Subname,
				Type:    key.Type,
				Records: nil,
				TTL:     rrSetTTL(rr),
			}
			rrsetMap[key] = rrset
		}
		rrset.Records = append(rrset.Records, rrSetRecord(rr))
	}

	// Build list of RRSets to pass to the API and list of libdns records
	// to return from the function
	rrsetList := make([]rrSet, 0, len(rrsetMap))
	ret := make([]libdns.Record, 0, len(records))
	for _, rrset := range rrsetMap {
		rrsetList = append(rrsetList, *rrset)
		records0, err := libdnsRecords(*rrset)
		if err != nil {
			return nil, fmt.Errorf("parsing RRSet: %v", err)
		}
		ret = append(ret, records0...)
	}

	if err := p.putRRSets(ctx, zone, rrsetList); err != nil {
		return nil, fmt.Errorf("writing RRSets: %v", err)
	}
	return ret, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	err := acquireWriteToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for inflight requests to finish: %v", err)
	}
	defer releaseWriteToken()

	rrsets := make(map[rrKey]*rrSet)
	dirty := make(map[rrKey]struct{})
	var ret []libdns.Record

	// Fetch rrsets with records requested for deletion.
	for _, r := range records {
		rr := r.RR()
		key := rrKey{rrSetSubname(rr), rr.Type}
		rrset := rrsets[key]
		if rrset == nil {
			rrset0, err := p.getRRSet(ctx, zone, key)
			switch {
			case errors.Is(err, errNotFound):
				continue
			case err != nil:
				return nil, fmt.Errorf("retrieving RRSet: %v", err)
			}
			rrsets[key] = &rrset0
			rrset = &rrset0
		}

		// Delete the record if it exists and mark the rrset as dirty, only dirty rrsets will be
		// updated.
		v := rrSetRecord(rr)
		if i := slices.Index(rrset.Records, v); i >= 0 {
			rrset.Records = slices.Delete(rrset.Records, i, i+1)
			dirty[key] = struct{}{}
			ret = append(ret, r)
		}
	}

	update := make([]rrSet, 0, len(dirty))
	for key := range dirty {
		update = append(update, *rrsets[key])
	}
	if err := p.putRRSets(ctx, zone, update); err != nil {
		return nil, fmt.Errorf("writing RRSets: %v", err)
	}
	return ret, nil
}

// https://desec.readthedocs.io/en/latest/dns/rrsets.html#rrset-field-reference
type rrSet struct {
	Subname string   `json:"subname"`
	Type    string   `json:"type"`
	Records []string `json:"records"`
	TTL     int      `json:"ttl,omitempty"`
}

// rrKey uniquely identifies an rrSet within a zone.
type rrKey struct {
	Subname string
	Type    string
}

func rrsetRecordsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a1, b1 := slices.Clone(a), slices.Clone(b)
	slices.Sort(a1)
	slices.Sort(b1)
	return slices.Equal(a1, b1)
}

func (rrs *rrSet) equal(other *rrSet) bool {
	return rrs.Subname == other.Subname && rrs.Type == other.Type && rrs.TTL == other.TTL && rrsetRecordsEqual(rrs.Records, other.Records)
}

// libdnsName returns the rrSet subname converted to libdns conventions.
//
// deSEC represents the zone itself using an empty (or missing) subname, libdns
// uses "@"
func libdnsName(rrs rrSet) string {
	if rrs.Subname == "" {
		return "@"
	}
	return rrs.Subname
}

// rrSetSubname returns the libdns name converted to deSEC conventions.
//
// deSEC represents the zone itself using an empty (or missing) subname, libdns
// uses "@"
func rrSetSubname(rr libdns.RR) string {
	// deSEC represents the zone itself using an empty (or missing) subname, libdns
	// uses "@"
	if rr.Name == "@" {
		return ""
	}
	return rr.Name
}

// libdnsTTL returns a valid libdns.Record.TTL value.
func libdnsTTL(rrs rrSet) time.Duration {
	return time.Duration(rrs.TTL) * time.Second
}

// rrSetTTL returns a valid rrSetTTL value for the given record.
//
// deSEC has a minimum TTL of 3600 seconds, if the record TTL
// is shorter than 3600 seconds, 3600 seconds is returned.
func rrSetTTL(rr libdns.RR) int {
	ttl := int(rr.TTL / time.Second)
	if ttl < 3600 {
		return 3600
	}
	return ttl
}

// rrSetRecord returns the libdns record value in deSEC conventions.
func rrSetRecord(rr libdns.RR) string {
	switch rr.Type {
	case "TXT":
		// deSEC requires custom quoting for TXT records
		var sb strings.Builder
		sb.WriteRune('"')
		for _, t := range rr.Data {
			switch t {
			case '\\':
				sb.WriteString("\\\\")
			case '"':
				sb.WriteString("\\\"")
			default:
				sb.WriteRune(t)
			}
		}
		sb.WriteRune('"')
		return sb.String()
	default:
		return rr.Data
	}
}

// libdnsRecords returns the libdns.Records corresponding to a given rrSet.
func libdnsRecords(rrs rrSet) ([]libdns.Record, error) {
	records := make([]libdns.Record, 0, len(rrs.Records))
	name := libdnsName(rrs)
	ttl := libdnsTTL(rrs)
	for _, data := range rrs.Records {
		switch rrs.Type {
		case "TXT":
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return nil, fmt.Errorf("parsing %v record value %q: not in quotes", rrs.Type, data)
			}
			// deSEC requires custom quoting for TXT records
			var sb strings.Builder
			for i := 1; i < len(data)-1; {
				t, sz := utf8.DecodeRuneInString(data[i:])
				switch t {
				case '\\':
					i += sz
					t, sz = utf8.DecodeRuneInString(data[i:])
					if t != '\\' && t != '"' {
						return nil, fmt.Errorf("parsing %v record value %q: invalid escape sequence", rrs.Type, data)
					}
					sb.WriteRune(t)
				default:
					sb.WriteRune(t)
				}
				i += sz
			}
			data = sb.String()
		}
		record, err := libdns.RR{
			Name: name,
			Type: rrs.Type,
			TTL:  ttl,
			Data: data,
		}.Parse()
		if err != nil {
			fmt.Println(name)
			return nil, fmt.Errorf("parsing %v record value %q: %v", rrs.Type, data, err)
		}
		records = append(records, record)
	}
	return records, nil
}

type statusError struct {
	code   int
	header http.Header
	body   []byte
}

func (err *statusError) Error() string {
	return fmt.Sprintf("unexpected status code %d: %v", err.code, string(err.body))
}

var errNotFound = errors.New("not found")

func (p *Provider) httpDo0(ctx context.Context, method, url string, in []byte) ([]byte, error) {
	var r io.Reader
	if len(in) > 0 {
		r = bytes.NewReader(in)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}
	req.Header.Set("Authorization", "Token "+p.Token)
	req.Header.Set("Accept", "application/json; charset=utf-8")
	if len(in) > 0 {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		return body, nil
	default:
		return nil, &statusError{code: res.StatusCode, header: res.Header, body: body}
	}
}

func (p *Provider) httpDo(ctx context.Context, method, url string, in []byte) ([]byte, error) {
	for {
		out, err := p.httpDo0(ctx, method, url, in)
		if s := (*statusError)(nil); errors.As(err, &s) && s.code == http.StatusTooManyRequests {
			// rate limited, wait until the next request can be send
			retryAfterHeader := s.header.Get("Retry-After")
			retryAfter, err := strconv.Atoi(retryAfterHeader)
			if err != nil {
				return nil, fmt.Errorf("parsing Retry-After header %q: %v", retryAfterHeader, err)
			}
			select {
			case <-time.After(time.Duration(retryAfter) * time.Second):
			case <-ctx.Done():
				return nil, fmt.Errorf("waiting for cooldown to end: %v", ctx.Err())
			}
			continue // try again
		}
		return out, err
	}
}

func (p *Provider) getRRSet(ctx context.Context, zone string, key rrKey) (rrSet, error) {
	// https://desec.readthedocs.io/en/latest/dns/rrsets.html#retrieving-a-specific-rrset
	subname := key.Subname
	if subname == "" {
		subname = "@"
	}
	domain := url.PathEscape(strings.TrimSuffix(zone, "."))
	url := fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/%s/%s", domain, url.PathEscape(subname), url.PathEscape(key.Type))
	outb, err := p.httpDo(ctx, "GET", url, nil)
	if err != nil {
		if status, ok := err.(*statusError); ok {
			if status.code == http.StatusNotFound {
				return rrSet{}, errNotFound
			}
		}
		return rrSet{}, err
	}

	var out rrSet
	if err := json.Unmarshal(outb, &out); err != nil {
		return rrSet{}, fmt.Errorf("decoding json: %v", err)
	}
	return out, nil
}

func (p *Provider) listRRSets(ctx context.Context, zone string) ([]rrSet, error) {
	// https://desec.readthedocs.io/en/latest/dns/rrsets.html#retrieving-all-rrsets-in-a-zone
	domain := url.PathEscape(strings.TrimSuffix(zone, "."))
	url := fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/", domain)
	buf, err := p.httpDo(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var out []rrSet
	if err := json.Unmarshal(buf, &out); err != nil {
		return nil, fmt.Errorf("decoding json: %v", err)
	}
	return out, nil
}

func (p *Provider) putRRSets(ctx context.Context, zone string, rrs []rrSet) error {
	if len(rrs) == 0 {
		return nil
	}

	// https://desec.readthedocs.io/en/latest/dns/rrsets.html#bulk-modification-of-rrsets
	domain := url.PathEscape(strings.TrimSuffix(zone, "."))
	url := fmt.Sprintf("https://desec.io/api/v1/domains/%s/rrsets/", domain)

	var buf []byte
	var err error
	buf, err = json.Marshal(rrs)
	if err != nil {
		return fmt.Errorf("encoding json: %v", err)
	}

	_, err = p.httpDo(ctx, "PUT", url, buf)
	if err != nil {
		return err
	}
	return nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
