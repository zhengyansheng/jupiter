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

package server

import (
	"context"
	"fmt"

	"github.com/zhengyansheng/jupiter/pkg"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
)

type Option func(c *ServiceInfo)

// ConfigInfo ServiceConfigurator represents service configurator
type ConfigInfo struct {
	Routes []Route
}

// ServiceInfo represents service info
type ServiceInfo struct {
	Name     string               `json:"name"`
	AppID    string               `json:"appId"`
	Scheme   string               `json:"scheme"`
	Address  string               `json:"address"`
	Weight   float64              `json:"weight"`
	Enable   bool                 `json:"enable"`
	Healthy  bool                 `json:"healthy"`
	Metadata map[string]string    `json:"metadata"`
	Region   string               `json:"region"`
	Zone     string               `json:"zone"`
	Kind     constant.ServiceKind `json:"kind"`
	Version  string               `json:"version"`
	Mode     string               `json:"mode"`
	Hostname string               `json:"hostname"`
	// Deployment 部署组: 不同组的流量隔离
	// 比如某些服务给内部调用和第三方调用，可以配置不同的deployment,进行流量隔离
	Deployment string `json:"deployment"`
	// Group 流量组: 流量在Group之间进行负载均衡
	Group string `json:"group"`
}

// RegistryName returns the registry name of the service
func (s *ServiceInfo) RegistryName() string {
	return fmt.Sprintf("%s:%s:%s:%s/%s", s.Scheme, s.Name, "v1", pkg.AppMode(), s.Address)
}

func (s *ServiceInfo) ServicePrefix() string {
	return fmt.Sprintf("%s:%s:%s:%s/", s.Scheme, s.Name, "v1", pkg.AppMode())
}

// Label ...
func (si ServiceInfo) Label() string {
	return fmt.Sprintf("%s://%s", si.Scheme, si.Address)
}

// Equal allows the values to be compared by Attributes.Equal, this change is in order
// to fit the change in grpc-go:
// attributes: add Equal method; resolver: add AddressMap and State.BalancerAttributes (#4855)
func (si ServiceInfo) Equal(o interface{}) bool {
	oa, ok := o.(ServiceInfo)
	return ok &&
		oa.Name == si.Name &&
		oa.Address == si.Address &&
		oa.Kind == si.Kind &&
		oa.Group == si.Group
}

// Server ...
type Server interface {
	Serve() error
	Stop() error
	GracefulStop(ctx context.Context) error
	Info() *ServiceInfo
	Healthz() bool
}

// Route ...
type Route struct {
	// 权重组，按照
	WeightGroups []WeightGroup
	// 方法名
	Method string
}

// WeightGroup ...
type WeightGroup struct {
	Group  string
	Weight int
}

func ApplyOptions(options ...Option) ServiceInfo {
	info := defaultServiceInfo()
	for _, option := range options {
		option(&info)
	}
	return info
}

func WithMetaData(key, value string) Option {
	return func(c *ServiceInfo) {
		c.Metadata[key] = value
	}
}

func WithScheme(scheme string) Option {
	return func(c *ServiceInfo) {
		c.Scheme = scheme
	}
}

func WithAddress(address string) Option {
	return func(c *ServiceInfo) {
		c.Address = address
	}
}

func WithKind(kind constant.ServiceKind) Option {
	return func(c *ServiceInfo) {
		c.Kind = kind
	}
}

func defaultServiceInfo() ServiceInfo {
	si := ServiceInfo{
		Name:       pkg.Name(),
		AppID:      pkg.AppID(),
		Weight:     100,
		Enable:     true,
		Healthy:    true,
		Metadata:   make(map[string]string),
		Region:     pkg.AppRegion(),
		Zone:       pkg.AppZone(),
		Mode:       pkg.AppMode(),
		Hostname:   pkg.AppHost(),
		Version:    "v1",
		Kind:       0,
		Deployment: "",
		Group:      "",
	}
	si.Metadata["startTime"] = pkg.StartTime()
	si.Metadata["buildTime"] = pkg.BuildTime()
	si.Metadata["appVersion"] = pkg.AppVersion()
	si.Metadata["jupiterVersion"] = pkg.JupiterVersion()
	return si
}
