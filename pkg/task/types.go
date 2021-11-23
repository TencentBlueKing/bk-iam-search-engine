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

	"engine/pkg/config"
	"engine/pkg/types"

	"github.com/sirupsen/logrus"
)

const (
	oneDay  = 24 * 60 * 60
	oneHour = 60 * 60

	// for full sync
	fullPoolSize  = 10
	fullBatchSize = 500

	// for incr sync
	incrPoolSize = 10
	// 必须 < 200
	incrBatchSize = 100

	// for delete sync
	deleteBatchSize = 1000

	// for index
	indexChannelBufferSize = 10000
	indexPoolSize          = 10
	indexBatchSize         = 100
)

// Syncer ...
type Syncer interface {
	Start(ctx context.Context, idx *Indexer)
	OnSuccess(func()) Syncer
}

// Snapshoter ...
type Snapshoter interface {
	Start(ctx context.Context, interval int64)

	Dump() error
	Load(cfg *config.Index) error
	Exists() bool
}

// IndexOperator ...
type IndexOperator interface {
	Start(ctx context.Context, cfg *config.Index)

	Add(p types.Policy)
	BulkAdd(ps []types.Policy)

	Delete(id int64)
	BulkDelete(ids []int64)
}

type deleteEvent interface {
	Delete(logger *logrus.Entry)
}
