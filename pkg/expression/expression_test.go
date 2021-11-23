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
	"github.com/TencentBlueKing/iam-go-sdk/expression"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Expression", func() {

	var anyExpr expression.ExprCell
	var eqExpr expression.ExprCell
	var inExpr expression.ExprCell
	var startsWithExpr expression.ExprCell
	var andExpr expression.ExprCell
	var orExpr expression.ExprCell
	var gtExpr expression.ExprCell

	BeforeEach(func() {
		anyExpr = expression.ExprCell{
			OP:      operator.Any,
			Content: nil,
			Field:   "",
			Value:   nil,
		}
		eqExpr = expression.ExprCell{
			OP:    operator.Eq,
			Field: "biz.id",
			Value: "123",
		}
		inExpr = expression.ExprCell{
			OP:    operator.In,
			Field: "biz.id",
			Value: []interface{}{"456"},
		}
		startsWithExpr = expression.ExprCell{
			OP:    operator.StartsWith,
			Field: "host._bk_iam_path_",
			Value: "/biz,1/set,*/",
		}

		andExpr = expression.ExprCell{
			OP:      operator.AND,
			Content: []expression.ExprCell{eqExpr, startsWithExpr},
		}

		orExpr = expression.ExprCell{
			OP:      operator.OR,
			Content: []expression.ExprCell{eqExpr, startsWithExpr},
		}

		gtExpr = expression.ExprCell{
			OP:    operator.Gt,
			Field: "box.size",
			Value: "123",
		}

	})

	Describe("indexer.isAny", func() {
		It("true", func() {
			assert.True(GinkgoT(), isAny(&anyExpr))
		})

		It("false", func() {
			assert.False(GinkgoT(), isAny(&eqExpr))
			assert.False(GinkgoT(), isAny(&orExpr))
		})
	})

	Describe("indexer.isSingleEqOrIn", func() {
		Describe("true", func() {
			It("eq", func() {
				assert.True(GinkgoT(), isSingleEqOrIn(&eqExpr))
			})
			It("in", func() {
				assert.True(GinkgoT(), isSingleEqOrIn(&inExpr))
			})
		})

		Describe("false", func() {

			It("starts_with", func() {
				assert.False(GinkgoT(), isSingleEqOrIn(&startsWithExpr))
			})
			It("any", func() {
				assert.False(GinkgoT(), isSingleEqOrIn(&anyExpr))
			})
			It("AND", func() {
				assert.False(GinkgoT(), isSingleEqOrIn(&andExpr))
			})
		})
	})

	Describe("indexer.isBkIAMPathStartsWith", func() {
		It("true", func() {
			assert.True(GinkgoT(), isBkIAMPathStartsWith(&startsWithExpr))

		})

		It("false", func() {
			assert.False(GinkgoT(), isBkIAMPathStartsWith(&eqExpr))
			assert.False(GinkgoT(), isBkIAMPathStartsWith(&inExpr))
			assert.False(GinkgoT(), isBkIAMPathStartsWith(&anyExpr))
			assert.False(GinkgoT(), isBkIAMPathStartsWith(&andExpr))

		})
	})

	Describe("indexer.isAllOR", func() {

		Describe("false", func() {
			It("OP.AND", func() {
				assert.False(GinkgoT(), isAllOR(&andExpr))
			})
			It("OP.Any", func() {
				assert.False(GinkgoT(), isAllOR(&anyExpr))
			})
		})

		Describe("OP.OR", func() {
			It("true, simple", func() {
				expr := expression.ExprCell{
					OP:      operator.OR,
					Content: []expression.ExprCell{eqExpr, startsWithExpr},
				}
				assert.True(GinkgoT(), isAllOR(&expr))
			})

			It("true, nested", func() {
				expr := expression.ExprCell{
					OP:      operator.OR,
					Content: []expression.ExprCell{eqExpr, startsWithExpr, orExpr},
				}
				assert.True(GinkgoT(), isAllOR(&expr))
			})

			It("false", func() {
				expr := expression.ExprCell{
					OP:      operator.OR,
					Content: []expression.ExprCell{eqExpr, andExpr},
				}
				assert.False(GinkgoT(), isAllOR(&expr))
			})

		})

		It("OP.others", func() {
			assert.True(GinkgoT(), isAllOR(&eqExpr))
			assert.True(GinkgoT(), isAllOR(&inExpr))
			assert.True(GinkgoT(), isAllOR(&startsWithExpr))
		})
	})

	Describe("indexer.isSameObjectWithDifferentFields", func() {
		It("not op.OR", func() {
			assert.False(GinkgoT(), isSameObjectWithDifferentFields(&eqExpr))
			assert.False(GinkgoT(), isSameObjectWithDifferentFields(&andExpr))
		})

		Describe("is op.OR", func() {
			It("not the same object", func() {
				expr := expression.ExprCell{
					OP: operator.OR,
					Content: []expression.ExprCell{
						{
							OP:    operator.Eq,
							Field: "biz.id",
							Value: "123",
						},
						{
							OP:    operator.Eq,
							Field: "set.id",
							Value: "456",
						},
					},
				}

				assert.False(GinkgoT(), isSameObjectWithDifferentFields(&expr))
			})

			It("same object, not support op", func() {
				expr := expression.ExprCell{
					OP: operator.OR,
					Content: []expression.ExprCell{
						{
							OP:    operator.Eq,
							Field: "biz.id",
							Value: "123",
						},
						{
							OP:    operator.Gt,
							Field: "biz.size",
							Value: "456",
						},
					},
				}
				assert.False(GinkgoT(), isSameObjectWithDifferentFields(&expr))
			})

			It("same object, starts_with field not _bk_iam_path_ ", func() {
				expr := expression.ExprCell{
					OP: operator.OR,
					Content: []expression.ExprCell{
						{
							OP:    operator.Eq,
							Field: "biz.id",
							Value: "123",
						},
						{
							OP:    operator.StartsWith,
							Field: "biz.mypath",
							Value: "456",
						},
					},
				}
				assert.False(GinkgoT(), isSameObjectWithDifferentFields(&expr))
			})

			It("same object", func() {
				expr := expression.ExprCell{
					OP: operator.OR,
					Content: []expression.ExprCell{
						{
							OP:    operator.Eq,
							Field: "biz.id",
							Value: "123",
						},
						{
							OP:    operator.Eq,
							Field: "biz.attr",
							Value: "456",
						},
					},
				}

				assert.True(GinkgoT(), isSameObjectWithDifferentFields(&expr))
			})
		})

	})

	Describe("indexer.flattenORExpr", func() {
		It("simple", func() {
			result := flattenORExpr(eqExpr)
			assert.Equal(GinkgoT(), []expression.ExprCell{eqExpr}, result)

			result = flattenORExpr(inExpr)
			assert.Equal(GinkgoT(), []expression.ExprCell{inExpr}, result)

			result = flattenORExpr(startsWithExpr)
			assert.Equal(GinkgoT(), []expression.ExprCell{startsWithExpr}, result)
		})

		It("or no nested", func() {
			expr := expression.ExprCell{
				OP: operator.OR,
				Content: []expression.ExprCell{
					eqExpr,
					inExpr,
					startsWithExpr,
				},
			}

			result := flattenORExpr(expr)
			assert.Len(GinkgoT(), result, 3)
		})

		It("or nested", func() {
			expr := expression.ExprCell{
				OP: operator.OR,
				Content: []expression.ExprCell{
					eqExpr,
					{
						OP: operator.OR,
						Content: []expression.ExprCell{
							inExpr,
							startsWithExpr,
						},
					},
				},
			}

			result := flattenORExpr(expr)
			assert.Len(GinkgoT(), result, 3)

		})

	})

	Describe("indexer.mergeORExpressions", func() {
		// NOTE: here, case by case
		Describe("single", func() {
			It("eq", func() {
				result := mergeORExpressions([]expression.ExprCell{eqExpr})
				// NOTE: the operator turned into `in`
				assert.Equal(GinkgoT(), operator.In, result.OP)
				assert.Equal(GinkgoT(), eqExpr.Field, result.Field)
				assert.Equal(GinkgoT(), []interface{}{eqExpr.Value}, result.Value)
			})

			It("in", func() {
				result := mergeORExpressions([]expression.ExprCell{inExpr})
				assert.Equal(GinkgoT(), operator.In, result.OP)
				assert.Equal(GinkgoT(), inExpr.Field, result.Field)
				assert.Equal(GinkgoT(), inExpr.Value, result.Value)
			})

			It("starts_with", func() {
				result := mergeORExpressions([]expression.ExprCell{startsWithExpr})
				assert.Equal(GinkgoT(), operator.StartsWith, result.OP)
				assert.Equal(GinkgoT(), startsWithExpr.Field, result.Field)
				assert.Equal(GinkgoT(), []interface{}{startsWithExpr.Value}, result.Value)

			})

			It("gt", func() {
				result := mergeORExpressions([]expression.ExprCell{gtExpr})
				assert.Equal(GinkgoT(), operator.Gt, result.OP)
				assert.Equal(GinkgoT(), gtExpr.Field, result.Field)
				assert.Equal(GinkgoT(), []interface{}{gtExpr.Value}, result.Value)

			})
		})

		Describe("merged", func() {
			It("all eq", func() {
				exprs := []expression.ExprCell{
					{
						OP:    operator.Eq,
						Field: "host.id",
						Value: "123",
					},
					{
						OP:    operator.Eq,
						Field: "host.id",
						Value: "456",
					},
				}

				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.In, result.OP)
				assert.Equal(GinkgoT(), "host.id", result.Field)
				assert.Equal(GinkgoT(), []interface{}{"123", "456"}, result.Value)
			})
			It("all in", func() {
				exprs := []expression.ExprCell{
					{
						OP:    operator.In,
						Field: "host.id",
						Value: []interface{}{"123", "456"},
					},
					{
						OP:    operator.In,
						Field: "host.id",
						Value: []interface{}{"789"},
					},
				}

				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.In, result.OP)
				assert.Equal(GinkgoT(), "host.id", result.Field)
				assert.Equal(GinkgoT(), []interface{}{"123", "456", "789"}, result.Value)
			})

			// a._bk_iam_path_ starts_with x OR a._bk_iam_path starts_with y
			// => a._bk_iam_path_ starts_with [x, y]
			It("all starts_with", func() {
				exprs := []expression.ExprCell{
					{
						Field: "host._bk_iam_path_",
						OP:    operator.StartsWith,
						Value: "/biz,100609/",
					},
					{
						Field: "host._bk_iam_path_",
						OP:    operator.StartsWith,
						Value: "/biz,100609/biz_custom_query,*/",
					},
				}

				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.StartsWith, result.OP)
				assert.Equal(GinkgoT(), "host._bk_iam_path_", result.Field)
				assert.Equal(GinkgoT(), []interface{}{"/biz,100609/", "/biz,100609/biz_custom_query,*/"}, result.Value)
			})

			It("1 eq, 1 in", func() {
				exprs := []expression.ExprCell{
					{
						OP:    operator.Eq,
						Field: "host.id",
						Value: "123",
					},
					{
						OP:    operator.In,
						Field: "host.id",
						Value: []interface{}{"456", "789"},
					},
				}

				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.In, result.OP)
				assert.Equal(GinkgoT(), "host.id", result.Field)
				assert.Equal(GinkgoT(), []interface{}{"123", "456", "789"}, result.Value)
			})

		})

		Describe("unmerged", func() {

			It("1 eq, 1 starts_with", func() {
				expr1 := expression.ExprCell{
					OP:    operator.Eq,
					Field: "host.id",
					Value: "123",
				}
				wantExpr1 := expression.ExprCell{
					OP:    operator.In,
					Field: "host.id",
					Value: []interface{}{"123"},
				}

				expr2 := expression.ExprCell{
					OP:    operator.StartsWith,
					Field: "host.id",
					Value: "/biz,123/",
				}

				wantExpr2 := expression.ExprCell{
					OP:    operator.StartsWith,
					Field: "host.id",
					Value: []interface{}{"/biz,123/"},
				}

				exprs := []expression.ExprCell{
					expr1,
					expr2,
				}

				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.OR, result.OP)
				assert.Len(GinkgoT(), result.Content, 2)
				assert.Contains(GinkgoT(), result.Content, wantExpr1)
				assert.Contains(GinkgoT(), result.Content, wantExpr2)

			})

			It("different attrs", func() {
				expr1 := expression.ExprCell{
					OP:    operator.In,
					Field: "task.id",
					Value: []interface{}{"110192", "110198", "110309", "110345"},
				}
				expr2 := expression.ExprCell{
					OP:    operator.Eq,
					Field: "task.iam_resource_owner",
					Value: "aaa",
				}
				wantExpr2 := expression.ExprCell{
					OP:    operator.In,
					Field: "task.iam_resource_owner",
					Value: []interface{}{"aaa"},
				}

				exprs := []expression.ExprCell{
					expr1,
					expr2,
				}
				result := mergeORExpressions(exprs)

				assert.Equal(GinkgoT(), operator.OR, result.OP)
				assert.Len(GinkgoT(), result.Content, 2)
				assert.Contains(GinkgoT(), result.Content, expr1)
				assert.Contains(GinkgoT(), result.Content, wantExpr2)
			})

		})

	})

	Describe("indexer.flattenAndMergeAllOR", func() {

		It("not all OR", func() {
			result := flattenAndMergeAllOR(andExpr)
			assert.Equal(GinkgoT(), andExpr, result)
		})

		It("ok", func() {
			expr := expression.ExprCell{
				OP:      operator.OR,
				Content: []expression.ExprCell{eqExpr, startsWithExpr},
			}
			result := flattenAndMergeAllOR(expr)

			assert.Equal(GinkgoT(), operator.OR, result.OP)
			assert.Len(GinkgoT(), result.Content, 2)
		})

	})

})
