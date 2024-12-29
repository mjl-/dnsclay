package porkbun

import (
	"bytes"
	"context"
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
	"time"
)

const ApiBase = "https://api.porkbun.com/api/json/v3"

// LibdnsZoneToPorkbunDomain Strips the trailing dot from a Zone
func LibdnsZoneToPorkbunDomain(zone string) string {
	return strings.TrimSuffix(zone, ".")
}

// CheckCredentials allows verifying credentials work in test scripts
func (p *Provider) CheckCredentials(_ context.Context) (string, error) {
	credentialJson, err := json.Marshal(p.getCredentials())
	if err != nil {
		return "", err
	}

	response, err := MakeApiRequest("/ping", bytes.NewReader(credentialJson), pkbnPingResponse{})

	if err != nil {
		return "", err
	}

	if response.Status != "SUCCESS" {
		return "", err
	}

	return response.YourIP, nil
}

func (p *Provider) getCredentials() ApiCredentials {
	return ApiCredentials{p.APIKey, p.APISecretKey}
}

func (p *Provider) getMatchingRecord(r libdns.Record, zone string) ([]libdns.Record, error) {
	var recs []libdns.Record
	trimmedZone := LibdnsZoneToPorkbunDomain(zone)

	credentialJson, err := json.Marshal(p.getCredentials())
	if err != nil {
		return recs, err
	}

	relativeName := libdns.RelativeName(r.Name, zone)
	trimmedName := relativeName
	if relativeName == "@" {
		trimmedName = ""
	}

	endpoint := fmt.Sprintf("/dns/retrieveByNameType/%s/%s/%s", trimmedZone, r.Type, trimmedName)
	response, err := MakeApiRequest(endpoint, bytes.NewReader(credentialJson), pkbnRecordsResponse{})

	if err != nil {
		return recs, err
	}

	recs = make([]libdns.Record, 0, len(response.Records))
	for _, rec := range response.Records {
		recs = append(recs, rec.toLibdnsRecord(zone))
	}
	return recs, nil
}

// UpdateRecords adds records to the zone. It returns the records that were added.
func (p *Provider) updateRecords(_ context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	credentials := p.getCredentials()
	trimmedZone := LibdnsZoneToPorkbunDomain(zone)

	var createdRecords []libdns.Record

	for _, record := range records {
		if record.TTL/time.Second < 600 {
			record.TTL = 600 * time.Second
		}
		ttlInSeconds := int(record.TTL / time.Second)
		relativeName := libdns.RelativeName(record.Name, zone)
		trimmedName := relativeName
		if relativeName == "@" {
			trimmedName = ""
		}

		reqBody := pkbnRecordPayload{&credentials, record.Value, trimmedName, strconv.Itoa(ttlInSeconds), record.Type}
		reqJson, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		response, err := MakeApiRequest(fmt.Sprintf("/dns/edit/%s/%s", trimmedZone, record.ID), bytes.NewReader(reqJson), pkbnResponseStatus{})
		if err != nil {
			return nil, err
		}

		if response.Status != "SUCCESS" {
			return nil, err
		}
		createdRecords = append(createdRecords, record)
	}

	return createdRecords, nil
}

func MakeApiRequest[T any](endpoint string, body io.Reader, responseType T) (T, error) {
	client := http.Client{}

	fullUrl := ApiBase + endpoint
	u, err := url.Parse(fullUrl)
	if err != nil {
		return responseType, err
	}

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return responseType, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return responseType, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("Couldn't close body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = errors.New("Invalid http response status, " + string(bodyBytes))
		return responseType, err
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return responseType, err
	}

	err = json.Unmarshal(result, &responseType)

	if err != nil {
		return responseType, err
	}

	return responseType, nil
}
