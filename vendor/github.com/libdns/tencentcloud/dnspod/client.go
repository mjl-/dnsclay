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
	"context"
	"errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2021-03-23"

type Client struct {
	common.Client
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewCreateRecordRequest() (request *CreateRecordRequest) {
	request = &CreateRecordRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("dnspod", APIVersion, "CreateRecord")

	return
}

func NewCreateRecordResponse() (response *CreateRecordResponse) {
	response = &CreateRecordResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// CreateRecordWithContext
// 添加记录
//
// 备注：新添加的解析记录存在短暂的索引延迟，如果查询不到新增记录，请在 30 秒后重试
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_DNSSECINCOMPLETECLOSED = "FailedOperation.DNSSECIncompleteClosed"
//	FAILEDOPERATION_DOMAINISLOCKED = "FailedOperation.DomainIsLocked"
//	FAILEDOPERATION_DOMAINISSPAM = "FailedOperation.DomainIsSpam"
//	FAILEDOPERATION_LOGINAREANOTALLOWED = "FailedOperation.LoginAreaNotAllowed"
//	FAILEDOPERATION_LOGINFAILED = "FailedOperation.LoginFailed"
//	FAILEDOPERATION_MUSTADDDEFAULTLINEFIRST = "FailedOperation.MustAddDefaultLineFirst"
//	FAILEDOPERATION_UNKNOWERROR = "FailedOperation.UnknowError"
//	INVALIDPARAMETER_ACCOUNTISBANNED = "InvalidParameter.AccountIsBanned"
//	INVALIDPARAMETER_CUSTOMMESSAGE = "InvalidParameter.CustomMessage"
//	INVALIDPARAMETER_DNSSECADDCNAMEERROR = "InvalidParameter.DnssecAddCnameError"
//	INVALIDPARAMETER_DOMAINIDINVALID = "InvalidParameter.DomainIdInvalid"
//	INVALIDPARAMETER_DOMAININVALID = "InvalidParameter.DomainInvalid"
//	INVALIDPARAMETER_DOMAINISALIASER = "InvalidParameter.DomainIsAliaser"
//	INVALIDPARAMETER_DOMAINNOTALLOWEDMODIFYRECORDS = "InvalidParameter.DomainNotAllowedModifyRecords"
//	INVALIDPARAMETER_DOMAINNOTBEIAN = "InvalidParameter.DomainNotBeian"
//	INVALIDPARAMETER_DOMAINRECORDEXIST = "InvalidParameter.DomainRecordExist"
//	INVALIDPARAMETER_EMAILNOTVERIFIED = "InvalidParameter.EmailNotVerified"
//	INVALIDPARAMETER_INVALIDWEIGHT = "InvalidParameter.InvalidWeight"
//	INVALIDPARAMETER_LOGINTOKENIDERROR = "InvalidParameter.LoginTokenIdError"
//	INVALIDPARAMETER_LOGINTOKENNOTEXISTS = "InvalidParameter.LoginTokenNotExists"
//	INVALIDPARAMETER_LOGINTOKENVALIDATEFAILED = "InvalidParameter.LoginTokenValidateFailed"
//	INVALIDPARAMETER_MOBILENOTVERIFIED = "InvalidParameter.MobileNotVerified"
//	INVALIDPARAMETER_MXINVALID = "InvalidParameter.MxInvalid"
//	INVALIDPARAMETER_RECORDLINEINVALID = "InvalidParameter.RecordLineInvalid"
//	INVALIDPARAMETER_RECORDTYPEINVALID = "InvalidParameter.RecordTypeInvalid"
//	INVALIDPARAMETER_RECORDVALUEINVALID = "InvalidParameter.RecordValueInvalid"
//	INVALIDPARAMETER_RECORDVALUELENGTHINVALID = "InvalidParameter.RecordValueLengthInvalid"
//	INVALIDPARAMETER_REQUESTIPLIMITED = "InvalidParameter.RequestIpLimited"
//	INVALIDPARAMETER_SUBDOMAININVALID = "InvalidParameter.SubdomainInvalid"
//	INVALIDPARAMETER_UNREALNAMEUSER = "InvalidParameter.UnrealNameUser"
//	INVALIDPARAMETER_URLVALUEILLEGAL = "InvalidParameter.UrlValueIllegal"
//	INVALIDPARAMETER_USERNOTEXISTS = "InvalidParameter.UserNotExists"
//	INVALIDPARAMETERVALUE_DOMAINNOTEXISTS = "InvalidParameterValue.DomainNotExists"
//	INVALIDPARAMETERVALUE_USERIDINVALID = "InvalidParameterValue.UserIdInvalid"
//	LIMITEXCEEDED_AAAACOUNTLIMIT = "LimitExceeded.AAAACountLimit"
//	LIMITEXCEEDED_ATNSRECORDLIMIT = "LimitExceeded.AtNsRecordLimit"
//	LIMITEXCEEDED_FAILEDLOGINLIMITEXCEEDED = "LimitExceeded.FailedLoginLimitExceeded"
//	LIMITEXCEEDED_HIDDENURLEXCEEDED = "LimitExceeded.HiddenUrlExceeded"
//	LIMITEXCEEDED_NSCOUNTLIMIT = "LimitExceeded.NsCountLimit"
//	LIMITEXCEEDED_RECORDTTLLIMIT = "LimitExceeded.RecordTtlLimit"
//	LIMITEXCEEDED_SRVCOUNTLIMIT = "LimitExceeded.SrvCountLimit"
//	LIMITEXCEEDED_SUBDOMAINLEVELLIMIT = "LimitExceeded.SubdomainLevelLimit"
//	LIMITEXCEEDED_SUBDOMAINROLLLIMIT = "LimitExceeded.SubdomainRollLimit"
//	LIMITEXCEEDED_SUBDOMAINWCARDLIMIT = "LimitExceeded.SubdomainWcardLimit"
//	LIMITEXCEEDED_URLCOUNTLIMIT = "LimitExceeded.UrlCountLimit"
//	OPERATIONDENIED_DOMAINOWNERALLOWEDONLY = "OperationDenied.DomainOwnerAllowedOnly"
//	OPERATIONDENIED_IPINBLACKLISTNOTALLOWED = "OperationDenied.IPInBlacklistNotAllowed"
//	OPERATIONDENIED_NOPERMISSIONTOOPERATEDOMAIN = "OperationDenied.NoPermissionToOperateDomain"
//	OPERATIONDENIED_NOTADMIN = "OperationDenied.NotAdmin"
//	OPERATIONDENIED_NOTAGENT = "OperationDenied.NotAgent"
//	OPERATIONDENIED_NOTMANAGEDUSER = "OperationDenied.NotManagedUser"
//	REQUESTLIMITEXCEEDED_REQUESTLIMITEXCEEDED = "RequestLimitExceeded.RequestLimitExceeded"
func (c *Client) CreateRecordWithContext(ctx context.Context, request *CreateRecordRequest) (response *CreateRecordResponse, err error) {
	if request == nil {
		request = NewCreateRecordRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("CreateRecord require credential")
	}

	request.SetContext(ctx)

	response = NewCreateRecordResponse()
	err = c.Send(request, response)
	return
}

func NewDescribeRecordListRequest() (request *DescribeRecordListRequest) {
	request = &DescribeRecordListRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("dnspod", APIVersion, "DescribeRecordList")

	return
}

func NewDescribeRecordListResponse() (response *DescribeRecordListResponse) {
	response = &DescribeRecordListResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// DescribeRecordListWithContext
// 获取某个域名下的解析记录列表
//
// 备注：
//
// 1. 新添加的解析记录存在短暂的索引延迟，如果查询不到新增记录，请在 30 秒后重试
//
// 2.  API获取的记录总条数会比控制台多2条，原因是： 为了防止用户误操作导致解析服务不可用，对2021-10-29 14:24:26之后添加的域名，在控制台都不显示这2条NS记录。
//
// 可能返回的错误码:
//
//	AUTHFAILURE = "AuthFailure"
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_NOTDOMAINOWNER = "FailedOperation.NotDomainOwner"
//	FAILEDOPERATION_NOTREALNAMEDUSER = "FailedOperation.NotRealNamedUser"
//	FAILEDOPERATION_UNKNOWERROR = "FailedOperation.UnknowError"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_DOMAINIDINVALID = "InvalidParameter.DomainIdInvalid"
//	INVALIDPARAMETER_DOMAININVALID = "InvalidParameter.DomainInvalid"
//	INVALIDPARAMETER_DOMAINISALIASER = "InvalidParameter.DomainIsAliaser"
//	INVALIDPARAMETER_OPERATEFAILED = "InvalidParameter.OperateFailed"
//	INVALIDPARAMETER_PARAMINVALID = "InvalidParameter.ParamInvalid"
//	INVALIDPARAMETER_RECORDLINEINVALID = "InvalidParameter.RecordLineInvalid"
//	INVALIDPARAMETER_RECORDTYPEINVALID = "InvalidParameter.RecordTypeInvalid"
//	INVALIDPARAMETER_RESULTMORETHAN500 = "InvalidParameter.ResultMoreThan500"
//	INVALIDPARAMETERVALUE_DOMAINNOTEXISTS = "InvalidParameterValue.DomainNotExists"
//	INVALIDPARAMETERVALUE_LIMITINVALID = "InvalidParameterValue.LimitInvalid"
//	OPERATIONDENIED_NOPERMISSIONTOOPERATEDOMAIN = "OperationDenied.NoPermissionToOperateDomain"
//	REQUESTLIMITEXCEEDED = "RequestLimitExceeded"
//	REQUESTLIMITEXCEEDED_REQUESTLIMITEXCEEDED = "RequestLimitExceeded.RequestLimitExceeded"
//	RESOURCENOTFOUND_NODATAOFRECORD = "ResourceNotFound.NoDataOfRecord"
//	UNAUTHORIZEDOPERATION = "UnauthorizedOperation"
func (c *Client) DescribeRecordListWithContext(ctx context.Context, request *DescribeRecordListRequest) (response *DescribeRecordListResponse, err error) {
	if request == nil {
		request = NewDescribeRecordListRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeRecordList require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeRecordListResponse()
	err = c.Send(request, response)
	return
}

func NewModifyRecordRequest() (request *ModifyRecordRequest) {
	request = &ModifyRecordRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("dnspod", APIVersion, "ModifyRecord")

	return
}

func NewModifyRecordResponse() (response *ModifyRecordResponse) {
	response = &ModifyRecordResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// ModifyRecordWithContext
// 修改记录
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_DNSSECINCOMPLETECLOSED = "FailedOperation.DNSSECIncompleteClosed"
//	FAILEDOPERATION_DOMAINISLOCKED = "FailedOperation.DomainIsLocked"
//	FAILEDOPERATION_DOMAINISSPAM = "FailedOperation.DomainIsSpam"
//	FAILEDOPERATION_FREQUENCYLIMIT = "FailedOperation.FrequencyLimit"
//	FAILEDOPERATION_LOGINAREANOTALLOWED = "FailedOperation.LoginAreaNotAllowed"
//	FAILEDOPERATION_LOGINFAILED = "FailedOperation.LoginFailed"
//	FAILEDOPERATION_MUSTADDDEFAULTLINEFIRST = "FailedOperation.MustAddDefaultLineFirst"
//	FAILEDOPERATION_UNKNOWERROR = "FailedOperation.UnknowError"
//	INVALIDPARAMETER_ACCOUNTISBANNED = "InvalidParameter.AccountIsBanned"
//	INVALIDPARAMETER_CUSTOMMESSAGE = "InvalidParameter.CustomMessage"
//	INVALIDPARAMETER_DNSSECADDCNAMEERROR = "InvalidParameter.DnssecAddCnameError"
//	INVALIDPARAMETER_DOMAINIDINVALID = "InvalidParameter.DomainIdInvalid"
//	INVALIDPARAMETER_DOMAININVALID = "InvalidParameter.DomainInvalid"
//	INVALIDPARAMETER_DOMAINISALIASER = "InvalidParameter.DomainIsAliaser"
//	INVALIDPARAMETER_DOMAINNOTALLOWEDMODIFYRECORDS = "InvalidParameter.DomainNotAllowedModifyRecords"
//	INVALIDPARAMETER_DOMAINNOTBEIAN = "InvalidParameter.DomainNotBeian"
//	INVALIDPARAMETER_DOMAINRECORDEXIST = "InvalidParameter.DomainRecordExist"
//	INVALIDPARAMETER_EMAILNOTVERIFIED = "InvalidParameter.EmailNotVerified"
//	INVALIDPARAMETER_INVALIDWEIGHT = "InvalidParameter.InvalidWeight"
//	INVALIDPARAMETER_LOGINTOKENIDERROR = "InvalidParameter.LoginTokenIdError"
//	INVALIDPARAMETER_LOGINTOKENNOTEXISTS = "InvalidParameter.LoginTokenNotExists"
//	INVALIDPARAMETER_LOGINTOKENVALIDATEFAILED = "InvalidParameter.LoginTokenValidateFailed"
//	INVALIDPARAMETER_MOBILENOTVERIFIED = "InvalidParameter.MobileNotVerified"
//	INVALIDPARAMETER_MXINVALID = "InvalidParameter.MxInvalid"
//	INVALIDPARAMETER_OPERATEFAILED = "InvalidParameter.OperateFailed"
//	INVALIDPARAMETER_RECORDIDINVALID = "InvalidParameter.RecordIdInvalid"
//	INVALIDPARAMETER_RECORDLINEINVALID = "InvalidParameter.RecordLineInvalid"
//	INVALIDPARAMETER_RECORDTYPEINVALID = "InvalidParameter.RecordTypeInvalid"
//	INVALIDPARAMETER_RECORDVALUEINVALID = "InvalidParameter.RecordValueInvalid"
//	INVALIDPARAMETER_RECORDVALUELENGTHINVALID = "InvalidParameter.RecordValueLengthInvalid"
//	INVALIDPARAMETER_REQUESTIPLIMITED = "InvalidParameter.RequestIpLimited"
//	INVALIDPARAMETER_SUBDOMAININVALID = "InvalidParameter.SubdomainInvalid"
//	INVALIDPARAMETER_UNREALNAMEUSER = "InvalidParameter.UnrealNameUser"
//	INVALIDPARAMETER_URLVALUEILLEGAL = "InvalidParameter.UrlValueIllegal"
//	INVALIDPARAMETER_USERNOTEXISTS = "InvalidParameter.UserNotExists"
//	INVALIDPARAMETERVALUE_DOMAINNOTEXISTS = "InvalidParameterValue.DomainNotExists"
//	INVALIDPARAMETERVALUE_USERIDINVALID = "InvalidParameterValue.UserIdInvalid"
//	LIMITEXCEEDED_AAAACOUNTLIMIT = "LimitExceeded.AAAACountLimit"
//	LIMITEXCEEDED_ATNSRECORDLIMIT = "LimitExceeded.AtNsRecordLimit"
//	LIMITEXCEEDED_FAILEDLOGINLIMITEXCEEDED = "LimitExceeded.FailedLoginLimitExceeded"
//	LIMITEXCEEDED_HIDDENURLEXCEEDED = "LimitExceeded.HiddenUrlExceeded"
//	LIMITEXCEEDED_NSCOUNTLIMIT = "LimitExceeded.NsCountLimit"
//	LIMITEXCEEDED_RECORDTTLLIMIT = "LimitExceeded.RecordTtlLimit"
//	LIMITEXCEEDED_SRVCOUNTLIMIT = "LimitExceeded.SrvCountLimit"
//	LIMITEXCEEDED_SUBDOMAINLEVELLIMIT = "LimitExceeded.SubdomainLevelLimit"
//	LIMITEXCEEDED_SUBDOMAINROLLLIMIT = "LimitExceeded.SubdomainRollLimit"
//	LIMITEXCEEDED_SUBDOMAINWCARDLIMIT = "LimitExceeded.SubdomainWcardLimit"
//	LIMITEXCEEDED_URLCOUNTLIMIT = "LimitExceeded.UrlCountLimit"
//	OPERATIONDENIED_DOMAINOWNERALLOWEDONLY = "OperationDenied.DomainOwnerAllowedOnly"
//	OPERATIONDENIED_IPINBLACKLISTNOTALLOWED = "OperationDenied.IPInBlacklistNotAllowed"
//	OPERATIONDENIED_NOPERMISSIONTOOPERATEDOMAIN = "OperationDenied.NoPermissionToOperateDomain"
//	OPERATIONDENIED_NOTADMIN = "OperationDenied.NotAdmin"
//	OPERATIONDENIED_NOTAGENT = "OperationDenied.NotAgent"
//	OPERATIONDENIED_NOTMANAGEDUSER = "OperationDenied.NotManagedUser"
//	REQUESTLIMITEXCEEDED_REQUESTLIMITEXCEEDED = "RequestLimitExceeded.RequestLimitExceeded"
func (c *Client) ModifyRecordWithContext(ctx context.Context, request *ModifyRecordRequest) (response *ModifyRecordResponse, err error) {
	if request == nil {
		request = NewModifyRecordRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyRecord require credential")
	}

	request.SetContext(ctx)

	response = NewModifyRecordResponse()
	err = c.Send(request, response)
	return
}

func NewDeleteRecordRequest() (request *DeleteRecordRequest) {
	request = &DeleteRecordRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("dnspod", APIVersion, "DeleteRecord")

	return
}

func NewDeleteRecordResponse() (response *DeleteRecordResponse) {
	response = &DeleteRecordResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// DeleteRecordWithContext
// 删除记录
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	FAILEDOPERATION_DOMAINISLOCKED = "FailedOperation.DomainIsLocked"
//	FAILEDOPERATION_DOMAINISSPAM = "FailedOperation.DomainIsSpam"
//	FAILEDOPERATION_FREQUENCYLIMIT = "FailedOperation.FrequencyLimit"
//	FAILEDOPERATION_LOGINAREANOTALLOWED = "FailedOperation.LoginAreaNotAllowed"
//	FAILEDOPERATION_LOGINFAILED = "FailedOperation.LoginFailed"
//	FAILEDOPERATION_UNKNOWERROR = "FailedOperation.UnknowError"
//	INVALIDPARAMETER_DOMAINIDINVALID = "InvalidParameter.DomainIdInvalid"
//	INVALIDPARAMETER_DOMAININVALID = "InvalidParameter.DomainInvalid"
//	INVALIDPARAMETER_DOMAINISALIASER = "InvalidParameter.DomainIsAliaser"
//	INVALIDPARAMETER_DOMAINNOTALLOWEDMODIFYRECORDS = "InvalidParameter.DomainNotAllowedModifyRecords"
//	INVALIDPARAMETER_LOGINTOKENIDERROR = "InvalidParameter.LoginTokenIdError"
//	INVALIDPARAMETER_LOGINTOKENNOTEXISTS = "InvalidParameter.LoginTokenNotExists"
//	INVALIDPARAMETER_LOGINTOKENVALIDATEFAILED = "InvalidParameter.LoginTokenValidateFailed"
//	INVALIDPARAMETER_OPERATEFAILED = "InvalidParameter.OperateFailed"
//	INVALIDPARAMETER_RECORDIDINVALID = "InvalidParameter.RecordIdInvalid"
//	INVALIDPARAMETER_REQUESTIPLIMITED = "InvalidParameter.RequestIpLimited"
//	INVALIDPARAMETER_UNREALNAMEUSER = "InvalidParameter.UnrealNameUser"
//	INVALIDPARAMETER_USERNOTEXISTS = "InvalidParameter.UserNotExists"
//	INVALIDPARAMETERVALUE_DOMAINNOTEXISTS = "InvalidParameterValue.DomainNotExists"
//	INVALIDPARAMETERVALUE_USERIDINVALID = "InvalidParameterValue.UserIdInvalid"
//	LIMITEXCEEDED_FAILEDLOGINLIMITEXCEEDED = "LimitExceeded.FailedLoginLimitExceeded"
//	OPERATIONDENIED_DOMAINOWNERALLOWEDONLY = "OperationDenied.DomainOwnerAllowedOnly"
//	OPERATIONDENIED_NOPERMISSIONTOOPERATEDOMAIN = "OperationDenied.NoPermissionToOperateDomain"
//	OPERATIONDENIED_NOTADMIN = "OperationDenied.NotAdmin"
//	OPERATIONDENIED_NOTAGENT = "OperationDenied.NotAgent"
//	OPERATIONDENIED_NOTMANAGEDUSER = "OperationDenied.NotManagedUser"
//	REQUESTLIMITEXCEEDED_REQUESTLIMITEXCEEDED = "RequestLimitExceeded.RequestLimitExceeded"
func (c *Client) DeleteRecordWithContext(ctx context.Context, request *DeleteRecordRequest) (response *DeleteRecordResponse, err error) {
	if request == nil {
		request = NewDeleteRecordRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DeleteRecord require credential")
	}

	request.SetContext(ctx)

	response = NewDeleteRecordResponse()
	err = c.Send(request, response)
	return
}
