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
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"engine/pkg/server"
	"engine/pkg/task"
)

// cmd for iam
var cfgFile string

func init() {
	// cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml;required)")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")

	_ = rootCmd.MarkFlagRequired("config")
	viper.SetDefault("author", "blueking-paas")
}

var rootCmd = &cobra.Command{
	Use:   "bk-iam-search-engine",
	Short: "bi-iam-search-engine is the helper service of IAM",
	Long:  `bi-iam-search-engine is the helper service of IAM`,

	Run: func(cmd *cobra.Command, args []string) {
		Start()
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Start ...
func Start() {

	fmt.Println("It's IAM Engine")

	// init rand
	rand.Seed(time.Now().UnixNano())

	// 0. init config
	if cfgFile != "" {
		// Use config file from the flag.
		log.Infof("Load config file: %s", cfgFile)
		viper.SetConfigFile(cfgFile)
	}
	initConfig()

	if globalConfig.Debug {
		fmt.Println(globalConfig)
	}

	//begin := time.Now()
	//log.Info("begin to load all system's policies")
	//demo.LoadAllLocalPolicies(&globalConfig.Index)
	//log.Info("done all load! tooks:", time.Since(begin))
	// 1. init
	initLogger()
	// initSentry()
	initMetrics()
	initBackend()
	initStoragePath()
	initSentryEventReport(globalConfig.Sentry.Enable)
	initGlobalIndex()
	initSuperAppCode()
	initRedis()
	initRedisKeys()
	initCaches()

	// 2. watch the signal
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		interrupt(cancelFunc)
	}()

	// start the sync
	task.StartSync(ctx, globalConfig)

	// 3. start the server
	httpServer := server.NewServer(globalConfig)
	httpServer.Run(ctx)
}

// a context canceled when SIGINT or SIGTERM are notified
func interrupt(onSignal func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for s := range c {
		log.Infof("Caught signal %s. Exiting.", s)
		onSignal()
		close(c)
	}
}
