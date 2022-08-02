/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心检索引擎
 * (BlueKing-IAM-Search-Engine) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package components

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	policyAPITypeAbac = "abac"
	policyAPITypeRbac = "rbac"
)

var (
	globalIAMHost   = ""
	globalAppCode   = ""
	globalAppSecret = ""

	// NOTE: we want to use the same config file, but different instance, so here we use env
	// api param type={engineAPIType}
	policyAPIType = "abac"
)

// InitComponentClients ...
func InitComponentClients(iamHost string, appCode string, appSecret string) {
	globalIAMHost = iamHost
	globalAppCode = appCode
	globalAppSecret = appSecret

	policyAPITypeFromEnv := os.Getenv("POLICY_API_TYPE")
	if policyAPITypeFromEnv == policyAPITypeAbac || policyAPITypeFromEnv == policyAPITypeRbac {
		policyAPIType = policyAPITypeFromEnv
	}
	log.Infof("init Component with policyAPIType=%s", policyAPIType)
}

// NewIAMClient ...
func NewIAMClient() IAMBackendClient {
	return NewIAMBackendClient(globalIAMHost, globalAppCode, globalAppSecret)
}
