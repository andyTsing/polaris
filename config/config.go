/**
 * Tencent is pleased to support the open source community by making Polaris available.
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

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/polarismesh/polaris-server/apiserver"
	"github.com/polarismesh/polaris-server/common/log"
	"github.com/polarismesh/polaris-server/naming"
	"github.com/polarismesh/polaris-server/naming/cache"
	"github.com/polarismesh/polaris-server/plugin"
	"github.com/polarismesh/polaris-server/store"
	yaml "gopkg.in/yaml.v2"
)

// Config 配置
type Config struct {
	Bootstrap  Bootstrap          `yaml:"bootstrap"`
	APIServers []apiserver.Config `yaml:"apiservers"`
	Cache      cache.Config       `yaml:"cache"`
	Naming     naming.Config      `yaml:"naming"`
	Store      store.Config       `yaml:"store"`
	Plugin     plugin.Config      `yaml:"plugin"`
}

// Bootstrap 启动引导配置
type Bootstrap struct {
	Logger         log.Options
	StartInOrder   map[string]interface{} `yaml:"startInOrder"`
	PolarisService PolarisService         `yaml:"polaris_service"`
}

// PolarisService polaris-server的自注册配置
type PolarisService struct {
	EnableRegister bool       `yaml:"enable_register"`
	ProbeAddress   string     `yaml:"probe_address"`
	Isolated       bool       `yaml:"isolated"`
	Services       []*Service `yaml:"services"`
}

// Service 服务的自注册的配置
type Service struct {
	Name      string   `yaml:"name"`
	Namespace string   `yaml:"namespace"`
	Protocols []string `yaml:"protocols"`
}

// APIEntries 对外提供的apiServers
type APIEntries struct {
	Name      string   `yaml:"name"`
	Protocols []string `yaml:"protocols"`
}

const (
	// DefaultPolarisName default polaris name
	DefaultPolarisName = "polaris-server"
	// DefaultPolarisNamespace default namespace
	DefaultPolarisNamespace = "Polaris"
	// DefaultFilePath default file path
	DefaultFilePath = "polaris-server.yaml"
)

// Load 加载配置
func Load(filePath string) (*Config, error) {
	if filePath == "" {
		err := errors.New("invalid config file path")
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	fmt.Printf("[INFO] load config from %v\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return nil, err
	}

	return config, nil
}
