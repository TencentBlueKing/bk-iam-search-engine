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
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"

	"engine/pkg/components"
	"engine/pkg/logging"
	"engine/pkg/util"
)

// IAM Backend
// 1. 最大时间间隔是 3600s (1 hour)
// 2. 每次接口能批量拉最大 200 个 ID

// NOTE: 每个 N s 启动一个增量同步, 但是同时有且仅有一个任务在跑

// TODO: 如果堵住了? 那么下一次发起的时候, 和上一次同步的中间间隔的数据, 没有被同步
//       怎么处理这种情况? 相当于 大一点的 增量, 如果纯靠  当前时间之前的几分钟, 会丢数据的
// 先这样, 暂时不引入复杂度
// TODO: 被删除的ID, 需要批量清空
// TODO: 队列满会被阻塞
// TODO: worker pool 参数配置

// leadInSeconds 每次开始增量同步的提前量
const leadInSeconds int64 = 1

// IncrSyncer will sync the upsertPolicies from now, each interval seconds.
type IncrSyncer struct {
	interval      int64 // second
	onSuccessFunc func()
}

// NewIncrSyncer ...
func NewIncrSyncer(interval int64) Syncer {
	return &IncrSyncer{
		interval:      interval,
		onSuccessFunc: func() {},
	}
}

// OnSuccess ...
func (s *IncrSyncer) OnSuccess(f func()) Syncer {
	s.onSuccessFunc = f
	return s
}

type timeGap struct {
	beginUpdatedAt int64
	endUpdatedAt   int64
}

// Start ...
func (s *IncrSyncer) Start(ctx context.Context, idx *Indexer) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "incr_sync",
	})

	entry.Infof("start a incr task with interval = %v seconds", s.interval)

	// 一个goroutine 定时生成 需要处理的增量时间差, 一个goroutine 消费
	// 如果消费比较慢, 会有延迟, 但是不会丢

	ticker := time.NewTicker(time.Duration(s.interval) * time.Second)
	timeGapChan := make(chan timeGap, 120)
	go func() {
		for {
			select {
			case <-ticker.C:
				endUpdatedAt := time.Now().Unix()
				beginUpdatedAt := endUpdatedAt - s.interval - leadInSeconds

				timeGapChan <- timeGap{
					beginUpdatedAt: beginUpdatedAt,
					endUpdatedAt:   endUpdatedAt,
				}

			case <-ctx.Done():
				logger.Info("context done, the incr trigger will stop running")
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case tg := <-timeGapChan:
				// 如果 channel  buffer 满了, 将会堵在这里
				err := syncWithMetrics(incrSyncType, func() error {
					return syncBetweenUpdatedAt(idx, tg.beginUpdatedAt, tg.endUpdatedAt, entry)
				})
				if err == nil {
					s.onSuccessFunc()
				}
				// do stuff
			case <-ctx.Done():
				logger.Info("context done, the incr syncer will stop running")
				ticker.Stop()
				return
			}
		}
	}()
}

func syncBetweenUpdatedAt(idx *Indexer, beginUpdatedAt int64, endUpdatedAt int64, logger *logrus.Entry) error {
	taskInfo := fmt.Sprintf("[id=%s, updated_at %d to %d, poolSize=%d, batchSize=%d]",
		util.RandString(16), beginUpdatedAt, endUpdatedAt, incrPoolSize, incrBatchSize)

	logger.Infof("do the sync task %s", taskInfo)

	var wg sync.WaitGroup
	// 1. get ids before last 5 minutes
	ids, err := components.NewIAMClient().ListPolicyIDBetweenUpdateAt(beginUpdatedAt, endUpdatedAt)
	if err != nil {
		logger.WithError(err).Errorf("ListPolicyIDByBetweenUpdateAt begin_updated_at=`%d`, end_updated_at=`%d` fail",
			beginUpdatedAt, endUpdatedAt)
		return fmt.Errorf("sync between update at list policy fail: %w", err)
	}

	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := ants.NewPoolWithFunc(incrPoolSize, func(i interface{}) {
		defer wg.Done()

		policies, err1 := components.NewIAMClient().ListPolicyByIDs(i.([]int64))
		if err1 != nil {
			logger.WithError(err).Errorf("ListPolicyByIDs ids=`%+v` fail", i.([]int64))
			return
		}

		// NOTE: 如果channel满了, 这里将会导致整体stuck, 无法close channel或者整体退出
		// ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		// defer cancel()

		// add to chan for indexer
		idx.BulkAdd(policies)
	}, ants.WithExpiryDuration(2*time.Second))
	defer p.Release()

	// Submit tasks one by one.
	maxIndex := len(ids)
	for i := 0; i < maxIndex; i += incrBatchSize {
		beginIndex := i
		endIndex := i + incrBatchSize
		if endIndex > maxIndex {
			endIndex = maxIndex
		}

		wg.Add(1)
		_ = p.Invoke(ids[beginIndex:endIndex])
	}

	wg.Wait()

	logger.Infof("done the sync task %s", taskInfo)

	return nil
}
