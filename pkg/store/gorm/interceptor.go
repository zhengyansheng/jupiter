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

package gorm

import (
	"errors"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/zhengyansheng/jupiter/pkg/core/sentinel"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"gorm.io/gorm"
)

type (
	Handler     func(*gorm.DB)
	Interceptor func(dsn *DSN, op string, options *Config, next Handler) Handler
)

var errSlowCommand = errors.New("mysql slow command")

func metricInterceptor() Interceptor {
	return func(dsn *DSN, op string, options *Config, next Handler) Handler {
		return func(scope *gorm.DB) {
			_ = scope
		}
	}
}

func logSQL(sql string, args []interface{}, containArgs bool) string {
	if containArgs {
		return bindSQL(sql, args)
	}
	return sql
}

func traceInterceptor() Interceptor {

	return func(dsn *DSN, op string, options *Config, next Handler) Handler {
		return func(scope *gorm.DB) {
			if ctx := scope.Statement.Context; ctx != nil {

				return
			}

			next(scope)
		}
	}
}

func sentinelInterceptor() Interceptor {
	return func(dsn *DSN, op string, options *Config, next Handler) Handler {
		return func(scope *gorm.DB) {
			entry, blockerr := sentinel.Entry(dsn.Addr,
				sentinel.WithResourceType(base.ResTypeDBSQL),
				sentinel.WithTrafficType(base.Outbound),
			)
			if blockerr != nil {
				_ = scope.AddError(blockerr)

				return
			}

			next(scope)

			entry.Exit(sentinel.WithError(scope.Error))
		}
	}
}
