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
	"fmt"
	"strings"

	"engine/pkg/cache"
	"engine/pkg/components"
)

func retrieveSystemClients(k cache.Key) (interface{}, error) {
	k1 := k.(cache.StringKey)

	systemID := k1.Key()

	system, err := components.NewIAMClient().GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	return strings.Split(system.Clients, ","), nil
}

// GetSystemClients ...
func GetSystemClients(systemID string) (clients []string, err error) {
	key := cache.NewStringKey(systemID)

	var value interface{}
	value, err = LocalSystemClientsCache.Get(key)
	if err != nil {
		err = fmt.Errorf("GetSystemClients: LocalSystemClientsCache.Get key=`%s` fail", key.Key())
		return
	}

	var ok bool
	clients, ok = value.([]string)
	if !ok {
		err = fmt.Errorf(
			"GetSystemClients: LocalSystemClientsCache.Get key=`%s` fail, not []string in cache", systemID)
		return
	}

	return clients, nil
}
