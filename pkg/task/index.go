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

package task

import (
	"context"
	"math/rand"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/indexer"
	"engine/pkg/logging"
	"engine/pkg/types"
	"engine/pkg/util"
)

// Indexer ...
type Indexer struct {
	upsertPolicies chan types.Policy
	deleteIDs      chan int64
	deleteEvents   chan deleteEvent
	interval       int64
}

// NewIndexer ...
func NewIndexer(interval int64) *Indexer {
	return &Indexer{
		upsertPolicies: make(chan types.Policy, indexChannelBufferSize),
		deleteIDs:      make(chan int64, indexChannelBufferSize),
		deleteEvents:   make(chan deleteEvent, indexChannelBufferSize),
		interval:       interval,
	}
}

// Add ...
func (i *Indexer) Add(p types.Policy) {
	i.upsertPolicies <- p
}

// BulkAdd ...
func (i *Indexer) BulkAdd(ps []types.Policy) {
	for _, p := range ps {
		i.upsertPolicies <- p
	}
}

// Delete ...
func (i *Indexer) Delete(id int64) {
	i.deleteIDs <- id
}

// BulkDelete ...
func (i *Indexer) BulkDelete(ids []int64) {
	for _, id := range ids {
		i.deleteIDs <- id
	}
}

// BulkDeleteByEvent ...
func (i *Indexer) BulkDeleteByEvent(event deleteEvent) {
	i.deleteEvents <- event
}

// Start ...
func (i *Indexer) Start(ctx context.Context, cfg *config.Index) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "index",
	})

	go i.run(ctx, cfg, entry)
}

func (i *Indexer) run(ctx context.Context, cfg *config.Index, logger *logrus.Entry) {
	// start an goroutine worker pool to consume
	logger.Info("start indexer, begin do indexing")

	pu, _ := ants.NewPoolWithFunc(indexPoolSize, func(i interface{}) {
		// TODO: if the es stuck? will stuck all goroutines?
		indexer.BulkUpsert(i.([]types.Policy), logger)
	})
	defer pu.Release()

	pd, _ := ants.NewPoolWithFunc(indexPoolSize, func(i interface{}) {
		// TODO: if the es stuck? will stuck all goroutines?
		switch v := i.(type) {
		case []int64:
			indexer.BulkDelete(v, logger)
		case deleteEvent:
			v.Delete(logger)
		}
	})
	defer pd.Release()

	idleTimeout := time.NewTicker(time.Duration(i.interval) * time.Second)
	defer idleTimeout.Stop()

	batchUpsertData := make([]types.Policy, 0, indexBatchSize)
	batchDeleteData := make([]int64, 0, indexBatchSize)
	for {
		select {
		case policy := <-i.upsertPolicies:
			batchUpsertData = append(batchUpsertData, policy)

			if len(batchUpsertData) == indexBatchSize {
				logger.WithField("op", "upsert").Infof("got %d records, do index upsert", indexBatchSize)
				_ = pu.Invoke(batchUpsertData)
				batchUpsertData = make([]types.Policy, 0, indexBatchSize)
			}

		case id := <-i.deleteIDs:
			batchDeleteData = append(batchDeleteData, id)

			if len(batchDeleteData) == indexBatchSize {
				logger.WithField("op", "delete").Infof("got %d records, do index delete", indexBatchSize)
				_ = pd.Invoke(batchDeleteData)
				batchDeleteData = make([]int64, 0, indexBatchSize)
			}

		case event := <-i.deleteEvents:
			// NOTE 基于事件的删除本身就是批量删除, 所以这里不再做buffer批量
			_ = pd.Invoke(event)

		case <-idleTimeout.C:
			batchUpsertSize := len(batchUpsertData)
			batchDeleteSize := len(batchDeleteData)

			if batchUpsertSize+batchDeleteSize > 0 {
				logger.Infof(
					"timeout and may do index, upsert size=%d, delete size=%d",
					len(batchUpsertData),
					len(batchDeleteData),
				)
			} else {
				if rand.Intn(10) == 0 {
					logger.Infof("alive")
				}
			}

			if batchUpsertSize > 0 {
				_ = pu.Invoke(batchUpsertData)
				batchUpsertData = make([]types.Policy, 0, indexBatchSize)
			}

			if batchDeleteSize > 0 {
				_ = pd.Invoke(batchDeleteData)
				batchDeleteData = make([]int64, 0, indexBatchSize)
			}

		case <-ctx.Done():
			logger.Info("context done, the indexer will stop running")
			return
		}
	}
}
