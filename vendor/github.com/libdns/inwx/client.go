package inwx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pquerna/otp/totp"
)

type client struct {
	httpClient  *http.Client
	endpointUrl string
}

type response struct {
	Code         int    `json:"code"`
	Message      string `json:"msg"`
	ReasonCode   string `json:"reasonCode"`
	Reason       string `json:"reason"`
	ResponseData any    `json:"resData"`
}

type errorResponse struct {
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	ReasonCode string `json:"reasonCode"`
	Reason     string `json:"reason"`
}

type nameserverInfoRequest struct {
	Domain   string `json:"domain,omitempty"`
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Content  string `json:"content,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
	Priority uint   `json:"prio,omitempty"`
}

type nameserverInfoResponse struct {
	RoID    int                `mapstructure:"roId"`
	Domain  string             `mapstructure:"domain"`
	Type    string             `mapstructure:"type"`
	Count   int                `mapstructure:"count"`
	Records []nameserverRecord `mapstructure:"record"`
}

type nameserverCreateRecordRequest struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority uint   `json:"prio"`
}

type nameserverCreateRecordResponse struct {
	ID int `mapstructure:"id"`
}

type nameserverUpdateRecordRequest struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority uint   `json:"prio"`
}

type nameserverDeleteRecordRequest struct {
	ID int `mapstructure:"id"`
}

type nameserverRecord struct {
	ID       int    `mapstructure:"id"`
	Name     string `mapstructure:"name"`
	Type     string `mapstructure:"type"`
	Content  string `mapstructure:"content"`
	TTL      int    `mapstructure:"ttl"`
	Priority uint   `mapstructure:"prio"`
}

type nameserverCreateRequest struct {
	Domain string   `json:"domain"`
	Type   string   `json:"type"`
	NS     []string `json:"ns"`
}

type nameserverDeleteRequest struct {
	Domain string `json:"domain"`
}

type accountLoginRequest struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type accountLoginResponse struct {
	TFA string `mapstructure:"tfa"`
}

type accountUnlockRequest struct {
	TAN string `json:"tan"`
}

const endpointURL = "https://api.domrobot.com/jsonrpc/"

func newClient(endpointURL string) (*client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{DisableCompression: true},
		Jar:       jar,
	}

	return &client{httpClient, endpointURL}, nil
}

func (c *client) getRecords(ctx context.Context, domain string) ([]nameserverRecord, error) {
	response, err := c.call(ctx, "nameserver.info", nameserverInfoRequest{
		Domain: domain,
	})

	if err != nil {
		return nil, err
	}

	data := nameserverInfoResponse{}
	err = mapstructure.Decode(response, &data)

	if err != nil {
		return nil, err
	}

	return data.Records, nil
}

func (c *client) findRecords(ctx context.Context, record nameserverRecord, domain string, matchContent bool) ([]nameserverRecord, error) {
	request := nameserverInfoRequest{
		Domain: domain,
		Type:   record.Type,
		Name:   record.Name,
	}

	if matchContent {
		request.Content = record.Content
	}

	response, err := c.call(ctx, "nameserver.info", request)

	if err != nil {
		return nil, err
	}

	data := nameserverInfoResponse{}
	err = mapstructure.Decode(response, &data)

	if err != nil {
		return nil, err
	}

	return data.Records, nil
}

func (c *client) createRecord(ctx context.Context, record nameserverRecord, domain string) (int, error) {
	response, err := c.call(ctx, "nameserver.createRecord", nameserverCreateRecordRequest{
		Domain:   domain,
		Name:     record.Name,
		Type:     record.Type,
		Content:  record.Content,
		TTL:      ensureMinTTL(record.TTL),
		Priority: record.Priority,
	})

	if err != nil {
		return 0, err
	}

	data := nameserverCreateRecordResponse{}
	err = mapstructure.Decode(response, &data)

	if err != nil {
		return 0, err
	}

	return data.ID, nil
}

func (c *client) updateRecord(ctx context.Context, record nameserverRecord) error {
	if record.ID == 0 {
		return fmt.Errorf("record cannot be updated because the ID is not set")
	}

	_, err := c.call(ctx, "nameserver.updateRecord", nameserverUpdateRecordRequest{
		ID:       record.ID,
		Name:     record.Name,
		Type:     record.Type,
		Content:  record.Content,
		TTL:      ensureMinTTL(record.TTL),
		Priority: record.Priority,
	})

	return err
}

func (c *client) deleteRecord(ctx context.Context, record nameserverRecord) error {
	_, err := c.call(ctx, "nameserver.deleteRecord", nameserverDeleteRecordRequest{
		ID: record.ID,
	})

	return err
}

func (c *client) createNameserver(ctx context.Context, domain string, _type string, nameservers []string) error {
	_, err := c.call(ctx, "nameserver.create", nameserverCreateRequest{
		Domain: domain,
		Type:   _type,
		NS:     nameservers,
	})

	return err
}

func (c *client) deleteNameserver(ctx context.Context, domain string) error {
	_, err := c.call(ctx, "nameserver.delete", nameserverDeleteRequest{
		Domain: domain,
	})

	return err
}

func (c *client) login(ctx context.Context, username string, password string, sharedSecret string) error {
	response, err := c.call(ctx, "account.login", accountLoginRequest{
		User: username,
		Pass: password,
	})

	if err != nil {
		return err
	}

	data := accountLoginResponse{}
	err = mapstructure.Decode(response, &data)

	if err != nil {
		return err
	}

	if data.TFA == "GOOGLE-AUTH" {
		tan, err := totp.GenerateCode(sharedSecret, time.Now())

		if err != nil {
			return err
		}

		err = c.unlock(ctx, tan)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) logout(ctx context.Context) error {
	_, err := c.call(ctx, "account.logout", nil)

	return err
}

func (c *client) unlock(ctx context.Context, tan string) error {
	_, err := c.call(ctx, "account.unlock", accountUnlockRequest{
		TAN: tan,
	})

	return err
}

func (c *client) call(ctx context.Context, method string, params any) (any, error) {
	requestBody := map[string]interface{}{}
	requestBody["method"] = method
	requestBody["params"] = params
	requestJsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, "POST", c.endpointUrl, bytes.NewReader(requestJsonBody))
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set("content-type", "application/json; charset=UTF-8")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var response response
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return nil, err
	}

	return response.ResponseData, checkResponse(response)
}

func (r *errorResponse) Error() string {
	if r.Reason != "" {
		return fmt.Sprintf("(%d) %s. Reason: (%s) %s", r.Code, r.Message, r.ReasonCode, r.Reason)
	}

	return fmt.Sprintf("(%d) %s", r.Code, r.Message)
}

func checkResponse(r response) error {
	if c := r.Code; c >= 1000 && c <= 1500 {
		return nil
	}

	return &errorResponse{
		Code:       r.Code,
		Message:    r.Message,
		Reason:     r.Reason,
		ReasonCode: r.ReasonCode,
	}
}

// Make sure that the TTL is at least 300 seconds, because INWX does
// not allow smaller TTL values.
// https://kb.inwx.com/en-us/3-nameserver/120-can-i-change-the-ttl
func ensureMinTTL(ttl int) int {
	if ttl < 300 {
		return 300
	}

	return ttl
}
