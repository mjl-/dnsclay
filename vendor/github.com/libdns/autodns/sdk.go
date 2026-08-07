package autodns

import (
	"context"
	"fmt"
	"net/http"
)

// checkZone servers as a check if the zone exists and returns the required data
// (origin and nameserver to retrieve the actual zone).
func (p *Provider) checkZone(ctx context.Context, zone string) (*ResponseSearchItem, error) {
	filter := map[string]string{
		"key":      "name",
		"operator": "EQUAL",
		"value":    zone,
	}

	payload := map[string][]map[string]string{}
	payload["filters"] = make([]map[string]string, 0)
	payload["filters"] = append(payload["filters"], filter)

	req, err := p.buildRequest(ctx, http.MethodPost, p.buildURL("zone/_search"), payload)
	if err != nil {
		return nil, err
	}

	resp, err := p.makeRequest(req)
	if err != nil {
		return nil, err
	}

	var result ResponseSearch

	if err := p.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("checkZone: %s", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("checkZone: %q not found", zone)
	}

	return &result.Data[0], nil
}

// getZone returns the zone.
func (p *Provider) getZone(ctx context.Context, origin, nameserver, zone string) (*ResponseZone, error) {
	req, err := p.buildRequest(ctx, http.MethodGet, p.buildURL("zone/"+origin+"/"+nameserver), RequestZone{
		Domain: zone,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.makeRequest(req)
	if err != nil {
		return nil, err
	}

	var result ResponseZone
	if err := p.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("getZone: %s", err)
	}

	if result.Status.Type == "ERROR" {
		if result.Messages == nil {
			return nil, fmt.Errorf("unknown error: %#v", result)
		}

		return nil, NewError(result.AutoDNSResponse)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("getZone: could not find %q", zone)
	}

	if len(result.Data) != 1 {
		return nil, fmt.Errorf("getZone: ambigous result for %q", zone)
	}

	return &result, nil
}

// updateZone updates the zone.
func (p *Provider) updateZone(ctx context.Context, origin, nameserver string, zone ZoneItem) error {
	req, err := p.buildRequest(ctx, http.MethodPut, p.buildURL("zone/"+origin+"/"+nameserver), zone)
	if err != nil {
		return err
	}

	resp, err := p.makeRequest(req)
	if err != nil {
		return err
	}

	var result AutoDNSResponse
	if err := p.parseResponse(resp, &result); err != nil {
		return fmt.Errorf("updateZone: %s", err)
	}

	if result.Status.Type == "ERROR" {
		return NewError(result)
	}
	return nil
}
