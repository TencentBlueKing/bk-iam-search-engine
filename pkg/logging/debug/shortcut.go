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

package debug

// WithValue ...
func WithValue(e *Entry, key string, value interface{}) {
	if e == nil {
		return
	}

	e.WithValue(key, value)
}

// WithValues ...
func WithValues(e *Entry, data map[string]interface{}) {
	if e == nil {
		return
	}

	e.WithValues(data)
}

// WithError ...
func WithError(e *Entry, err error) {
	if e == nil {
		return
	}

	e.WithError(err)
}

// WithEsQuery ...
func WithEsQuery(e *Entry, _type string, query interface{}) {
	if e == nil {
		return
	}

	e.WithEsQuery(_type, NewEsQuery(query))
}

// WithEsQuerySubjects ...
func WithEsQuerySubjects(e *Entry, _type string, subjects interface{}) {
	if e == nil {
		return
	}

	e.WithEsQuerySubjects(_type, subjects)
}

// AddPolicy ...
func AddPolicy(e *Entry, policy interface{}) {
	if e == nil {
		return
	}

	e.AddPolicy(policy)
}

// AddStep ...
func AddStep(e *Entry, name string) {
	if e == nil {
		return
	}

	e.AddStep(NewStep(name))
}

// ReleaseDebugEntry ...
func ReleaseDebugEntry(e *Entry) {
	globalEntryPool.Put(e)
}

// NewDebugEntry ...
func NewDebugEntry() *Entry {
	return globalEntryPool.Get()
}

// NewDebugEntryWithFixedSubEntries ...
func NewDebugEntryWithFixedSubEntries(subSize int) *Entry {
	entry := globalEntryPool.Get()
	for i := 0; i < subSize; i++ {
		entry.AddSubDebug(globalEntryPool.Get())
	}
	return entry
}

// GetSubEntryByIndex ...
func GetSubEntryByIndex(e *Entry, index int) *Entry {
	if e == nil {
		return nil
	}

	if index >= len(e.SubDebugs) {
		return nil
	}

	return e.SubDebugs[index]
}
