// Copyright 2022 zhengyansheng
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

package xlog

import (
	"fmt"
	"log"
	"time"

	"github.com/zhengyansheng/jupiter/pkg"
	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	conf.OnLoaded(func(c *conf.Configuration) {
		prefix := constant.GetConfigPrefix()
		// compatible with app.logger
		if conf.Get("app.logger") != nil {
			prefix = "app"
		}

		log.Print("hook config, init loggers")

		key := prefix + ".logger.default"
		log.Printf("reload default logger with configKey: %s", key)
		SetDefault(RawConfig(key).Build())

		key = prefix + ".logger.jupiter"
		log.Printf("reload jupiter logger with configKey: %s", key)
		SetJupiter(jupiterConfig(prefix).Build())
	})
}

// Config ...
type Config struct {
	// Dir 日志输出目录
	Dir string
	// Name 日志文件名称
	Name string
	// Level 日志初始等级
	Level string
	// 日志初始化字段
	Fields []zap.Field
	// 是否添加调用者信息
	AddCaller bool
	// 日志前缀
	Prefix string
	// 日志输出文件最大长度，超过改值则截断
	MaxSize   int
	MaxAge    int
	MaxBackup int
	// 日志磁盘刷盘间隔
	Interval      time.Duration
	CallerSkip    int
	Async         bool
	Queue         bool
	QueueSleep    time.Duration
	Core          zapcore.Core
	Debug         bool
	EncoderConfig *zapcore.EncoderConfig
	configKey     string
}

// Filename ...
func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	config, _ = conf.UnmarshalWithExpect(key, config).(*Config)
	config.configKey = key
	return config
}

// StdConfig Jupiter Standard logger config
func StdConfig(prefix, name string) *Config {
	return RawConfig(prefix + ".logger." + name)
}

// DefaultConfig for application.
func DefaultConfig() *Config {
	return &Config{
		Name:          "jupiter_default.json",
		Dir:           pkg.LogDir(),
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		Interval:      24 * time.Hour,
		CallerSkip:    0,
		AddCaller:     true,
		Async:         true,
		Queue:         false,
		QueueSleep:    100 * time.Millisecond,
		EncoderConfig: DefaultZapConfig(),
		Fields: []zap.Field{
			String("aid", pkg.AppID()),
			String("iid", pkg.AppInstance()),
		},
	}
}

// jupiterConfig for framework.
func jupiterConfig(prefix string) *Config {
	config := DefaultConfig()
	config.Name = "jupiter_framework.sys"
	config, _ = conf.UnmarshalWithExpect(prefix+".logger.jupiter", config).(*Config)

	return config
}

// Build ...
func (config Config) Build() *Logger {
	if config.EncoderConfig == nil {
		config.EncoderConfig = DefaultZapConfig()
	}
	if config.Debug {
		config.EncoderConfig.EncodeLevel = DebugEncodeLevel
	}
	logger := newLogger(&config)

	return logger
}
