package westcn

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const defaultEndpoint = "https://api.west.cn/api/v2"

// Client West.cn API 客户端
type Client struct {
	username string
	password string

	encoder  *encoding.Encoder
	decoder  *encoding.Decoder
	endpoint *url.URL
}

// NewClient 创建新 West.cn API 客户端
func NewClient(username, password string) *Client {
	endpoint, _ := url.Parse(defaultEndpoint)

	return &Client{
		username: username,
		password: password,
		encoder:  simplifiedchinese.GBK.NewEncoder(),
		decoder:  simplifiedchinese.GBK.NewDecoder(),
		endpoint: endpoint,
	}
}

// AppendRecord 添加一条记录
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#37u3001u6dfbu52a0u57dfu540du89e3u67900a3ca20id3d37u3001u6dfbu52a0u57dfu540du89e3u67903e203ca3e
func (c *Client) AppendRecord(ctx context.Context, zone string, record Record) (int, error) {
	values := url.Values{}
	values.Set("domain", strings.TrimSuffix(zone, "."))
	values.Set("host", record.Host)
	values.Set("type", record.Type)
	values.Set("value", record.Value)
	values.Set("ttl", strconv.Itoa(record.TTL))
	values.Set("level", strconv.Itoa(record.Level))

	req, err := c.newReq(ctx, "domain", "adddnsrecord", values)
	if err != nil {
		return 0, err
	}

	raw, err := c.do(req)
	if err != nil {
		return 0, err
	}
	var resp RecordIDResponse
	if err = c.decodeResp(raw, &resp); err != nil {
		return 0, err
	}
	if resp.Result != http.StatusOK {
		return 0, fmt.Errorf("%s, error code: %d", resp.Msg, resp.ErrorCode)
	}

	return resp.Data.ID, nil
}

// GetRecords 获取所有记录
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#310u3001u83b7u53d6u57dfu540du89e3u6790u8bb0u5f550a3ca20id3d310u3001u83b7u53d6u57dfu540du89e3u6790u8bb0u5f553e203ca3e
func (c *Client) GetRecords(ctx context.Context, zone string) ([]Record, error) {
	values := url.Values{}
	values.Set("domain", strings.TrimSuffix(zone, "."))
	values.Set("limit", "1000")

	req, err := c.newReq(ctx, "domain", "getdnsrecord", values)
	if err != nil {
		return nil, err
	}

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var resp RecordsResponse
	if err = c.decodeResp(raw, &resp); err != nil {
		return nil, err
	}
	if resp.Result != http.StatusOK {
		return nil, fmt.Errorf("%s, error code: %d", resp.Msg, resp.ErrorCode)
	}

	return resp.Data.Records, nil
}

// DeleteRecord 删除一条记录
// https://www.west.cn/CustomerCenter/doc/domain_v2.html#39u3001u5220u9664u57dfu540du89e3u67900a3ca20id3d39u3001u5220u9664u57dfu540du89e3u67903e203ca3e
func (c *Client) DeleteRecord(ctx context.Context, zone string, recordID int) error {
	values := url.Values{}
	values.Set("domain", strings.TrimSuffix(zone, "."))
	values.Set("id", strconv.Itoa(recordID))

	req, err := c.newReq(ctx, "domain", "deldnsrecord", values)
	if err != nil {
		return err
	}

	raw, err := c.do(req)
	if err != nil {
		return err
	}
	var resp APIResponse
	if err = c.decodeResp(raw, &resp); err != nil {
		return err
	}
	if resp.Result != http.StatusOK {
		return fmt.Errorf("%s, error code: %d", resp.Msg, resp.ErrorCode)
	}

	return nil
}

// newReq 创建新请求
func (c *Client) newReq(ctx context.Context, p, act string, form url.Values) (*http.Request, error) {
	// 签名请求
	// https://www.west.cn/CustomerCenter/doc/apiv2.html#12u3001u8eabu4efdu9a8cu8bc10a3ca20id3d12u3001u8eabu4efdu9a8cu8bc13e203ca3e
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sum := md5.Sum([]byte(c.username + c.password + timestamp))
	form.Set("username", c.username)
	form.Set("time", timestamp)
	form.Set("token", hex.EncodeToString(sum[:]))

	values, err := c.encodeValues(form)
	if err != nil {
		return nil, err
	}

	endpoint := c.endpoint.JoinPath(p, "/")

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

// do 发送请求
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return io.ReadAll(resp.Body)
}

// encodeValues GBK 编码 URL 参数
func (c *Client) encodeValues(values url.Values) (url.Values, error) {
	result := make(url.Values)

	for k, vs := range values {
		encKey, err := c.encoder.String(k)
		if err != nil {
			return nil, err
		}
		for _, v := range vs {
			encValue, err := c.encoder.String(v)
			if err != nil {
				return nil, err
			}
			result.Add(encKey, encValue)
		}
	}

	return result, nil
}

// decodeResp GBK 解码响应
func (c *Client) decodeResp(raw []byte, v any) error {
	return json.NewDecoder(transform.NewReader(bytes.NewBuffer(raw), c.decoder)).Decode(v)
}
