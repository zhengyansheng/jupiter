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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zhengyansheng/jupiter/pkg/client/etcdv3"
	"github.com/zhengyansheng/jupiter/pkg/core/constant"
	"github.com/zhengyansheng/jupiter/pkg/registry"
	"github.com/zhengyansheng/jupiter/pkg/server"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

func Test_etcdv3Registry(t *testing.T) {
	etcdConfig := etcdv3.DefaultConfig()
	etcdConfig.Endpoints = []string{"localhost:2379"}
	registry, err := newETCDRegistry(&Config{
		Config:      etcdConfig,
		ReadTimeout: time.Second * 10,
		Prefix:      "jupiter",
		logger:      xlog.Jupiter(),
	})

	assert.Nil(t, err)
	assert.Nil(t, registry.RegisterService(context.Background(), &server.ServiceInfo{
		Name:       "service_1",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9091",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceProvider,
		Deployment: "default",
		Group:      "",
	}))

	services, err := registry.ListServices(context.Background(), "grpc:service_1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, "10.10.10.1:9091", services[0].Address)

	go func() {
		si := &server.ServiceInfo{
			Name:       "service_1",
			Scheme:     "grpc",
			Address:    "10.10.10.1:9092",
			Enable:     true,
			Healthy:    true,
			Metadata:   map[string]string{},
			Region:     "default",
			Zone:       "default",
			Deployment: "default",
		}
		time.Sleep(time.Second)
		assert.Nil(t, registry.RegisterService(context.Background(), si))
		assert.Nil(t, registry.UnregisterService(context.Background(), si))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		endpoints, err := registry.WatchServices(ctx, "grpc:service_1")
		assert.Nil(t, err)
		for msg := range endpoints {
			t.Logf("watch service: %+v\n", msg)
			// 	assert.Equal(t, "10.10.10.2:9092", msg)
		}
	}()

	time.Sleep(time.Second * 3)
	cancel()
	_ = registry.Close()
	time.Sleep(time.Second * 1)
}

func Test_etcdv3registry_UpdateAddressList(t *testing.T) {
	etcdConfig := etcdv3.DefaultConfig()
	etcdConfig.Endpoints = []string{"localhost:2379"}
	reg, err := newETCDRegistry(&Config{
		Config:      etcdConfig,
		ReadTimeout: time.Second * 10,
		Prefix:      "jupiter",
		logger:      xlog.Jupiter(),
	})

	assert.Nil(t, err)

	var routeConfig = registry.RouteConfig{
		ID:         "1",
		Scheme:     "grpc",
		Host:       "",
		Deployment: "openapi",
		URI:        "/hello",
		Upstream: registry.Upstream{
			Nodes: map[string]int{
				"10.10.10.1:9091": 1,
				"10.10.10.1:9092": 10,
			},
		},
	}
	_, err = reg.client.Put(context.Background(), "/jupiter/service_1/configurators/grpc:///routes/1", routeConfig.String())
	assert.Nil(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		services, err := reg.WatchServices(ctx, "grpc:service_1")
		assert.Nil(t, err)
		fmt.Printf("len(services) = %+v\n", len(services))
		for service := range services {
			fmt.Printf("service = %+v\n", service)
		}
	}()
	time.Sleep(time.Second * 3)
	cancel()
	_ = reg.Close()
	time.Sleep(time.Second * 1)
}

func TestKeepalive(t *testing.T) {
	etcdConfig := etcdv3.DefaultConfig()
	etcdConfig.Endpoints = []string{"localhost:2379"}
	reg, err := newETCDRegistry(&Config{
		Config:      etcdConfig,
		ReadTimeout: time.Second * 10,
		Prefix:      "jupiter",
		logger:      xlog.Jupiter(),
		ServiceTTL:  time.Second,
	})
	assert.Nil(t, err)
	assert.Nil(t, reg.RegisterService(context.Background(), &server.ServiceInfo{
		Name:       "service_2",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9091",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceProvider,
		Deployment: "default",
		Group:      "",
	}))
	assert.Nil(t, reg.RegisterService(context.Background(), &server.ServiceInfo{
		Name:       "service_2",
		AppID:      "",
		Scheme:     "grpc",
		Address:    "10.10.10.1:9092",
		Weight:     0,
		Enable:     true,
		Healthy:    true,
		Metadata:   map[string]string{},
		Region:     "default",
		Zone:       "default",
		Kind:       constant.ServiceGovernor,
		Deployment: "default",
		Group:      "",
	}))

	lease := reg.getLeaseID()
	reg.client.Revoke(context.Background(), lease)

	assert.Eventually(t, func() bool {
		return reg.getLeaseID() != 0 && lease != reg.getLeaseID()
	}, 5*time.Second, time.Second)

	ttl, err := reg.client.TimeToLive(context.Background(), lease)
	assert.Nil(t, err)
	assert.Equal(t, int64(-1), ttl.TTL)

	ttl, err = reg.client.TimeToLive(context.Background(), reg.getLeaseID())
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ttl.TTL)

	_ = reg.Close()
	time.Sleep(time.Second * 1)
}
