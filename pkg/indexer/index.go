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

package indexer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	log "github.com/sirupsen/logrus"

	"engine/pkg/client"
	"engine/pkg/config"
	"engine/pkg/engine/doc"
	"engine/pkg/engine/eval"
	"engine/pkg/expression"
	"engine/pkg/logging/debug"
	"engine/pkg/types"
)

// Index ...
type Index struct {
	EsEngine   types.Engine
	EvalEngine types.Engine
}

// NewIndex ...
func NewIndex(cfg *config.Index) (*Index, error) {
	var err error

	/*
		NOTE: 总共有一下几种引擎
		1. 本地Eval 用于 not searchable 策略
		2. ES 用于 any 与 search able 策略
	*/
	// doc engine ES
	var esEngine types.Engine
	esEngine, err = doc.NewEsEngine(cfg)
	if err != nil {
		return nil, err
	}

	// 简单策略: id in/startwith _bk_iam_path_
	var evalEngine types.Engine
	evalEngine, err = eval.NewEvalEngine()
	if err != nil {
		return nil, err
	}

	return &Index{
		EsEngine:   esEngine,
		EvalEngine: evalEngine,
	}, nil
}

// Search ...
func (i *Index) Search(ctx context.Context, req *types.SearchRequest, entry *debug.Entry) ([]types.Subject, error) {
	subjects := make([]types.Subject, 0, 5)
	allowedSubjectUIDs := set.NewFixedLengthStringSet(10)

	// 记录debug上下文
	debug.WithValues(entry, types.H{
		"system":       req.System,
		"action":       req.Action,
		"resource":     req.Resource,
		"subject_type": req.SubjectType,
	})

	// TODO: did timeout
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	/*
		NOTE: 总共有一下几种策略
		1. 直接就是Any的策略
		2. 直接就是 eq/in/startwith _bk_iam_path_ 的策略
		3. 全是OR, 展开以后是 eq/in/startwith _bk_iam_path_ 组合的策略
		4. 其它策略
	*/

	debug.AddStep(entry, "execute es query")
	esResult, err := i.EsEngine.Search(ctx, req, entry)
	if err != nil {
		return nil, err
	}
	subjects = append(subjects, esResult.GetSubjects(allowedSubjectUIDs)...)

	// reach the limit, truncate and return
	if types.ResourceCountReachLimit(req, allowedSubjectUIDs) {
		return subjects[:req.Limit], nil
	}

	// 3. search toEval
	debug.AddStep(entry, "execute eval policies")
	evalResult, err := i.EvalEngine.Search(ctx, req, entry)
	if err != nil {
		return nil, err
	}
	subjects = append(subjects, evalResult.GetSubjects(allowedSubjectUIDs)...)

	// reach the limit, truncate and return
	if types.ResourceCountReachLimit(req, allowedSubjectUIDs) {
		return subjects[:req.Limit], nil
	}

	return subjects, nil
}

// Stats ...
func (i *Index) Stats(system, action string) map[string]uint64 {
	docSize := i.EsEngine.Size(system, action)
	evalSize := i.EvalEngine.Size(system, action)

	return map[string]uint64{
		"total": docSize + evalSize,
		"doc":   docSize,
		"eval":  evalSize,
	}
}

// BulkUpsert ...
func (i *Index) BulkUpsert(policies []types.Policy, logger *log.Entry) {
	evalPolicies, esPolicies := expression.SplitPoliciesWithExpressionType(policies)

	evalPolicyIDs := make([]int64, 0, len(evalPolicies))
	for _, p := range evalPolicies {
		evalPolicyIDs = append(evalPolicyIDs, p.ID)
	}

	err := i.EvalEngine.BulkAdd(evalPolicies)
	if err != nil {
		logger.WithError(err).Error("indexer BulkUpsert EvalEngine.BulkAdd error")
	}

	if len(evalPolicyIDs) > 0 {
		// delete eval policies from es engine
		err = i.EsEngine.BulkDelete(evalPolicyIDs, logger)
		if err != nil {
			logger.WithError(err).Error("indexer BulkUpsert EsEngine.BulkDelete error")
		}
	}

	esPolicyIDs := make([]int64, 0, len(esPolicies))
	for _, p := range esPolicies {
		esPolicyIDs = append(esPolicyIDs, p.ID)
	}

	if len(esPolicies) > 0 {
		err = i.EsEngine.BulkAdd(esPolicies)
		if err != nil {
			logger.WithError(err).Error("indexer BulkUpsert EsEngine.BulkAdd error")
		}
	}

	if len(esPolicyIDs) > 0 {
		// delete es policies from eval engine
		err = i.EvalEngine.BulkDelete(esPolicyIDs, logger)
		if err != nil {
			logger.WithError(err).Error("indexer BulkUpsert EvalEngine.BulkDelete error")
		}
	}
}

// BulkDelete ...
func (i *Index) BulkDelete(ids []int64, logger *log.Entry) {
	if len(ids) == 0 {
		return
	}
	err := i.EsEngine.BulkDelete(ids, logger)
	if err != nil {
		logger.WithError(err).Error("indexer BulkDelete EsEngine.BulkDelete error")
	}
	err = i.EvalEngine.BulkDelete(ids, logger)
	if err != nil {
		logger.WithError(err).Error("indexer BulkUpsert EvalEngine.BulkDelete error")
	}
}

// BulkDeleteBySubjects ...
func (i *Index) BulkDeleteBySubjects(beforeUpdatedAt int64, subjects []types.Subject, logger *log.Entry) {
	if len(subjects) == 0 {
		return
	}
	err := i.EsEngine.BulkDeleteBySubjects(beforeUpdatedAt, subjects, logger)
	if err != nil {
		logger.WithError(err).Error("indexer BulkDeleteBySubjects EsEngine.BulkDeleteBySubjects error")
	}
	err = i.EvalEngine.BulkDeleteBySubjects(beforeUpdatedAt, subjects, logger)
	if err != nil {
		logger.WithError(err).Error("indexer BulkDeleteBySubjects EvalEngine.BulkDeleteBySubjects error")
	}
}

// TotalStats ...
func (i *Index) TotalStats() map[string]uint64 {
	docSize := i.EsEngine.Total()
	evalSize := i.EvalEngine.Total()

	return map[string]uint64{
		"total": docSize + evalSize,
		"doc":   docSize,
		"eval":  evalSize,
	}
}

// BatchSearch ...
func (i *Index) BatchSearch(
	ctx context.Context,
	requests []*types.SearchRequest,
	entry *debug.Entry,
) ([][]types.Subject, error) {
	// TODO: did timeout
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	esSearchResults, err := i.EsEngine.BatchSearch(ctx, requests, entry)
	if err != nil {
		return nil, err
	}

	results := make([][]types.Subject, 0, len(requests))

	for idx, req := range requests {
		subjects := make([]types.Subject, 0, 5)
		allowedSubjectUIDs := set.NewFixedLengthStringSet(10)

		esQuerySubjects := esSearchResults[idx]
		subjects = append(subjects, esQuerySubjects.GetSubjects(allowedSubjectUIDs)...)

		// reach the limit, truncate
		if types.ResourceCountReachLimit(req, allowedSubjectUIDs) {
			results = append(results, subjects[:req.Limit])
			continue
		}

		subEntry := debug.GetSubEntryByIndex(entry, idx)
		evaResult, err := i.EvalEngine.Search(ctx, req, subEntry)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, evaResult.GetSubjects(allowedSubjectUIDs)...)

		// reach the limit, truncate
		if types.ResourceCountReachLimit(req, allowedSubjectUIDs) {
			results = append(results, subjects[:req.Limit])
			continue
		}

		results = append(results, subjects)
	}

	return results, nil
}

func creatIndexIfNotExists(cfg *config.Index) error {
	esClient, err := client.NewEsClient(&cfg.ElasticSearch)
	if err != nil {
		return fmt.Errorf("new es client error:%w", err)
	}

	resp, err := esClient.IndexExists(cfg.ElasticSearch.IndexName)
	if err != nil {
		return fmt.Errorf("query index: [%s] exists error:%w", cfg.ElasticSearch.IndexName, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// 动态mapping, match string 类型时索引转换为 keyword 类型
		mapping := `{
			"mappings": {
				"dynamic_templates": [
					{
						"strings": {
							"match_mapping_type": "string",
							"mapping": {
								"type": "keyword"
							}
						}
					}
				]
			}
		}`
		_, err = esClient.CreateIndex(cfg.ElasticSearch.IndexName, mapping)
		if err != nil {
			return fmt.Errorf("create index: [%s] error:%w", cfg.ElasticSearch.IndexName, err)
		}
	}

	return nil
}
