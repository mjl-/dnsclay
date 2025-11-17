package tencentcloud

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

const (
	endpoint = "https://dnspod.tencentcloudapi.com"

	DescribeRecordList = "DescribeRecordList"
	CreateRecord       = "CreateRecord"
	ModifyRecord       = "ModifyRecord"
	DeleteRecord       = "DeleteRecord"
)

func (p *Provider) listRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	domain := strings.TrimSuffix(zone, ".")

	requestData := FindRecordRequest{
		Domain:     domain,
		RecordLine: "默认",
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	resp, err := p.sendRequest(ctx, DescribeRecordList, string(payload))
	if err != nil {
		return nil, err
	}

	var response Response
	if err = json.Unmarshal(resp, &response); err != nil {
		return nil, err
	}

	list := make([]libdns.Record, 0, len(response.Response.RecordList))
	for _, txRecord := range response.Response.RecordList {
		rr := record{
			Type:  txRecord.Type,
			Name:  txRecord.Name,
			Value: txRecord.Value,
			TTL:   time.Duration(txRecord.TTL) * time.Second,
			MX:    txRecord.MX,
		}
		libdnsRecord, err := rr.libdnsRecord()
		if err != nil {
			return nil, err
		}
		list = append(list, libdnsRecord)
	}

	return list, nil
}

func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) error {
	domain := strings.TrimSuffix(zone, ".")
	r := fromLibdnsRecord(record)
	requestData := CreateModifyRecordRequest{
		Domain:     domain,
		SubDomain:  r.Name,
		RecordType: r.Type,
		RecordLine: "默认",
		Value:      r.Value,
		TTL:        int64(r.TTL.Seconds()),
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	resp, err := p.sendRequest(ctx, CreateRecord, string(payload))
	if err != nil {
		return err
	}

	var response Response
	if err := json.Unmarshal(resp, &response); err != nil {
		return err
	}

	if response.Response.RecordId == 0 {
		return ErrNotValid
	}

	return nil
}

func (p *Provider) modifyRecord(ctx context.Context, id uint64, zone string, record libdns.Record) error {
	domain := strings.TrimSuffix(zone, ".")
	r := fromLibdnsRecord(record)
	requestData := CreateModifyRecordRequest{
		Domain:     domain,
		SubDomain:  r.Name,
		RecordType: r.Type,
		RecordLine: "默认",
		Value:      r.Value,
		TTL:        int64(r.TTL.Seconds()),
		RecordId:   id,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	_, err = p.sendRequest(ctx, ModifyRecord, string(payload))
	return err
}

func (p *Provider) deleteRecord(ctx context.Context, id uint64, zone string) error {
	domain := strings.TrimSuffix(zone, ".")

	requestData := DeleteRecordRequest{
		Domain:   domain,
		RecordId: id,
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	_, err = p.sendRequest(ctx, DeleteRecord, string(payload))
	return err
}

func (p *Provider) findRecord(ctx context.Context, zone string, record libdns.Record) (uint64, error) {
	domain := strings.TrimSuffix(zone, ".")
	r := fromLibdnsRecord(record)
	requestData := FindRecordRequest{
		Domain:     domain,
		RecordType: r.Type,
		RecordLine: "默认",
		Subdomain:  r.Name,
		Limit:      3000,
	}
	payload, err := json.Marshal(requestData)
	if err != nil {
		return 0, err
	}

	resp, err := p.sendRequest(ctx, DescribeRecordList, string(payload))
	if err != nil {
		return 0, err
	}

	var response Response
	if err = json.Unmarshal(resp, &response); err != nil {
		return 0, err
	}
	var recordId uint64
	for _, item := range response.Response.RecordList {
		if item.Name == r.Name && item.Type == r.Type {
			if r.Value != "" && item.Value != r.Value {
				continue
			}
			recordId = uint64(item.RecordId)
			break
		}
	}

	if recordId == 0 {
		return 0, ErrRecordNotFound
	}

	return recordId, nil
}

func (p *Provider) sendRequest(ctx context.Context, action string, data string) ([]byte, error) {
	endpointUrl := endpoint
	if p.Region != "" {
		endpointUrl = "https://dnspod." + p.Region + ".tencentcloudapi.com"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpointUrl, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Version", "2021-03-23")

	SignRequest(p.SecretId, p.SecretKey, p.SessionToken, req, action, data)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
