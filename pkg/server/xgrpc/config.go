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

package xgrpc

import (
	"fmt"

	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"github.com/zhengyansheng/jupiter/pkg/core/ecode"
	"github.com/zhengyansheng/jupiter/pkg/flag"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config ...
type Config struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Deployment string `json:"deployment"`
	// Network network type, tcp4 by default
	Network string `json:"network" toml:"network"`
	// EnableAccessLog enable Access Interceptor, true by default
	EnableAccessLog bool
	// DisableTrace disable Trace Interceptor, false by default
	DisableTrace bool
	// DisableMetric disable Metric Interceptor, false by default
	DisableMetric bool
	// DisableSentinel disable Sentinel Interceptor, false by default
	DisableSentinel bool
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64
	// ServiceAddress service address in registry info, default to 'Host:Port'
	ServiceAddress string
	// EnableTLS
	EnableTLS bool
	// CaFile
	CaFile string
	// CertFile
	CertFile string
	// PrivateFile
	PrivateFile string

	Labels map[string]string `json:"labels"`

	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor

	logger *xlog.Logger
}

// StdConfig represents Standard gRPC Server config
// which will parse config by conf package,
// panic if no config key found in conf
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil {
		config.logger.Panic("grpc server parse config panic",
			xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr),
			xlog.FieldErr(err), xlog.FieldKey(key),
			xlog.FieldValueAny(config),
		)
	}
	return config
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      flag.String("host"),
		Port:                      9092,
		Deployment:                constant.DefaultDeployment,
		EnableAccessLog:           true,
		DisableMetric:             false,
		DisableTrace:              false,
		EnableTLS:                 false,
		SlowQueryThresholdInMilli: 500,
		logger:                    xlog.Jupiter().Named(ecode.ModGrpcServer),
		serverOptions:             []grpc.ServerOption{},
		streamInterceptors:        []grpc.StreamServerInterceptor{},
		unaryInterceptors:         []grpc.UnaryServerInterceptor{},
	}
}

// WithServerOption inject server option to grpc server
// User should not inject interceptor option, which is recommend by WithStreamInterceptor
// and WithUnaryInterceptor
func (config *Config) WithServerOption(options ...grpc.ServerOption) *Config {
	if config.serverOptions == nil {
		config.serverOptions = make([]grpc.ServerOption, 0)
	}
	config.serverOptions = append(config.serverOptions, options...)
	return config
}

// WithStreamInterceptor inject stream interceptors to server option
func (config *Config) WithStreamInterceptor(intes ...grpc.StreamServerInterceptor) *Config {
	if config.streamInterceptors == nil {
		config.streamInterceptors = make([]grpc.StreamServerInterceptor, 0)
	}

	config.streamInterceptors = append(config.streamInterceptors, intes...)
	return config
}

// WithUnaryInterceptor inject unary interceptors to server option
func (config *Config) WithUnaryInterceptor(intes ...grpc.UnaryServerInterceptor) *Config {
	if config.unaryInterceptors == nil {
		config.unaryInterceptors = make([]grpc.UnaryServerInterceptor, 0)
	}

	config.unaryInterceptors = append(config.unaryInterceptors, intes...)
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		xlog.Jupiter().Panic("build xgrpc server", zap.Error(err))
	}
	return server
}

// Build ...
func (config *Config) Build() (*Server, error) {

	return newServer(config)
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// Address ...
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
