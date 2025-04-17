package westcn

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	querystring "github.com/google/go-querystring/query"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const defaultBaseURL = "https://api.west.cn/api/v2"

// Client the West.cn API client.
type Client struct {
	username string
	password string

	encoder *encoding.Encoder

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		username:   username,
		password:   password,
		encoder:    simplifiedchinese.GBK.NewEncoder(),
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddRecord adds a record.
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#37u3001u6dfbu52a0u57dfu540du89e3u67900a3ca20id3d37u3001u6dfbu52a0u57dfu540du89e3u67903e203ca3e
func (c *Client) AddRecord(ctx context.Context, record Record) (int, error) {
	values, err := querystring.Values(record)
	if err != nil {
		return 0, err
	}

	req, err := c.newRequest(ctx, "domain", "adddnsrecord", values)
	if err != nil {
		return 0, err
	}

	results := &APIResponse[RecordID]{}
	if err = c.do(req, results); err != nil {
		return 0, err
	}

	if results.Result != http.StatusOK {
		return 0, results
	}

	return results.Data.ID, nil
}

// GetRecords gets all records.
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#310u3001u83b7u53d6u57dfu540du89e3u6790u8bb0u5f550a3ca20id3d310u3001u83b7u53d6u57dfu540du89e3u6790u8bb0u5f553e203ca3e
func (c *Client) GetRecords(ctx context.Context, domain string) ([]Item, error) {
	values := url.Values{}
	values.Set("domain", domain)
	values.Set("limit", "1000")

	req, err := c.newRequest(ctx, "domain", "getdnsrecord", values)
	if err != nil {
		return nil, err
	}

	results := &APIResponse[Records]{}
	if err = c.do(req, results); err != nil {
		return nil, err
	}

	if results.Result != http.StatusOK {
		return nil, results
	}

	return results.Data.Items, nil
}

// DeleteRecord deleted a record.
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#39u3001u5220u9664u57dfu540du89e3u67900a3ca20id3d39u3001u5220u9664u57dfu540du89e3u67903e203ca3e
func (c *Client) DeleteRecord(ctx context.Context, domain string, recordID int) error {
	values := url.Values{}
	values.Set("domain", domain)
	values.Set("id", strconv.Itoa(recordID))

	req, err := c.newRequest(ctx, "domain", "deldnsrecord", values)
	if err != nil {
		return err
	}

	results := &APIResponse[any]{}
	if err = c.do(req, results); err != nil {
		return err
	}

	if results.Result != http.StatusOK {
		return results
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, p, act string, form url.Values) (*http.Request, error) {
	if form == nil {
		form = url.Values{}
	}

	c.sign(form, time.Now())

	values, err := c.convertURLValues(form)
	if err != nil {
		return nil, err
	}

	endpoint := c.baseURL.JoinPath(p, "/")

	query := endpoint.Query()
	query.Set("act", act)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

// sign
// https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
func (c *Client) sign(form url.Values, now time.Time) {
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	sum := md5.Sum([]byte(c.username + c.password + timestamp))

	form.Set("username", c.username)
	form.Set("time", timestamp)
	form.Set("token", hex.EncodeToString(sum[:]))
}

func (c *Client) do(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return parseError(req, resp)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = gbkDecoder(raw).Decode(result); err != nil {
		return err
	}

	return nil
}

func (c *Client) convertURLValues(values url.Values) (url.Values, error) {
	results := make(url.Values)

	for key, vs := range values {
		encKey, err := c.encoder.String(key)
		if err != nil {
			return nil, err
		}

		for _, value := range vs {
			encValue, err := c.encoder.String(value)
			if err != nil {
				return nil, err
			}

			results.Add(encKey, encValue)
		}
	}

	return results, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)
	result := &APIResponse[any]{}

	if err := gbkDecoder(raw).Decode(result); err != nil {
		return err
	}

	return result
}

func gbkDecoder(raw []byte) *json.Decoder {
	return json.NewDecoder(transform.NewReader(bytes.NewBuffer(raw), simplifiedchinese.GBK.NewDecoder()))
}
