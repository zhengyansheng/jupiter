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

package server

import (
	"net/http"
	"time"

	"github.com/gogf/gf/net/ghttp"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zhengyansheng/jupiter/pkg/client/resty"
	"github.com/zhengyansheng/jupiter/pkg/server/xgoframe"
	"github.com/zhengyansheng/jupiter/test/e2e/framework"
)

var _ = ginkgo.Describe("[xgoframe] e2e test", func() {
	var server *xgoframe.Server

	ginkgo.BeforeEach(func() {
		server = xgoframe.DefaultConfig().MustBuild()
		server.BindHandler("/test", func(r *ghttp.Request) {
			r.Response.Writeln("hello")
		})
		go func() {
			err := server.Serve()
			assert.Nil(ginkgo.GinkgoT(), err)
		}()
		time.Sleep(time.Second)
	})

	ginkgo.AfterEach(func() {
		_ = server.Stop()
	})

	ginkgo.DescribeTable("xgoframe", func(htc framework.HTTPTestCase) {
		framework.RunHTTPTestCase(htc)
	}, ginkgo.Entry("normal case", framework.HTTPTestCase{
		Conf: &resty.Config{
			Addr: "http://localhost:8099",
		},
		Method:       "GET",
		Path:         "/test",
		ExpectStatus: http.StatusOK,
		ExpectBody:   "hello",
	}))
})
