package svc

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"zeroim/common/discovery"
	"zeroim/imrpc/internal/config"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceContext struct {
	Config    config.Config
	BizRedis  *redis.Redis
	QueueList *QueueList
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 第一次初始化
	queueList := GetQueueList(c.QueueEtcd)
	threading.GoSafe(func() {
		go discovery.QueueDiscoveryProc(c.QueueEtcd, queueList)
	})
	rds, err := redis.NewRedis(redis.RedisConf{
		Host: c.BizRedis.Host,
		Pass: c.BizRedis.Pass,
		Type: c.BizRedis.Type,
	})
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:    c,
		QueueList: queueList,
		BizRedis:  rds,
	}
}

type QueueList struct {
	kqs map[string]*kq.Pusher
	l   sync.Mutex
}

func NewQueueList() *QueueList {
	return &QueueList{
		kqs: make(map[string]*kq.Pusher),
	}
}

func (q *QueueList) Update(key string, data kq.KqConf) {
	edgeQueue := kq.NewPusher(data.Brokers, data.Topic)
	q.l.Lock()
	q.kqs[key] = edgeQueue
	q.l.Unlock()
}

func (q *QueueList) Delete(key string) {
	q.l.Lock()
	delete(q.kqs, key)
	q.l.Unlock()
}

func (q *QueueList) Load(key string) (*kq.Pusher, bool) {
	q.l.Lock()
	defer q.l.Unlock()

	edgeQueue, ok := q.kqs[key]
	return edgeQueue, ok
}

func GetQueueList(conf discov.EtcdConf) *QueueList {
	ql := NewQueueList()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Hosts,
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := cli.Get(ctx, conf.Key, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	for _, kv := range res.Kvs {
		var data kq.KqConf
		if err := json.Unmarshal(kv.Value, &data); err != nil {
			logx.Errorf("invalid data key is: %s value is: %s", string(kv.Key), string(kv.Value))
			continue
		}
		if len(data.Brokers) == 0 || len(data.Topic) == 0 {
			continue
		}
		edgeQueue := kq.NewPusher(data.Brokers, data.Topic)

		ql.l.Lock()
		ql.kqs[string(kv.Key)] = edgeQueue
		ql.l.Unlock()
	}

	return ql
}
