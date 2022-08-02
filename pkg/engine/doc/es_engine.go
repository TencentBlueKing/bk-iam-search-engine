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
	"fmt"

	"engine/pkg/types"
)

// NOTE 当前ES的索引都是动态mapping, 所以text类型字段在默认查询时都会分词, 但是term要求精确匹配, 所以必须加上keyword
// https://segmentfault.com/q/1010000017312707
func genDocQuery(req *types.SearchRequest) types.H {
	var sqs []types.H

	system := req.System
	action := req.Action.ID

	// NOTE: 这里req.Resource可能是多个, 理论上, 跨系统资源依赖不应该出现到doc search
	for _, resourceNode := range req.Resource {
		for key, value := range resourceNode.Attribute {
			// 如果key是 _bk_iam_path_ 那么需要 拆分成多个 OR 关系去直接匹配(不走前缀匹配)
			if key == types.BkIAMPathKey {
				// if  x._bk_iam_path_ starts_with /biz,1/cluster,2/ =>  x._bk_iam_path_ in [/biz,1/, /biz,1/cluster,2/]
				paths := generateBkIAMPathList(value)
				term := fmt.Sprintf("resource.%s.%s.%s", system, resourceNode.Type, key)
				for _, path := range paths {
					sqs = append(sqs, types.H{
						"term": types.H{term: path},
					})
				}

				// if x._bk_iam_path_ starts_with /biz,1/cluster,2/ =>  x._bk_iam_path_contains_ in [/biz,1/, /cluster,2/]
				nodes := generateBkIAMPathNodes(value)
				containsTerm := fmt.Sprintf("resource.%s.%s.%s", system, resourceNode.Type, types.BkIAMPathContainsKey)
				for _, node := range nodes {
					sqs = append(sqs, types.H{
						"term": types.H{containsTerm: node},
					})
				}

				continue
			}

			fieldName := fmt.Sprintf("resource.%s.%s.%s", system, resourceNode.Type, key)
			sqs = append(sqs, types.H{
				"term": types.H{fieldName: value},
			})

		}
	}

	var subQuery types.H
	switch len(sqs) {
	case 0:
		return nil
	case 1:
		subQuery = sqs[0]
	default:
		subQuery = types.H{
			"bool": types.H{
				"should": sqs,
			},
		}
	}

	// action.id = create or actions.id = create
	actionSubQuery := types.H{
		"bool": types.H{
			"should": []types.H{
				{
					"action.id": action,
				},
				{
					"actions.id": action,
				},
			},
		},
	}

	must := []interface{}{
		subQuery,
		types.H{
			"range": types.H{
				"expired_at": types.H{
					"gte": req.NowTimestamp,
				},
			},
		},
		types.H{"term": types.H{"system": system}},
		actionSubQuery,
		// types.H{"term": types.H{"action.id": action}},
		types.H{"term": types.H{"type": string(types.Doc)}},
	}

	// filter the subject not the req.SubjectType
	if req.SubjectType != types.SubjectTypeAll {
		must = append(must, types.H{
			"term": types.H{"subject.type": req.SubjectType},
		})
	}

	// subQuery AND expired_at > now
	query := types.H{
		"query": types.H{
			"bool": types.H{
				"must": must,
			},
		},
	}

	return query
}

// genSubjectsQuery ...
func genSubjectsQuery(timestamp int64, subjects []types.Subject) types.H {
	subQuery := genSubjectsBoolCondition(subjects)
	query := types.H{
		"query": types.H{
			"bool": types.H{
				"must": []interface{}{
					subQuery,
					types.H{
						"range": types.H{
							"updated_at": types.H{
								"lt": timestamp,
							},
						},
					},
				},
			},
		},
	}

	return query
}

// genTemplateSubjectsQuery ...
func genTemplateSubjectsQuery(timestamp int64, templateID int64, subjects []types.Subject) types.H {
	subQuery := genSubjectsBoolCondition(subjects)
	query := types.H{
		"query": types.H{
			"bool": types.H{
				"must": []interface{}{
					types.H{"term": types.H{"template_id": templateID}},
					subQuery,
					types.H{
						"range": types.H{
							"updated_at": types.H{
								"lt": timestamp,
							},
						},
					},
				},
			},
		},
	}

	return query
}

func genSubjectsBoolCondition(subjects []types.Subject) types.H {
	var sqs []types.H
	for _, subject := range subjects {
		sq := types.H{
			"bool": types.H{
				"must": []interface{}{
					types.H{"term": types.H{"subject.type": subject.Type}},
					types.H{"term": types.H{"subject.id": subject.ID}},
				},
			},
		}
		sqs = append(sqs, sq)
	}

	var subQuery types.H
	if len(sqs) == 1 {
		subQuery = sqs[0]
	} else {
		// a OR b OR c
		subQuery = types.H{
			"bool": types.H{
				"should": sqs,
			},
		}
	}

	return subQuery
}

func genAnyQuery(req *types.SearchRequest) types.H {
	system := req.System
	action := req.Action.ID

	filter := []interface{}{
		types.H{
			"range": types.H{
				"expired_at": types.H{
					"gte": req.NowTimestamp,
				},
			},
		},
		types.H{"term": types.H{"system": system}},
		types.H{"term": types.H{"action.id": action}},
		types.H{"term": types.H{"type": string(types.Any)}},
	}

	// filter the subject not the req.SubjectType
	if req.SubjectType != types.SubjectTypeAll {
		filter = append(filter, types.H{
			"term": types.H{"subject.type": req.SubjectType},
		})
	}

	// use the `filter` replace the `must`
	query := types.H{
		"query": types.H{
			"bool": types.H{
				"filter": filter,
			},
		},
	}
	return query
}
