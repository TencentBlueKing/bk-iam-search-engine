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

	"engine/pkg/types"
)

// SplitPoliciesWithExpressionType ...
func SplitPoliciesWithExpressionType(policies []types.Policy) (
	evalPolicies []*types.Policy,
	esPolicies []*types.Policy) {

	if len(policies) == 0 {
		return
	}

	esPolicies = make([]*types.Policy, 0, len(policies))

	// TODO: 确保这里的判断及合并的正确性!!!!!!
	// NOTE: any + eval engine add/batchAdd will return error = nil
	for _, p := range policies {
		p := p

		// NOTE: 这里必须初始化填充 uid, 方便后面计算不需要重复获取 => TODO: 更新index也需要初始化!!!
		err := p.FillUniqueFields()
		if err != nil {
			// TODO: 这里单条报错怎么处理?
			fmt.Println("err:", err)
			continue
		}

		// 1. 单个any, 特殊; 直接获得权限, 不需要计算
		// NOTE: 这里是否也需要批量add
		if isAny(&p.Expression) {
			p.ExpressionType = types.Any
			esPolicies = append(esPolicies, &p)
			continue
		}

		// 2. `biz.id eq 1` 及 `biz.id in [1,2,3]`;   `biz._bk_iam_path_ starts_with x`
		if isSingleEqOrIn(&p.Expression) || isBkIAMPathStartsWith(&p.Expression) {
			p.ExpressionType = types.Doc
			esPolicies = append(esPolicies, &p)
			continue
		}

		// NOTE: 这里的逻辑很复杂了, 需要保证正确性!!!!!!
		// 3. 如果是 `(A or B) or (C or (D or F))` => `A or B or C or D or F`
		if isAllOR(&p.Expression) {
			expr := flattenAndMergeAllOR(p.Expression)

			// 3.1 打平后如果是  eq / in / _bk_iam_path_ starts_with
			if isSingleEqOrIn(&expr) || isBkIAMPathStartsWith(&expr) {
				p.Expression = expr
				p.ExpressionType = types.Doc
				esPolicies = append(esPolicies, &p)
				continue
			}

			// 3.2 same object with different values in all OR content
			// file相同, op为 in startwith _bk_iam_path_也可以搜索
			// 注意这里有个前提, isAllOR, 然后打平->合并后的表达式, 才能保证正确性
			if isSameObjectWithDifferentFields(&expr) {
				p.Expression = expr
				p.ExpressionType = types.Doc
				esPolicies = append(esPolicies, &p)
				continue
			}
		}

		// 暂时不支持的表达式, 全部需要执行
		// 1. 所有包含and, 包括跨系统资源依赖的 and 关系
		p.ExpressionType = types.Eval
		evalPolicies = append(evalPolicies, &p)
	}

	return
}
