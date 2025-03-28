// Copyright 2023 Huawei Technologies Co.,Ltd.
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package sdkerr

import (
	"github.com/libdns/huaweicloud/sdk/core/request"
	"github.com/libdns/huaweicloud/sdk/core/response"
)

type ErrorHandler interface {
	HandleError(req *request.DefaultHttpRequest, resp *response.DefaultHttpResponse) error
}

type DefaultErrorHandler struct {
}

func (h DefaultErrorHandler) HandleError(req *request.DefaultHttpRequest, resp *response.DefaultHttpResponse) error {
	if resp.GetStatusCode() < 400 {
		return nil
	}
	return NewServiceResponseError(resp.Response)
}
