/**
 * Tencent is pleased to support the open source community by making CL5 available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cache

/**
 * Config 缓存配置
 */
type Config struct {
	Open      bool `yaml:"open"`
	Resources []ConfigEntry
}

/*
 * ConfigEntry 单个缓存资源配置
 */
type ConfigEntry struct {
	Name   string                 `yaml:"name"`
	Option map[string]interface{} `yaml:"option"`
}

var (
	config *Config
)

/**
 * SetCacheConfig 设置缓存配置
 */
func SetCacheConfig(conf *Config) {
	config = conf
}
