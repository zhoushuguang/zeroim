package logic

import (
	"context"
	"github.com/zhoushuguang/zeroim/common/libnet"
	"github.com/zhoushuguang/zeroim/common/session"
	"github.com/zhoushuguang/zeroim/common/socket"
	"github.com/zhoushuguang/zeroim/common/socketio"
	"github.com/zhoushuguang/zeroim/edge/internal/svc"
	"github.com/zhoushuguang/zeroim/imrpc/imrpc"

	"github.com/golang/protobuf/proto"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type MqLogic struct {
	ctx      context.Context
	svcCtx   *svc.ServiceContext
	server   *socket.Server
	wsServer *socketio.Server
	logx.Logger
}

func NewMqLogic(ctx context.Context, svcCtx *svc.ServiceContext, srv *socket.Server, wsSrv *socketio.Server) *MqLogic {
	return &MqLogic{
		ctx:      ctx,
		svcCtx:   svcCtx,
		server:   srv,
		wsServer: wsSrv,
		Logger:   logx.WithContext(ctx),
	}
}

func (l *MqLogic) Consume(_, val string) error {
	var msg imrpc.PostMsg
	err := proto.Unmarshal([]byte(val), &msg)
	if err != nil {
		logx.Errorf("[Consume] proto.Unmarshal val: %s error: %v", val, err)
		return err
	}
	logx.Infof("[Consume] succ msg: %+v body: %s", msg, msg.Msg)

	if len(msg.ToToken) > 0 {
		sessions := l.server.Manager.GetTokenSessions(msg.ToToken)
		for i := range sessions {
			s := sessions[i]
			if s == nil {
				logx.Errorf("[Consume] session not found, msg: %v", msg)
				continue
			}
			err := s.Send(makeMessage(&msg))
			if err != nil {
				logx.Errorf("[Consume] session send error, msg: %v, err: %v", msg, err)
			}
		}
	} else {
		sess := l.server.Manager.GetSession(session.FromString(msg.SessionId))
		wsSess := l.wsServer.Manager.GetSession(session.FromString(msg.SessionId))
		if sess == nil && wsSess == nil {
			logx.Errorf("[Consume] session not found, msg: %+v", &msg)
			return nil
		}
		if sess != nil {
			err = sess.Send(makeMessage(&msg))
			if err != nil {
				logx.Errorf("[Consume] session send error, msg: %+v, err: %v", &msg, err)
			}
		}
		if wsSess != nil {
			err = wsSess.Send(makeMessage(&msg))
			if err != nil {
				logx.Errorf("[Consume] wsSession send error, msg: %+v, err: %v", &msg, err)
			}
		}
	}

	return err
}

func Consumers(ctx context.Context, svcCtx *svc.ServiceContext, srv *socket.Server, wsSrv *socketio.Server) []service.Service {
	return []service.Service{
		kq.MustNewQueue(svcCtx.Config.KqConf, NewMqLogic(ctx, svcCtx, srv, wsSrv)),
	}
}

func makeMessage(msg *imrpc.PostMsg) libnet.Message {
	var message libnet.Message
	message.Version = uint8(msg.Version)
	message.Status = uint8(msg.Status)
	message.ServiceId = uint16(msg.ServiceId)
	message.Cmd = uint16(msg.Cmd)
	message.Seq = msg.Seq
	message.Body = []byte(msg.Msg)
	return message
}
