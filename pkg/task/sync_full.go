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

type betweenArgs struct {
	ExpiredAt int64
	BeginID   int64
	EndID     int64
}

// FullSyncer will sync all policies from iam backend
type FullSyncer struct {
	onSuccessFunc func()
}

// NewFullSyncer ...
func NewFullSyncer() Syncer {
	return &FullSyncer{
		onSuccessFunc: func() {},
	}
}

// OnSuccess ...
func (s *FullSyncer) OnSuccess(f func()) Syncer {
	s.onSuccessFunc = f
	return s
}

// Start ...
func (s *FullSyncer) Start(ctx context.Context, idx *Indexer) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "full_sync",
	})

	go func() {
		err := syncWithMetrics(fullSyncType, func() error {
			return fullSync(idx, entry)
		})
		if err == nil {
			s.onSuccessFunc()
		}
	}()
}

func fullSync(idx *Indexer, logger *logrus.Entry) error {
	taskInfo := fmt.Sprintf("[poolSize=%d, batchSize=%d]", fullPoolSize, fullBatchSize)
	logger.Infof("start a full sync %s", taskInfo)

	var wg sync.WaitGroup

	// 1. get max id
	nowTs := time.Now().Unix()
	maxID, err := components.NewIAMClient().GetMaxIDBeforeUpdate(nowTs)
	if err != nil {
		logger.WithError(err).Errorf("GetMaxIDBeforeUpdate updated_at=`%d` fail", nowTs)
		return fmt.Errorf("full sync get max id fail: %w", err)
	}

	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := ants.NewPoolWithFunc(fullPoolSize, func(i interface{}) {
		defer wg.Done()

		args := i.(betweenArgs)
		policies, err1 := components.NewIAMClient().ListPolicyBetweenID(args.ExpiredAt, args.BeginID, args.EndID)
		if err1 != nil {
			// TODO 任务池中的 error 怎样能优雅的暴露到调用方
			logger.WithError(err1).Errorf("ListPolicyBetweenID minID=`%d`, maxID=`%d` fail", args.BeginID, args.EndID)
			return
		}

		// add to chan for indexer
		idx.BulkAdd(policies)

		// 404 or expired, should be deleted
		existedPIDs := util.NewFixedLengthInt64Set(len(policies))
		for _, p := range policies {
			existedPIDs.Add(p.ID)
		}

		batchDeleteIDs := util.NewInt64Set()
		for i := args.BeginID; i <= args.EndID; i++ {
			if !existedPIDs.Has(i) {
				batchDeleteIDs.Add(i)
			}
		}
		idx.BulkDelete(batchDeleteIDs.ToSlice())

	}, ants.WithExpiryDuration(2*time.Second))
	defer p.Release()

	// Submit tasks one by one.
	for i := int64(1); i <= maxID; i += fullBatchSize {
		beginID := i
		endID := i + fullBatchSize
		if endID > maxID {
			endID = maxID
		}

		wg.Add(1)
		// TODO 错误处理, 待读了ants的代码后
		_ = p.Invoke(betweenArgs{
			ExpiredAt: nowTs,
			BeginID:   beginID,
			EndID:     endID,
		})
	}

	wg.Wait()

	logger.Infof("done the full sync %s", taskInfo)

	return nil
}
