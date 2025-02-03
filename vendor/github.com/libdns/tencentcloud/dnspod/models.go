// Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dnspod

import (
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/json"
)

type CreateRecordRequest struct {
	*tchttp.BaseRequest

	// 域名
	Domain *string `json:"Domain,omitnil,omitempty" name:"Domain"`

	// 记录类型，通过 API 记录类型获得，大写英文，比如：A 。
	RecordType *string `json:"RecordType,omitnil,omitempty" name:"RecordType"`

	// 记录线路，通过 API 记录线路获得，中文，比如：默认。
	RecordLine *string `json:"RecordLine,omitnil,omitempty" name:"RecordLine"`

	// 记录值，如 IP : 200.200.200.200， CNAME : cname.dnspod.com.， MX : mail.dnspod.com.。
	Value *string `json:"Value,omitnil,omitempty" name:"Value"`

	// 域名 ID 。参数 DomainId 优先级比参数 Domain 高，如果传递参数 DomainId 将忽略参数 Domain 。
	DomainId *uint64 `json:"DomainId,omitnil,omitempty" name:"DomainId"`

	// 主机记录，如 www，如果不传，默认为 @。
	SubDomain *string `json:"SubDomain,omitnil,omitempty" name:"SubDomain"`

	// 线路的 ID，通过 API 记录线路获得，英文字符串，比如：10=1。参数RecordLineId优先级高于RecordLine，如果同时传递二者，优先使用RecordLineId参数。
	RecordLineId *string `json:"RecordLineId,omitnil,omitempty" name:"RecordLineId"`

	// MX 优先级，当记录类型是 MX 时有效，范围1-20，MX 记录时必选。
	MX *uint64 `json:"MX,omitnil,omitempty" name:"MX"`

	// TTL，范围1-604800，不同套餐域名最小值不同。
	TTL *uint64 `json:"TTL,omitnil,omitempty" name:"TTL"`

	// 权重信息，0到100的整数。0 表示关闭，不传该参数，表示不设置权重信息。
	Weight *uint64 `json:"Weight,omitnil,omitempty" name:"Weight"`

	// 记录初始状态，取值范围为 ENABLE 和 DISABLE 。默认为 ENABLE ，如果传入 DISABLE，解析不会生效，也不会验证负载均衡的限制。
	Status *string `json:"Status,omitnil,omitempty" name:"Status"`

	// 备注
	Remark *string `json:"Remark,omitnil,omitempty" name:"Remark"`

	// 开启DNSSEC时，强制添加CNAME/URL记录
	DnssecConflictMode *string `json:"DnssecConflictMode,omitnil,omitempty" name:"DnssecConflictMode"`

	// 记录分组 Id。可以通过接口 DescribeRecordGroupList 接口 GroupId 字段获取。
	GroupId *uint64 `json:"GroupId,omitnil,omitempty" name:"GroupId"`
}

func (r *CreateRecordRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateRecordRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Domain")
	delete(f, "RecordType")
	delete(f, "RecordLine")
	delete(f, "Value")
	delete(f, "DomainId")
	delete(f, "SubDomain")
	delete(f, "RecordLineId")
	delete(f, "MX")
	delete(f, "TTL")
	delete(f, "Weight")
	delete(f, "Status")
	delete(f, "Remark")
	delete(f, "DnssecConflictMode")
	delete(f, "GroupId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "CreateRecordRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type CreateRecordResponseParams struct {
	// 记录ID
	RecordId *uint64 `json:"RecordId,omitnil,omitempty" name:"RecordId"`

	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

type CreateRecordResponse struct {
	*tchttp.BaseResponse
	Response *CreateRecordResponseParams `json:"Response"`
}

func (r *CreateRecordResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *CreateRecordResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRecordListRequest struct {
	*tchttp.BaseRequest

	// 要获取的解析记录所属的域名
	Domain *string `json:"Domain,omitnil,omitempty" name:"Domain"`

	// 要获取的解析记录所属的域名Id，如果传了DomainId，系统将会忽略Domain参数。 可以通过接口DescribeDomainList查到所有的Domain以及DomainId
	DomainId *uint64 `json:"DomainId,omitnil,omitempty" name:"DomainId"`

	// 解析记录的主机头，如果传了此参数，则只会返回此主机头对应的解析记录
	Subdomain *string `json:"Subdomain,omitnil,omitempty" name:"Subdomain"`

	// 获取某种类型的解析记录，如 A，CNAME，NS，AAAA，显性URL，隐性URL，CAA，SPF等
	RecordType *string `json:"RecordType,omitnil,omitempty" name:"RecordType"`

	// 获取某条线路名称的解析记录。可以通过接口DescribeRecordLineList查看当前域名允许的线路信息
	RecordLine *string `json:"RecordLine,omitnil,omitempty" name:"RecordLine"`

	// 获取某个线路Id对应的解析记录，如果传RecordLineId，系统会忽略RecordLine参数。可以通过接口DescribeRecordLineList查看当前域名允许的线路信息
	RecordLineId *string `json:"RecordLineId,omitnil,omitempty" name:"RecordLineId"`

	// 获取某个分组下的解析记录时，传这个分组Id。
	GroupId *uint64 `json:"GroupId,omitnil,omitempty" name:"GroupId"`

	// 通过关键字搜索解析记录，当前支持搜索主机头和记录值
	Keyword *string `json:"Keyword,omitnil,omitempty" name:"Keyword"`

	// 排序字段，支持 name,line,type,value,weight,mx,ttl,updated_on 几个字段。
	SortField *string `json:"SortField,omitnil,omitempty" name:"SortField"`

	// 排序方式，正序：ASC，逆序：DESC。默认值为ASC。
	SortType *string `json:"SortType,omitnil,omitempty" name:"SortType"`

	// 偏移量，默认值为0。
	Offset *uint64 `json:"Offset,omitnil,omitempty" name:"Offset"`

	// 限制数量，当前Limit最大支持3000。默认值为100。
	Limit *uint64 `json:"Limit,omitnil,omitempty" name:"Limit"`
}

func (r *DescribeRecordListRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRecordListRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Domain")
	delete(f, "DomainId")
	delete(f, "Subdomain")
	delete(f, "RecordType")
	delete(f, "RecordLine")
	delete(f, "RecordLineId")
	delete(f, "GroupId")
	delete(f, "Keyword")
	delete(f, "SortField")
	delete(f, "SortType")
	delete(f, "Offset")
	delete(f, "Limit")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DescribeRecordListRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DescribeRecordListResponseParams struct {
	// 记录的数量统计信息
	RecordCountInfo *RecordCountInfo `json:"RecordCountInfo,omitnil,omitempty" name:"RecordCountInfo"`

	// 获取的记录列表
	RecordList []*RecordListItem `json:"RecordList,omitnil,omitempty" name:"RecordList"`

	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

type DescribeRecordListResponse struct {
	*tchttp.BaseResponse
	Response *DescribeRecordListResponseParams `json:"Response"`
}

func (r *DescribeRecordListResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeRecordListResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type ModifyRecordRequest struct {
	*tchttp.BaseRequest

	// 域名
	Domain *string `json:"Domain,omitnil,omitempty" name:"Domain"`

	// 记录类型，通过 API 记录类型获得，大写英文，比如：A 。
	RecordType *string `json:"RecordType,omitnil,omitempty" name:"RecordType"`

	// 记录线路，通过 API 记录线路获得，中文，比如：默认。
	RecordLine *string `json:"RecordLine,omitnil,omitempty" name:"RecordLine"`

	// 记录值，如 IP : 200.200.200.200， CNAME : cname.dnspod.com.， MX : mail.dnspod.com.。
	Value *string `json:"Value,omitnil,omitempty" name:"Value"`

	// 记录 ID 。可以通过接口DescribeRecordList查到所有的解析记录列表以及对应的RecordId
	RecordId *uint64 `json:"RecordId,omitnil,omitempty" name:"RecordId"`

	// 域名 ID 。参数 DomainId 优先级比参数 Domain 高，如果传递参数 DomainId 将忽略参数 Domain 。可以通过接口DescribeDomainList查到所有的Domain以及DomainId
	DomainId *uint64 `json:"DomainId,omitnil,omitempty" name:"DomainId"`

	// 主机记录，如 www，如果不传，默认为 @。
	SubDomain *string `json:"SubDomain,omitnil,omitempty" name:"SubDomain"`

	// 线路的 ID，通过 API 记录线路获得，英文字符串，比如：10=1。参数RecordLineId优先级高于RecordLine，如果同时传递二者，优先使用RecordLineId参数。
	RecordLineId *string `json:"RecordLineId,omitnil,omitempty" name:"RecordLineId"`

	// MX 优先级，当记录类型是 MX 时有效，范围1-20，MX 记录时必选。
	MX *uint64 `json:"MX,omitnil,omitempty" name:"MX"`

	// TTL，范围1-604800，不同等级域名最小值不同。
	TTL *uint64 `json:"TTL,omitnil,omitempty" name:"TTL"`

	// 权重信息，0到100的整数。0 表示关闭，不传该参数，表示不设置权重信息。
	Weight *uint64 `json:"Weight,omitnil,omitempty" name:"Weight"`

	// 记录初始状态，取值范围为 ENABLE 和 DISABLE 。默认为 ENABLE ，如果传入 DISABLE，解析不会生效，也不会验证负载均衡的限制。
	Status *string `json:"Status,omitnil,omitempty" name:"Status"`

	// 记录的备注信息。传空删除备注。
	Remark *string `json:"Remark,omitnil,omitempty" name:"Remark"`

	// 开启DNSSEC时，强制将其它记录修改为CNAME/URL记录
	DnssecConflictMode *string `json:"DnssecConflictMode,omitnil,omitempty" name:"DnssecConflictMode"`
}

func (r *ModifyRecordRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyRecordRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Domain")
	delete(f, "RecordType")
	delete(f, "RecordLine")
	delete(f, "Value")
	delete(f, "RecordId")
	delete(f, "DomainId")
	delete(f, "SubDomain")
	delete(f, "RecordLineId")
	delete(f, "MX")
	delete(f, "TTL")
	delete(f, "Weight")
	delete(f, "Status")
	delete(f, "Remark")
	delete(f, "DnssecConflictMode")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "ModifyRecordRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type ModifyRecordResponseParams struct {
	// 记录ID
	RecordId *uint64 `json:"RecordId,omitnil,omitempty" name:"RecordId"`

	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

type ModifyRecordResponse struct {
	*tchttp.BaseResponse
	Response *ModifyRecordResponseParams `json:"Response"`
}

func (r *ModifyRecordResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *ModifyRecordResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type DeleteRecordRequest struct {
	*tchttp.BaseRequest

	// 域名
	Domain *string `json:"Domain,omitnil,omitempty" name:"Domain"`

	// 记录 ID 。可以通过接口DescribeRecordList查到所有的解析记录列表以及对应的RecordId
	RecordId *uint64 `json:"RecordId,omitnil,omitempty" name:"RecordId"`

	// 域名 ID 。参数 DomainId 优先级比参数 Domain 高，如果传递参数 DomainId 将忽略参数 Domain 。可以通过接口DescribeDomainList查到所有的Domain以及DomainId
	DomainId *uint64 `json:"DomainId,omitnil,omitempty" name:"DomainId"`
}

func (r *DeleteRecordRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteRecordRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	delete(f, "Domain")
	delete(f, "RecordId")
	delete(f, "DomainId")
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "DeleteRecordRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type DeleteRecordResponseParams struct {
	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

type DeleteRecordResponse struct {
	*tchttp.BaseResponse
	Response *DeleteRecordResponseParams `json:"Response"`
}

func (r *DeleteRecordResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DeleteRecordResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type RecordCountInfo struct {
	// 子域名数量
	SubdomainCount *uint64 `json:"SubdomainCount,omitnil,omitempty" name:"SubdomainCount"`

	// 列表返回的记录数
	ListCount *uint64 `json:"ListCount,omitnil,omitempty" name:"ListCount"`

	// 总的记录数
	TotalCount *uint64 `json:"TotalCount,omitnil,omitempty" name:"TotalCount"`
}

type RecordListItem struct {
	// 记录Id
	RecordId *uint64 `json:"RecordId,omitnil,omitempty" name:"RecordId"`

	// 记录值
	Value *string `json:"Value,omitnil,omitempty" name:"Value"`

	// 记录状态，启用：ENABLE，暂停：DISABLE
	Status *string `json:"Status,omitnil,omitempty" name:"Status"`

	// 更新时间
	UpdatedOn *string `json:"UpdatedOn,omitnil,omitempty" name:"UpdatedOn"`

	// 主机名
	Name *string `json:"Name,omitnil,omitempty" name:"Name"`

	// 记录线路
	Line *string `json:"Line,omitnil,omitempty" name:"Line"`

	// 线路Id
	LineId *string `json:"LineId,omitnil,omitempty" name:"LineId"`

	// 记录类型
	Type *string `json:"Type,omitnil,omitempty" name:"Type"`

	// 记录权重，用于负载均衡记录
	// 注意：此字段可能返回 null，表示取不到有效值。
	Weight *uint64 `json:"Weight,omitnil,omitempty" name:"Weight"`

	// 记录监控状态，正常：OK，告警：WARN，宕机：DOWN，未设置监控或监控暂停则为空
	MonitorStatus *string `json:"MonitorStatus,omitnil,omitempty" name:"MonitorStatus"`

	// 记录备注说明
	Remark *string `json:"Remark,omitnil,omitempty" name:"Remark"`

	// 记录缓存时间
	TTL *uint64 `json:"TTL,omitnil,omitempty" name:"TTL"`

	// MX值，只有MX记录有
	// 注意：此字段可能返回 null，表示取不到有效值。
	MX *uint64 `json:"MX,omitnil,omitempty" name:"MX"`

	// 是否是默认的ns记录
	DefaultNS *bool `json:"DefaultNS,omitnil,omitempty" name:"DefaultNS"`
}
