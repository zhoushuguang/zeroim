package config

import (
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	BizRedis   redis.RedisConf
	AuthConfig struct {
		AccessSecret string
	}
	QueueEtcd discov.EtcdConf
}
