package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/zhoushuguang/zeroim/common/libnet"
	"github.com/zhoushuguang/zeroim/common/socket"
	"github.com/zhoushuguang/zeroim/edge/internal/config"
	"github.com/zhoushuguang/zeroim/edge/internal/logic"
	"github.com/zhoushuguang/zeroim/edge/internal/server"
	"github.com/zhoushuguang/zeroim/edge/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/threading"
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
	protobuf := libnet.NewIMProtocol()

	tcpServer.Server, err = socket.NewServe(c.Name, c.TCPListenOn, protobuf, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting tcp server at %s...\n", c.TCPListenOn)

	threading.GoSafe(func() {
		tcpServer.HandleRequest()
	})
	threading.GoSafe(func() {
		tcpServer.KqHeart()
	})

	serviceGroup := zeroservice.NewServiceGroup()
	defer serviceGroup.Stop()

	for _, mq := range logic.Consumers(context.Background(), srvCtx, tcpServer.Server) {
		serviceGroup.Add(mq)
	}
	serviceGroup.Start()
}
