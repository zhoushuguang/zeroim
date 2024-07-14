package svc

import (
	"github.com/zhoushuguang/zeroim/imapi/internal/config"
	"github.com/zhoushuguang/zeroim/imrpc/imrpc"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	IMRpc  imrpc.ImrpcClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	userClient := zrpc.MustNewClient(c.ImRPC)
	return &ServiceContext{
		Config: c,
		IMRpc:  imrpc.NewImrpcClient(userClient.Conn()),
	}
}
