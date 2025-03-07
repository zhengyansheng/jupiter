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

package rocketmq

import (
	"context"
	"strings"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zhengyansheng/jupiter/pkg/core/imeta"
	"github.com/zhengyansheng/jupiter/pkg/core/istats"
	"github.com/zhengyansheng/jupiter/pkg/core/sentinel"
	"github.com/zhengyansheng/jupiter/pkg/util/xdebug"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type FlowInfo struct {
	Name      string   `json:"name"`
	Addr      []string `json:"addr"`
	Topic     string   `json:"topic"`
	Group     string   `json:"group"`
	GroupType string   `json:"groupType"` // 类型， consumer 消费者， producer 生产者
	istats.FlowInfoBase
}

func consumeResultStr(result consumer.ConsumeResult) string {
	switch result {
	case consumer.ConsumeSuccess:
		return "success"
	case consumer.ConsumeRetryLater:
		return "retryLater"
	case consumer.Commit:
		return "commit"
	case consumer.Rollback:
		return "rollback"
	case consumer.SuspendCurrentQueueAMoment:
		return "suspendCurrentQueueAMoment"
	default:
		return "unknown"
	}
}

func consumerMetricInterceptor() primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		beg := time.Now()
		msgs, _ := req.([]*primitive.MessageExt)

		err := next(ctx, msgs, reply)
		if reply == nil {
			return err
		}

		holder := reply.(*consumer.ConsumeResultHolder)
		xdebug.PrintObject("consume", map[string]interface{}{
			"err":    err,
			"count":  len(msgs),
			"result": consumeResultStr(holder.ConsumeResult),
		})

		// 消息处理结果统计
		for _, msg := range msgs {
			host := msg.StoreHost
			topic := msg.Topic
			result := consumeResultStr(holder.ConsumeResult)
			if err != nil {
				xlog.Jupiter().Error("push consumer",
					xlog.String("topic", topic),
					xlog.String("host", host),
					xlog.String("result", result),
					xlog.Any("err", err))
			} else {
				xlog.Jupiter().Debug("push consumer",
					xlog.String("topic", topic),
					xlog.String("host", host),
					xlog.String("result", result),
				)
			}
		}
		return err
	}
}

func consumerSlowInterceptor(topic string, rwTimeout time.Duration) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		beg := time.Now()
		msgs, _ := req.([]*primitive.MessageExt)

		err := next(ctx, msgs, reply)
		if reply == nil {
			return err
		}

		holder := reply.(*consumer.ConsumeResultHolder)
		xdebug.PrintObject("consume", map[string]interface{}{
			"err":    err,
			"count":  len(msgs),
			"result": consumeResultStr(holder.ConsumeResult),
		})

		if rwTimeout > time.Duration(0) {
			if time.Since(beg) > rwTimeout {
				xlog.Jupiter().Error("slow",
					xlog.String("topic", topic),
					xlog.String("result", consumeResultStr(holder.ConsumeResult)),
					xlog.Any("cost", time.Since(beg).Seconds()),
				)
			}
		}

		return err
	}
}

func consumerSentinelInterceptor(add []string) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		msgs, _ := req.([]*primitive.MessageExt)

		entry, blockerr := sentinel.Entry(add[0],
			sentinel.WithResourceType(base.ResTypeMQ),
			sentinel.WithTrafficType(base.Inbound))
		if blockerr != nil {
			return blockerr
		}

		err := next(ctx, msgs, reply)
		entry.Exit(sentinel.WithError(err))

		return err
	}
}

func produceResultStr(result primitive.SendStatus) string {
	switch result {
	case primitive.SendOK:
		return "sendOk"
	case primitive.SendFlushDiskTimeout:
		return "sendFlushDiskTimeout"
	case primitive.SendFlushSlaveTimeout:
		return "sendFlushSlaveTimeout"
	case primitive.SendSlaveNotAvailable:
		return "sendSlaveNotAvailable"
	default:
		return "unknown"
	}
}

func producerDefaultInterceptor(producer *Producer) primitive.Interceptor {

	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		beg := time.Now()
		realReq := req.(*primitive.Message)
		realReply := reply.(*primitive.SendResult)

		var span trace.Span

		if producer.EnableTrace {

			md := metadata.New(nil)

			defer span.End()

			for k, v := range md {
				realReq.WithProperty(strings.ToLower(k), strings.Join(v, ","))
			}
		}

		err := next(ctx, realReq, realReply)
		if realReply == nil || realReply.MessageQueue == nil {
			return err
		}

		xdebug.PrintObject("produce", map[string]interface{}{
			"err":     err,
			"message": realReq,
			"result":  realReply.String(),
		})

		if producer.EnableTrace {
			if err != nil {
				span := trace.SpanFromContext(ctx)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		}

		// 消息处理结果统计
		topic := producer.Topic
		if err != nil {
			xlog.Jupiter().Error("produce",
				xlog.String("topic", topic),
				xlog.String("queue", ""),
				xlog.String("result", realReply.String()),
				xlog.Any("err", err),
			)
		} else {
			xlog.Jupiter().Debug("produce",
				xlog.String("topic", topic),
				xlog.Any("queue", realReply.MessageQueue),
				xlog.String("result", produceResultStr(realReply.Status)),
			)
		}

		if producer.RwTimeout > time.Duration(0) {
			if time.Since(beg) > producer.RwTimeout {
				xlog.Jupiter().Error("slow",
					xlog.String("topic", topic),
					xlog.String("result", realReply.String()),
					xlog.Any("cost", time.Since(beg).Seconds()),
				)
			}
		}

		return err
	}
}

// 统一 metadata 传递.
func producerMDInterceptor(producer *Producer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		if md, ok := imeta.FromContext(ctx); ok {
			realReq := req.(*primitive.Message)
			for k, v := range md {
				realReq.WithProperty(k, strings.Join(v, ","))
			}
		}
		err := next(ctx, req, reply)
		return err
	}
}

func producerSentinelInterceptor(producer *Producer) primitive.Interceptor {
	return func(ctx context.Context, req, reply interface{}, next primitive.Invoker) error {
		entry, blockerr := sentinel.Entry(producer.Addr[0],
			sentinel.WithResourceType(base.ResTypeMQ),
			sentinel.WithTrafficType(base.Outbound))
		if blockerr != nil {
			return blockerr
		}

		err := next(ctx, req, reply)
		entry.Exit(sentinel.WithError(err))

		return err
	}
}
