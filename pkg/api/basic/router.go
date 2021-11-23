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
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "engine/docs" // docs is generated by Swag CLI, you have to import it.
	"engine/pkg/config"
)

// Register ...
func Register(cfg *config.Config, r *gin.Engine) {
	r.GET("/ping", Pong)
	r.GET("/version", Version)
	r.GET("/healthz", NewHealthzHandleFunc(cfg))

	// metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	pprofRouter := r.Group("/debug/pprof")
	if !cfg.Debug {
		pprofRouter.Use(gin.BasicAuth(gin.Accounts{
			"bk-iam": "DebugModel@bk",
		}))
	}
	{
		pprofRouter.GET("/", pprofHandler(pprof.Index))
		pprofRouter.GET("/cmdline", pprofHandler(pprof.Cmdline))
		pprofRouter.GET("/profile", pprofHandler(pprof.Profile))
		pprofRouter.POST("/symbol", pprofHandler(pprof.Symbol))
		pprofRouter.GET("/symbol", pprofHandler(pprof.Symbol))
		pprofRouter.GET("/trace", pprofHandler(pprof.Trace))
		pprofRouter.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
		pprofRouter.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
		pprofRouter.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
		pprofRouter.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
		pprofRouter.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
		pprofRouter.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	}

	// swagger docs
	if cfg.Debug {
		url := ginSwagger.URL("/swagger/doc.json")
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	}

}

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Metrics godoc
// @Summary prometheus metrics
// @Description /metrics
// @ID metrics
// @Tags basic
// @Accept json
// @Produce json
// @Success 200 {string} string metrics_text
// @Header 200 {string} X-Request-Id "the request id"
// @Router /metrics [get]
