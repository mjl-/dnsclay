// Package hosttech implements methods for manipulating Hosttech.ch DNS records with the libdns interfaces.
// Manipulation is achieved with the Hosttech API at https://api.ns1.hosttech.eu/api/documentation/#/.
package hosttech

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Hosttech.ch.
type Provider struct {
	APIToken string `json:"api_token,omitempty"`
}

// The URL for the Hosttech API connection
const apiHost = "https://api.ns1.hosttech.eu/api/user/v1"

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	reqURL := fmt.Sprintf("%s/zones/%s/records", apiHost, RemoveTrailingDot(zone))

	responseBody, err := p.makeApiCall(ctx, http.MethodGet, reqURL, nil)

	//If there's an error return an empty slice
	if err != nil {
		return []libdns.Record{}, err
	}

	var parsedResponse = HosttechListResponseWrapper{}
	err = json.Unmarshal(responseBody, &parsedResponse)

	if err != nil {
		return []libdns.Record{}, err
	}

	var libdnsRecords []libdns.Record
	for _, record := range parsedResponse.Data {
		libdnsRecords = append(libdnsRecords, record.toLibdnsRecord(zone))
	}

	return libdnsRecords, nil
}

// AppendRecords adds records to the zone. It returns all records that were added.
// If an error occurs while records are being added, the already successfully added records will be returned along with an error.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	reqURL := fmt.Sprintf("%s/zones/%s/records", apiHost, RemoveTrailingDot(zone))

	successfullyAppendedRecords := []libdns.Record{}
	for _, record := range records {

		hosttechRecord, err := LibdnsRecordToHosttechRecordWrapper(record)
		if err != nil {
			return nil, err
		}

		bodyBytes, err := json.Marshal(hosttechRecord)
		if err != nil {
			return nil, err
		}

		responseBody, err := p.makeApiCall(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return successfullyAppendedRecords, err
		}

		var parsedResponse = HosttechSingleResponseWrapper{}
		err = json.Unmarshal(responseBody, &parsedResponse)

		if err != nil {
			return []libdns.Record{}, err
		}

		successfullyAppendedRecords = append(successfullyAppendedRecords, parsedResponse.Data.toLibdnsRecord(zone))
	}

	return successfullyAppendedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	successfullyUpdatedRecords := []libdns.Record{}
	for _, record := range records {

		hosttechRecord, err := LibdnsRecordToHosttechRecordWrapper(record)
		if err != nil {
			return nil, err
		}
		bodyBytes, err := json.Marshal(hosttechRecord)

		if err != nil {
			return nil, err
		}

		reqURL := fmt.Sprintf("%s/zones/%s/records/%s", apiHost, RemoveTrailingDot(zone), record.ID)

		responseBody, err := p.makeApiCall(ctx, http.MethodPut, reqURL, bytes.NewReader(bodyBytes))

		if err != nil {
			//If the error doesn't have anything to do with the api, return
			apiError, ok := err.(ApiError)
			if !ok {
				return successfullyUpdatedRecords, err
			}

			//If the error isn't a 404, return
			if apiError.ErrorCode != 404 {
				return successfullyUpdatedRecords, err
			}

			//If the error was a 404, the record could not be updated because it didn't exist. So we create a new one
			appendedRecords, err := p.AppendRecords(ctx, zone, []libdns.Record{record})
			if err != nil {
				return successfullyUpdatedRecords, err
			}

			successfullyUpdatedRecords = append(successfullyUpdatedRecords, appendedRecords...)
			continue
		}

		var parsedResponse = HosttechSingleResponseWrapper{}
		err = json.Unmarshal(responseBody, &parsedResponse)
		if err != nil {
			return []libdns.Record{}, err
		}

		successfullyUpdatedRecords = append(successfullyUpdatedRecords, parsedResponse.Data.toLibdnsRecord(zone))
	}

	return successfullyUpdatedRecords, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
// If an error occurs while records are being deleted, the already successfully deleted records will be returned along with an error.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	successfullyDeletedRecords := []libdns.Record{}
	for _, record := range records {
		reqUrl := fmt.Sprintf("%s/zones/%s/records/%s", apiHost, RemoveTrailingDot(zone), record.ID)
		_, err := p.makeApiCall(ctx, http.MethodDelete, reqUrl, nil)

		if err != nil {
			return successfullyDeletedRecords, err
		}

		successfullyDeletedRecords = append(successfullyDeletedRecords, record)
	}

	return successfullyDeletedRecords, nil
}

// List all available zones
func (p *Provider) ListZones(ctx context.Context) ([]libdns.Zone, error) {
	reqUrl := fmt.Sprintf("%s/zones", apiHost)
	responseBody, err := p.makeApiCall(ctx, http.MethodGet, reqUrl, nil)

	if err != nil {
		return nil, err
	}

	var parsedResponse = HosttechZoneListResponseWrapper{}
	err = json.Unmarshal(responseBody, &parsedResponse)

	if err != nil {
		return nil, err
	}

	var libdnsZones []libdns.Zone
	for _, zone := range parsedResponse.Data {
		libdnsZones = append(libdnsZones, zone.toLibdnsZone())
	}

	return libdnsZones, nil
}

func (p *Provider) makeApiCall(ctx context.Context, httpMethod string, reqUrl string, body io.Reader) (response []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, reqUrl, body)
	req.Header.Set("Authorization", "Bearer "+p.APIToken)
	req.Header.Set("Content-Type", "application/json")

	//Return nil if there's an error
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)

	//Return an empty slice if there's an error
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ApiError{
			s:         fmt.Sprintf("call to API was not successful, returned the status code '%s'", resp.Status),
			ErrorCode: resp.StatusCode,
		}
	}

	return io.ReadAll(resp.Body)
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
	_ libdns.ZoneLister     = (*Provider)(nil)
)
