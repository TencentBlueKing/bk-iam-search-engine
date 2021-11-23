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

package impls

import (
	"time"

	"engine/pkg/cache/memory"
)

// LocalAppCodeAppSecretCache ...
var (
	LocalAPIGatewayJWTClientIDCache memory.Cache
	LocalSystemClientsCache         memory.Cache
	LocalAppCodeAppSecretCache      memory.Cache
)

// Cache should only know about get/retrieve data
// ! DO NOT CARE ABOUT WHAT THE DATA WILL BE USED FOR
func InitCaches(disabled bool) {
	LocalAppCodeAppSecretCache = memory.NewCache(
		"app_code_app_secret",
		disabled,
		retrieveAppCodeAppSecret,
		12*time.Hour,
	)

	LocalAPIGatewayJWTClientIDCache = memory.NewCache(
		"local_apigw_jwt_client_id",
		disabled,
		retrieveAPIGatewayJWTClientID,
		30*time.Second,
	)

	LocalSystemClientsCache = memory.NewCache(
		"local_system_clients",
		disabled,
		retrieveSystemClients,
		10*time.Minute,
	)
}
