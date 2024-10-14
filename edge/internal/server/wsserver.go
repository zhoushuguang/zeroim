package server

import (
	"net/http"

	"github.com/zhoushuguang/zeroim/common/socketio"
	"github.com/zhoushuguang/zeroim/edge/client"
	"github.com/zhoushuguang/zeroim/edge/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/websocket"
)

type WSServer struct {
	svcCtx *svc.ServiceContext
	Server *socketio.Server
}

func NewWSServer(svcCtx *svc.ServiceContext) *WSServer {
	return &WSServer{svcCtx: svcCtx}
}

func (ws *WSServer) Start() {
	err := http.ListenAndServe(ws.Server.Address, nil)
	if err != nil {
		panic(err)
	}
}

func (ws *WSServer) HandleRequest(conn *websocket.Conn) {
	session, err := ws.Server.Accept(conn)
	if err != nil {
		panic(err)
	}
	cli := client.NewClient(ws.Server.Manager, session, ws.svcCtx.IMRpc)
	ws.sessionLoop(cli)
}

func (ws *WSServer) sessionLoop(client *client.Client) {
	message, err := client.Receive()
	if err != nil {
		logx.Errorf("[ws:sessionLoop] client.Receive error: %v", err)
		_ = client.Close()
		return
	}
	// login check
	err = client.Login(message)
	if err != nil {
		logx.Errorf("[ws:sessionLoop] client.Login error: %v", err)
		_ = client.Close()
		return
	}

	//client.HeartBeat()

	for {
		message, err = client.Receive()
		if err != nil {
			logx.Errorf("[ws:sessionLoop] client.Receive error: %v", err)
			_ = client.Close()
			return
		}
		err = client.HandlePackage(message)
		if err != nil {
			logx.Errorf("[ws:sessionLoop] client.HandleMessage error: %v", err)
		}
	}
}
