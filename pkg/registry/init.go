// Copyright 2021 rex lv
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

package registry

import (
	"log"

	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

// var _registerers = sync.Map{}
var registryBuilder = make(map[string]Builder)

type Config map[string]struct {
	Kind          string `json:"kind" description:"底层注册器类型, eg: etcdv3, consul"`
	ConfigKey     string `json:"configKey" description:"底册注册器的配置键"`
	DeplaySeconds int    `json:"deplaySeconds" description:"延迟注册"`
}

// default register
var DefaultRegisterer Registry = &Local{}

func init() {
	// 初始化注册中心
	conf.OnLoaded(func(c *conf.Configuration) {
		xlog.Jupiter().Sugar().Info("hook config, init registry")
		var config Config
		if err := c.UnmarshalKey(constant.ConfigKey("registry"), &config); err != nil {
			xlog.Jupiter().Sugar().Infof("hook config, read registry config failed: %v", err)
			return
		}

		for name, item := range config {
			var itemKind = item.Kind
			if itemKind == "" {
				itemKind = "etcdv3"
			}

			if item.ConfigKey == "" {
				item.ConfigKey = constant.ConfigKey("registry.default")
			}

			build, ok := registryBuilder[itemKind]
			if !ok {
				xlog.Jupiter().Sugar().Infof("invalid registry kind: %s", itemKind)
				continue
			}

			xlog.Jupiter().Sugar().Infof("build registrerer %s with config: %s", name, item.ConfigKey)
			DefaultRegisterer = build(item.ConfigKey)
		}
	})
}

type Builder func(string) Registry

type BuildFunc func(string) (Registry, error)

func RegisterBuilder(kind string, build Builder) {
	if _, ok := registryBuilder[kind]; ok {
		log.Panicf("duplicate register registry builder: %s", kind)
	}
	registryBuilder[kind] = build
}
