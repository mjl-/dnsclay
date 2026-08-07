// Implements the libdns interfaces for name.com
// https://www.name.com/api-docs
package namedotcom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/libdns/libdns"
	"github.com/pkg/errors"
)

// default timeout for the http request handler (seconds)
const HTTP_TIMEOUT = 30

type (
	nameDotCom struct {
		Server string `json:"server,omitempty"`
		User   string `json:"user,omitempty"`
		Token  string `json:"token,omitempty"`
		client *http.Client
	}

	// listRecordsResponse contains the response for the ListRecords function.
	listRecordsResponse struct {
		Records  []nameDotComRecord `json:"records,omitempty"`
		NextPage int32              `json:"nextPage,omitempty"`
		LastPage int32              `json:"lastPage,omitempty"`
	}

	// nameDotComRecord is an individual DNS resource record for name.com.
	nameDotComRecord struct {
		ID         int32  `json:"id,omitempty"`
		DomainName string `json:"domainName,omitempty"`
		Host       string `json:"host,omitempty"`
		Fqdn       string `json:"fqdn,omitempty"`
		Type       string `json:"type,omitempty"`
		Answer     string `json:"answer,omitempty"`
		TTL        uint32 `json:"ttl,omitempty"`
		Priority   uint32 `json:"priority,omitempty"`
	}
)

type (
	// errorResponse is what is returned if the HTTP status code is not 200.
	errorResponse struct {
		// Message is the error message.
		Message string `json:"message,omitempty"`
		// Details may have some additional details about the error.
		Details string `json:"details,omitempty"`
	}
)

func (er errorResponse) Error() string {
	return er.Message + ": " + er.Details
}

// errorResponse - provides a more verbose stderr
func (n *nameDotCom) errorResponse(resp *http.Response) error {
	er := &errorResponse{}
	err := json.NewDecoder(resp.Body).Decode(er)
	if err != nil {
		return errors.Wrap(err, "api returned unexpected response")
	}

	return errors.WithStack(er)
}

// doRequest is the base http request handler including a request context.
func (n *nameDotCom) doRequest(ctx context.Context, method, endpoint string, post io.Reader) (io.Reader, error) {
	uri := n.Server + endpoint
	req, err := http.NewRequestWithContext(ctx, method, uri, post) // the offical name.com go client does not implement ctx
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(n.User, n.Token)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, n.errorResponse(resp)
	}

	return resp.Body, nil
}

// fromLibDNSRecord maps a name.com record from a libdns record
func (n *nameDotComRecord) fromLibDNSRecord(record libdns.Record, zone string) {
	var id int64
	if record.ID != "" {
		id, _ = strconv.ParseInt(record.ID, 10, 32)
	}
	n.ID = int32(id)
	n.Type = record.Type
	n.Host = n.toSanitized(record, zone)
	n.Answer = record.Value
	n.TTL = uint32(record.TTL.Seconds())
}

// toLibDNSRecord maps a name.com record to a libdns record
func (n *nameDotComRecord) toLibDNSRecord() libdns.Record {
	return libdns.Record{
		ID:    fmt.Sprint(n.ID),
		Type:  n.Type,
		Name:  n.Host,
		Value: n.Answer,
		TTL:   time.Duration(n.TTL) * time.Second,
	}
}

// name.com's api server expects the sub domain name to be relavtive and have no trailing period
// , e.g. "sub.zone." -> "sub"
func (n *nameDotComRecord) toSanitized(libdnsRecord libdns.Record, zone string) string {
	return strings.TrimSuffix(strings.Replace(libdnsRecord.Name, zone, "", -1), ".")
}

// NewNameDotComClient returns a new name.com client struct
func NewNameDotComClient(ctx context.Context, token, user, server string) (*nameDotCom, error) {
	re := regexp.MustCompile(`^https://.+\.com$`)
	validURL := re.MatchString(server)
	if !validURL {
		return nil, errors.New("invalid url scheme, expecting https:// prefix")
	}

	httpClient := &http.Client{Timeout: HTTP_TIMEOUT * time.Second}

	return &nameDotCom{
		server, user, token,
		httpClient,
	}, nil
}
