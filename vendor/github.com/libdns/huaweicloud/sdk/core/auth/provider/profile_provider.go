// Copyright 2022 Huawei Technologies Co.,Ltd.
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

package provider

import (
	"errors"
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/auth"
)

const (
	credentialsFileEnvName = "HUAWEICLOUD_SDK_CREDENTIALS_FILE"
	defaultDir             = ".huaweicloud"
	defaultFile            = "credentials"

	akName            = "ak"
	skName            = "sk"
	projectIdName     = "project_id"
	domainIdName      = "domain_id"
	securityTokenName = "security_token"
	iamEndpointName   = "iam_endpoint"
	idpIdName         = "idp_id"
	idTokenFileName   = "id_token_file"
)

type ProfileCredentialProvider struct {
	credentialType string
}

// NewProfileCredentialProvider return a profile credential provider
// Supported credential types: basic, global
func NewProfileCredentialProvider(credentialType string) *ProfileCredentialProvider {
	return &ProfileCredentialProvider{credentialType: strings.ToLower(credentialType)}
}

// BasicCredentialProfileProvider return a profile provider for basic.Credentials
func BasicCredentialProfileProvider() *ProfileCredentialProvider {
	return NewProfileCredentialProvider(basicCredentialType)
}

// GlobalCredentialProfileProvider return a profile provider for global.Credentials
func GlobalCredentialProfileProvider() *ProfileCredentialProvider {
	return NewProfileCredentialProvider(globalCredentialType)
}

// GetCredentials get basic.Credentials or global.Credentials from profile
func (p *ProfileCredentialProvider) GetCredentials() (auth.ICredential, error) {
	return nil, errors.New("not support")
}
