/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云 - 权限中心检索引擎
 * (BlueKing-IAM-Search-Engine) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package redis

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"engine/pkg/config"
	"engine/pkg/util"
)

const (
	ModeStandalone = "standalone"
	ModeSentinel   = "sentinel"
)

var mqRedisClient *redis.Client

var mqRedisClientInitOnce sync.Once

// InitRedisClient ...
func InitRedisClient(debugMode bool, redisConfig *config.Redis) {
	if mqRedisClient == nil {
		mqRedisClientInitOnce.Do(func() {
			switch redisConfig.Type {
			case ModeStandalone:
				mqRedisClient = newStandaloneClient(redisConfig)
			case ModeSentinel:
				mqRedisClient = newSentinelClient(redisConfig)
			default:
				panic(fmt.Errorf("no support redis config type: %s", redisConfig.Type))
			}

			_, err := mqRedisClient.Ping(context.TODO()).Result()
			if err != nil {
				log.WithError(err).Error("connect to redis fail")
				// redis is important
				if !debugMode {
					panic(err)
				}
			}
		})
	}
}

func newStandaloneClient(cfg *config.Redis) *redis.Client {
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	// set default options
	opt.DialTimeout = time.Duration(2) * time.Second
	opt.ReadTimeout = time.Duration(1) * time.Second
	opt.WriteTimeout = time.Duration(1) * time.Second
	opt.PoolSize = 20 * runtime.NumCPU()
	opt.MinIdleConns = 10 * runtime.NumCPU()
	opt.IdleTimeout = time.Duration(3) * time.Minute

	// set custom options, from config.yaml
	if cfg.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(cfg.DialTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}

	if cfg.PoolSize > 0 {
		opt.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opt.MinIdleConns = cfg.MinIdleConns
	}

	// TLS configuration
	if cfg.TLS.Enabled {
		tlsConfig, err := util.NewTLSConfig(
			cfg.TLS.CertCaFile, cfg.TLS.CertFile, cfg.TLS.CertKeyFile, cfg.TLS.InsecureSkipVerify,
		)
		if err != nil {
			log.Fatalf("redis tls config init: %s", err)
		}
		opt.TLSConfig = tlsConfig
	}

	log.Infof(
		"connect to redis: "+
			"%s [db=%d, dialTimeout=%s, readTimeout=%s, writeTimeout=%s, poolSize=%d, minIdleConns=%d, idleTimeout=%s]",
		opt.Addr,
		opt.DB,
		opt.DialTimeout,
		opt.ReadTimeout,
		opt.WriteTimeout,
		opt.PoolSize,
		opt.MinIdleConns,
		opt.IdleTimeout,
	)

	return redis.NewClient(opt)
}

func newSentinelClient(cfg *config.Redis) *redis.Client {
	sentinelAddrs := strings.Split(cfg.SentinelAddr, ",")
	opt := &redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: sentinelAddrs,
		DB:            cfg.DB,
		Password:      cfg.Password,
	}

	if cfg.SentinelPassword != "" {
		opt.SentinelPassword = cfg.SentinelPassword
	}

	// set default options
	opt.DialTimeout = 2 * time.Second
	opt.ReadTimeout = 1 * time.Second
	opt.WriteTimeout = 1 * time.Second
	opt.PoolSize = 20 * runtime.NumCPU()
	opt.MinIdleConns = 10 * runtime.NumCPU()
	opt.IdleTimeout = 3 * time.Minute

	// set custom options, from config.yaml
	if cfg.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(cfg.DialTimeout) * time.Second
	}
	if cfg.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	if cfg.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(cfg.WriteTimeout) * time.Second
	}

	if cfg.PoolSize > 0 {
		opt.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opt.MinIdleConns = cfg.MinIdleConns
	}

	// TLS configuration
	// Note: TLS for Client To Sentinel、TLS for Client To Master are shared
	if cfg.TLS.Enabled {
		tlsConfig, err := util.NewTLSConfig(
			cfg.TLS.CertCaFile, cfg.TLS.CertFile, cfg.TLS.CertKeyFile, cfg.TLS.InsecureSkipVerify,
		)
		if err != nil {
			log.Fatalf("redis tls config init: %s", err)
		}
		opt.TLSConfig = tlsConfig
	}

	return redis.NewFailoverClient(opt)
}

// GetDefaultRedisClient 获取默认的 Redis 实例
func GetDefaultMQRedisClient() *redis.Client {
	return mqRedisClient
}
