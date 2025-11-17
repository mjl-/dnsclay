package directadmin

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/libdns/libdns"
	"go.uber.org/zap"
)

func (p *Provider) getZoneRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		p.getLogger().Error("Failed to parse server URL", zap.Error(err))
		return nil, err
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("json", "yes")
	queryString.Set("full_mx_records", "yes")
	queryString.Set("allow_dns_underscore", "yes")
	queryString.Set("ttl", "yes")
	queryString.Set("domain", zone)

	reqURL.RawQuery = queryString.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		p.getLogger().Error("Failed to build new request", zap.Error(err))
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
		p.getLogger().Error("Failed to execute request", zap.Error(err))
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			p.getLogger().Error("Failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			p.getLogger().Error("Failed to read response body", zap.Error(err))
			return nil, err
		}

		bodyString := string(bodyBytes)

		p.getLogger().Error("API returned a non-200 status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", bodyString))

		return nil, fmt.Errorf("api request failed with status code %d", resp.StatusCode)
	}

	var respData daZone
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		p.getLogger().Error("Failed to decode JSON response", zap.Error(err))
		return nil, err
	}

	recs := make([]libdns.Record, 0, len(respData.Records))
	for i := range respData.Records {
		libDnsRecord, err := respData.Records[i].libdnsRecord(zone)
		if err != nil {
			switch err {
			case ErrUnsupported:
				p.getLogger().Warn("Unsupported record conversion",
					zap.String("type", respData.Records[i].Type),
					zap.String("name", respData.Records[i].Name))
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
		p.getLogger().Error("Failed to parse server URL", zap.Error(err))
		return nil, err
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "add")
	queryString.Set("json", "yes")
	queryString.Set("full_mx_records", "yes")
	queryString.Set("allow_dns_underscore", "yes")
	queryString.Set("domain", zone)

	rr := record.RR()
	queryString.Set("type", rr.Type)
	queryString.Set("name", rr.Name)
	queryString.Set("value", rr.Data)

	if rr.Type != "NS" {
		queryString.Set("ttl", strconv.Itoa(int(rr.TTL.Seconds())))
	}

	reqURL.RawQuery = queryString.Encode()

	err = p.executeRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		p.getLogger().Error("DirectAdmin create record request failed", zap.Error(err))
		return nil, err
	}

	return &rr, nil
}

func (p *Provider) setZoneRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		p.getLogger().Error("Failed to parse server URL", zap.Error(err))
		return nil, err
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "edit")
	queryString.Set("json", "yes")
	queryString.Set("domain", zone)

	rr := record.RR()
	queryString.Set("type", rr.Type)
	queryString.Set("name", rr.Name)
	queryString.Set("value", rr.Data)

	if rr.Type != "NS" {
		queryString.Set("ttl", strconv.Itoa(int(rr.TTL.Seconds())))
	}

	existingRecords, err := p.getZoneRecords(ctx, zone)
	if err != nil {
		p.getLogger().Error("Could not get existing records while setting record", zap.Error(err))
		return nil, err
	}
	var existingRecordIndex = -1
	for i := range existingRecords {
		existingRR := existingRecords[i].RR()
		if existingRR.Name == rr.Name && existingRR.Type == rr.Type {
			existingRecordIndex = i
			break
		}
	}

	// If we're not -1, we found a matching existing record. This changes the API call
	// from create only to edit.
	if existingRecordIndex != -1 {
		editKey := fmt.Sprintf("%vrecs0", strings.ToLower(rr.Type))
		editValue := existingRecords[existingRecordIndex].RR().Data
		queryString.Set(editKey, editValue)
	}

	reqURL.RawQuery = queryString.Encode()

	err = p.executeRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		p.getLogger().Error("DirectAdmin set record request failed", zap.Error(err))
		return nil, err
	}

	return &rr, nil
}

func (p *Provider) deleteZoneRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		p.getLogger().Error("Failed to parse server URL", zap.Error(err))
		return nil, err
	}

	reqURL.Path = "/CMD_API_DNS_CONTROL"

	queryString := make(url.Values)
	queryString.Set("action", "select")
	queryString.Set("json", "yes")
	queryString.Set("domain", zone)

	rr := record.RR()
	editKey := fmt.Sprintf("%vrecs0", strings.ToLower(rr.Type))
	editValue := fmt.Sprintf("name=%v&value=%v", rr.Name, rr.Data)
	queryString.Set(editKey, editValue)

	reqURL.RawQuery = queryString.Encode()

	err = p.executeRequest(ctx, http.MethodGet, reqURL.String())
	if err != nil {
		p.getLogger().Error("DirectAdmin delete record request failed", zap.Error(err))
		return nil, err
	}

	return record, nil
}

func (p *Provider) executeRequest(ctx context.Context, method, url string) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		p.getLogger().Error("Failed to build new request", zap.Error(err))
		return err
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
		p.getLogger().Error("Failed to execute request", zap.Error(err))
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			p.getLogger().Error("Failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	var respData daResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		p.getLogger().Error("Failed to decode JSON response", zap.Error(err))
		return err
	}

	if len(respData.Error) > 0 {
		trimmedResult := strings.Split(respData.Result, "\n")[0]
		p.getLogger().Error("API response error",
			zap.String("error", respData.Error),
			zap.String("result", trimmedResult))
		return fmt.Errorf("api error: %s, result: %s", respData.Error, trimmedResult)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			p.getLogger().Error("Failed to read response body", zap.Error(err))
			return err
		}

		bodyString := string(bodyBytes)

		p.getLogger().Error("API returned a non-200 status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", bodyString))

		return fmt.Errorf("api request failed with status code %d", resp.StatusCode)
	}

	return nil
}

func (p *Provider) getDomainList(ctx context.Context) ([]string, error) {
	reqURL, err := url.Parse(p.ServerURL)
	if err != nil {
		p.getLogger().Error("Failed to parse server URL", zap.Error(err))
		return nil, err
	}

	reqURL.Path = "/CMD_API_SHOW_DOMAINS"

	queryString := make(url.Values)
	queryString.Set("json", "yes")

	reqURL.RawQuery = queryString.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		p.getLogger().Error("Failed to build new request", zap.Error(err))
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
		p.getLogger().Error("Failed to execute request", zap.Error(err))
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			p.getLogger().Error("Failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			p.getLogger().Error("Failed to read response body", zap.Error(err))
			return nil, err
		}

		bodyString := string(bodyBytes)

		p.getLogger().Error("API returned a non-200 status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("body", bodyString))

		return nil, fmt.Errorf("api request failed with status code %d", resp.StatusCode)
	}

	var respData daDomainList
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		p.getLogger().Error("Failed to decode JSON response", zap.Error(err))
		return nil, err
	}

	return respData, nil
}

func (p *Provider) findManageableZone(ctx context.Context, requestedZone string) (string, error) {
	p.getLogger().Debug("findManageableZone called", zap.String("zone", requestedZone))

	// Remove trailing dot if present
	requestedZone = strings.TrimSuffix(requestedZone, ".")

	// Get list of domains we can manage
	domains, err := p.getDomainList(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get domain list: %v", err)
	}

	p.getLogger().Debug("Available domains", zap.Strings("domains", domains))

	// Try the requested zone first (exact match)
	for _, domain := range domains {
		if strings.EqualFold(requestedZone, domain) {
			p.getLogger().Debug("Found exact match", zap.String("domain", domain))
			return domain, nil
		}
	}

	// If no exact match, traverse backwards through the FQDN to find parent zones
	parts := strings.Split(requestedZone, ".")
	for i := 1; i < len(parts); i++ {
		candidateZone := strings.Join(parts[i:], ".")
		p.getLogger().Debug("Checking candidate zone", zap.String("candidate_zone", candidateZone))

		for _, domain := range domains {
			if strings.EqualFold(candidateZone, domain) {
				p.getLogger().Debug("Found manageable parent zone",
					zap.String("parent_zone", domain),
					zap.String("requested_zone", requestedZone))
				return domain, nil
			}
		}
	}

	return "", fmt.Errorf("no manageable zone found for %s in available domains: %v", requestedZone, domains)
}
