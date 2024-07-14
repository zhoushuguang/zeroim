package discovery

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type QueueWorker struct {
	key    string
	kqConf kq.KqConf
	client *clientv3.Client
}

func NewQueueWorker(key string, endpoints []string, kqConf kq.KqConf) *QueueWorker {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 3,
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		panic(err)
	}
	return &QueueWorker{
		key:    key,
		client: etcdClient,
		kqConf: kqConf,
	}
}

func (q *QueueWorker) HeartBeat() {
	value, err := json.Marshal(q.kqConf)
	if err != nil {
		panic(err)
	}
	q.register(string(value))
}

func (q *QueueWorker) register(value string) {
	//申请一个45秒的租约
	leaseGrantResp, err := q.client.Grant(context.TODO(), 45)
	if err != nil {
		panic(err)
	}
	//拿到租约的id
	leaseId := leaseGrantResp.ID
	logx.Infof("查看leaseId:%x", leaseId)

	//获得kv api子集
	kv := clientv3.NewKV(q.client)

	//put一个kv，让它与租约关联起来，从而实现10秒后自动过期
	putResp, err := kv.Put(context.TODO(), q.key, value, clientv3.WithLease(leaseId))
	if err != nil {
		panic(err)
	}

	//(自动续租)当我们申请了租约之后，我们就可以启动一个续租
	keepRespChan, err := q.client.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		panic(err)
	}

	//处理续租应答的协程
	go func() {
		for {
			select {
			case keepResp, ok := <-keepRespChan:
				if !ok {
					logx.Infof("租约已经失效:%x", leaseId)
					q.register(value)
					return
				} else { //每秒会续租一次，所以就会受到一次应答
					logx.Infof("收到自动续租应答:%x", keepResp.ID)
				}
			}
		}
	}()

	logx.Info("写入成功:", putResp.Header.Revision)
}
