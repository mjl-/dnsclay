package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Editing string

const (
	NotEditable Editing = "0"

	Editable Editing = "1"
)

type RecordType string

const (
	CNAMERecordType RecordType = "CNAME"
	ARecordType     RecordType = "A"
	NSRecordType    RecordType = "NS"
	NAPTRRecordType RecordType = "NAPTR"
	SRVRecordType   RecordType = "SRV"
	TXTRecordType   RecordType = "TXT"
	AAAARecordType  RecordType = "AAAA"
)

type DNSRecord struct {
	Comment   string     `json:"comment"`
	AccountID string     `json:"account_id"` // nolint:tagliatelle
	Zone      string     `json:"zone"`
	Record    string     `json:"record"`
	Value     string     `json:"value"`
	Type      RecordType `json:"type"`
	Editable  Editing    `json:"editable"`
}

type DNSRecordInput struct {
	Record string     `url:"record" json:"record"`
	Value  string     `url:"value" json:"value"`
	Type   RecordType `url:"type" json:"type"`
}

const (
	successResult = "success"
	errorResult   = "error" //nolint:varcheck,deadcode
)

type apiResponse struct {
	Data   interface{} `json:"data"`
	Result string      `json:"result"`
}

type listDNSRecordsResponse struct {
	Records []DNSRecord `json:"data"`
	Result  string      `json:"result"`
}

const (
	addRecordCmd    = "dns-add_record"
	listRecordsCmd  = "dns-list_records"
	removeRecordCmd = "dns-remove_record"
)

func (c *Client) AddDNSRecord(ctx context.Context, record DNSRecordInput) error {
	params := dnsRecordParams{
		baseParams: baseParams{
			APIKey:  c.apiKey,
			Command: addRecordCmd,
			Format:  string(jsonFormat),
		},
		DNSRecordInput: record,
	}
	req, err := newRequest(ctx, params)
	if err != nil {
		return errors.Wrap(err, "could not create client")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if _, err := getProcessedRespBody(*resp); err != nil {
		return errors.Wrap(err, "failed to process response body")
	}

	return nil
}

func (c *Client) ListDNSRecords(ctx context.Context) ([]DNSRecord, error) {
	params := baseParams{
		APIKey:  c.apiKey,
		Command: listRecordsCmd,
		Format:  string(jsonFormat),
	}
	req, err := newRequest(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "could not create client")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	bodyStr, err := getProcessedRespBody(*resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to process response body")
	}

	dnsResp := &listDNSRecordsResponse{}
	err = json.Unmarshal(bodyStr, dnsResp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal list dns response")
	}

	if dnsResp.Result != successResult {
		return nil, fmt.Errorf("operation failed - response: %s", bodyStr)
	}

	// DreamHost API's response ordering is nondeterministic.
	// Integrations can depend on API response stability, so let's sort it.
	records := dnsResp.Records
	sort.Slice(records, func(i, j int) bool {
		return dnsResp.Records[i].Record < dnsResp.Records[j].Record
	})

	return dnsResp.Records, nil
}

func (c *Client) RemoveDNSRecord(ctx context.Context, record DNSRecordInput) error {
	params := dnsRecordParams{
		baseParams: baseParams{
			APIKey:  c.apiKey,
			Command: removeRecordCmd,
			Format:  string(jsonFormat),
		},
		DNSRecordInput: record,
	}
	req, err := newRequest(ctx, params)
	if err != nil {
		return errors.Wrap(err, "could not create client")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()

	if _, err := getProcessedRespBody(*resp); err != nil {
		return errors.Wrap(err, "failed to process response body")
	}

	return nil
}

func getProcessedRespBody(resp http.Response) ([]byte, error) {
	bodyDump, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of response")
	}

	if err := checkSuccessStatus(bodyDump); err != nil {
		return nil, err
	}

	bodyStr := string(bodyDump)
	return []byte(strings.ReplaceAll(bodyStr, "\\\"", "\"")), nil
}

func checkSuccessStatus(bodyStr []byte) error {
	apiResp := apiResponse{}
	err := json.Unmarshal(bodyStr, &apiResp)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal API response")
	}
	if apiResp.Result != successResult {
		return fmt.Errorf("operation failed - response: %s", apiResp.Data)
	}

	return nil
}
