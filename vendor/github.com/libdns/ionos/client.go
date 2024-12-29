// libdns client for IONOS DNS API
package ionos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	APIEndpoint = "https://api.hosting.ionos.com/dns/v1"
)

type getAllZonesResponse struct {
	Zones []zoneDescriptor
}

type zoneDescriptor struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

type getZoneResponse struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Type    string       `json:"type"`
	Records []zoneRecord `json:"records"`
}

type zoneRecord struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RootName   string `json:"rootName"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	ChangeDate string `json:"changeDate"`
	TTL        int    `json:"ttl"`
	Prio       int    `json:"prio"`
	Disabled   bool   `json:"disabled"`
}

type record struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	TTL      *int   `json:"ttl,omitempty"`
	Prio     int    `json:"prio"`
	Disabled bool   `json:"disabled,omitempty"` // TODO default=true
}

// IONOS does not accept TTL values < 60, and returns status 400. If the
// TTL is 0, we leave the field empty, by setting the struct value to nil.
func ionosTTL(ttl float64) *int {
	var intTTL *int
	if ttl > 0 {
		tmp := int(ttl)
		intTTL = &tmp
	}
	return intTTL
}

func doRequest(token string, request *http.Request) ([]byte, error) {
	request.Header.Add("X-API-Key", token)
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{} // no timeout set because request is w/ context
	//	fmt.Printf(">>> HTTP req: %+v\n\n", request)
	response, err := client.Do(request)

	//	fmt.Printf("<<< HTTP res: %+v\n\n", response)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s (%d)", http.StatusText(response.StatusCode), response.StatusCode)
	}
	return io.ReadAll(response.Body)

}

// GET /v1/zones
func ionosGetAllZones(ctx context.Context, token string) (getAllZonesResponse, error) {
	uri := fmt.Sprintf("%s/zones", APIEndpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	data, err := doRequest(token, req)

	if err != nil {
		return getAllZonesResponse{}, err
	}

	// parse top-level JSON array
	zones := make([]zoneDescriptor, 0)
	err = json.Unmarshal(data, &zones)
	return getAllZonesResponse{zones}, err
}

// ionosGetZone reads the contents of zone by it's IONOS zoneID, optionally filtering for
// a specific recordType and recordName.
// /v1/zones/{zoneId}
func ionosGetZone(ctx context.Context, token string, zoneID string, recordType, recordName string) (getZoneResponse, error) {
	u, err := url.Parse(APIEndpoint)
	if err != nil {
		return getZoneResponse{}, err
	}
	u = u.JoinPath("zones", zoneID)
	queryString := u.Query()
	if recordType != "" {
		queryString.Set("recordType", recordType)
	}
	if recordName != "" {
		queryString.Set("recordName", recordName)
	}
	u.RawQuery = queryString.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	data, err := doRequest(token, req)
	var result getZoneResponse
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(data, &result)
	return result, err
}

// ionosFindRecordsInZone is a convenience function to search all records in the
// given zone for a record with the given name and type and returns this record
// on success
func ionosFindRecordsInZone(ctx context.Context, token string, zoneID, typ, name string) ([]zoneRecord, error) {
	resp, err := ionosGetZone(ctx, token, zoneID, typ, name)
	if err != nil {
		return nil, err
	}
	if len(resp.Records) < 1 {
		return nil, fmt.Errorf("record not found in zone")
	}
	return resp.Records, nil
}

// ionosDeleteRecord deletes the given record
// DELETE /v1/zones/{zoneId}/records/{recordId}
func ionosDeleteRecord(ctx context.Context, token string, zoneID, id string) error {

	if id == "" {
		return fmt.Errorf("no record id provided")
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("%s/zones/%s/records/%s", APIEndpoint, zoneID, id), nil)

	if err != nil {
		return err
	}
	_, err = doRequest(token, req)
	return err
}

// ionosCreateRecord creates a batch of DNS record in the given zone
// POST /v1/zones/{zoneId}/records
func ionosCreateRecords(
	ctx context.Context,
	token string,
	zoneID string,
	records []record) ([]zoneRecord, error) {

	reqBuffer, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s/zones/%s/records", APIEndpoint, zoneID)
	req, err := http.NewRequestWithContext(ctx, "POST", uri, bytes.NewBuffer(reqBuffer))
	if err != nil {
		return nil, err
	}

	// as result of the POST, a zoneRecord array is returned
	res, err := doRequest(token, req)
	if err != nil {
		return nil, err
	}

	var zoneRecords []zoneRecord
	if err = json.Unmarshal(res, &zoneRecords); err != nil {
		return nil, err
	}
	return zoneRecords, nil

}

// ionosUpdateRecord updates the record with id `id` in the given zone
// TODO check TTL
// PUT /v1/zones/{zoneId}/records/{recordId}
func ionosUpdateRecord(ctx context.Context, token string, zoneID, id string, r record) error {
	if id == "" {
		return fmt.Errorf("no record id provided")
	}

	reqBuffer, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal record for update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT",
		fmt.Sprintf("%s/zones/%s/records/%s", APIEndpoint, zoneID, id),
		bytes.NewBuffer(reqBuffer))

	if err != nil {
		return err
	}

	// according to API doc, no response returned here
	_, err = doRequest(token, req)
	return err
}
