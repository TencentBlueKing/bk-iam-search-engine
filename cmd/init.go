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

package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"engine/pkg/cache/impls"
	"engine/pkg/components"
	"engine/pkg/config"
	"engine/pkg/errorx"
	"engine/pkg/indexer"
	"engine/pkg/logging"
	"engine/pkg/metric"
	"engine/pkg/redis"
	"engine/pkg/storage"
	"engine/pkg/task"
)

var globalConfig *config.Config

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		panic("Config file missing")
	}
	// Use config file from the flag.
	// viper.SetConfigFile(cfgFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Using config file: %s, read fail: err=%v", viper.ConfigFileUsed(), err))
	}
	var err error
	globalConfig, err = config.Load(viper.GetViper())
	if err != nil {
		panic(fmt.Sprintf("Could not load configurations from file, error: %v", err))
	}
}

func initMetrics() {
	metric.InitMetrics()
	log.Info("init Metrics success")
}

func initBackend() {
	backend := globalConfig.Backend
	if backend.Addr == "" {
		panic("backend addr should be configured")
	}

	if backend.Authorization.AppCode == "" || backend.Authorization.AppSecret == "nil" {
		panic("backend authorization app_code and app_secret should not be empty")
	}

	components.InitComponentClients(
		backend.Addr, backend.Authorization.AppCode, backend.Authorization.AppSecret,
	)

	log.Info("init Hosts success")
}

func initLogger() {
	logging.InitLogger(&globalConfig.Logger)
}

func initStoragePath() {
	storage.InitStoragePath(globalConfig.Storage.Path)

	log.Info("init local data path success")
}

func initSentryEventReport(sentryEnabled bool) {
	errorx.InitErrorReport(sentryEnabled)
}

func initGlobalIndex() {
	indexer.InitGlobalIndex(&globalConfig.Index)

}

func initCaches() {
	impls.InitCaches(false)
}

func initSuperAppCode() {
	config.InitSuperAppCode(globalConfig.SuperAppCode)
}

func initRedis() {
	redis.InitRedisClient(false, &globalConfig.Redis)
}

func initRedisKeys() {
	keys := make(map[string]string, len(globalConfig.RedisKeys))
	for _, redisKey := range globalConfig.RedisKeys {
		keys[redisKey.ID] = redisKey.Key
	}
	task.InitDeleteQueueKey(keys)
}
