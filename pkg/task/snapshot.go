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

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/indexer"
	"engine/pkg/logging"
	"engine/pkg/metric"
	"engine/pkg/storage"
	"engine/pkg/types"
	"engine/pkg/util"
)

// Snapshot ...
type Snapshot struct {
	mu sync.RWMutex
}

// NewSnapshot ...
func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

// Dump ...
func (s *Snapshot) Dump() error {
	s.mu.Lock()

	data := indexer.TakeSnapshot()
	bs, err := jsoniter.Marshal(data)
	if err != nil {
		return err
	}

	err = storage.SyncSnapshotStorage.SaveSnapshot(bs)
	if err != nil {
		return err
	}

	s.mu.Unlock()
	return nil
}

// Load ...
func (s *Snapshot) Load(cfg *config.Index) error {
	s.mu.RLock()

	bs, err := storage.SyncSnapshotStorage.GetSnapshot()
	if err != nil {
		return err
	}

	var data []types.SnapRecord
	err = jsoniter.Unmarshal(bs, &data)
	if err != nil {
		return err
	}

	// NOTE: 这里必须初始化填充 uid, 方便后面计算不需要重复获取
	for i := 0; i < len(data); i++ {
		err = data[i].FillPoliciesUniqueFields()
		if err != nil {
			return fmt.Errorf("snapshot FillPoliciesUniqueFields error: %w", err)
		}
	}

	err = indexer.LoadSnapshot(data)
	if err != nil {
		return err
	}

	s.mu.RUnlock()
	return nil
}

// Exists ...
func (s *Snapshot) Exists() bool {
	return storage.SyncSnapshotStorage.ExistSnapshot()
}

// Start ...
func (s *Snapshot) Start(ctx context.Context, interval int64) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "snapshot",
	})

	go s.run(ctx, interval, entry)

}

func (s *Snapshot) run(ctx context.Context, interval int64, logger *logrus.Entry) {
	logger.Infof("start a snapshot with interval = %v seconds", interval)

	// NOTE: take the snapshot immediately
	err := s.Dump()
	if err != nil {
		metric.SnapshotDumpFail.Inc()
		logger.WithError(err).Error("take snapshot fail!")
	} else {
		logger.Info("take snapshot success!")
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err1 := s.Dump()
			if err1 != nil {
				metric.SnapshotDumpFail.Inc()
				logger.WithError(err1).Error("take snapshot fail!")
			} else {
				logger.Info("take snapshot success!")
			}

		case <-ctx.Done():
			logger.Info("context done, the snapshot will stop running")
			ticker.Stop()
			return
		}
	}
}
