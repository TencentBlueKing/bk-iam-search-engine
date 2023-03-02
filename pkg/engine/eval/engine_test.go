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

package eval

// func BenchmarkSortPoliciesRaw(b *testing.B) {
//
// 	count := 1000000
// 	policies := make(map[int64]*types.Policy, count)
// 	for i := 0; i < count; i++ {
// 		policies[int64(i)] = &types.Policy{
// 			ID:               int64(i),
// 			Subject:          types.Subject{},
// 			ExpressionLength: rand.Intn(100),
// 		}
// 	}
//
// 	pToSort := make(policyArray, 0, count)
// 	for _, p := range policies {
// 		pToSort = append(pToSort, p)
// 	}
//
// 	for i := 0; i < b.N; i++ {
// 		sort.Sort(pToSort)
// 	}
// }
// func BenchmarkSortPoliciesBySlice(b *testing.B) {
// 	count := 1000000
// 	policies := make(map[int64]*types.Policy, count)
// 	for i := 0; i < count; i++ {
// 		policies[int64(i)] = &types.Policy{
// 			ID:               int64(i),
// 			Subject:          types.Subject{},
// 			ExpressionLength: rand.Intn(100),
// 		}
// 	}
//
// 	pToSort := make([]*types.Policy, 0, count)
// 	for _, p := range policies {
// 		pToSort = append(pToSort, p)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		sort.Slice(pToSort, func(a, b int) bool {
// 			return pToSort[a].ExpressionLength < pToSort[b].ExpressionLength
// 		})
//
// 	}
// }
