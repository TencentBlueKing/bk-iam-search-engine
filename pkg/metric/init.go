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

package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	serviceName = "iamSearchEngine"
)

var (
	// RequestCount api状态计数 + server_ip的请求数量和状态
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "bkiam_engine_api_requests_total",
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "path", "status", "error", "client_id"},
	)

	// RequestDuration api响应时间分布
	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "bkiam_engine_api_request_duration_milliseconds",
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     []float64{50, 100, 200, 500, 1000, 2000, 5000},
	},
		[]string{"method", "path", "status", "client_id"},
	)

	// ClientRequestDuration 依赖 api 响应时间分布
	ClientRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "bkiam_engine_client_request_duration_milliseconds",
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     []float64{20, 50, 100, 200, 500, 1000, 2000, 5000},
	},
		[]string{"method", "path", "status", "component"},
	)

	// LastSyncTimestamp 记录最后更新时间戳 => 告警事项: 某个实例多久没有成功同步
	LastSyncTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "bkiam_engine_last_sync_timestamp",
		Help:        "Timestamp of last sync task.",
		ConstLabels: prometheus.Labels{"service": serviceName},
	},
		[]string{"type"},
	)

	// SyncFail 当前这次同步失败了, 检测到直接告警
	SyncFail = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "bkiam_engine_sync_fail",
		Help:        "Fail point of sync task.",
		ConstLabels: prometheus.Labels{"service": serviceName},
	},
		[]string{"type"},
	)

	// SyncTaskDuration 同步任务时长分布
	SyncTaskDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "bkiam_engine_sync_task_duration_seconds",
		Help:        "How long it took to process the sync task, partitioned by type.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     []float64{1, 5, 10, 20, 50, 100, 1000, 2000, 5000},
	},
		[]string{"type"},
	)

	// EsSearchDuration ElasticSearch request duration
	EsSearchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        "bkiam_engine_es_search_duration_milliseconds",
		Help:        "How long it took to process es search.",
		ConstLabels: prometheus.Labels{"service": serviceName},
		Buckets:     []float64{20, 50, 100, 200, 500, 1000, 2000, 5000},
	})

	// SnapshotDumpFail 当前这次同步失败了, 检测到直接告警
	SnapshotDumpFail = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "bkiam_engine_snapshot_dump_fail",
		Help:        "Fail point of snapshot dump.",
		ConstLabels: prometheus.Labels{"service": serviceName},
	})
)

// InitMetrics ...
func InitMetrics() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ClientRequestDuration)
	prometheus.MustRegister(LastSyncTimestamp)
	prometheus.MustRegister(SyncFail)
	prometheus.MustRegister(SyncTaskDuration)
	prometheus.MustRegister(EsSearchDuration)
	prometheus.MustRegister(SnapshotDumpFail)
}
