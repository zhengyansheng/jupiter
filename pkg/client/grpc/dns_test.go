package grpc

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	helloworldv1 "github.com/zhengyansheng/jupiter/proto/helloworld/v1"
)

func TestDNS(t *testing.T) {
	config := DefaultConfig()
	config.Addr = "dns:///localhost:9528"
	config.Debug = true
	cc := helloworldv1.NewGreeterServiceClient(lo.Must(config.Build()))

	res, err := cc.SayHello(context.Background(), &helloworldv1.SayHelloRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, res)
}
