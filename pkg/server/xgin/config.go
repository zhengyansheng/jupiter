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

package xgin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"github.com/zhengyansheng/jupiter/pkg/core/ecode"
	"github.com/zhengyansheng/jupiter/pkg/flag"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

// ModName ..
const ModName = "server.gin"

// Config HTTP config
type Config struct {
	Host          string
	Port          int
	Deployment    string
	Mode          string
	DisableMetric bool
	DisableTrace  bool
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string

	SlowQueryThresholdInMilli int64

	logger *xlog.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      flag.String("host"),
		Port:                      9091,
		Mode:                      gin.ReleaseMode,
		SlowQueryThresholdInMilli: 500, // 500ms
		logger:                    xlog.Jupiter().With(xlog.FieldMod(ModName)),
	}
}

// StdConfig Jupiter Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.FieldKey(key), xlog.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) MustBuild() *Server {
	server := newServer(config)
	server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))

	if !config.DisableMetric {
		server.Use(metricServerInterceptor())
	}

	if !config.DisableTrace {
		server.Use(traceServerInterceptor())
	}
	return server
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
