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

package types

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/TencentBlueKing/gopkg/collection/set"

	"engine/pkg/logging/debug"
)

// SubjectTypeAll ...
const (
	SubjectTypeAll = "all"
)

// ResourceNode ...
type ResourceNode struct {
	System    string                 `json:"system" binding:"required" example:"bk_paas"`
	Type      string                 `json:"type" binding:"required" example:"app"`
	ID        string                 `json:"id" binding:"required" example:"framework"`
	Attribute map[string]interface{} `json:"attribute" binding:"required"`
}

// Resource ...
type Resource []ResourceNode

// Action ...
type Action struct {
	ID string `json:"id" binding:"required" example:"edit"`
}

// SearchRequest ...
type SearchRequest struct {
	System   string   `json:"system" binding:"required" example:"bk_paas"`
	Action   Action   `json:"action" binding:"required"`
	Resource Resource `json:"resource" binding:"required"`

	SubjectType string `json:"subject_type" binding:"required,oneof=all group user" example:"all"`
	// ! we don't support pagination, we can only fetch limit subjects at once
	Limit int `json:"limit" binding:"min=-1,max=1000" example:"10"`

	NowTimestamp int64
}

// Engine ...
type Engine interface {
	Size(system, action string) uint64
	BulkAdd(policies []*Policy) error

	BulkDelete(ids []int64, logger *log.Entry) error
	BulkDeleteBySubjects(beforeUpdatedAt int64, subjects []Subject, logger *log.Entry) error
	BulkDeleteByTemplateSubjects(beforeUpdatedAt int64, templateID int64, subjects []Subject, logger *log.Entry) error

	Search(ctx context.Context, req *SearchRequest, entry *debug.Entry) (SearchResult, error)
	BatchSearch(ctx context.Context, requests []*SearchRequest, entry *debug.Entry) (results []SearchResult, err error)

	Total() uint64
	GetLastIndexTime() time.Time

	TakeSnapshot() []SnapRecord
	LoadSnapshot([]SnapRecord) error
}

// SearchResult ...
type SearchResult interface {
	GetSubjects(allowedSubjectUIDs *set.StringSet) []Subject
}
