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
	"time"

	"github.com/sirupsen/logrus"

	"engine/pkg/logging"
	"engine/pkg/storage"
	"engine/pkg/util"
)

// GapIncrSyncer will sync the upsertPolicies between beginUpdatedAt and endUpdatedAt.
type GapIncrSyncer struct {
	beginUpdatedAt int64
	endUpdatedAt   int64
	onSuccessFunc  func()
}

// NewGapIncrSyncer ...
func NewGapIncrSyncer(beginUpdatedAt int64, endUpdatedAt int64) Syncer {
	return &GapIncrSyncer{
		beginUpdatedAt: beginUpdatedAt,
		endUpdatedAt:   endUpdatedAt,
		onSuccessFunc:  func() {},
	}
}

// OnSuccess ...
func (s *GapIncrSyncer) OnSuccess(f func()) Syncer {
	s.onSuccessFunc = f
	return s
}

// Start ...
func (s *GapIncrSyncer) Start(ctx context.Context, idx *Indexer) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "gap_incr_sync",
	})

	go func() {
		_ = syncWithMetrics(gapSyncType, func() error {
			return s.run(ctx, idx, entry)
		})
	}()
}

func (s *GapIncrSyncer) run(ctx context.Context, idx *Indexer, logger *logrus.Entry) error {
	logger.Infof("start a gap incr sync [updated_at %d to %d]", s.beginUpdatedAt, s.endUpdatedAt)
	// do sync one hour by one hour, not parallel
	for i := s.beginUpdatedAt; i < s.endUpdatedAt; i += oneHour {
		begin := i
		end := i + oneHour
		err := syncBetweenUpdatedAt(idx, begin, end, logger)
		if err != nil {
			return fmt.Errorf("gap incr sync fail: %w", err)
		}
	}

	s.onSuccessFunc()
	logger.Infof("done the gap incr sync [updated_at %d to %d]", s.beginUpdatedAt, s.endUpdatedAt)

	return nil
}

type timingGapIncrSyncer struct {
	interval      int64 // second
	onSuccessFunc func()

	snapshot *Snapshot
}

// NewTimingGapIncrSyncer ...
func NewTimingGapIncrSyncer(interval int64, snapshot *Snapshot) Syncer {
	return &timingGapIncrSyncer{
		interval:      interval,
		onSuccessFunc: func() {},

		snapshot: snapshot,
	}
}

// OnSuccess ...
func (s *timingGapIncrSyncer) OnSuccess(f func()) Syncer {
	s.onSuccessFunc = f
	return s
}

// Start ...
func (s *timingGapIncrSyncer) Start(ctx context.Context, idx *Indexer) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "timing_gap_incr",
	})

	ticker := time.NewTicker(time.Duration(s.interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				endUpdatedAt := time.Now().Unix()
				beginUpdateAt, err := storage.SyncSnapshotStorage.GetFullSyncLastSyncTime()
				if err != nil {
					entry.WithError(err).Error("storage.SyncSnapshotStorage.GetFullSyncLastSyncTime fail")
					beginUpdateAt = endUpdatedAt - s.interval
				}

				// start gap incr
				NewGapIncrSyncer(beginUpdateAt, endUpdatedAt).OnSuccess(func() {
					err1 := storage.SyncSnapshotStorage.SetFullSyncLastSyncTime(endUpdatedAt)
					if err1 != nil {
						entry.WithError(err1).Error("storage.SyncSnapshotStorage.SetFullSyncLastSyncTime fail")
					}
					err1 = s.snapshot.Dump()
					if err1 != nil {
						entry.WithError(err1).Error("s.snapshot.Dump fail")
					}
				}).Start(ctx, idx)

				s.onSuccessFunc()
			case <-ctx.Done():
				logger.Info("context done, the incr trigger will stop running")
				ticker.Stop()
				return
			}
		}
	}()
}
