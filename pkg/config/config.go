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

package config

import (
	"github.com/spf13/viper"
)

// Server ...
type Server struct {
	Host string
	Port int

	GraceTimeout int64

	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// Sentry ...
type Sentry struct {
	Enable bool
	DSN    string
}

// ElasticSearch ...
type ElasticSearch struct {
	Addresses  []string // A list of Elasticsearch nodes to use.
	Username   string   // Username for HTTP Basic Authentication.
	Password   string   // Password for HTTP Basic Authentication.
	MaxRetries int      // Default: 3.

	IndexName string
}

// Index ...
type Index struct {
	ElasticSearch ElasticSearch
}

// Logger ...
type Logger struct {
	System    LogConfig
	API       LogConfig
	Sync      LogConfig
	Component LogConfig
	ES        LogConfig
}

// LogConfig ...
type LogConfig struct {
	Level    string
	Writer   string
	Settings map[string]string
}

// Authorization ...
type Authorization struct {
	AppCode   string
	AppSecret string
}

// Backend ...
type Backend struct {
	Addr          string
	Authorization Authorization
}

// Storage ...
type Storage struct {
	Path string
}

// Crypto store the keys for crypto
type Crypto struct {
	ID  string
	Key string
}

// Redis ...
type Redis struct {
	ID           string
	Type         string
	Addr         string
	Password     string
	DB           int
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
	PoolSize     int
	MinIdleConns int

	// mode=sentinel required
	SentinelAddr     string
	MasterName       string
	SentinelPassword string
}

// Config ...
type Config struct {
	Debug bool

	Index  Index
	Sentry Sentry
	Server Server

	Backend Backend

	Storage Storage

	Logger Logger

	SuperAppCode string

	Cryptos map[string]*Crypto

	Redis Redis // NOTE 需要扩展的时候变更为map
}

// Load 从viper中读取配置文件
func Load(v *viper.Viper) (*Config, error) {
	var cfg Config
	// 将配置信息绑定到结构体上
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
