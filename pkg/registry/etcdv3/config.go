// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcdv3

import (
	"time"

	"github.com/spf13/cast"
	"github.com/zhengyansheng/jupiter/pkg/client/etcdv3"
	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"github.com/zhengyansheng/jupiter/pkg/core/ecode"
	"github.com/zhengyansheng/jupiter/pkg/core/singleton"
	"github.com/zhengyansheng/jupiter/pkg/registry"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

// StdConfig ...
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("registry." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	// 解析最外层配置
	if err := conf.UnmarshalKey(key, &config); err != nil {
		xlog.Jupiter().Panic("unmarshal key", xlog.FieldMod("registry.etcd"), xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.String("key", key), xlog.Any("config", config))
	}
	// 解析嵌套配置
	if err := conf.UnmarshalKey(key, &config.Config); err != nil {
		xlog.Jupiter().Panic("unmarshal key", xlog.FieldMod("registry.etcd"), xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.String("key", key), xlog.Any("config", config))
	}
	return config
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Config:      etcdv3.DefaultConfig(),
		ReadTimeout: time.Second * 3,
		Prefix:      "wsd-reg",
		logger:      xlog.Jupiter().Named(ecode.ModRegistryETCD),
		ServiceTTL:  cast.ToDuration("60s"),
	}
}

// Config ...
type Config struct {
	*etcdv3.Config
	ReadTimeout time.Duration
	ConfigKey   string
	Prefix      string
	ServiceTTL  time.Duration
	logger      *xlog.Logger
}

// Build ...
func (config Config) Build() (registry.Registry, error) {
	if config.ConfigKey != "" {
		config.Config = etcdv3.RawConfig(config.ConfigKey)
	}
	return newETCDRegistry(&config)
}

func (config Config) MustBuild() registry.Registry {
	reg, err := config.Build()
	if err != nil {
		xlog.Jupiter().Panic("build registry failed", zap.Error(err))
	}
	return reg
}

func (config *Config) Singleton() (registry.Registry, error) {
	if val, ok := singleton.Load(constant.ModuleClientEtcd, config.ConfigKey); ok {
		return val.(registry.Registry), nil
	}

	reg, err := config.Build()
	if err != nil {
		return nil, err
	}

	singleton.Store(constant.ModuleClientEtcd, config.ConfigKey, reg)

	return reg, nil
}

func (config *Config) MustSingleton() registry.Registry {
	reg, err := config.Singleton()
	if err != nil {
		xlog.Jupiter().Panic("build registry failed", zap.Error(err))
	}

	return reg
}
