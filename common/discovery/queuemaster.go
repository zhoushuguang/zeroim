package discovery

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type QueueMaster struct {
	members  map[string]kq.KqConf
	cli      *clientv3.Client
	rootPath string
	observer QueueObserver
}

func NewQueueMaster(rootPath string, address []string) (*QueueMaster, error) {
	cfg := clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 3,
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &QueueMaster{
		members:  make(map[string]kq.KqConf),
		cli:      cli,
		rootPath: rootPath,
	}, nil
}

func (m *QueueMaster) Register(o QueueObserver) {
	m.observer = o
}

func (m *QueueMaster) notifyUpdate(key string, kqConf kq.KqConf) {
	m.observer.Update(key, kqConf)
}

func (m *QueueMaster) notifyDelete(key string) {
	m.observer.Delete(key)
}

func (m *QueueMaster) addQueueWorker(key string, kqConf kq.KqConf) {
	if len(kqConf.Brokers) == 0 || len(kqConf.Topic) == 0 {
		logx.Errorf("invalid kqConf: %+v", kqConf)
		return
	}
	m.members[key] = kqConf
	m.notifyUpdate(key, kqConf)
}

func (m *QueueMaster) updateQueueWorker(key string, kqConf kq.KqConf) {
	if len(kqConf.Brokers) == 0 || len(kqConf.Topic) == 0 {
		logx.Errorf("invalid kqConf: %+v", kqConf)
		return
	}
	m.members[key] = kqConf
	m.notifyUpdate(key, kqConf)
}

func (m *QueueMaster) deleteQueueWorker(key string) {
	delete(m.members, key)
	m.notifyDelete(key)
}

func (m *QueueMaster) WatchQueueWorkers() {
	rch := m.cli.Watch(context.Background(), m.rootPath, clientv3.WithPrefix())
	for wresp := range rch {
		if wresp.Err() != nil {
			logx.Severe(wresp.Err())
		}
		if wresp.Canceled {
			logx.Severe("watch is canceled")
		}
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				var kqConf kq.KqConf
				if err := json.Unmarshal(ev.Kv.Value, &kqConf); err != nil {
					logx.Error(err)
					continue
				}
				if ev.IsCreate() {
					m.addQueueWorker(string(ev.Kv.Key), kqConf)
				} else if ev.IsModify() {
					m.updateQueueWorker(string(ev.Kv.Key), kqConf)
				}
			case clientv3.EventTypeDelete:
				m.deleteQueueWorker(string(ev.Kv.Key))
			}
		}
	}
}
