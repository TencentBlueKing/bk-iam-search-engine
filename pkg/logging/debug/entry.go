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

import (
	"time"
)

// Fields ...
type Fields map[string]interface{}

// Step ...
type Step struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// EsQuery ...
type EsQuery struct {
	Expression interface{} `json:"expression"`
	Subjects   interface{} `json:"subjects"`
}

// Entry ...
type Entry struct {
	Time      time.Time     `json:"time"`
	Context   Fields        `json:"context"`
	Steps     []Step        `json:"steps"`
	AnyQuery  *EsQuery      `json:"any_query"`
	DocQuery  *EsQuery      `json:"doc_query"`
	Policies  []interface{} `json:"policies"`
	Error     string        `json:"error"`
	SubDebugs []*Entry      `json:"sub_debugs"`
}

// WithValue ...
func (e *Entry) WithValue(key string, value interface{}) {
	e.Context[key] = value
}

// WithValues ...
func (e *Entry) WithValues(data map[string]interface{}) {
	for key, value := range data {
		e.Context[key] = value
	}
}

// WithAnyQuery ...
func (e *Entry) WithEsQuery(_type string, query *EsQuery) {
	switch _type {
	case "any":
		e.AnyQuery = query
	case "doc":
		e.DocQuery = query
	}
}

// WithAnyQuery ...
func (e *Entry) WithEsQuerySubjects(_type string, subjects interface{}) {
	switch _type {
	case "any":
		if e.AnyQuery != nil {
			e.AnyQuery.Subjects = subjects
		}
	case "doc":
		if e.DocQuery != nil {
			e.DocQuery.Subjects = subjects
		}
	}
}

// WithError ...
func (e *Entry) WithError(err error) {
	if err != nil {
		e.Error = err.Error()
	}
}

// AddPolicy ...
func (e *Entry) AddPolicy(policy interface{}) {
	e.Policies = append(e.Policies, policy)
}

// AddStep ...
func (e *Entry) AddStep(step Step) {
	if e.Steps == nil {
		e.Steps = make([]Step, 0, 5)
	}

	step.Index = len(e.Steps) + 1

	e.Steps = append(e.Steps, step)
}

// NewStep ...
func NewStep(name string) Step {
	return Step{
		Name: name,
	}
}

// AddSubDebug ...
func (e *Entry) AddSubDebug(debug *Entry) {
	if debug == nil {
		return
	}

	if e.SubDebugs == nil {
		e.SubDebugs = make([]*Entry, 0, 5)
	}

	e.SubDebugs = append(e.SubDebugs, debug)
}

// NewEsQuery ...
func NewEsQuery(query interface{}) *EsQuery {
	return &EsQuery{
		Expression: query,
	}
}
