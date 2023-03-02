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

package main

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"engine/pkg/types"

	"github.com/TencentBlueKing/iam-go-sdk"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
	"github.com/stretchr/testify/assert"
)

func TestRunRun(t *testing.T) {
	system := "bk_sops"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System: system,
			Type:   "biz_custom_query",
			ID:     "bt7nv2qevc81ff76hpk0",
			Attribute: map[string]interface{}{
				"_bk_iam_path_": "/biz,0/set,2/",
			},
		},
	})

	expr := expression.ExprCell{
		OP: operator.AND,
		Content: []expression.ExprCell{
			{
				OP:    operator.Eq,
				Field: "biz_custom_query.id",
				Value: "bt7nv2qevc81ff76hpk0",
			},
			{
				OP:    operator.StartsWith,
				Field: "biz_custom_query._bk_iam_path_",
				Value: "/biz,0/",
			},
		},
	}

	assert.True(t, expr.Eval(objSet))
}

func BenchmarkEvalOnceEqID(b *testing.B) {
	system := "bk_sops"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System:    system,
			Type:      "project",
			ID:        "5000805",
			Attribute: map[string]interface{}{},
		},
	})

	// "data/bk_sops_flow_create.txt"
	expr := expression.ExprCell{
		OP:    operator.Eq,
		Field: "project.id",
		Value: "5000805",
	}

	for i := 0; i < b.N; i++ {
		expr.Eval(objSet)
	}
}

func BenchmarkEvalIn10000ID(b *testing.B) {
	system := "bk_sops"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System:    system,
			Type:      "project",
			ID:        "5000805",
			Attribute: map[string]interface{}{},
		},
	})

	ids := make([]string, 0, 10000)
	for i := 0; i < 9999; i++ {
		ids = append(ids, strconv.Itoa(i))
	}
	ids = append(ids, "5000805")

	// "data/bk_sops_flow_create.txt"
	expr := expression.ExprCell{
		OP:    operator.In,
		Field: "project.id",
		Value: ids,
	}

	for i := 0; i < b.N; i++ {
		expr.Eval(objSet)
	}
}

func BenchmarkEvalStartsWith(b *testing.B) {
	system := "bk_sops"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System: system,
			Type:   "project",
			ID:     "5000805",
			Attribute: map[string]interface{}{
				"_bk_iam_path_": "/biz,1/set,2/module,3/",
			},
		},
	})

	// "data/bk_sops_flow_create.txt"
	expr := expression.ExprCell{
		OP:    operator.StartsWith,
		Field: "project._bk_iam_path_",
		Value: "/biz,1/set,2/",
	}

	for i := 0; i < b.N; i++ {
		expr.Eval(objSet)
	}
}

func BenchmarkEval100000EqID(b *testing.B) {
	system := "bk_sops"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System:    system,
			Type:      "project",
			ID:        "5000805",
			Attribute: map[string]interface{}{},
		},
	})

	// "data/bk_sops_flow_create.txt"
	expr := expression.ExprCell{
		OP:    operator.Eq,
		Field: "project.id",
		Value: "5000805",
	}

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100000; j++ {
			expr.Eval(objSet)
		}
	}
}

func BenchmarkEval100000Complex(b *testing.B) {
	system := "bk_cmdb"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System:    system,
			Type:      "biz",
			ID:        "100382",
			Attribute: map[string]interface{}{},
		},
		{
			System:    system,
			Type:      "sys_resource_pool_directory",
			ID:        "5000415",
			Attribute: map[string]interface{}{},
		},
	})

	// "data/bk_cmdb_unassign_biz_host.txt"
	expr := expression.ExprCell{
		OP: operator.AND,
		Content: []expression.ExprCell{
			{
				OP:    operator.Eq,
				Field: "biz.id",
				Value: "100368",
			},
			{
				OP:    operator.Eq,
				Field: "sys_resource_pool_directory.id",
				Value: "5000415",
			},
		},
	}

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100000; j++ {
			expr.Eval(objSet)
		}
	}
}

func BenchmarkEval100000Complex2(b *testing.B) {
	system := "bk_job"
	objSet := iam.NewObjectSet([]iam.ResourceNode{
		{
			System: system,
			Type:   "job_template",
			ID:     "",
			Attribute: map[string]interface{}{
				"_bk_iam_path_": "/biz,100368/",
			},
		},
		{
			System: system,
			Type:   "host",
			ID:     "",
			Attribute: map[string]interface{}{
				"_bk_iam_path_": "/biz,100368/",
			},
		},
	})

	// "data/bk_job_debug_job_template.txt"
	expr := expression.ExprCell{
		OP: operator.AND,
		Content: []expression.ExprCell{
			{
				OP:    operator.StartsWith,
				Field: "job_template._bk_iam_path_",
				Value: "/biz,100368/",
			},
			{
				OP: operator.OR,
				Content: []expression.ExprCell{
					{
						OP:    operator.StartsWith,
						Field: "host._bk_iam_path_",
						Value: "/biz,100368/",
					},
					{
						OP:    operator.StartsWith,
						Field: "host._bk_iam_path_",
						Value: "/biz,100368/biz_custom_query,*/",
					},
				},
			},
		},
	}

	for i := 0; i < b.N; i++ {
		for j := 0; j < 100000; j++ {
			expr.Eval(objSet)
		}
	}
}

func BenchmarkLoopMap(b *testing.B) {
	subjects := make(map[string]types.Subject, 10)
	for i := 0; i < 300; i++ {
		subjects[fmt.Sprintf("%d", i)] = types.Subject{
			Type: "user",
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("%d", i),
		}
	}

	for i := 0; i < b.N; i++ {
		// keys := make([]string, 0, len(subjects))
		// for k, _ := range subjects {
		// 	keys = append(keys, k)
		// }
		// sort.Strings(keys)
		for _, s := range subjects {
			s.Name = ""
		}
	}
}

func BenchmarkLoopOrderedMap(b *testing.B) {
	subjects := make(map[string]types.Subject, 10)
	for i := 0; i < 300; i++ {
		// subjects = append(subjects, types.Subject{
		// 	Type: "user",
		// 	ID:   fmt.Sprintf("%d", i),
		// 	Name: fmt.Sprintf("%d", i),
		// })
		subjects[fmt.Sprintf("%d", i)] = types.Subject{
			Type: "user",
			ID:   fmt.Sprintf("%d", i),
			Name: fmt.Sprintf("%d", i),
		}
	}

	for i := 0; i < b.N; i++ {
		keys := make([]string, 0, len(subjects))
		for k := range subjects {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, s := range subjects {
			s.Name = ""
		}
	}
}
