package logic

import (
	"context"

	"github.com/zhoushuguang/zeroim/common/session"
	"github.com/zhoushuguang/zeroim/imrpc/imrpc"
	"github.com/zhoushuguang/zeroim/imrpc/internal/svc"

	"github.com/golang/protobuf/proto"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
)

type PostMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostMessageLogic {
	return &PostMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PostMessageLogic) PostMessage(in *imrpc.PostMessageRequest) (*imrpc.PostMessageReponse, error) {
	var (
		allDevice bool
		name      string
		token     string
		id        uint64
	)
	if len(in.Token) != 0 {
		allDevice = true
		token = in.Token
	} else {
		name, token, id = session.FromString(in.SessionId).Info()
	}
	sessionIds, err := l.svcCtx.BizRedis.Zrange(token, 0, -1)
	if err != nil {
		return nil, err
	}
	if len(sessionIds) == 0 {
		return nil, err
	}
	data, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}
	set := collection.NewSet()
	for _, sessionId := range sessionIds {
		respName, _, respId := session.FromString(sessionId).Info()
		if allDevice {
			set.Add(respName)
		} else {
			if name == respName && id == respId {
				edgeQueue, ok := l.svcCtx.QueueList.Load(respName)
				if !ok {
					logx.Severe("invalid session")
				} else {
					err = edgeQueue.Push(string(data))
					if err != nil {
						logx.Errorf("[PostMessage] push data: %s error: %v", string(data), err)
						return nil, err
					}
				}
			} else {
				logx.Severe("invalid session")
			}
		}
	}
	if set.Count() > 0 {
		logx.Infof("send to %d devices", set.Count())
	}

	for _, respName := range set.KeysStr() {
		edgeQueue, ok := l.svcCtx.QueueList.Load(respName)
		if !ok {
			logx.Errorf("invalid session")
		} else {
			err = edgeQueue.Push(string(data))
			if err != nil {
				return nil, err
			}
		}
	}

	return &imrpc.PostMessageReponse{}, nil
}
