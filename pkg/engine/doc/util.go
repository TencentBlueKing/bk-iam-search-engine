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

package doc

import (
	"errors"
	"reflect"
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"

	"engine/pkg/types"
	"engine/pkg/util"
)

func splitBKIAMPath(value string) (paths []string) {
	// value = /biz,1/set,2/module,3/host,4/
	if value == "" {
		return
	}

	// 常见的path < 3层, 5个
	results := make([]string, 0, 5)

	// biz,1/set,2/module,3/host,4
	trimmedValue := strings.Trim(value, "/")

	// make prefix
	parts := strings.Split(trimmedValue, "/")
	for i := 1; i <= len(parts); i++ {
		a := strings.Join(parts[:i], "/")
		results = append(results, "/"+a+"/")

		// make prefix *
		if i > 1 {
			b := a[:strings.LastIndexByte(a, ',')] + ",*"
			results = append(results, "/"+b+"/")
		}
	}
	return results
}

func generateBkIAMPathList(value interface{}) []string {
	// NOTE: /biz,1/set,*/ 和  /biz,1/set,123/ 都能命中  /biz,1/set,123/host,3/
	//       如果资源有多个_bk_iam_path_, 那么这里的组合会变得非常大
	var paths []string
	if util.IsValueTypeArray(value) {
		pathSet := set.NewStringSet()

		// TODO: 未经测试
		listValue := reflect.ValueOf(value)
		for i := 0; i < listValue.Len(); i++ {
			ps := splitBKIAMPath(listValue.Index(i).Interface().(string))
			pathSet.Append(ps...)
		}
		paths = pathSet.ToSlice()
	} else {
		paths = splitBKIAMPath(value.(string))
	}
	return paths
}

func generateBkIAMPathNodes(value interface{}) []string {
	// NOTE: /biz,1/set,*/ 和  /biz,1/set,123/ 都能命中  /biz,1/set,123/host,3/
	//       如果资源有多个_bk_iam_path_, 那么这里的组合会变得非常大
	nodes := set.NewStringSet()

	if util.IsValueTypeArray(value) {
		// TODO: 未经测试
		listValue := reflect.ValueOf(value)
		for i := 0; i < listValue.Len(); i++ {
			ps := splitBKIAMPathIntoNodes(listValue.Index(i).Interface().(string))

			nodes.Append(ps...)
		}
	} else {
		ps := splitBKIAMPathIntoNodes(value.(string))
		nodes.Append(ps...)
	}
	return nodes.ToSlice()
}

func splitBKIAMPathIntoNodes(value string) (nodes []string) {
	// value = /biz,1/set,2/module,3/host,4/
	if value == "" {
		return
	}

	value = strings.Trim(value, "/")

	// /biz,1/set,2/module,3/host,4/
	parts := strings.Split(value, "/")
	for _, part := range parts {
		nodes = append(nodes, "/"+part+"/")
	}
	return nodes
}

// makeDoc ...
func makeDoc(docType types.ExpressionType, policy *types.Policy) (map[string]interface{}, error) {
	object := types.H{}

	system := policy.System
	// action := policy.Action.ID

	actions := make([]interface{}, 0, len(policy.Actions))
	for _, a := range policy.Actions {
		actions = append(actions, types.H{"id": a.ID})
	}

	// if docType == "any", do nothing
	// else, build the object for eval
	if docType == "doc" {
		if policy.Expression.OP == operator.Eq {
			object[policy.Expression.Field] = []interface{}{policy.Expression.Value}
		} else if policy.Expression.OP == operator.In {
			object[policy.Expression.Field] = policy.Expression.Value
		} else if policy.Expression.OP == operator.StartsWith {
			if util.IsValueTypeArray(policy.Expression.Value) {
				object[policy.Expression.Field] = policy.Expression.Value
			} else {
				object[policy.Expression.Field] = []interface{}{policy.Expression.Value}
			}
		} else if policy.Expression.OP == operator.StringContains {
			// a._bk_iam_path_ string_contains "/project,1/"
			// trans to:
			// a._bk_iam_path_contains_ = ["/project,1/"]
			field := types.ConvertBKIAMPathSuffixToBKIAMPathContainsSuffix(policy.Expression.Field)
			object[field] = []interface{}{policy.Expression.Value}
		} else if policy.Expression.OP == operator.OR {
			// NOTE: not a simple expression, but the OR expression with same object different fields
			// TODO: 考虑 写得更通用些, or A.a = 1 or A.a =2 or A.c = 3 能够正常合并
			for _, c := range policy.Expression.Content {
				object[c.Field] = c.Value
			}
		} else {
			return nil, errors.New("not a simple expression")
		}
	}

	// TODO: 这里没有包含 跨系统资源依赖的情况, resourceType chain来自于不同系统
	doc := types.H{
		"type":    docType,
		"id":      policy.ID,
		"version": policy.Version,

		"system": system,
		// "action": types.H{
		// 	"id": action,
		// },
		"actions": actions,

		"subject": types.H{
			"id":   policy.Subject.ID,
			"type": policy.Subject.Type,
			"name": policy.Subject.Name,
			"uid":  policy.Subject.UID,
		},

		"template_id": policy.TemplateID,

		// NOTE: 这里的system目的是为了避免不同系统的同一个resourceType名字一样类型不一样导致索引失败
		"resource": types.H{
			system: object,
		},

		"expired_at": policy.ExpiredAt,
		"updated_at": policy.UpdatedAt,
	}
	return doc, nil
}

type esSearchQueryFunc func(req *types.SearchRequest) types.H

func genActionSubQuery(action string) types.H {
	return types.H{
		"bool": types.H{
			"should": []types.H{
				{"action.id": action},
				{"actions.id": action},
			},
		},
	}
}
