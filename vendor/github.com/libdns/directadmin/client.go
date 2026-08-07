package directadmin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libdns/libdns"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (p *Provider) getZoneRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		fmt.Printf("failed to parse server url: %v\n", err)
		return nil, err
	}

	rootZone, err := p.findRoot(ctx, zone)
	if err != nil {
		rootZone = zone
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("json", "yes")
	queryString.Set("full_mx_records", "yes")
	queryString.Set("allow_dns_underscore", "yes")
	queryString.Set("ttl", "yes")
	queryString.Set("domain", rootZone)

	reqURL.RawQuery = queryString.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		fmt.Printf("failed to build new request: %v\n", err)
		return nil, err
	}

	req.SetBasicAuth(p.User, p.LoginKey)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: p.InsecureRequests,
			},
		}}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to execute request: %v\n", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("api response error, status code: %v\n", resp.StatusCode)
		return nil, err
	}

	var respData daZone
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		fmt.Printf("failed to json decode response: %v\n", err)
		return nil, err
	}

	recs := make([]libdns.Record, 0, len(respData.Records))
	for i := range respData.Records {
		libDnsRecord, err := respData.Records[i].libdnsRecord(zone)
		if err != nil {
			switch {
			case errors.Is(err, ErrUnsupported):
				fmt.Printf("unsupported record conversion of type %v: %v\n", libDnsRecord.Type, libDnsRecord.Name)
				continue
			default:
				return nil, err
			}
		}
		recs = append(recs, libDnsRecord)
	}

	return recs, nil
}

func (p *Provider) appendZoneRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		fmt.Printf("failed to parse server url: %v\n", err)
		return libdns.Record{}, err
	}

	rootZone, err := p.findRoot(ctx, zone)
	if err != nil {
		rootZone = zone
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "add")
	queryString.Set("json", "yes")
	queryString.Set("full_mx_records", "yes")
	queryString.Set("allow_dns_underscore", "yes")
	queryString.Set("domain", rootZone)
	queryString.Set("type", record.Type)
	queryString.Set("name", record.Name)
	queryString.Set("value", record.Value)

	if record.Type != "NS" {
		queryString.Set("ttl", strconv.Itoa(int(record.TTL.Seconds())))
	}

	reqURL.RawQuery = queryString.Encode()

	err = p.executeJsonRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		return libdns.Record{}, err
	}

	record.ID = fmt.Sprintf("name=%v&value=%v", record.Name, record.Value)

	return record, nil
}

func (p *Provider) setZoneRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		fmt.Printf("failed to parse server url: %v\n", err)
		return libdns.Record{}, err
	}

	rootZone, err := p.findRoot(ctx, zone)
	if err != nil {
		rootZone = zone
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "edit")
	queryString.Set("json", "yes")
	queryString.Set("domain", rootZone)
	queryString.Set("type", record.Type)
	queryString.Set("name", record.Name)
	queryString.Set("value", record.Value)

	if record.Type != "NS" {
		queryString.Set("ttl", strconv.Itoa(int(record.TTL.Seconds())))
	}

	existingRecords, _ := p.getZoneRecords(ctx, zone)
	var existingRecordIndex = -1
	for i := range existingRecords {
		if existingRecords[i].Name == record.Name && existingRecords[i].Type == record.Type {
			existingRecordIndex = i
			break
		}
	}

	// If we're not -1, we found a matching existing record. This changes the API call
	// from create only to edit.
	if existingRecordIndex != -1 {
		editKey := fmt.Sprintf("%vrecs0", strings.ToLower(record.Type))
		editValue := existingRecords[existingRecordIndex].ID
		queryString.Set(editKey, editValue)
	}

	reqURL.RawQuery = queryString.Encode()

	err = p.executeJsonRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		return libdns.Record{}, err
	}

	record.ID = fmt.Sprintf("name=%v&value=%v", record.Name, record.Value)

	return record, nil
}

func (p *Provider) deleteZoneRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		fmt.Printf("failed to parse server url: %v\n", err)
		return libdns.Record{}, err
	}

	rootZone, err := p.findRoot(ctx, zone)
	if err != nil {
		rootZone = zone
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "select")
	queryString.Set("json", "yes")
	queryString.Set("domain", rootZone)

	editKey := fmt.Sprintf("%vrecs0", strings.ToLower(record.Type))
	editValue := fmt.Sprintf("name=%v&value=%v", record.Name, record.Value)
	queryString.Set(editKey, editValue)

	reqURL.RawQuery = queryString.Encode()

	err = p.executeJsonRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}

func (p *Provider) findRoot(ctx context.Context, zone string) (string, error) {
	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		fmt.Printf("failed to parse server url: %v\n", err)
		return "", err
	}

	reqURL.Path = "/CMD_API_SHOW_DOMAINS"

	resp, err := p.executeQueryRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		return "", err
	}

	zoneParts := strings.Split(zone, ".")

	// Limit to 100 rounds
	for i := 0; i < 100; i++ {
		for _, value := range resp {
			if value == strings.Join(zoneParts, ".") {
				return value, nil
			}
		}

		zoneParts = zoneParts[1:]
	}

	return "", errors.New("root zone not found")
}

func (p *Provider) executeJsonRequest(ctx context.Context, method, requestUrl string) error {
	resp, err := p.doRequest(ctx, method, requestUrl)
	if err != nil {
		return err
	}

	var respData daResponse
	err = json.Unmarshal(resp, &respData)
	if err != nil {
		fmt.Printf("failed to json decode response: %v\n", err)
		return err
	}

	if len(respData.Error) > 0 {
		trimmedResult := strings.Split(respData.Result, "\n")[0]
		fmt.Printf("api response error: %v: %v\n", respData.Error, trimmedResult)
		return fmt.Errorf("api response error: %v: %v\n", respData.Error, trimmedResult)
	}

	return nil
}

func (p *Provider) executeQueryRequest(ctx context.Context, method, requestUrl string) ([]string, error) {
	resp, err := p.doRequest(ctx, method, requestUrl)
	if err != nil {
		return nil, err
	}

	params, err := url.ParseQuery(string(resp))
	if err != nil {
		return nil, err
	}

	var domains []string
	for _, param := range params {
		for _, domain := range param {
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

func (p *Provider) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		fmt.Printf("failed to build new request: %v\n", err)
		return nil, err
	}

	req.SetBasicAuth(p.User, p.LoginKey)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: p.InsecureRequests,
			},
		}}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("failed to execute request: %v\n", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read response body: %v\n", err)
			return nil, err
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)

		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}
