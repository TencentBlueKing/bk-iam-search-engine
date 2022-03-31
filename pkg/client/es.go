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

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/getsentry/sentry-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/errorx"
	"engine/pkg/logging"
	"engine/pkg/metric"
	"engine/pkg/types"
	"engine/pkg/util"
)

const slowRequestSeconds = 2

// EsClient ...
type EsClient struct {
	client *elasticsearch.Client
}

// NewEsClient ...
func NewEsClient(cfg *config.ElasticSearch) (*EsClient, error) {
	// client, err := elasticsearch.NewDefaultClient()
	// if err != nil {
	// 	err = fmt.Errorf("error creating the client: %w", err)
	// 	return nil, err
	// }
	retryBackoff := backoff.NewExponentialBackOff()
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:  cfg.Addresses,
		Username:   cfg.Username,
		Password:   cfg.Password,
		MaxRetries: cfg.MaxRetries,

		// Retry on 429 TooManyRequests statuses
		RetryOnStatus: []int{502, 503, 504, 429},

		EnableRetryOnTimeout: true,

		// Configure the backoff function
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//MaxRetries: 3,

		// EnableDebugLogger: true,
	})
	if err != nil {
		err = fmt.Errorf("error creating the client: %w", err)
		return nil, err
	}

	return &EsClient{
		client,
	}, nil

}

// NewEsPingClient ...
func NewEsPingClient(cfg *config.ElasticSearch) (*EsClient, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		err = fmt.Errorf("error creating the client: %w", err)
		return nil, err
	}

	return &EsClient{
		client,
	}, nil

}

// Index ...
func (c *EsClient) Index(indexName string, docID string, doc map[string]interface{}) error {
	// speed up the marshal
	data, err := jsoniter.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc fail: %w", err)
	}

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("[%s] Error indexing document ID=%s", res.Status(), docID)
	}
	return nil
}

// DeleteByQuery ...
func (c *EsClient) DeleteByQuery(indexName string, query types.H) error {
	// speed up the marshal
	data, err := jsoniter.Marshal(query)
	if err != nil {
		return fmt.Errorf("marshal doc fail: %w", err)
	}

	refresh := true
	// Set up the request object.
	req := esapi.DeleteByQueryRequest{
		Index:   []string{indexName},
		Body:    bytes.NewReader(data),
		Refresh: &refresh,
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("[%s] Error delete by query", res.Status())
	}
	return nil
}

// BulkIndex ...
func (c *EsClient) BulkIndex(indexName string, docs []types.H) error {
	return c.bulk(indexName, docs, "index")
}

// BulkDelete ...
func (c *EsClient) BulkDelete(indexName string, docs []types.H) error {
	return c.bulk(indexName, docs, "delete")
}

func (c *EsClient) bulk(indexName string, docs []types.H, action string) error {
	logger := logging.GetESLogger()
	// Create the BulkIndexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:  indexName, // The default Index name
		Client: c.client,  // The Elasticsearch client
		// NumWorkers:    numWorkers,       // The number of worker goroutines
		// FlushBytes:    int(flushBytes),  // The flush threshold in bytes
		NumWorkers:    runtime.NumCPU(),
		FlushBytes:    5e+6,             // 5M
		FlushInterval: 30 * time.Second, // The periodic flush interval
		OnError: func(ctx context.Context, err error) {
			if err != nil {
				logger.WithError(err).Error("bulk index err")

				ev := sentry.NewEvent()
				ev.Message = "bulk index error"
				ev.Level = "error"
				ev.Timestamp = time.Now()
				ev.Extra = map[string]interface{}{
					"error": err,
				}
				errorx.ReportEvent(ev)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("error creating the indexer: %w", err)
	}

	start := time.Now().UTC()

	var numNotFound int64 = 0 // 用于记录action为delete时出现的404的数量
	// Loop over the collection
	for _, d := range docs {
		var body io.Reader = nil
		if action != "delete" {
			data, err := jsoniter.Marshal(d)
			if err != nil {
				return fmt.Errorf("marshal doc fail: %w", err)
			}
			body = bytes.NewReader(data)
		}

		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (Index, create, delete, update)
				Action: action,

				// DocumentID is the (optional) document ID
				DocumentID: strconv.FormatInt(d["id"].(int64), 10), // strconv.Itoa(a.ID),

				// Body is an `io.Reader` with the payload
				Body: body,
				// Body: esutil.NewJSONReader(d),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					// atomic.AddUint64(&countSuccessful, 1)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						logger.WithError(err).Errorf("elasticsearch index error: item=`%+v`", item)

						ev := sentry.NewEvent()
						ev.Message = "elasticsearch index error"
						ev.Level = "error"
						ev.Timestamp = time.Now()
						ev.Extra = map[string]interface{}{
							"item":     item,
							"response": res,
							"error":    err,
						}
						errorx.ReportEvent(ev)
					} else {
						if res.Error.Type != "" ||
							res.Error.Reason != "" ||
							(res.Error.Cause.Type != "" || res.Error.Cause.Reason != "" ||
								(res.Status > http.StatusCreated && res.Status != http.StatusNotFound)) {
							logger.Errorf("elasticsearch index response error: item=`%+v`, response=`%+v`", item, res)

							ev := sentry.NewEvent()
							ev.Message = "elasticsearch index response error"
							ev.Level = "error"
							ev.Timestamp = time.Now()
							ev.Extra = map[string]interface{}{
								"item":     item,
								"response": res,
							}
							errorx.ReportEvent(ev)
						}

						// 删除时如果有404 记录数量
						if res.Status == http.StatusNotFound {
							numNotFound += 1
						}
					}
				},
			},
		)
		if err != nil {
			return fmt.Errorf("unexpected error: %w", err)
		}
	}

	// TODO: 如果有一个失败了怎么办?  what if errors > 0

	// close the indexer
	if err := bi.Close(context.Background()); err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}

	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	biStats := bi.Stats()
	dur := time.Since(start)
	if int64(biStats.NumFailed)-numNotFound > 0 {
		logger.Errorf(
			"Indexed [%d] documents with [%d] errors in %s (%d docs/sec)",
			int64(biStats.NumFlushed),
			int64(biStats.NumFailed)-numNotFound, // 删除时如果有404不计入失败数
			dur.Truncate(time.Millisecond),
			int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed)),
		)
	} else {
		logger.Infof(
			"Sucessfuly indexed [%d] documents in %s (%d docs/sec)",
			int64(biStats.NumFlushed),
			dur.Truncate(time.Millisecond),
			int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed)),
		)
	}
	return nil
}

func (c *EsClient) search(
	ctx context.Context,
	indexName string,
	query types.H, from, pageSize int,
	fields []string,
) (res *esapi.Response, err error) {
	// Build the request body.
	var buf bytes.Buffer
	if err = jsoniter.NewEncoder(&buf).Encode(query); err != nil {
		err = fmt.Errorf("encode query fail:%w", err)
		return
	}

	start := time.Now()
	res, err = c.client.Search(
		c.client.Search.WithContext(context.Background()),
		c.client.Search.WithIndex(indexName),
		c.client.Search.WithBody(&buf),
		c.client.Search.WithTrackTotalHits(true),
		c.client.Search.WithPretty(),
		c.client.Search.WithFrom(from),
		c.client.Search.WithSize(pageSize),
		c.client.Search.WithSource(fields...),
	)

	duration := time.Since(start)
	metric.EsSearchDuration.Observe(float64(duration / time.Millisecond))

	// gl 2 second
	if (duration / time.Second) > slowRequestSeconds {
		entry := logging.GetESLogger().WithField("request_id", ctx.Value(util.ContextRequestIDKey))

		durationMilliseconds := float64(duration/time.Millisecond) + 1

		entry.WithFields(logrus.Fields{
			"index":      indexName,
			"expression": query,
			"from":       from,
			"page_size":  pageSize,
			"fields":     fields,
			"duration":   durationMilliseconds,
		}).Warning("-")

		util.ReportToSentry("es client search slow request", map[string]interface{}{
			"index":      indexName,
			"expression": query,
			"from":       from,
			"page_size":  pageSize,
			"fields":     fields,
			"duration":   durationMilliseconds,
		})
	}

	return res, err
}

func (c *EsClient) msearch(
	ctx context.Context,
	indexName string,
	queries []types.H,
) (res *esapi.Response, err error) {
	// Build the request body.
	var buf bytes.Buffer
	for _, query := range queries {
		buf.WriteString("{}\n")
		if err = jsoniter.NewEncoder(&buf).Encode(query); err != nil {
			err = fmt.Errorf("encode query fail:%w", err)
			return
		}
	}

	start := time.Now()
	res, err = c.client.Msearch(
		&buf,
		c.client.Msearch.WithContext(context.Background()),
		c.client.Msearch.WithIndex(indexName),
		c.client.Msearch.WithPretty(),
	)

	duration := time.Since(start)
	metric.EsSearchDuration.Observe(float64(duration / time.Millisecond))

	// gl 2 second
	if (duration / time.Second) > slowRequestSeconds {
		entry := logging.GetESLogger().WithField("request_id", ctx.Value(util.ContextRequestIDKey))

		expression := buf.String()
		durationMilliseconds := float64(duration/time.Millisecond) + 1

		entry.WithFields(logrus.Fields{
			"index":      indexName,
			"expression": expression,
			"duration":   durationMilliseconds,
		}).Warning("-")

		util.ReportToSentry("es client msearch slow request", map[string]interface{}{
			"index":      indexName,
			"expression": expression,
			"duration":   durationMilliseconds,
		})
	}

	return res, err
}

// Search ...
func (c *EsClient) Search(
	ctx context.Context,
	indexName string,
	query types.H, from, pageSize int,
	fields []string,
) (r types.H, err error) {
	// Perform the search request.
	res, err := c.search(ctx, indexName, query, from, pageSize, fields)
	if err != nil {
		err = fmt.Errorf("error getting response: %w", err)
		return
	}
	defer res.Body.Close()
	return decodeSearchResponse(res)
}

// Msearch ...
func (c *EsClient) Msearch(
	ctx context.Context,
	indexName string,
	queries []types.H,
) (r types.H, err error) {
	// Perform the search request.
	res, err := c.msearch(ctx, indexName, queries)
	if err != nil {
		err = fmt.Errorf("error getting response: %w", err)
		return
	}
	defer res.Body.Close()
	return decodeSearchResponse(res)
}

func decodeSearchResponse(res *esapi.Response) (r types.H, err error) {
	if res.IsError() {
		var e types.H
		if err = jsoniter.NewDecoder(res.Body).Decode(&e); err != nil {
			err = fmt.Errorf("response error and error parsing the response body: %w", err)
			return
		} else {
			err = fmt.Errorf("error [%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
			return
		}
	}

	err = jsoniter.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		err = fmt.Errorf("error parsing the response body: %w", err)
		return
	}

	return r, nil
}

// Ping ...
func (c *EsClient) Ping() (*esapi.Response, error) {
	return c.client.Ping()

}

// CreateIndex ...
func (c *EsClient) CreateIndex(index string) (*esapi.Response, error) {
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

	return c.client.Indices.Create(index, func(req *esapi.IndicesCreateRequest) {
		req.Body = strings.NewReader(mapping)
	})
}

// IndexExists ...
func (c *EsClient) IndexExists(index string) (*esapi.Response, error) {
	return c.client.Indices.Exists([]string{index})
}
