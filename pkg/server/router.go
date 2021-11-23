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

package server

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"engine/pkg/api/basic"
	"engine/pkg/api/search"
	"engine/pkg/config"
	"engine/pkg/middleware"
)

// NewRouter ...
func NewRouter(cfg *config.Config) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// disable console log color
	gin.DisableConsoleColor()

	//router := gin.Default()
	router := gin.New()

	// MW: gin default logger
	router.Use(gin.Logger())
	// MW: gin default recovery
	router.Use(gin.Recovery())
	// MW: request_id
	router.Use(middleware.RequestID())

	// https://github.com/getsentry/sentry-go/tree/master/gin
	if cfg.Sentry.Enable {
		// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: cfg.Sentry.DSN,
		}); err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}

		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	// basic apis
	basic.Register(cfg, router)
	//
	// apis
	apiRouter := router.Group("/api/v1")
	apiRouter.Use(middleware.Metrics())
	apiRouter.Use(middleware.APILogger())
	apiRouter.Use(middleware.NewClientAuthMiddleware(cfg))
	search.Register(apiRouter)

	return router
}
