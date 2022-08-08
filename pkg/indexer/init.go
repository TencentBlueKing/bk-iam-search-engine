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

package indexer

import (
	"context"

	"github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/logging/debug"
	"engine/pkg/types"
)

var globalIndex *Index

// InitGlobalIndex ...
func InitGlobalIndex(cfg *config.Index) {
	var err error
	err = creatIndexIfNotExists(cfg)
	if err != nil {
		panic(err)
	}

	globalIndex, err = NewIndex(cfg)
	if err != nil {
		panic(err)
	}
}

// TakeSnapshot ...
func TakeSnapshot() []types.SnapRecord {
	return globalIndex.EvalEngine.TakeSnapshot()
}

// LoadSnapshot ...
func LoadSnapshot(data []types.SnapRecord) error {
	return globalIndex.EvalEngine.LoadSnapshot(data)
}

// BulkUpsert ...
func BulkUpsert(policies []types.Policy, logger *logrus.Entry) {
	globalIndex.BulkUpsert(policies, logger)
}

// BulkDelete ...
func BulkDelete(ids []int64, logger *logrus.Entry) {
	globalIndex.BulkDelete(ids, logger)
}

// BulkDeleteBySubjects ...
func BulkDeleteBySubjects(beforeUpdatedAt int64, subjects []types.Subject, logger *logrus.Entry) {
	globalIndex.BulkDeleteBySubjects(beforeUpdatedAt, subjects, logger)
}

// Search ...
func Search(ctx context.Context, req *types.SearchRequest, entry *debug.Entry) ([]types.Subject, error) {
	return globalIndex.Search(ctx, req, entry)
}

// BatchSearch ...
func BatchSearch(ctx context.Context, requests []*types.SearchRequest, entry *debug.Entry) ([][]types.Subject, error) {
	return globalIndex.BatchSearch(ctx, requests, entry)
}

// Stats ...
func Stats(system, action string) map[string]uint64 {
	return globalIndex.Stats(system, action)
}

// TotalStats ...
func TotalStats() map[string]uint64 {
	return globalIndex.TotalStats()
}
