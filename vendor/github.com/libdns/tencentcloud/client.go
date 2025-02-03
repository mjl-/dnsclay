package tencentcloud

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/libdns/tencentcloud/dnspod"
)

// getClient gets the client for Tencent Cloud DNS
func (p *Provider) getClient() (*dnspod.Client, error) {
	client := sync.OnceValues(func() (*dnspod.Client, error) {
		credential := common.NewCredential(
			p.SecretId,
			p.SecretKey,
		)
		cpf := tp.NewClientProfile()
		cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
		client, err := dnspod.NewClient(credential, "", cpf)
		if err != nil {
			return nil, err
		}
		return client, nil
	})
	return client()
}

// describeRecordList describes the records for a zone
func (p *Provider) describeRecordList(ctx context.Context, zone string) ([]libdns.Record, error) {
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	var list []libdns.Record
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(strings.Trim(zone, "."))
	request.Offset = common.Uint64Ptr(0)
	request.Limit = common.Uint64Ptr(3000)

	totalCount := uint64(100)
	for *request.Offset < totalCount {
		response, err := client.DescribeRecordListWithContext(ctx, request)
		if err != nil {
			return nil, err
		}
		if response.Response.RecordList != nil && len(response.Response.RecordList) > 0 {
			for _, record := range response.Response.RecordList {
				list = append(list, libdns.Record{
					ID:    strconv.Itoa(int(*record.RecordId)),
					Type:  *record.Type,
					Name:  *record.Name,
					Value: *record.Value,
					TTL:   time.Duration(*record.TTL) * time.Second,
				})
			}
		}
		totalCount = *response.Response.RecordCountInfo.TotalCount
		request.Offset = common.Uint64Ptr(*request.Offset + uint64(len(response.Response.RecordList)))
	}
	return list, err
}

// createRecord creates a record for a zone
func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (string, error) {
	client, err := p.getClient()
	if err != nil {
		return "", err
	}

	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(strings.Trim(zone, "."))
	request.SubDomain = common.StringPtr(record.Name)
	request.RecordType = common.StringPtr(record.Type)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(record.Value)
	response, err := client.CreateRecordWithContext(ctx, request)

	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(*response.Response.RecordId)), nil
}

// modifyRecord modifies a record for a zone
func (p *Provider) modifyRecord(ctx context.Context, zone string, record libdns.Record) error {
	client, err := p.getClient()
	if err != nil {
		return err
	}

	recordId, _ := strconv.Atoi(record.ID)
	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(strings.Trim(zone, "."))
	request.SubDomain = common.StringPtr(record.Name)
	request.RecordType = common.StringPtr(record.Type)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(record.Value)
	request.RecordId = common.Uint64Ptr(uint64(recordId))

	_, err = client.ModifyRecordWithContext(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

// deleteRecord deletes a record for a zone
func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) error {
	client, err := p.getClient()
	if err != nil {
		return err
	}

	recordId, _ := strconv.Atoi(record.ID)
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(strings.Trim(zone, "."))
	request.RecordId = common.Uint64Ptr(uint64(recordId))

	_, err = client.DeleteRecordWithContext(ctx, request)
	if err != nil {
		return err
	}
	return nil
}
