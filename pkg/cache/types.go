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

package cache

import (
	"strconv"
)

// Key ...
type Key interface {
	Key() string
}

// StringKey ...
type StringKey struct {
	key string
}

// NewStringKey ...
func NewStringKey(key string) StringKey {
	return StringKey{
		key: key,
	}
}

// Key ...
func (s StringKey) Key() string {
	return s.key
}

// Int64Key ...
type Int64Key struct {
	key int64
}

// NewInt64Key ...
func NewInt64Key(key int64) Int64Key {
	return Int64Key{
		key: key,
	}
}

// Key ...
func (k Int64Key) Key() string {
	return strconv.FormatInt(k.key, 10)
}
