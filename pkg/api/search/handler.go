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

package search

import (
	"time"

	"github.com/gin-gonic/gin"

	"engine/pkg/indexer"
	"engine/pkg/logging/debug"
	"engine/pkg/storage"
	"engine/pkg/task"
	"engine/pkg/types"
	"engine/pkg/util"
)

// ! limit default is 10, but you can set limit to maximum 1000; OR set to 0(unlimited); but the timeout is set to 10s
// ! will stop search while:
// !    - reach the limit count if limit > 0
// !    - reach the timeout

// search godoc
// @Summary search subjects by system/action/resource
// @Description search the subjects who have the permission of that system/action/resource
// @ID api-search
// @Tags api
// @Accept json
// @Produce json
// @Param params body types.SearchRequest true "the list request"
// @Success 200 {object} map[string]interface{}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/search [post]
func search(c *gin.Context) {

	var req types.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check system
	systemID := req.System
	clientID := util.GetClientID(c)
	if !isSuperClient(clientID) {
		if err := validateSystemMatchClient(systemID, clientID); err != nil {
			util.BadRequestErrorJSONResponse(c, err.Error())
			return
		}
	}

	req.NowTimestamp = time.Now().Unix()
	for _, rn := range req.Resource {
		if rn.Attribute == nil {
			rn.Attribute = make(map[string]interface{})
		}
		rn.Attribute["id"] = rn.ID
	}

	// enable debug
	var entry *debug.Entry
	_, isDebug := c.GetQuery("debug")
	if isDebug {
		entry = debug.NewDebugEntry()
		defer debug.ReleaseDebugEntry(entry)
	}

	subjects, err := indexer.Search(util.GetContextWithRequestID(c), &req, entry)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponseWithDebug(c, "ok", subjects, entry)
}

// batchSearch godoc
// @Summary batch search subjects by system/action/resource
// @Description batch search the subjects who have the permission of that system/action/resource
// @ID api-batch-search
// @Tags api
// @Accept json
// @Produce json
// @Param params body []types.SearchRequest true "the list request"
// @Success 200 {object} map[string]interface{}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/batch-search [post]
func batchSearch(c *gin.Context) {
	var body []*types.SearchRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// check system
	clientID := util.GetClientID(c)
	if !isSuperClient(clientID) {
		systemIDs := util.NewStringSet()
		for _, req := range body {
			systemIDs.Add(req.System)
		}

		for _, systemID := range systemIDs.ToSlice() {
			if err := validateSystemMatchClient(systemID, clientID); err != nil {
				util.BadRequestErrorJSONResponse(c, err.Error())
				return
			}
		}
	}

	now := time.Now().Unix()
	for _, req := range body {
		for _, rn := range req.Resource {
			if rn.Attribute == nil {
				rn.Attribute = make(map[string]interface{})
			}
			rn.Attribute["id"] = rn.ID
		}
		req.NowTimestamp = now
	}

	// enable debug
	var entry *debug.Entry
	_, isDebug := c.GetQuery("debug")
	if isDebug {
		entry = debug.NewDebugEntryWithFixedSubEntries(len(body))
		defer debug.ReleaseDebugEntry(entry)
	}

	ctx := util.GetContextWithRequestID(c)
	results, err := indexer.BatchSearch(ctx, body, entry)
	if err != nil {
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponseWithDebug(c, "ok", gin.H{"results": results}, entry)
}

// batchSearch godoc
// @Summary get iam search engine stats
// @Description get iam search engine stats
// @ID api-stats
// @Tags api
// @Accept json
// @Produce json
// @Param system path string false "System ID"
// @Param action path string false "Action ID"
// @Success 200 {object} map[string]interface{}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/stats [get]
func stats(c *gin.Context) {
	system, _ := c.GetQuery("system")
	action, _ := c.GetQuery("action")

	var stats map[string]uint64
	if system != "" && action != "" {
		clientID := util.GetClientID(c)
		if !isSuperClient(clientID) {
			if err := validateSystemMatchClient(system, clientID); err != nil {
				util.BadRequestErrorJSONResponse(c, err.Error())
				return
			}
		}

		stats = indexer.Stats(system, action)
	} else {
		clientID := util.GetClientID(c)
		if !isSuperClient(clientID) {
			util.ForbiddenJSONResponse(c, "only supper app code can access the global stats")
			return
		}

		stats = indexer.TotalStats()

		// 查询最近的同步时间
		fullSyncLastTime, _ := storage.SyncSnapshotStorage.GetFullSyncLastSyncTime()
		incrSyncLastTime, _ := storage.SyncSnapshotStorage.GetIncrSyncLastSyncTime()

		stats["full_sync_last_time"] = uint64(fullSyncLastTime)
		stats["incr_sync_last_time"] = uint64(incrSyncLastTime)
	}

	util.SuccessJSONResponse(c, "ok", stats)
}

// batchSearch godoc
// @Summary trigger iam search engine full sync task
// @Description trigger iam search engine full sync task
// @ID api-full-sync
// @Tags api
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v1/full-sync [post]
func fullSync(c *gin.Context) {
	// 敏感操作, 只有超级app code能
	clientID := util.GetClientID(c)
	if !isSuperClient(clientID) {
		util.ForbiddenJSONResponse(c, "")
		return
	}

	// 触发全量同步
	select {
	case task.FullSyncSignal <- struct{}{}:
	default:
		util.ConflictJSONResponse(c, "")
		return
	}

	util.SuccessJSONResponse(c, "ok", nil)
}
