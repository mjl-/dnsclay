// Copyright 2020 Huawei Technologies Co.,Ltd.
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

package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/auth"
	"github.com/libdns/huaweicloud/sdk/core/auth/provider"
	"github.com/libdns/huaweicloud/sdk/core/config"
	"github.com/libdns/huaweicloud/sdk/core/impl"
	"github.com/libdns/huaweicloud/sdk/core/region"
	"github.com/libdns/huaweicloud/sdk/core/sdkerr"
)

type HcHttpClientBuilder struct {
	CredentialsType        []string
	derivedAuthServiceName string
	credentials            auth.ICredential
	endpoints              []string
	httpConfig             *config.HttpConfig
	region                 *region.Region
	errorHandler           sdkerr.ErrorHandler
}

func NewHcHttpClientBuilder() *HcHttpClientBuilder {
	hcHttpClientBuilder := &HcHttpClientBuilder{
		CredentialsType: []string{"basic.Credentials"},
		errorHandler:    sdkerr.DefaultErrorHandler{},
	}
	return hcHttpClientBuilder
}

func (builder *HcHttpClientBuilder) WithCredentialsType(credentialsType string) *HcHttpClientBuilder {
	builder.CredentialsType = strings.Split(credentialsType, ",")
	return builder
}

func (builder *HcHttpClientBuilder) WithDerivedAuthServiceName(derivedAuthServiceName string) *HcHttpClientBuilder {
	builder.derivedAuthServiceName = derivedAuthServiceName
	return builder
}

func (builder *HcHttpClientBuilder) WithEndpoints(endpoints []string) *HcHttpClientBuilder {
	builder.endpoints = endpoints
	return builder
}

func (builder *HcHttpClientBuilder) WithRegion(region *region.Region) *HcHttpClientBuilder {
	builder.region = region
	return builder
}

func (builder *HcHttpClientBuilder) WithHttpConfig(httpConfig *config.HttpConfig) *HcHttpClientBuilder {
	builder.httpConfig = httpConfig
	return builder
}

func (builder *HcHttpClientBuilder) WithCredential(iCredential auth.ICredential) *HcHttpClientBuilder {
	builder.credentials = iCredential
	return builder
}

func (builder *HcHttpClientBuilder) WithErrorHandler(errorHandler sdkerr.ErrorHandler) *HcHttpClientBuilder {
	builder.errorHandler = errorHandler
	return builder
}

func (builder *HcHttpClientBuilder) SafeBuild() (client *HcHttpClient, err error) {
	if builder.httpConfig == nil {
		builder.httpConfig = config.DefaultHttpConfig()
	}

	defaultHttpClient := impl.NewDefaultHttpClient(builder.httpConfig)

	if builder.credentials == nil {
		p := provider.DefaultCredentialProviderChain(builder.CredentialsType[0])
		credentials, err := p.GetCredentials()
		if err != nil {
			return nil, err
		}
		builder.credentials = credentials
	}

	t := reflect.TypeOf(builder.credentials)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	givenCredentialsType := t.String()
	match := false
	for _, credentialsType := range builder.CredentialsType {
		if credentialsType == givenCredentialsType {
			match = true
			break
		}
	}
	if !match {
		return nil, fmt.Errorf("need credential type is %s, actually is %s", builder.CredentialsType, givenCredentialsType)
	}

	if builder.region != nil {
		builder.endpoints = builder.region.Endpoints
		builder.credentials.ProcessAuthParams(defaultHttpClient, builder.region.Id)

		if credential, ok := builder.credentials.(auth.IDerivedCredential); ok {
			credential.ProcessDerivedAuthParams(builder.derivedAuthServiceName, builder.region.Id)
		}
	}

	for index, endpoint := range builder.endpoints {
		if !strings.HasPrefix(endpoint, "http") {
			builder.endpoints[index] = "https://" + endpoint
		}
	}

	hcHttpClient := NewHcHttpClient(defaultHttpClient).
		WithEndpoints(builder.endpoints).
		WithCredential(builder.credentials).
		WithErrorHandler(builder.errorHandler)
	return hcHttpClient, nil
}
