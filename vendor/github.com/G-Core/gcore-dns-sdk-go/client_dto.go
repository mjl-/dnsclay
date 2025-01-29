package dnssdk

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
)

// ListZones dto to read list of zones from API
type ListZones struct {
	Zones       []Zone `json:"zones"`
	TotalAmount int    `json:"total_amount"`
	Error       string `json:"error,omitempty"`
}

// Zone dto to read info from API
type Zone struct {
	Name    string       `json:"name"`
	Records []ZoneRecord `json:"records"`
}

// AddZone dto to create new zone
type AddZone struct {
	Name string `json:"name"`
}

// CreateResponse dto to create new zone
type CreateResponse struct {
	ID    uint64 `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type RRSetMeta map[string]any

// RRSet dto as part of zone info from API
type RRSet struct {
	Type    string           `json:"type"`
	TTL     int              `json:"ttl"`
	Records []ResourceRecord `json:"resource_records"`
	Filters []RecordFilter   `json:"filters"`
	Meta    RRSetMeta        `json:"meta"` // this one for failover, not Meta property inside Records
	// according to https://api.gcore.com/docs/dns#tag/rrsets/operation/CreateRRSet
	/*
	   asn (array of int)
	   continents (array of string)
	   countries (array of string)
	   latlong (array of float64, latitude and longitude)
	   fallback (bool)
	   backup (bool)
	   notes (string)
	   weight (float)
	   ip (string)
	   failover (object, beta feature, might be changed in the future) can have fields 10.1. protocol (string, required, HTTP, TCP, UDP, ICMP) 10.2. port (int, required, 1-65535) 10.3. frequency (int, required, in seconds 10-3600) 10.4. timeout (int, required, in seconds 1-10), 10.5. method (string, only for protocol=HTTP) 10.6. command (string, bytes to be sent only for protocol=TCP/UDP) 10.7. url (string, only for protocol=HTTP) 10.8. tls (bool, only for protocol=HTTP) 10.9. regexp (string regex to match, only for non-ICMP) 10.10. http_status_code (int, only for protocol=HTTP) 10.11. host (string, only for protocol=HTTP)
	*/
}

// SetMetaAsn
func (r *RRSet) SetMetaAsn(asns []int) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["asn"] = asns
	return r
}

// SetMetaContinents
func (r *RRSet) SetMetaContinents(continents []string) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["continents"] = continents
	return r
}

// SetMetaCountries
func (r *RRSet) SetMetaCountries(countries []string) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["countries"] = countries
	return r
}

// SetMetaLatLong
func (r *RRSet) SetMetaLatLong(lat, long float64) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["latlong"] = []float64{lat, long}
	return r
}

// SetMetaFallback
func (r *RRSet) SetMetaFallback(fallback bool) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["fallback"] = fallback
	return r
}

// SetMetaBackup
func (r *RRSet) SetMetaBackup(backup bool) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["backup"] = backup
	return r
}

// SetMetaNotes
func (r *RRSet) SetMetaNotes(notes string) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["notes"] = notes
	return r
}

// SetMetaWeight
func (r *RRSet) SetMetaWeight(weight float64) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["weight"] = weight
	return r
}

// SetMetaIP
func (r *RRSet) SetMetaIP(ip string) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["ip"] = ip
	return r
}

// NewRRSetMetaFailoverFromMap for failover
func NewRRSetMetaFailoverFromMap(failover map[string]any) *RRSetMeta {
	return &RRSetMeta{
		"failover": failover,
	}
}

// SetMetaFailover
func (r *RRSet) SetMetaFailover(failover map[string]any) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["failover"] = failover
	return r
}

// NewRRSetMetaGeodnsLink

func NewRRSetMetaGeodnsLink(link string) *RRSetMeta {
	return &RRSetMeta{
		"geodns_link": link,
	}
}

// SetMetaGeodnsLink
func (r *RRSet) SetMetaGeodnsLink(link string) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["geodns_link"] = link
	return r
}

// FailoverHttpCheck for failover meta property with protocol=HTTP
type FailoverHttpCheck struct {
	Protocol  string `json:"protocol"` // HTTP
	Port      uint16 `json:"port"`
	Frequency uint16 `json:"frequency"`
	Timeout   uint16 `json:"timeout"`
	// HTTP only
	Method         string  `json:"method,omitempty"` // GET, POST, PUT, DELETE, PATCH
	URL            string  `json:"url,omitempty"`    // without / prefix
	Host           *string `json:"host,omitempty"`
	HttpStatusCode *uint16 `json:"http_status_code,omitempty"` // 100-599
	Regexp         *string `json:"regexp,omitempty"`
	TLS            bool    `json:"tls"`
}

// FailoverTcpUdpCheck for failover meta property with protocol=TCP|UDP
type FailoverTcpUdpCheck struct {
	Protocol  string `json:"protocol"` // TCP or UDP
	Port      uint16 `json:"port"`
	Frequency uint16 `json:"frequency"`
	Timeout   uint16 `json:"timeout"`
	// TCP/UDP only
	Command *string `json:"command"` // bytes to sent
	Regexp  *string `json:"regexp,omitempty"`
}

// FailoverIcmpCheck for failover meta property with protocol=ICMP
type FailoverIcmpCheck struct {
	Protocol  string `json:"protocol"` // ICMP
	Port      uint16 `json:"port"`
	Frequency uint16 `json:"frequency"`
	Timeout   uint16 `json:"timeout"`
}

// NewRRSetMetaMetaFailoverFromHttp for failover
func NewRRSetMetaMetaFailoverFromHttp(failover FailoverHttpCheck) *RRSetMeta {
	return &RRSetMeta{
		"failover": failover,
	}
}

// SetMetaFailoverHttp for failover
func (r *RRSet) SetMetaFailoverHttp(check FailoverHttpCheck) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["failover"] = check
	return r
}

// NewRRSetMetaMetaFailoverFromTcpUdp for TCP/DUP failover
func NewRRSetMetaMetaFailoverFromTcpUdp(failover FailoverTcpUdpCheck) *RRSetMeta {
	return &RRSetMeta{
		"failover": failover,
	}
}

// SetMetaFailoverTcpUdp for TCP/DUP failover
func (r *RRSet) SetMetaFailoverTcpUdp(check FailoverTcpUdpCheck) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["failover"] = check
	return r
}

// NewRRSetMetaMetaFailoverFromIcmp for ICMP failover
func NewRRSetMetaMetaFailoverFromIcmp(failover FailoverIcmpCheck) *RRSetMeta {
	return &RRSetMeta{
		"failover": failover,
	}
}

// SetMetaFailoverIcmp for ICMP failover
func (r *RRSet) SetMetaFailoverIcmp(check FailoverIcmpCheck) *RRSet {
	if r.Meta == nil {
		r.Meta = RRSetMeta{}
	}
	r.Meta["failover"] = check
	return r
}

type RRSets struct {
	RRSets []RRSet `json:"rrsets"`
}

// ResourceRecord dto describe records in RRSet
type ResourceRecord struct {
	Content []any          `json:"content"`
	Meta    map[string]any `json:"meta"`
	Enabled bool           `json:"enabled"`
}

// ContentToString as short value
// example from:
// tls-ech.dev.            899     IN      HTTPS   1 . ech=AEn+DQBFKwAgACABWIHUGj4u+PIggYXcR5JF0gYk3dCRioBW8uJq9H4mKAAIAAEAAQABAANAEnB1YmxpYy50bHMtZWNoLmRldgAA
// clickhouse.com.         899     IN      HTTPS   1 . alpn="h3,h3-29,h2" ipv4hint=172.66.40.249,172.66.43.7 ipv6hint=2606:4700:3108::ac42:28f9,2606:4700:3108::ac42:2b07
func (r ResourceRecord) ContentToString() string {
	parts := make([]string, len(r.Content))
	for i := range r.Content {
		if arr, ok := r.Content[i].([]any); ok {
			if len(arr) > 0 {
				key := fmt.Sprint(arr[0])
				if key == "alpn" { // only alpn quoted
					parts[i] = fmt.Sprintf(`%s="%s"`, key, valuesJoin(arr[1:]))
				} else if len(arr) == 1 {
					parts[i] = key
				} else {
					parts[i] = fmt.Sprintf(`%s=%s`, key, valuesJoin(arr[1:]))
				}
			}
		} else {
			parts[i] = fmt.Sprint(r.Content[i])
		}
	}
	return strings.Join(parts, " ")
}

// join https kv-params
func valuesJoin(vs []any) (res string) {
	for _, v := range vs {
		res += "," + fmt.Sprint(v)
	}
	res = strings.Trim(res, ",")
	return res
}

// RecordFilter describe Filters in RRSet
type RecordFilter struct {
	Limit  uint   `json:"limit"`
	Type   string `json:"type"`
	Strict bool   `json:"strict"`
}

// NewGeoDNSFilter for RRSet
func NewGeoDNSFilter(limit uint, strict bool) RecordFilter {
	return RecordFilter{
		Limit:  limit,
		Type:   "geodns",
		Strict: strict,
	}
}

// NewGeoDistanceFilter for RRSet
func NewGeoDistanceFilter(limit uint, strict bool) RecordFilter {
	return RecordFilter{
		Limit:  limit,
		Type:   "geodistance",
		Strict: strict,
	}
}

// NewDefaultFilter for RRSet
func NewDefaultFilter(limit uint, strict bool) RecordFilter {
	return RecordFilter{
		Limit:  limit,
		Type:   "default",
		Strict: strict,
	}
}

// NewFirstNFilter for RRSet
func NewFirstNFilter(limit uint, strict bool) RecordFilter {
	return RecordFilter{
		Limit:  limit,
		Type:   "first_n",
		Strict: strict,
	}
}

// RecordType contract
type RecordType interface {
	ToContent() []any
}

// RecordTypeMX as type of record
type RecordTypeMX string

// ToContent convertor
func (mx RecordTypeMX) ToContent() []any {
	parts := strings.Split(string(mx), " ")
	// nolint: gomnd
	if len(parts) != 2 {
		return nil
	}
	content := make([]any, len(parts))
	// nolint: gomnd
	content[1] = parts[1]
	// nolint: gomnd
	content[0], _ = strconv.ParseInt(parts[0], 10, 64)

	return content
}

// RecordTypeCAA as type of record
type RecordTypeCAA string

// ToContent convertor
func (caa RecordTypeCAA) ToContent() []any {
	parts := strings.Split(string(caa), " ")
	// nolint: gomnd
	if len(parts) < 3 {
		return nil
	}
	content := make([]any, len(parts))
	// nolint: gomnd
	content[1] = parts[1]
	// nolint: gomnd
	content[2] = strings.Join(parts[2:], " ")
	// nolint: gomnd
	content[0], _ = strconv.ParseInt(parts[0], 10, 64)

	return content
}

// RecordTypeHTTPS_SCVB as type of record
type RecordTypeHTTPS_SCVB string

// function to parse uint16
func tryParseUint16(x string) any {
	v, err := strconv.ParseFloat(x, 64)
	if err != nil {
		return x
	}
	if v > math.MaxUint16 || v < 0 {
		return v
	}
	u := uint16(v)
	if float64(u) != v { // contains fractional
		return v
	}
	return u
}

// ToContent convertor
func (r RecordTypeHTTPS_SCVB) ToContent() (res []any) {
	arr := strings.Split(string(r), ` `)
	if len(arr) == 0 {
		return []any{}
	}
	if len(arr) >= 1 { // try parse priority
		res = append(res, tryParseUint16(arr[0]))
	}
	if len(arr) >= 2 { // try parse port
		res = append(res, arr[1])
	}
	for i := 2; i < len(arr); i++ { // try parse params
		kvParam := arr[i]
		idx := strings.Index(kvParam, `=`) + 1
		if idx <= 0 {
			idx = len(kvParam) + 1
		}
		param := []any{}
		k := kvParam[:idx-1]
		param = append(param, k)
		if idx <= len(kvParam) {
			vStr := kvParam[idx:]
			vStr = strings.Trim(vStr, `"`) // remove quote
			vArr := strings.Split(vStr, `,`)
			for _, v := range vArr {
				if k == `port` { // try parse to number
					param = append(param, tryParseUint16(v))
					continue
				}
				param = append(param, v)
			}
		}
		res = append(res, param)
	}
	return res
}

// RecordTypeSRV as type of record
type RecordTypeSRV string

// ToContent convertor
func (srv RecordTypeSRV) ToContent() []any {
	parts := strings.Split(string(srv), " ")
	// nolint: gomnd
	if len(parts) != 4 {
		return nil
	}
	content := make([]any, len(parts))
	// nolint: gomnd
	content[0], _ = strconv.ParseInt(parts[0], 10, 64)
	// nolint: gomnd
	content[1], _ = strconv.ParseInt(parts[1], 10, 64)
	// nolint: gomnd
	content[2], _ = strconv.ParseInt(parts[2], 10, 64)
	// nolint: gomnd
	content[3] = parts[3]

	return content
}

// RecordTypeAny as type of record
type RecordTypeAny string

// ToContent convertor
func (x RecordTypeAny) ToContent() []any {
	return []any{string(x)}
}

// ToRecordType builder
func ToRecordType(rType, content string) RecordType {
	switch strings.ToLower(rType) {
	case "mx":
		return RecordTypeMX(content)
	case "caa":
		return RecordTypeCAA(content)
	case "srv":
		return RecordTypeSRV(content)
	case "https", "scvb":
		return RecordTypeHTTPS_SCVB(content)
	}
	return RecordTypeAny(content)
}

// ContentFromValue convertor from flat value to valid for api
func ContentFromValue(recordType, content string) []any {
	rt := ToRecordType(recordType, content)
	if rt == nil {
		return nil
	}
	return rt.ToContent()
}

// ResourceMeta for ResourceRecord
type ResourceMeta struct {
	name     string
	value    any
	validErr error
}

// Valid error
func (rm ResourceMeta) Valid() error {
	return rm.validErr
}

// NewResourceMetaIP for ip meta
func NewResourceMetaIP(ips ...string) ResourceMeta {
	for _, v := range ips {
		_, _, err := net.ParseCIDR(v)
		if err != nil {
			if ip := net.ParseIP(v); ip == nil {
				// nolint: goerr113
				return ResourceMeta{validErr: fmt.Errorf("wrong ip: %v", err)}
			}
		}
	}
	return ResourceMeta{
		name:  "ip",
		value: ips,
	}
}

// NewResourceMetaAsn for asn meta
func NewResourceMetaAsn(asn ...uint64) ResourceMeta {
	return ResourceMeta{
		name:  "asn",
		value: asn,
	}
}

// NewResourceMetaLatLong for lat long meta
func NewResourceMetaLatLong(latlong string) ResourceMeta {
	latlong = strings.TrimLeft(latlong, "(")
	latlong = strings.TrimLeft(latlong, "[")
	latlong = strings.TrimLeft(latlong, "{")
	latlong = strings.TrimRight(latlong, ")")
	latlong = strings.TrimRight(latlong, "]")
	latlong = strings.TrimRight(latlong, "}")
	parts := strings.Split(strings.ReplaceAll(latlong, " ", ""), ",")
	// nolint: gomnd
	if len(parts) != 2 {
		// nolint: goerr113
		return ResourceMeta{validErr: fmt.Errorf("latlong invalid format")}
	}
	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		// nolint: goerr113
		return ResourceMeta{validErr: fmt.Errorf("lat is invalid: %w", err)}
	}
	long, err := strconv.ParseFloat(parts[1], 64)
	// nolint: goerr113
	if err != nil {
		return ResourceMeta{validErr: fmt.Errorf("long is invalid: %w", err)}
	}

	return ResourceMeta{
		name:  "latlong",
		value: []float64{lat, long},
	}
}

// NewResourceMetaNotes for notes meta
func NewResourceMetaNotes(notes ...string) ResourceMeta {
	return ResourceMeta{
		name:  "notes",
		value: notes,
	}
}

// NewResourceMetaCountries for Countries meta
func NewResourceMetaCountries(countries ...string) ResourceMeta {
	return ResourceMeta{
		name:  "countries",
		value: countries,
	}
}

// NewResourceMetaContinents for continents meta
func NewResourceMetaContinents(continents ...string) ResourceMeta {
	return ResourceMeta{
		name:  "continents",
		value: continents,
	}
}

// NewResourceMetaDefault for default meta
func NewResourceMetaDefault() ResourceMeta {
	return ResourceMeta{
		name:  "default",
		value: true,
	}
}

// NewResourceMetaBackup for backup meta
func NewResourceMetaBackup() ResourceMeta {
	return ResourceMeta{
		name:  "backup",
		value: true,
	}
}

// NewResourceMetaFallback for fallback meta
func NewResourceMetaFallback() ResourceMeta {
	return ResourceMeta{
		name:  "fallback",
		value: true,
	}
}

// NewResourceMetaWeight for fallback meta
func NewResourceMetaWeight(weight int) ResourceMeta {
	return ResourceMeta{
		name:  "weight",
		value: weight,
	}
}

// SetContent to ResourceRecord
func (r *ResourceRecord) SetContent(recordType, val string) *ResourceRecord {
	r.Content = ContentFromValue(recordType, val)
	return r
}

// AddMeta to ResourceRecord
func (r *ResourceRecord) AddMeta(meta ResourceMeta) *ResourceRecord {
	if meta.validErr != nil {
		return r
	}
	if meta.name == "" || meta.value == "" {
		return r
	}
	if r.Meta == nil {
		r.Meta = map[string]any{}
	}
	r.Meta[meta.name] = meta.value
	return r
}

// AddFilter to RRSet
func (rr *RRSet) AddFilter(filters ...RecordFilter) *RRSet {
	if rr.Filters == nil {
		rr.Filters = make([]RecordFilter, 0)
	}
	rr.Filters = append(rr.Filters, filters...)
	return rr
}

// ZoneRecord dto describe records in Zone
type ZoneRecord struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	TTL          uint     `json:"ttl"`
	ShortAnswers []string `json:"short_answers"`
}

// DNSSecDS dto describe DS records that returned from API on DNSSec requests
type DNSSecDS struct {
	Algorithm       string `json:"algorithm"`
	Digest          string `json:"digest"`
	DigestAlgorithm string `json:"digest_algorithm"`
	DigestType      string `json:"digest_type"`
	Ds              string `json:"ds"`
	Flags           int    `json:"flags"`
	KeyTag          int    `json:"key_tag"`
	KeyType         string `json:"key_type"`
	PublicKey       string `json:"public_key"`
	Uuid            string `json:"uuid"`
}

// APIError customization for API calls
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error,omitempty"`
}

// Error implementation
func (a APIError) Error() string {
	return fmt.Sprintf("%d: %s", a.StatusCode, a.Message)
}
