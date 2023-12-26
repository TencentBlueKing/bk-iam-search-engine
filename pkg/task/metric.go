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
	"time"

	"engine/pkg/metric"
)

const (
	fullSyncType = "full_sync"
	gapSyncType  = "gap_sync"
	incrSyncType = "incr_sync"
)

// 记录任务中的metric信息
func syncWithMetrics(_type string, fn func() error) error {
	start := time.Now()
	metric.LastSyncTimestamp.WithLabelValues(_type).SetToCurrentTime()

	err := fn()
	if err != nil {
		metric.SyncFail.WithLabelValues(_type).Inc()
		return err
	}

	duration := time.Since(start)
	metric.SyncTaskDuration.WithLabelValues(_type).Observe(float64(duration / time.Second))

	return nil
}
