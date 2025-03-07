// Copyright 2020 zhengyansheng
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

package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/fatih/color"
	"github.com/zhengyansheng/jupiter/pkg"
	"github.com/zhengyansheng/jupiter/pkg/core/ecode"
	"github.com/zhengyansheng/jupiter/pkg/core/sentinel"
	"github.com/zhengyansheng/jupiter/pkg/util/xstring"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	errSlowCommand = errors.New("grpc unary slow command")
)

// metric统计
func metricUnaryClientInterceptor(name string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		return err
	}
}

func sentinelUnaryClientInterceptor(name string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		entry, blockerr := sentinel.Entry(name,
			api.WithResourceType(base.ResTypeRPC),
			api.WithTrafficType(base.Outbound))
		if blockerr != nil {
			return blockerr
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		entry.Exit(base.WithError(err))

		return err
	}
}

func debugUnaryClientInterceptor(addr string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var p peer.Peer
		prefix := fmt.Sprintf("[%s]", addr)
		if remote, ok := peer.FromContext(ctx); ok && remote.Addr != nil {
			prefix = prefix + "(" + remote.Addr.String() + ")"
		}

		fmt.Printf("%-50s[%s] => %s\n", color.GreenString(prefix), time.Now().Format("04:05.000"), color.GreenString("Send: "+method+" | "+xstring.Json(req)))
		err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Peer(&p))...)
		if err != nil {
			fmt.Printf("%-50s[%s] => %s\n", color.RedString(prefix), time.Now().Format("04:05.000"), color.RedString("Erro: "+err.Error()))
		} else {
			fmt.Printf("%-50s[%s] => %s\n", color.GreenString(prefix), time.Now().Format("04:05.000"), color.GreenString("Recv: "+xstring.Json(reply)))
		}

		return err
	}
}

func TraceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {

		return err
	}
}

func aidUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		clientAidMD := metadata.Pairs("aid", pkg.AppID())
		if ok {
			md = metadata.Join(md, clientAidMD)
		} else {
			md = clientAidMD
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// timeoutUnaryClientInterceptor gRPC客户端超时拦截器
func timeoutUnaryClientInterceptor(_logger *xlog.Logger, timeout time.Duration, slowThreshold time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		now := time.Now()
		// 若无自定义超时设置，默认设置超时
		_, ok := ctx.Deadline()
		if !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		du := time.Since(now)
		remoteIP := "unknown"
		if remote, ok := peer.FromContext(ctx); ok && remote.Addr != nil {
			remoteIP = remote.Addr.String()
		}

		if slowThreshold > time.Duration(0) && du > slowThreshold {
			_logger.Error("slow",
				xlog.FieldErr(errSlowCommand),
				xlog.FieldMethod(method),
				xlog.FieldName(cc.Target()),
				xlog.FieldCost(du),
				xlog.FieldAddr(remoteIP),
			)
		}
		return err
	}
}

// loggerUnaryClientInterceptor gRPC客户端日志中间件
func loggerUnaryClientInterceptor(_logger *xlog.Logger, name string, accessInterceptorLevel string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)

		spbStatus := ecode.ExtractCodes(err)
		if err != nil {
			// 只记录系统级别错误
			if spbStatus.Code < ecode.EcodeNum {
				// 只记录系统级别错误
				_logger.Error(
					"access",
					xlog.FieldType("unary"),
					xlog.FieldCode(spbStatus.Code),
					xlog.FieldStringErr(spbStatus.Message),
					xlog.FieldName(name),
					xlog.FieldMethod(method),
					xlog.FieldCost(time.Since(beg)),
					xlog.Any("req", json.RawMessage(xstring.Json(req))),
					xlog.Any("reply", json.RawMessage(xstring.Json(reply))),
				)
			} else {
				// 业务报错只做warning
				_logger.Warn(
					"access",
					xlog.FieldType("unary"),
					xlog.FieldCode(spbStatus.Code),
					xlog.FieldStringErr(spbStatus.Message),
					xlog.FieldName(name),
					xlog.FieldMethod(method),
					xlog.FieldCost(time.Since(beg)),
					xlog.Any("req", json.RawMessage(xstring.Json(req))),
					xlog.Any("reply", json.RawMessage(xstring.Json(reply))),
				)
			}
			return err
		} else {
			if accessInterceptorLevel == "info" {
				_logger.Info(
					"access",
					xlog.FieldType("unary"),
					xlog.FieldCode(spbStatus.Code),
					xlog.FieldName(name),
					xlog.FieldMethod(method),
					xlog.FieldCost(time.Since(beg)),
					xlog.Any("req", json.RawMessage(xstring.Json(req))),
					xlog.Any("reply", json.RawMessage(xstring.Json(reply))),
				)
			}
		}

		return nil
	}
}
