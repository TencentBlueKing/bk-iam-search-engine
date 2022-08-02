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

package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"

	"engine/pkg/types"
)

// FIXME: 这里应该合并 isSingleEqOrIn / isBkIAMPathStartsWith / isBKIAMPathStringContains
// FIXME: 如果是单个number cmp, 也应该是simple的, 未来表达式中加入需要支持

// isAny will check the expression, return true if only one expression with operator `any`
func isAny(expr *expression.ExprCell) bool {
	return expr.OP == operator.Any
}

// isSingleEqOrIn will check the expression, return true if `obj.attr eq x` or  `obj.attr in []`
func isSingleEqOrIn(expr *expression.ExprCell) bool {
	return expr.OP == operator.Eq || expr.OP == operator.In
}

// isBkIAMPathStartsWith will check the expression, return true if `obj._bk_iam_path starts_with x`
func isBkIAMPathStartsWith(expr *expression.ExprCell) bool {
	return expr.OP == operator.StartsWith && strings.HasSuffix(expr.Field, types.BkIAMPathSuffix)
}

func isBkIAMPathStringContains(expr *expression.ExprCell) bool {
	return expr.OP == operator.StringContains && strings.HasSuffix(expr.Field, types.BkIAMPathSuffix)
}

// isAllOR will check the expression, return true if all the expression in content or nested expression are `OR`
func isAllOR(expr *expression.ExprCell) bool {
	// NOTE: single any already processed before
	// TODO: remove any from here, any in all_OR should be merged

	if expr.OP == operator.AND || expr.OP == operator.Any {
		return false
	}

	// OR and nested OR, all OR
	if expr.OP == operator.OR {
		for _, c := range expr.Content {
			if !isAllOR(&c) {
				return false
			}
		}
		return true
	}

	// NOTE: 当前只支持 _bk_iam_path_ 的 starts_with
	if expr.OP == operator.StartsWith && !strings.HasSuffix(expr.Field, types.BkIAMPathSuffix) {
		return false
	}
	// other operator
	return true
}

// sameObjectMergeableOPs ...
var sameObjectMergeableOPs = map[operator.OP]bool{
	operator.Eq:         true,
	operator.In:         true,
	operator.StartsWith: true,
}

// isSameObjectWithDifferentFields will check the expression,
// return true if all OR content, the field's objects are the same
func isSameObjectWithDifferentFields(expr *expression.ExprCell) bool {
	if expr.OP != operator.OR {
		return false
	}

	s := set.NewStringSet()
	for _, c := range expr.Content {
		// currently, only can merge into same object
		// 1. eq
		// 2. in
		// 3. x._bk_iam_path_ starts_with
		_, mergeable := sameObjectMergeableOPs[c.OP]
		if !mergeable {
			return false
		}

		// NOTE: 这里不允许, ._bk_iam_path_ 配置其他操作符
		// 只支持starts_with  *._bk_iam_path_
		if c.OP == operator.StartsWith && !strings.HasSuffix(c.Field, "._bk_iam_path_") {
			return false
		}

		// TODO: 未来
		//     - 其他操作符号怎么办? 目前 eq/in/starts_with特殊处理的
		//     - 未来, field结构改变了, 会变成什么样子的?
		parts := strings.Split(c.Field, ".")
		s.Add(parts[0])
	}

	// size == 1 表示是同一个对象
	return s.Size() == 1
}

// flattenORExpr will flat the nested OR expression into flatten expression list.
// the input expr should be all OR expression!
// e.g (A or B) or (C or (D or F)) => A or B or C or D or F
func flattenORExpr(expr expression.ExprCell) (exprs []expression.ExprCell) {
	// NOTE: the expr should be all_OR expr
	if expr.OP == operator.OR {
		for _, c := range expr.Content {
			exprs = append(exprs, flattenORExpr(c)...)
		}
		return exprs
	}

	return []expression.ExprCell{expr}
}

// mergeORExpressions will merge the same field and operator into one OR expression
func mergeORExpressions(exprs []expression.ExprCell) expression.ExprCell {
	type tmpMergeExpr struct {
		Operator operator.OP
		Value    []interface{}
	}

	uniqFieldValues := map[string]*tmpMergeExpr{}

	for _, expr := range exprs {
		// field in / eq are are the same
		var key string
		if expr.OP == operator.Eq {
			key = fmt.Sprintf("%s:%s", expr.Field, operator.In)
		} else {
			key = fmt.Sprintf("%s:%s", expr.Field, expr.OP)
		}

		e, ok := uniqFieldValues[key]
		if ok {
			if expr.OP == operator.In {
				value := []interface{}{}
				listValue := reflect.ValueOf(expr.Value)
				for i := 0; i < listValue.Len(); i++ {
					value = append(value, listValue.Index(i).Interface())
				}

				e.Value = append(e.Value, value...)
			} else {
				// Eq / StartsWith / other ops
				// TODO: expr.Value is an array?
				e.Value = append(e.Value, expr.Value)
			}
		} else {
			if expr.OP == operator.Eq {
				uniqFieldValues[key] = &tmpMergeExpr{
					// NOTE: to in
					Operator: operator.In,
					Value:    []interface{}{expr.Value},
				}
			} else if expr.OP == operator.StartsWith {
				// key = fmt.Sprintf("%s:%s", expr.Field, operator.StartsWith)
				uniqFieldValues[key] = &tmpMergeExpr{
					Operator: operator.StartsWith,
					Value:    []interface{}{expr.Value},
				}
			} else if expr.OP == operator.In {
				// key = fmt.Sprintf("%s:%s", expr.Field, operator.In)
				value := []interface{}{}
				listValue := reflect.ValueOf(expr.Value)
				for i := 0; i < listValue.Len(); i++ {
					value = append(value, listValue.Index(i).Interface())
				}
				uniqFieldValues[key] = &tmpMergeExpr{
					Operator: operator.In,
					Value:    value,
				}
			} else {
				// TODO: expr.Value is an array?
				uniqFieldValues[key] = &tmpMergeExpr{
					Operator: expr.OP,
					Value:    []interface{}{expr.Value},
				}
			}
		}
	}
	// TODO: 如果field出现两次并使用两个不同的operator(any除外, 那么这里无法处理)
	if len(uniqFieldValues) == 1 {
		for k, v := range uniqFieldValues {
			parts := strings.Split(k, ":")
			return expression.ExprCell{
				OP:    v.Operator,
				Field: parts[0],
				Value: v.Value,
			}

		}
	}

	content := make([]expression.ExprCell, 0, len(uniqFieldValues))
	for k, e := range uniqFieldValues {
		parts := strings.Split(k, ":")

		content = append(content, expression.ExprCell{
			OP:    e.Operator,
			Field: parts[0],
			// NOTE: 这里, 所有操作符的value都是列表
			Value: e.Value,
		})
	}

	// TODO: if got one any, just return any
	return expression.ExprCell{
		OP:      operator.OR,
		Content: content,
	}
}

// flattenAndMergeAllOR ...
func flattenAndMergeAllOR(expr expression.ExprCell) expression.ExprCell {
	// (A or B) or (C or (D or F)) => A or B or C or D or F
	if !isAllOR(&expr) {
		return expr
	}

	// 1. flatten
	// should be all_OR expression
	flattenExprs := flattenORExpr(expr)

	// 2. do merge the op with the same field+operator
	mergedExpr := mergeORExpressions(flattenExprs)
	return mergedExpr
}
