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

import "github.com/zhengyansheng/jupiter/pkg/server"

type service struct {
	server.Server
}

func Service(srv server.Server) server.Server {
	return &service{Server: srv}
}

func (s *service) Serve() error {

	return s.Server.Serve()
}
