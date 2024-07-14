package server

import (
	"github.com/zhoushuguang/zeroim/common/discovery"
	"github.com/zhoushuguang/zeroim/common/socket"
	"github.com/zhoushuguang/zeroim/edge/client"
	"github.com/zhoushuguang/zeroim/edge/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TCPServer struct {
	svcCtx *svc.ServiceContext
	Server *socket.Server
}

func NewTCPServer(svcCtx *svc.ServiceContext) *TCPServer {
	return &TCPServer{svcCtx: svcCtx}
}

func (srv *TCPServer) HandleRequest() {
	for {
		session, err := srv.Server.Accept()
		if err != nil {
			panic(err)
		}
		cli := client.NewClient(srv.Server.Manager, session, srv.svcCtx.IMRpc)
		go srv.sessionLoop(cli)
	}
}

func (srv *TCPServer) sessionLoop(client *client.Client) {
	message, err := client.Receive()
	if err != nil {
		logx.Errorf("[sessionLoop] client.Receive error: %v", err)
		_ = client.Close()
		return
	}

	// 登录校验
	err = client.Login(message)
	if err != nil {
		logx.Errorf("[sessionLoop] client.Login error: %v", err)
		_ = client.Close()
		return
	}

	go client.HeartBeat()

	for {
		message, err := client.Receive()
		if err != nil {
			logx.Errorf("[sessionLoop] client.Receive error: %v", err)
			_ = client.Close()
			return
		}
		err = client.HandlePackage(message)
		if err != nil {
			logx.Errorf("[sessionLoop] client.HandleMessage error: %v", err)
		}
	}
}

func (srv *TCPServer) KqHeart() {
	work := discovery.NewQueueWorker(srv.svcCtx.Config.Etcd.Key, srv.svcCtx.Config.Etcd.Hosts, srv.svcCtx.Config.KqConf)
	work.HeartBeat()
}
