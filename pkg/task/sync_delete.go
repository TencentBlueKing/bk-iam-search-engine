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
	"encoding/json"
	"engine/pkg/indexer"
	"engine/pkg/logging"
	"engine/pkg/redis"
	"engine/pkg/types"
	"engine/pkg/util"
	"time"

	rds "github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

// TypePolicy ...
const (
	TypePolicy          = "policy"
	TypeSubject         = "subject"
	TypeSubjectTemplate = "subject_template"
)

// DeleteQueueKey ...
var DeleteQueueKey string

// Event ...
type Event struct {
	Type      string          `json:"type"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// PolicyIDsEvent ...
type PolicyIDsEvent struct {
	PolicyIDs []int64 `json:"policy_ids"`
}

// Delete ...
func (e *PolicyIDsEvent) Delete(logger *logrus.Entry) {
	indexer.BulkDelete(e.PolicyIDs, logger)
}

// SubjectsEvent ...
type SubjectsEvent struct {
	Subjects  []types.Subject `json:"subjects"`
	Timestamp int64
}

// Delete ...
func (e *SubjectsEvent) Delete(logger *logrus.Entry) {
	indexer.BulkDeleteBySubjects(e.Timestamp, e.Subjects, logger)
}

// TemplateSubjectEvent ...
type TemplateSubjectEvent struct {
	TemplateSubjects []TemplateSubject `json:"subject_templates"`
	Timestamp        int64
}

// TemplateSubject ...
type TemplateSubject struct {
	Subject    types.Subject `json:"subject"`
	TemplateID int64         `json:"template_id"`
}

// Delete ...
func (e *TemplateSubjectEvent) Delete(logger *logrus.Entry) {
	// NOTE 基于template的删除都是同一个模板, 不存在批量的情况
	for _, templateSubject := range e.TemplateSubjects {
		indexer.BulkDeleteByTemplateSubjects(
			e.Timestamp,
			templateSubject.TemplateID,
			[]types.Subject{templateSubject.Subject},
			logger,
		)
	}
}

// DeleteSyncer ...
type DeleteSyncer struct {
	interval      int64 // second
	onSuccessFunc func()
}

// NewDeleteSyncer ...
func NewDeleteSyncer(interval int64) Syncer {
	return &DeleteSyncer{
		interval:      interval,
		onSuccessFunc: func() {},
	}
}

// OnSuccess ...
func (s *DeleteSyncer) OnSuccess(f func()) Syncer {
	s.onSuccessFunc = f
	return s
}

// Start ...
func (s *DeleteSyncer) Start(ctx context.Context, idx *Indexer) {
	logger := logging.GetSyncLogger()
	taskID := util.RandString(16)
	entry := logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"type":    "delete_sync",
	})

	entry.Infof("start a delete task with interval = %v seconds", s.interval)

	ticker := time.NewTicker(time.Duration(s.interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:

				// 解析事件格式, 根据格式调用对应的方法
				syncDeleteEvent(entry, idx)

				s.onSuccessFunc()
			case <-ctx.Done():
				logger.Info("context done, the incr trigger will stop running")
				ticker.Stop()
				return
			}
		}
	}()
}

func syncDeleteEvent(entry *logrus.Entry, idx *Indexer) {
	// 一次最多消费1000条
	for i := 0; i < deleteBatchSize; i++ {
		eventString, err := redis.RPop(DeleteQueueKey)
		if err != nil {
			// list empty
			if err == rds.Nil {
				return
			}
			entry.Errorf("event pop error: %s", err.Error())
			continue
		}

		// !!! 事件做持久化, 在Gap同步时可以考虑读取处理
		indexDeleteByEvent(idx, eventString, entry)
	}
}

func indexDeleteByEvent(idx *Indexer, eventString string, entry *logrus.Entry) {
	event := Event{}
	err := jsoniter.UnmarshalFromString(eventString, &event)
	if err != nil {
		entry.Errorf("unmarshal event `%s` error: %s", eventString, err.Error())
		return
	}

	var data deleteEvent
	switch event.Type {
	case TypePolicy:
		data = &PolicyIDsEvent{}
	case TypeSubject:
		data = &SubjectsEvent{
			Timestamp: event.Timestamp,
		}
	case TypeSubjectTemplate:
		data = &TemplateSubjectEvent{
			Timestamp: event.Timestamp,
		}
	default:
		entry.Errorf("unsupported event `%s`", eventString)
		return
	}

	err = jsoniter.Unmarshal(event.Data, data)
	if err != nil {
		entry.Errorf("unmarshal event `%s` error: %s", eventString, err.Error())
		return
	}
	idx.BulkDeleteByEvent(data)
}
