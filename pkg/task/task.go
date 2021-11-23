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
	"errors"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/logging"
	"engine/pkg/storage"
	"engine/pkg/util"
)

// FullSyncSignal 触发全量同步的信号
var FullSyncSignal chan struct{} = make(chan struct{})

// StartSync ...
func StartSync(ctx context.Context, cfg *config.Config) {
	syncLogger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	logger := syncLogger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "init",
	})

	// start the indexer, will keep do index in both full/incr sync
	indexer := NewIndexer(5)
	indexer.Start(ctx, &cfg.Index)

	lastFullSyncTime, err := storage.SyncSnapshotStorage.GetFullSyncLastSyncTime()
	if err != nil {
		if errors.Is(err, storage.ErrNoSyncBefore) {
			logger.WithError(err).Info("get last full sync time fail")
		} else {
			logger.WithError(err).Error("get last full sync time fail")
		}

		lastFullSyncTime = 0
	}

	now := time.Now().Unix()
	gap := now - lastFullSyncTime
	snapshot := NewSnapshot()

	shouldRunFullSync := true
	// NOTE: 使用gap sync的前提是, memory index有存一份并且启动的时候拉起来了
	// NOTE: snapshot.Start 启动的条件是, fullSync或gapIncrSync 成功执行完, 否则可能出现, 执行过程中被dump, 覆盖掉了现有的数据, 导致中断重启后数据有问题
	if gap < oneDay && snapshot.Exists() {
		logger.Infof("last sync time %d from now(%d) is %d, less than one day and snapshot exists, start a gap sync",
			lastFullSyncTime, now, gap)

		// load snapshot
		err := snapshot.Load(&cfg.Index)
		if err == nil {
			logger.Info("load the snapshot success, will start a gap inc sync")

			// start the gap sync
			NewGapIncrSyncer(lastFullSyncTime, now).OnSuccess(func() {
				err1 := storage.SyncSnapshotStorage.SetFullSyncLastSyncTime(now)
				if err1 != nil {
					logger.WithError(err1).Error("storage.SyncSnapshotStorage.SetFullSyncLastSyncTime fail")
				}
				snapshot.Start(ctx, 300)
			}).Start(ctx, indexer)

			// NOTE: use gap incr sync instead of full sync
			shouldRunFullSync = false
		} else {
			logger.WithError(err).Error("load the snapshot fail, will start a full sync")
		}
	}

	if shouldRunFullSync {
		logger.Info("will start a full sync",
			lastFullSyncTime)

		// start the full sync
		NewFullSyncer().OnSuccess(func() {
			err1 := storage.SyncSnapshotStorage.SetFullSyncLastSyncTime(now)
			if err1 != nil {
				logger.WithError(err1).Error("storage.SyncSnapshotStorage.SetFullSyncLastSyncTime fail")
			}
			snapshot.Start(ctx, 300)
		}).Start(ctx, indexer)
	}

	// start the incr sync, will sync every 30 seconds from now!
	NewIncrSyncer(30).OnSuccess(func() {
		err1 := storage.SyncSnapshotStorage.SetIncrSyncLastSyncTime(time.Now().Unix())
		if err1 != nil {
			logger.WithError(err1).Error("storage.SyncSnapshotStorage.SetIncrSyncLastSyncTime fail")
		}

	}).Start(ctx, indexer)

	// start delete event sync, will sync 5 seconds from now!
	NewDeleteSyncer(5).Start(ctx, indexer)

	// start timing grap incr, will sync 24 hour from now!
	NewTimingGapIncrSyncer(24*60*60, snapshot).Start(ctx, indexer)

	// 通过其它方式触发全量同步任务
	go waitFullSyncSignal(logger, ctx, indexer)
}

func waitFullSyncSignal(logger *logrus.Entry, ctx context.Context, indexer *Indexer) {
	var flag int32 // 限制并发
	for {
		select {
		case <-FullSyncSignal:
			// 同一时间只有一个全量同步任务能执行, 非阻塞锁
			if atomic.CompareAndSwapInt32(&flag, 0, 1) {
				// 全量同步开始的时间
				now := time.Now().Unix()

				NewFullSyncer().OnSuccess(func() {
					defer atomic.StoreInt32(&flag, 0) // 同步完成后释放锁

					err := storage.SyncSnapshotStorage.SetFullSyncLastSyncTime(now)
					if err != nil {
						logger.WithError(err).Error("storage.SyncSnapshotStorage.SetFullSyncLastSyncTime fail")
					}
				}).Start(ctx, indexer)
			}
		case <-ctx.Done():
			return
		}
	}
}
