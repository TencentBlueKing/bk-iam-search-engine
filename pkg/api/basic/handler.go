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

package basic

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"engine/pkg/client"
	"engine/pkg/components"
	"engine/pkg/config"
	"engine/pkg/version"
)

// Pong is the response for /ping
// Ping godoc
// @Summary ping-pong for alive test
// @Description /ping to get response from iam, make sure the server is alive
// @ID ping
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Header 200 {string} X-Request-Id "the request id"
// @Router /ping [get]
func Pong(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// Version godoc
// @Summary version for identify
// @Description /version to get the version of iam
// @ID version
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Header 200 {string} X-Request-Id "the request id"
// @Router /version [get]
func Version(c *gin.Context) {
	runEnv := os.Getenv("RUN_ENV")
	c.JSON(200, gin.H{
		"version":   version.Version,
		"commit":    version.Commit,
		"buildTime": version.BuildTime,
		"goVersion": version.GoVersion,
		"env":       runEnv,
	})
}

// NewHealthzHandleFunc create the handler of /healthz
// Healthz godoc
// @Summary healthz for server health check
// @Description /healthz to make sure the server is health
// @ID healthz
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {string} string message
// @Failure 500 {string} string message
// @Header 200 {string} X-Request-Id "the request id"
// @Router /healthz [get]
func NewHealthzHandleFunc(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// check the es is running
		esClient, err := client.NewEsPingClient(&cfg.Index.ElasticSearch)
		if err != nil {
			message := fmt.Sprintf("new es clinet fail: %s [address=%v]",
				err.Error(), cfg.Index.ElasticSearch.Addresses)
			c.String(http.StatusInternalServerError, message)
			return
		}

		_, err = esClient.Ping()
		if err != nil {
			message := fmt.Sprintf("ping elasticsearch fail: %s",
				err.Error())
			c.String(http.StatusInternalServerError, message)
			return
		}

		// check the iam backend
		err = components.NewIAMClient().Ping()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, "ok")
	}
}
