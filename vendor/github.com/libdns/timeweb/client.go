package timeweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libdns/libdns"
	"io"
	"net/http"
)

func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (SavedRecord, error) {
	body, err := json.Marshal(libdnsToRecord(record))
	reqURL := fmt.Sprintf("%s/domains/%s/dns-records", p.ApiURL, zone)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result, err
}

func (p *Provider) updateRecord(ctx context.Context, zone string, record libdns.Record) (SavedRecord, error) {
	body, err := json.Marshal(libdnsToRecord(record))
	reqURL := fmt.Sprintf("%s/domains/%s/dns-records/%s", p.ApiURL, zone, record.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))

	var result SavedRecord
	err = p.doAPIRequest(req, &result)

	return result, err
}

func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) error {
	reqURL := fmt.Sprintf("%s/domains/%s/dns-records/%s", p.ApiURL, zone, record.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)

	err = p.doAPIRequest(req, nil)

	return err
}

func (p *Provider) doAPIRequest(req *http.Request, result interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.ApiToken))

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if response.StatusCode >= 400 {
		return fmt.Errorf("got error status: HTTP %d: %+v", response.StatusCode, string(body))
	}

	if response.StatusCode == http.StatusNoContent {
		return err
	}

	err = json.Unmarshal(body, &result)

	return err
}
