package logic

import (
	"context"
	"time"

	"zeroim/imrpc/imrpc"
	"zeroim/imrpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *imrpc.LoginRequest) (*imrpc.LoginResponse, error) {
	// TODO jwt验证
	//err := jwt.NewReg(l.svcCtx.Config.AuthConfig.AccessSecret).VerifyToken(in.Token, in.Authorization)
	//if err != nil {
	//	logx.Errorf("[Login] jwt verify token req: %+v error: %v", in, err)
	//	return nil, err
	//}
	_, err := l.svcCtx.BizRedis.Zadd(in.Token, time.Now().UnixMilli(), in.SessionId)
	if err != nil {
		logx.Errorf("[Login] Zadd token: %s sessionId: %s  error: %v", in.Token, in.SessionId, err)
		return nil, err
	}
	_ = l.svcCtx.BizRedis.Expire(in.Token, 3600)

	return &imrpc.LoginResponse{}, nil
}
