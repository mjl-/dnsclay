package v2

import (
	httpclient "github.com/libdns/huaweicloud/sdk/core"
	"github.com/libdns/huaweicloud/sdk/services/dns/v2/model"
)

type DnsClient struct {
	HcClient *httpclient.HcHttpClient
}

func NewDnsClient(hcClient *httpclient.HcHttpClient) *DnsClient {
	return &DnsClient{HcClient: hcClient}
}

func DnsClientBuilder() *httpclient.HcHttpClientBuilder {
	builder := httpclient.NewHcHttpClientBuilder()
	return builder
}

// ListPublicZones 查询公网Zone列表
//
// 查询公网Zone列表
//
// Please refer to HUAWEI cloud API Explorer for details.
func (c *DnsClient) ListPublicZones(request *model.ListPublicZonesRequest) (*model.ListPublicZonesResponse, error) {
	requestDef := GenReqDefForListPublicZones()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListPublicZonesResponse), nil
	}
}

// ListRecordSetsByZone 查询单个Zone下Record Set列表
//
// 查询单个Zone下Record Set列表
//
// Please refer to HUAWEI cloud API Explorer for details.
func (c *DnsClient) ListRecordSetsByZone(request *model.ListRecordSetsByZoneRequest) (*model.ListRecordSetsByZoneResponse, error) {
	requestDef := GenReqDefForListRecordSetsByZone()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListRecordSetsByZoneResponse), nil
	}
}

// CreateRecordSet 创建单个Record Set
//
// 创建单个Record Set
//
// Please refer to HUAWEI cloud API Explorer for details.
func (c *DnsClient) CreateRecordSet(request *model.CreateRecordSetRequest) (*model.CreateRecordSetResponse, error) {
	requestDef := GenReqDefForCreateRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateRecordSetResponse), nil
	}
}

// UpdateRecordSet 修改单个Record Set
//
// 修改单个Record Set
//
// Please refer to HUAWEI cloud API Explorer for details.
func (c *DnsClient) UpdateRecordSet(request *model.UpdateRecordSetRequest) (*model.UpdateRecordSetResponse, error) {
	requestDef := GenReqDefForUpdateRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateRecordSetResponse), nil
	}
}

// DeleteRecordSet 删除单个Record Set
//
// 删除单个Record Set。删除有添加智能解析的记录集时，需要用Record Set多线路管理模块中删除接口进行删除。
//
// Please refer to HUAWEI cloud API Explorer for details.
func (c *DnsClient) DeleteRecordSet(request *model.DeleteRecordSetRequest) (*model.DeleteRecordSetResponse, error) {
	requestDef := GenReqDefForDeleteRecordSet()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteRecordSetResponse), nil
	}
}
