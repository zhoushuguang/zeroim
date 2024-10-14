package config

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Name         string
	TCPListenOn  string
	WSListenOn   string
	SendChanSize int
	IMRpc        zrpc.RpcClientConf
	KqConf       kq.KqConf
	Etcd         discov.EtcdConf
}
