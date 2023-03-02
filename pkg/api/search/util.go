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

package search

import (
	"fmt"

	"engine/pkg/cache/impls"
	"engine/pkg/config"
)

func validateSystemMatchClient(systemID, clientID string) error {
	if systemID == "" || clientID == "" {
		return fmt.Errorf("system_id or client_id do not allow empty")
	}

	validClients, err := impls.GetSystemClients(systemID)
	if err != nil {
		return fmt.Errorf("get system(%s) valid clients fail, err=%w", systemID, err)
	}

	for _, c := range validClients {
		if clientID == c {
			return nil
		}
	}

	return fmt.Errorf("client(%s) can not request system(%s)", clientID, systemID)
}

func isSuperClient(clientID string) bool {
	return config.SuperAppCodeSet.Has(clientID)
}
