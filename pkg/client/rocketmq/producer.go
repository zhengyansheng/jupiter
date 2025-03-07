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

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/samber/lo"
	"github.com/zhengyansheng/jupiter/pkg/core/hooks"
	"github.com/zhengyansheng/jupiter/pkg/util/xdebug"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

type Producer struct {
	started bool

	rocketmq.Producer
	name string
	ProducerConfig
	interceptors []primitive.Interceptor
}

func StdNewProducer(name string) *Producer {
	return StdProducerConfig(name).Build()
}

func (conf *ProducerConfig) Build() *Producer {
	name := conf.Name

	if xdebug.IsDevelopmentMode() {
		xdebug.PrettyJsonPrint("rocketmq's config: "+name, conf)
	}

	cc := &Producer{
		name:           name,
		ProducerConfig: *conf,
		interceptors:   []primitive.Interceptor{},
	}

	cc.interceptors = append(cc.interceptors,
		producerDefaultInterceptor(cc),
		producerMDInterceptor(cc),
		producerSentinelInterceptor(cc),
	)

	// 服务启动前先start
	hooks.Register(hooks.Stage_BeforeRun, func() {
		_ = cc.Start()
	})

	return cc
}

func (pc *Producer) Start() error {
	if pc.started {
		return nil
	}

	// 兼容配置
	client, err := rocketmq.NewProducer(
		producer.WithGroupName(pc.Group),
		producer.WithNameServer(pc.Addr),
		producer.WithRetry(pc.Retry),
		producer.WithInterceptor(pc.interceptors...),
		producer.WithInstanceName(pc.InstanceName),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: pc.AccessKey,
			SecretKey: pc.SecretKey,
		}),
	)
	if err != nil {
		xlog.Jupiter().Panic("create producer",
			xlog.FieldName(pc.name),
			xlog.FieldExtMessage(pc.ProducerConfig),
			xlog.Any("error", err),
		)
	}

	if err := client.Start(); err != nil {
		xlog.Jupiter().Panic("start producer",
			xlog.FieldName(pc.name),
			xlog.FieldExtMessage(pc.ProducerConfig),
			xlog.Any("error", err),
		)
	}

	pc.started = true
	pc.Producer = client
	// 进程退出时，producer不Close，避免消息发失败
	// defers.Register(pc.Close)
	return nil
}

// MustStart panics when error found.
func (pc *Producer) MustStart() {
	lo.Must0(pc.Start())
}

func (pc *Producer) WithInterceptor(fs ...primitive.Interceptor) *Producer {
	pc.interceptors = append(pc.interceptors, fs...)
	return pc
}

func (pc *Producer) Close() error {
	err := pc.Shutdown()
	if err != nil {
		xlog.Jupiter().Warn("consumer close fail", xlog.Any("error", err.Error()))
		return err
	}
	return nil
}

// SendWithContext 发送消息
func (pc *Producer) SendWithContext(ctx context.Context, msg []byte) error {
	m := primitive.NewMessage(pc.Topic, msg)
	_, err := pc.SendSync(ctx, m)
	if err != nil {
		xlog.Jupiter().Error("send message error", xlog.Any("msg", msg))
		return err
	}
	return nil
}

// SendWithMsg 发送消息,可以自定义选择tag
func (pc *Producer) SendWithMsg(ctx context.Context, msg *primitive.Message) error {
	msg.Topic = pc.Topic
	_, err := pc.SendSync(ctx, msg)
	if err != nil {
		xlog.Jupiter().Error("send message error", xlog.Any("msg", msg))
		return err
	}
	return nil
}

// SendWithResult rocket mq 发送消息,可以自定义选择 tag 及返回结果
func (pc *Producer) SendWithResult(ctx context.Context, msg []byte, tag string) (*primitive.SendResult, error) {
	m := primitive.NewMessage(pc.Topic, msg)
	if tag != "" {
		m.WithTag(tag)
	}

	res, err := pc.SendSync(ctx, m)
	if err != nil {
		xlog.Jupiter().Error("send message error", xlog.Any("msg", string(msg)))
		return res, err
	}
	return res, nil
}

// SendMsg... 自定义消息格式
func (pc *Producer) SendMsg(ctx context.Context, msg *primitive.Message) (*primitive.SendResult, error) {
	res, err := pc.SendSync(ctx, msg)
	if err != nil {
		xlog.Jupiter().Error("send message error", xlog.Any("msg", msg))
		return res, err
	}
	return res, nil
}
