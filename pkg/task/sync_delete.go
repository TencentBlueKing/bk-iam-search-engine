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
	"time"

	"engine/pkg/indexer"
	"engine/pkg/logging"
	"engine/pkg/types"
	"engine/pkg/util"

	"github.com/adjust/rmq/v4"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// TypePolicy ...
const (
	TypePolicy  = "policy"
	TypeSubject = "subject"
	// TypeSubjectTemplate = "subject_template"
)

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

	engineDeletionEventQueue.AddConsumerFunc(rmqConsumerTag, func(delivery rmq.Delivery) {
		// parse message
		payload := delivery.Payload()

		// process
		indexDeleteByEvent(idx, payload, entry)

		// ack
		if err := delivery.Ack(); err != nil {
			entry.WithError(err).Errorf("rmq ack payload `%s` fail", payload)
		}
	})

	err := engineDeletionEventQueue.StartConsuming(10, time.Second)
	if err != nil {
		log.WithError(err).Error("rmq queue start consuming fail")
		panic(err)
	}

	<-ctx.Done()                    // wait for signal
	<-connection.StopAllConsuming() // wait for all Consume() calls to finish
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
