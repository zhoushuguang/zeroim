package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/zhoushuguang/zeroim/common/libnet"
	"github.com/zhoushuguang/zeroim/common/socket"
	"github.com/zhoushuguang/zeroim/common/socketio"
	"github.com/zhoushuguang/zeroim/edge/internal/config"
	"github.com/zhoushuguang/zeroim/edge/internal/logic"
	"github.com/zhoushuguang/zeroim/edge/internal/server"
	"github.com/zhoushuguang/zeroim/edge/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"golang.org/x/net/websocket"
)

var configFile = flag.String("f", "etc/edge.yaml", "the config file")

func main() {
	flag.Parse()

	var err error
	var c config.Config
	conf.MustLoad(*configFile, &c)
	srvCtx := svc.NewServiceContext(c)

	logx.DisableStat()

	tcpServer := server.NewTCPServer(srvCtx)
	wsServer := server.NewWSServer(srvCtx)
	protobuf := libnet.NewIMProtocol()

	tcpServer.Server, err = socket.NewServe(c.Name, c.TCPListenOn, protobuf, c.SendChanSize)
	if err != nil {
		panic(err)
	}
	wsServer.Server, err = socketio.NewServe(c.Name, c.WSListenOn, protobuf, c.SendChanSize)
	if err != nil {
		panic(err)
	}
	http.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		wsServer.HandleRequest(conn)
	}))

	go wsServer.Start()
	go tcpServer.HandleRequest()
	go tcpServer.KqHeart() // 注册kq心跳

	fmt.Printf("Starting tcp server at %s, ws server at: %s...\n", c.TCPListenOn, c.WSListenOn)

	serviceGroup := zeroservice.NewServiceGroup()
	defer serviceGroup.Stop()

	for _, mq := range logic.Consumers(context.Background(), srvCtx, tcpServer.Server, wsServer.Server) {
		serviceGroup.Add(mq)
	}
	serviceGroup.Start()
}
