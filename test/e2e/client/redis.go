// Copyright 2022 Douyu
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

package client

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/zhengyansheng/jupiter/pkg/client/redis"
	tests "github.com/zhengyansheng/jupiter/test/e2e/framework"
)

var _ = ginkgo.Describe("[redis] e2e test", func() {
	ginkgo.DescribeTable("redis ", func(tc tests.RedisTestCase) {
		tests.RunRedisTestCase(tc)
	},
		ginkgo.Entry("set case", tests.RedisTestCase{
			Conf: &redis.Config{
				Master: struct {
					Addr string "json:\"addr\" toml:\"addr\""
				}{
					Addr: "localhost:6379",
				},
				Slaves: struct {
					Addr []string "json:\"addr\" toml:\"addr\""
				}{
					Addr: []string{"localhost:6379"},
				},
			},
			Args:         []interface{}{"Set", "test", "1"},
			OnMaster:     true,
			ExpectError:  nil,
			ExpectResult: interface{}("OK"),
		}),
		ginkgo.Entry("get case", tests.RedisTestCase{
			Conf: &redis.Config{
				Master: struct {
					Addr string "json:\"addr\" toml:\"addr\""
				}{
					Addr: "localhost:6379",
				},
				Slaves: struct {
					Addr []string "json:\"addr\" toml:\"addr\""
				}{
					Addr: []string{"localhost:6379"},
				},
			},
			Args:         []interface{}{"Get", "test"},
			OnMaster:     false,
			ExpectError:  nil,
			ExpectResult: interface{}("1"),
		}),
	)
})
