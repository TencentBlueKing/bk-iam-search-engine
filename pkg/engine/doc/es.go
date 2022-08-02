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
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	log "github.com/sirupsen/logrus"

	"engine/pkg/client"
	"engine/pkg/config"
	"engine/pkg/logging/debug"
	"engine/pkg/types"
)

// EsEngine ...
type EsEngine struct {
	client        *client.EsClient
	indexName     string
	lastIndexTime time.Time
}

// NewEsEngine ...
func NewEsEngine(cfg *config.Index) (types.Engine, error) {
	esClient, err := client.NewEsClient(&cfg.ElasticSearch)
	if err != nil {
		return nil, fmt.Errorf("new es client error:%w", err)
	}

	return &EsEngine{
		client:        esClient,
		indexName:     cfg.ElasticSearch.IndexName,
		lastIndexTime: time.Time{},
	}, nil
}

// Size ...
func (e *EsEngine) Size(system, action string) uint64 {
	count, err := e.getActionCount(system, action)
	if err != nil {
		return 0
	}
	return uint64(count)
}

// BulkAdd ...
func (e *EsEngine) BulkAdd(policies []*types.Policy) error {
	docs, err := e.makeDocs(policies)
	if err != nil {
		return err
	}
	err = e.client.BulkIndex(e.indexName, docs)
	if err != nil {
		return fmt.Errorf("bulk index fail: %w", err)
	}

	e.lastIndexTime = time.Now()
	return nil
}

// Search ...
func (e *EsEngine) Search(
	ctx context.Context,
	req *types.SearchRequest,
	entry *debug.Entry,
) (types.SearchResult, error) {
	queries := genQueriesByRequest(req, entry)
	r, err := e.client.Msearch(ctx, e.indexName, queries)
	if err != nil {
		return &EsSearchResult{}, fmt.Errorf("index search fail %w", err)
	}
	esQuerySubjects := getSearchResultByResponses(r["responses"].([]interface{}), entry)
	return esQuerySubjects, nil
}

func (e *EsEngine) makeDocs(policies []*types.Policy) (docs []types.H, err error) {
	docs = make([]types.H, 0, len(policies))
	for _, p := range policies {
		doc, err := makeDoc(p.ExpressionType, p)
		if err != nil {
			return nil, fmt.Errorf("make doc fail: %w", err)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// BatchSearch ...
func (e *EsEngine) BatchSearch(
	ctx context.Context,
	requests []*types.SearchRequest,
	entry *debug.Entry,
) ([]types.SearchResult, error) {
	queries := make([]types.H, 0, len(requests)*2)
	for idx, req := range requests {
		subEntry := debug.GetSubEntryByIndex(entry, idx)

		queries = append(queries, genQueriesByRequest(req, subEntry)...)
	}
	r, err := e.client.Msearch(ctx, e.indexName, queries)
	if err != nil {
		return nil, fmt.Errorf("index search fail %w", err)
	}

	searchResults := make([]types.SearchResult, 0, len(requests))

	responses := r["responses"].([]interface{})
	for i := 0; i < len(requests); i++ {
		subEntry := debug.GetSubEntryByIndex(entry, i)
		esQuerySubjects := getSearchResultByResponses(responses[i*2:i*2+2], subEntry)

		searchResults = append(searchResults, esQuerySubjects)
	}
	return searchResults, nil
}

// BulkDelete ...
func (e *EsEngine) BulkDelete(ids []int64, logger *log.Entry) error {
	docs := make([]types.H, 0, len(ids))
	for _, id := range ids {
		docs = append(docs, types.H{"id": id})
	}
	err := e.client.BulkDelete(e.indexName, docs)
	if err != nil {
		logger.WithError(err).Error("esClient.BulkDelete fail")
		return fmt.Errorf("bulk delete fail: %w", err)
	}

	e.lastIndexTime = time.Now()
	return nil
}

func (e *EsEngine) deleteByQuery(query types.H, logger *log.Entry) (err error) {
	err = e.client.DeleteByQuery(e.indexName, query)
	if err != nil {
		logger.WithError(err).WithFields(log.Fields{
			"index":      e.indexName,
			"expression": query,
		}).Error("esClient.DeleteByQuery fail")
	}
	e.lastIndexTime = time.Now()
	return
}

// BulkDeleteBySubjects ...
func (e *EsEngine) BulkDeleteBySubjects(beforeUpdatedAt int64, subjects []types.Subject, logger *log.Entry) error {
	query := genSubjectsQuery(beforeUpdatedAt, subjects)
	return e.deleteByQuery(query, logger)
}

// BulkDeleteByTemplateSubjects ...
func (e *EsEngine) BulkDeleteByTemplateSubjects(
	beforeUpdatedAt int64, templateID int64, subjects []types.Subject, logger *log.Entry,
) error {
	query := genTemplateSubjectsQuery(beforeUpdatedAt, templateID, subjects)
	return e.deleteByQuery(query, logger)
}

func (e *EsEngine) getActionCount(system, action string) (int, error) {
	query := types.H{
		"query": types.H{
			"bool": types.H{
				"filter": []interface{}{
					types.H{"term": types.H{"system": system}},
					// types.H{"term": types.H{"action.id": action}},
					types.H{
						"bool": types.H{
							"should": []types.H{
								{"action.id": action},
								{"actions.id": action},
							},
						},
					},
				},
			},
		},
	}

	return e.getCount(query)
}

// GetLastIndexTime ...
func (e *EsEngine) GetLastIndexTime() time.Time {
	return e.lastIndexTime
}

func (e *EsEngine) getCount(query types.H) (int, error) {
	result, err := e.client.Search(context.Background(), e.indexName, query, 0, 1, []string{})
	if err != nil {
		return 0, fmt.Errorf("es client search fail: %w", err)
	}

	return int(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)), nil
}

// LoadSnapshot ...
func (e *EsEngine) LoadSnapshot(data []types.SnapRecord) error {
	return nil
}

// TakeSnapshot ...
func (e *EsEngine) TakeSnapshot() []types.SnapRecord {
	return nil
}

// Total ...
func (e *EsEngine) Total() uint64 {
	query := types.H{
		"query": types.H{
			"match_all": types.H{},
		},
	}

	count, err := e.getCount(query)
	if err != nil {
		return 0
	}

	return uint64(count)
}

func genQuery(genFn esSearchQueryFunc, req *types.SearchRequest, size int) (query types.H) {
	query = genFn(req)
	// NOTE: 为了兼容批量查询, 如果查询的query返回nil, 则匹配一个占位为none的语句
	if query == nil {
		return types.H{
			"query": types.H{
				"match_none": types.H{},
			},
		}
	}

	query["from"] = 0
	query["size"] = size
	query["_source"] = "subject"
	query["track_total_hits"] = "true"
	return
}

func genQueriesByRequest(req *types.SearchRequest, entry *debug.Entry) []types.H {
	pageSize := 100
	if req.Limit > 0 {
		pageSize = req.Limit
	}

	queries := make([]types.H, 0, 2)

	// gen any query
	anyQuery := genQuery(genAnyQuery, req, pageSize)
	debug.WithEsQuery(entry, string(types.Any), anyQuery)
	queries = append(queries, anyQuery)

	// gen doc query
	docQuery := genQuery(genDocQuery, req, pageSize)
	debug.WithEsQuery(entry, string(types.Doc), docQuery)
	queries = append(queries, docQuery)

	return queries
}

// EsSearchResult ES 查询结果
type EsSearchResult struct {
	anySubjects []types.Subject
	docSubjects []types.Subject
}

// GetSubjects ...
func (e *EsSearchResult) GetSubjects(allowedSubjectUIDs *set.StringSet) []types.Subject {
	subjects := make([]types.Subject, 0, len(e.anySubjects)+len(e.docSubjects))
	for _, subject := range append(e.anySubjects, e.docSubjects...) {
		if allowedSubjectUIDs.Has(subject.UID) {
			continue
		}

		subjects = append(subjects, subject)
		allowedSubjectUIDs.Add(subject.UID)
	}
	return subjects
}

func getSearchResultByResponses(responses []interface{}, entry *debug.Entry) *EsSearchResult {
	esQuerySubjects := &EsSearchResult{}

	for i, r := range responses {
		subjects := make([]types.Subject, 0, 10)

		for _, hit := range r.(map[string]interface{})["hits"].(map[string]interface{})["hits"].([]interface{}) {
			hitDoc := hit.(map[string]interface{})

			source := hitDoc["_source"].(map[string]interface{})
			subject := source["subject"].(map[string]interface{})
			subjectUID := subject["uid"].(string)
			s := types.Subject{
				Type: subject["type"].(string),
				ID:   subject["id"].(string),
				Name: subject["name"].(string),
				UID:  subjectUID,
			}

			subjects = append(subjects, s)
		}

		switch i {
		case 0:
			esQuerySubjects.anySubjects = subjects
			debug.WithEsQuerySubjects(entry, string(types.Any), subjects)
		case 1:
			esQuerySubjects.docSubjects = subjects
			debug.WithEsQuerySubjects(entry, string(types.Doc), subjects)
		}
	}
	return esQuerySubjects
}
