package discovery

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
)

type QueueObserver interface {
	Update(string, kq.KqConf)
	Delete(string)
}

func QueueDiscoveryProc(conf discov.EtcdConf, qo QueueObserver) {
	master, err := NewQueueMaster(conf.Key, conf.Hosts)
	if err != nil {
		panic(err)
	}
	master.Register(qo)
	master.WatchQueueWorkers()
}
