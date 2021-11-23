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

package logging

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"engine/pkg/config"
)

var loggerInitOnce sync.Once

var apiLogger *zap.Logger

var componentLogger *logrus.Logger
var syncLogger *logrus.Logger
var esLogger *logrus.Logger

// InitLogger ...
func InitLogger(logger *config.Logger) {
	initSystemLogger(&logger.System)

	loggerInitOnce.Do(func() {

		apiLogger = newZapJSONLogger(&logger.API)

		syncLogger = logrus.New()
		initJSONLogger(syncLogger, &logger.Sync)

		componentLogger = logrus.New()
		initJSONLogger(componentLogger, &logger.Component)

		esLogger = logrus.New()
		initJSONLogger(esLogger, &logger.ES)
	})
}

// GetAPILogger api log
func GetAPILogger() *zap.Logger {
	// if not init yet, use system logger
	if apiLogger == nil {
		apiLogger, _ = zap.NewProduction()
		defer func() {
			_ = apiLogger.Sync()
		}()
	}

	return apiLogger
}

func newZapJSONLogger(cfg *config.LogConfig) *zap.Logger {
	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	//w := zapcore.AddSync(writer)
	w := &zapcore.BufferedWriteSyncer{
		WS:            zapcore.AddSync(writer),
		Size:          256 * 1024, // 256 kB
		FlushInterval: 30 * time.Second,
	}

	// 设置日志级别
	l, err := parseZapLogLevel(cfg.Level)
	if err != nil {
		fmt.Println("api logger settings level invalid, will use level: info")
		l = zap.InfoLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		l,
	)
	return zap.New(core)
}

// parseZapLogLevel takes a string level and returns the zap log level constant.
func parseZapLogLevel(lvl string) (zapcore.Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return zap.PanicLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "warn", "warning":
		return zap.WarnLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "debug":
		return zap.DebugLevel, nil
	}

	var l zapcore.Level
	return l, fmt.Errorf("not a valid logrus Level: %q", lvl)
}

func initSystemLogger(cfg *config.LogConfig) {
	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	// 日志输出到stdout
	logrus.SetOutput(writer)
	// 设置日志格式, 不需要颜色
	// logrus.SetFormatter(&logrus.TextFormatter{
	// 	DisableColors:   true,
	// 	FullTimestamp:   true,
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	// 设置日志级别
	l, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		fmt.Println("system logger settings level invalid, will use level: info")
		l = logrus.InfoLevel
		// panic(err)
	}
	logrus.SetLevel(l)
	// https://github.com/sirupsen/logrus#logging-method-name
	// DONT OPEN IT
	// 显示代码行数
	// logrus.SetReportCaller(true)
}

func initJSONLogger(jsonLogger *logrus.Logger, cfg *config.LogConfig) {
	writer, err := getWriter(cfg.Writer, cfg.Settings)
	if err != nil {
		panic(err)
	}
	jsonLogger.SetOutput(writer)

	// apiLogger.SetFormatter(&logrus.JSONFormatter{
	// 	TimestampFormat: "2006-01-02 15:04:05",
	// })
	jsonLogger.SetFormatter(&JSONFormatter{})
	// 设置日志级别
	l, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		fmt.Println("api logger settings level invalid, will use level: info")
		l = logrus.InfoLevel
		// panic(err)
	}
	jsonLogger.SetLevel(l)
}

// GetSystemLogger :
func GetSystemLogger() *logrus.Logger {
	return logrus.StandardLogger()
}

// GetSyncLogger ...
func GetSyncLogger() *logrus.Logger {
	// if not init yet, use system logger
	if syncLogger == nil {
		return logrus.StandardLogger()
	}
	return syncLogger
}

// GetComponentLogger ...
func GetComponentLogger() *logrus.Logger {
	// if not init yet, use system logger
	if componentLogger == nil {
		return logrus.StandardLogger()
	}
	return componentLogger
}

// GetESLogger ...
func GetESLogger() *logrus.Logger {
	// if not init yet, use system logger
	if esLogger == nil {
		return logrus.StandardLogger()
	}
	return esLogger
}
